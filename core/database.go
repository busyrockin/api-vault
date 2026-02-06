package core

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	_ "github.com/mutecomm/go-sqlcipher/v4"
	"golang.org/x/crypto/argon2"
)

// Argon2id parameters (RFC 9106 §7.3).
const (
	argonTime    = 1
	argonMemory  = 64 * 1024 // 64 MB
	argonThreads = 4
	argonKeyLen  = 32
	saltLen      = 16
	nonceLen     = 12
)

// Sentinel errors.
var (
	ErrNotFound    = errors.New("credential not found")
	ErrDuplicate   = errors.New("credential already exists")
	ErrDecryptFail = errors.New("decryption failed")
)

// Database is an encrypted credential store backed by SQLCipher.
type Database struct {
	db  *sql.DB
	key []byte // 32-byte AES-256-GCM key, in-memory only
	mu  sync.RWMutex
}

// Credential holds metadata about a stored credential. The API key itself
// is never included — it's returned only as a bare string from GetCredential.
type Credential struct {
	ID, Name, APIType, Metadata string
	CreatedAt, UpdatedAt        time.Time
}

// NewDatabase opens (or creates) an encrypted database at path, protected
// by password. SQLCipher encrypts the file on disk; an Argon2id-derived
// AES key adds a second layer for individual API key fields.
func NewDatabase(path, password string) (*Database, error) {
	dsn := fmt.Sprintf("%s?_pragma_key=%s", path, password)
	db, err := sql.Open("sqlite3", dsn)
	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("ping db: %w", err)
	}

	if _, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS config (
			key   TEXT PRIMARY KEY,
			value BLOB NOT NULL
		);
		CREATE TABLE IF NOT EXISTS credentials (
			id         TEXT PRIMARY KEY,
			name       TEXT UNIQUE NOT NULL,
			api_key    BLOB NOT NULL,
			api_type   TEXT,
			metadata   TEXT,
			created_at INTEGER NOT NULL,
			updated_at INTEGER NOT NULL
		);
	`); err != nil {
		db.Close()
		return nil, fmt.Errorf("schema: %w", err)
	}

	salt, err := loadOrCreateSalt(db)
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("salt: %w", err)
	}

	return &Database{
		db:  db,
		key: deriveKey(password, salt),
	}, nil
}

// AddCredential stores a new credential with an encrypted API key.
func (d *Database) AddCredential(name, apiKey, apiType string) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	blob, err := d.encrypt([]byte(apiKey))
	if err != nil {
		return err
	}

	now := time.Now().Unix()
	_, err = d.db.Exec(
		`INSERT INTO credentials (id, name, api_key, api_type, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?)`,
		newID(), name, blob, apiType, now, now,
	)
	if err != nil && isUniqueViolation(err) {
		return ErrDuplicate
	}
	return err
}

// GetCredential returns the decrypted API key for the given name.
func (d *Database) GetCredential(name string) (string, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	var blob []byte
	err := d.db.QueryRow(
		`SELECT api_key FROM credentials WHERE name = ?`, name,
	).Scan(&blob)
	if errors.Is(err, sql.ErrNoRows) {
		return "", ErrNotFound
	}
	if err != nil {
		return "", err
	}

	plain, err := d.decrypt(blob)
	if err != nil {
		return "", err
	}
	return string(plain), nil
}

// ListCredentials returns metadata for every stored credential.
// No secrets are included.
func (d *Database) ListCredentials() ([]Credential, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	rows, err := d.db.Query(
		`SELECT id, name, api_type, metadata, created_at, updated_at
		 FROM credentials ORDER BY name`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var creds []Credential
	for rows.Next() {
		var c Credential
		var apiType, meta sql.NullString
		var created, updated int64
		if err := rows.Scan(&c.ID, &c.Name, &apiType, &meta, &created, &updated); err != nil {
			return nil, err
		}
		c.APIType = apiType.String
		c.Metadata = meta.String
		c.CreatedAt = time.Unix(created, 0)
		c.UpdatedAt = time.Unix(updated, 0)
		creds = append(creds, c)
	}
	return creds, rows.Err()
}

// DeleteCredential removes a credential by name.
func (d *Database) DeleteCredential(name string) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	res, err := d.db.Exec(`DELETE FROM credentials WHERE name = ?`, name)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return ErrNotFound
	}
	return nil
}

// Close zeros the in-memory key and closes the database.
func (d *Database) Close() error {
	d.mu.Lock()
	defer d.mu.Unlock()

	for i := range d.key {
		d.key[i] = 0
	}
	return d.db.Close()
}

// --- unexported helpers ---

func deriveKey(password string, salt []byte) []byte {
	return argon2.IDKey([]byte(password), salt, argonTime, argonMemory, argonThreads, argonKeyLen)
}

func (d *Database) encrypt(plaintext []byte) ([]byte, error) {
	block, err := aes.NewCipher(d.key)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	nonce := make([]byte, nonceLen)
	if _, err := rand.Read(nonce); err != nil {
		return nil, err
	}
	return gcm.Seal(nonce, nonce, plaintext, nil), nil
}

func (d *Database) decrypt(data []byte) ([]byte, error) {
	if len(data) < nonceLen {
		return nil, ErrDecryptFail
	}
	block, err := aes.NewCipher(d.key)
	if err != nil {
		return nil, ErrDecryptFail
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, ErrDecryptFail
	}
	plain, err := gcm.Open(nil, data[:nonceLen], data[nonceLen:], nil)
	if err != nil {
		return nil, ErrDecryptFail
	}
	return plain, nil
}

func loadOrCreateSalt(db *sql.DB) ([]byte, error) {
	var salt []byte
	err := db.QueryRow(`SELECT value FROM config WHERE key = 'salt'`).Scan(&salt)
	if err == nil && len(salt) == saltLen {
		return salt, nil
	}

	salt = make([]byte, saltLen)
	if _, err := rand.Read(salt); err != nil {
		return nil, err
	}
	_, err = db.Exec(`INSERT OR REPLACE INTO config (key, value) VALUES ('salt', ?)`, salt)
	return salt, err
}

func newID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func isUniqueViolation(err error) bool {
	return err != nil && strings.Contains(err.Error(), "UNIQUE constraint failed")
}
