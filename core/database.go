package core

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"encoding/json"
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

// Credential holds metadata about a stored credential. V1 methods still work
// with ID/Name/APIType/Metadata. V2 methods use the full struct.
type Credential struct {
	ID, Name, APIType, Metadata string
	Environment                 *string
	PublicKey                   *string
	SecretKey                   *string
	URL                         *string
	Config                      map[string]string
	KeyID                       *string
	LastRotated                 *time.Time
	CreatedAt, UpdatedAt        time.Time
}

func (c *Credential) Validate() error {
	if c.Name == "" {
		return errors.New("name is required")
	}
	if c.SecretKey == nil && c.PublicKey == nil {
		return errors.New("at least one of secret or public key is required")
	}
	return nil
}

func (c *Credential) HasSecret() bool { return c.SecretKey != nil && *c.SecretKey != "" }
func (c *Credential) HasPublic() bool { return c.PublicKey != nil && *c.PublicKey != "" }

// RotationRecord is a single entry in the rotation audit trail.
type RotationRecord struct {
	ID            string
	RotatedFields []string
	NewKeyID      string
	PluginName    string
	RotatedAt     time.Time
	RotatedBy     string
	Metadata      map[string]string
}

// RotationResult carries the output of a rotation plugin. Defined here to
// avoid an import cycle between core and rotation packages.
type RotationResult struct {
	NewSecretKey *string
	NewPublicKey *string
	NewURL       *string
	KeyID        string
	OldKeyGrace  time.Duration
	Metadata     map[string]string
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

	if err := migrateV2(db); err != nil {
		db.Close()
		return nil, fmt.Errorf("migrate v2: %w", err)
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

// AddCredentialV2 stores a credential using the full V2 model.
func (d *Database) AddCredentialV2(cred *Credential) error {
	if err := cred.Validate(); err != nil {
		return err
	}

	d.mu.Lock()
	defer d.mu.Unlock()

	secretBlob := []byte{} // empty blob satisfies NOT NULL when no secret
	if cred.HasSecret() {
		var err error
		secretBlob, err = d.encrypt([]byte(*cred.SecretKey))
		if err != nil {
			return err
		}
	}

	var publicBlob []byte
	if cred.HasPublic() {
		var err error
		publicBlob, err = d.encrypt([]byte(*cred.PublicKey))
		if err != nil {
			return err
		}
	}

	var cfgJSON *string
	if len(cred.Config) > 0 {
		b, _ := json.Marshal(cred.Config)
		s := string(b)
		cfgJSON = &s
	}

	now := time.Now().Unix()
	_, err := d.db.Exec(
		`INSERT INTO credentials (id, name, api_key, api_type, environment, public_key, url, config, key_id, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		newID(), cred.Name, secretBlob, cred.APIType, cred.Environment, publicBlob, cred.URL, cfgJSON, cred.KeyID, now, now,
	)
	if err != nil && isUniqueViolation(err) {
		return ErrDuplicate
	}
	return err
}

// GetCredentialV2 returns the full credential struct with decrypted keys.
func (d *Database) GetCredentialV2(name string) (*Credential, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	var c Credential
	var apiType, meta, env, url, cfgJSON, keyID sql.NullString
	var secretBlob, publicBlob []byte
	var created, updated int64
	var lastRotated sql.NullInt64

	err := d.db.QueryRow(
		`SELECT id, name, api_key, api_type, metadata, environment, public_key, url, config, key_id, last_rotated, created_at, updated_at
		 FROM credentials WHERE name = ?`, name,
	).Scan(&c.ID, &c.Name, &secretBlob, &apiType, &meta, &env, &publicBlob, &url, &cfgJSON, &keyID, &lastRotated, &created, &updated)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	c.APIType = apiType.String
	c.Metadata = meta.String
	c.CreatedAt = time.Unix(created, 0)
	c.UpdatedAt = time.Unix(updated, 0)

	if env.Valid {
		c.Environment = &env.String
	}
	if url.Valid {
		c.URL = &url.String
	}
	if keyID.Valid {
		c.KeyID = &keyID.String
	}
	if lastRotated.Valid {
		t := time.Unix(lastRotated.Int64, 0)
		c.LastRotated = &t
	}

	if len(secretBlob) > 0 {
		plain, err := d.decrypt(secretBlob)
		if err != nil {
			return nil, err
		}
		s := string(plain)
		c.SecretKey = &s
	}
	if len(publicBlob) > 0 {
		plain, err := d.decrypt(publicBlob)
		if err != nil {
			return nil, err
		}
		s := string(plain)
		c.PublicKey = &s
	}

	if cfgJSON.Valid {
		c.Config = make(map[string]string)
		json.Unmarshal([]byte(cfgJSON.String), &c.Config)
	}

	return &c, nil
}

// RotateCredential atomically updates keys and logs the rotation.
func (d *Database) RotateCredential(name string, result *RotationResult, pluginName, rotatedBy string) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	ctx := context.Background()
	tx, err := d.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Update credential fields
	now := time.Now().Unix()
	var fields []string

	if result.NewSecretKey != nil {
		blob, err := d.encrypt([]byte(*result.NewSecretKey))
		if err != nil {
			return err
		}
		if _, err := tx.Exec(`UPDATE credentials SET api_key = ?, updated_at = ? WHERE name = ?`, blob, now, name); err != nil {
			return err
		}
		fields = append(fields, "secret_key")
	}

	if result.NewPublicKey != nil {
		blob, err := d.encrypt([]byte(*result.NewPublicKey))
		if err != nil {
			return err
		}
		if _, err := tx.Exec(`UPDATE credentials SET public_key = ?, updated_at = ? WHERE name = ?`, blob, now, name); err != nil {
			return err
		}
		fields = append(fields, "public_key")
	}

	if result.NewURL != nil {
		if _, err := tx.Exec(`UPDATE credentials SET url = ?, updated_at = ? WHERE name = ?`, *result.NewURL, now, name); err != nil {
			return err
		}
		fields = append(fields, "url")
	}

	if result.KeyID != "" {
		if _, err := tx.Exec(`UPDATE credentials SET key_id = ?, updated_at = ? WHERE name = ?`, result.KeyID, now, name); err != nil {
			return err
		}
	}

	if _, err := tx.Exec(`UPDATE credentials SET last_rotated = ?, updated_at = ? WHERE name = ?`, now, now, name); err != nil {
		return err
	}

	// Log rotation
	fieldsJSON, _ := json.Marshal(fields)
	var metaJSON *string
	if len(result.Metadata) > 0 {
		b, _ := json.Marshal(result.Metadata)
		s := string(b)
		metaJSON = &s
	}

	if _, err := tx.Exec(
		`INSERT INTO rotations (id, credential_name, rotated_fields, old_key_id, new_key_id, plugin_name, rotated_at, rotated_by, metadata)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		newID(), name, string(fieldsJSON), nil, result.KeyID, pluginName, now, rotatedBy, metaJSON,
	); err != nil {
		return err
	}

	return tx.Commit()
}

// GetRotationHistory returns the most recent rotation records for a credential.
func (d *Database) GetRotationHistory(name string, limit int) ([]RotationRecord, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	rows, err := d.db.Query(
		`SELECT id, rotated_fields, new_key_id, plugin_name, rotated_at, rotated_by, metadata
		 FROM rotations WHERE credential_name = ? ORDER BY rotated_at DESC LIMIT ?`,
		name, limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var records []RotationRecord
	for rows.Next() {
		var r RotationRecord
		var fieldsJSON string
		var newKeyID sql.NullString
		var rotatedAt int64
		var metaJSON sql.NullString

		if err := rows.Scan(&r.ID, &fieldsJSON, &newKeyID, &r.PluginName, &rotatedAt, &r.RotatedBy, &metaJSON); err != nil {
			return nil, err
		}
		json.Unmarshal([]byte(fieldsJSON), &r.RotatedFields)
		r.NewKeyID = newKeyID.String
		r.RotatedAt = time.Unix(rotatedAt, 0)
		if metaJSON.Valid {
			r.Metadata = make(map[string]string)
			json.Unmarshal([]byte(metaJSON.String), &r.Metadata)
		}
		records = append(records, r)
	}
	return records, rows.Err()
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

func migrateV2(db *sql.DB) error {
	// Idempotent: check if public_key column already exists
	var hasColumn bool
	rows, err := db.Query(`PRAGMA table_info(credentials)`)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var cid int
		var name, typ string
		var notnull int
		var dflt sql.NullString
		var pk int
		if err := rows.Scan(&cid, &name, &typ, &notnull, &dflt, &pk); err != nil {
			return err
		}
		if name == "public_key" {
			hasColumn = true
			break
		}
	}
	if hasColumn {
		return nil
	}

	for _, stmt := range []string{
		`ALTER TABLE credentials ADD COLUMN environment TEXT`,
		`ALTER TABLE credentials ADD COLUMN public_key TEXT`,
		`ALTER TABLE credentials ADD COLUMN url TEXT`,
		`ALTER TABLE credentials ADD COLUMN config TEXT`,
		`ALTER TABLE credentials ADD COLUMN key_id TEXT`,
		`ALTER TABLE credentials ADD COLUMN last_rotated INTEGER`,
	} {
		if _, err := db.Exec(stmt); err != nil {
			return fmt.Errorf("migrate: %w", err)
		}
	}

	if _, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS rotations (
			id TEXT PRIMARY KEY,
			credential_name TEXT NOT NULL,
			rotated_fields TEXT NOT NULL,
			old_key_id TEXT,
			new_key_id TEXT,
			plugin_name TEXT NOT NULL,
			rotated_at INTEGER NOT NULL,
			rotated_by TEXT NOT NULL,
			metadata TEXT,
			FOREIGN KEY (credential_name) REFERENCES credentials(name) ON DELETE CASCADE
		);
		CREATE INDEX IF NOT EXISTS idx_rotations_credential ON rotations(credential_name);
		CREATE INDEX IF NOT EXISTS idx_rotations_date ON rotations(rotated_at);
	`); err != nil {
		return fmt.Errorf("migrate rotations: %w", err)
	}

	return nil
}
