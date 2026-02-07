# Phase 2: Auto-Rotation with Flexible Credential Model

## Mission

Build auto-rotation system with **flexible credential storage** that handles real-world API patterns: secret-only (OpenAI), public-only (frontend apps), public+secret (full-stack), and multi-value configs (Firebase).

## Core Principle

**Not all credentials are created equal.** Some APIs only have secrets. Some only have public keys. Some have both. Some have URLs and config values. The vault must handle ALL patterns while keeping sensitive data (secrets) heavily encrypted and non-sensitive data (public keys, URLs) accessible.

## New Credential Model

### Database Schema

```sql
-- Updated credentials table
CREATE TABLE credentials (
    id TEXT PRIMARY KEY,
    name TEXT UNIQUE NOT NULL,
    api_type TEXT NOT NULL,
    environment TEXT,              -- NEW: optional (dev/staging/prod)

    -- Keys (at least one required)
    public_key TEXT,               -- NEW: optional, plain text (meant to be public)
    secret_key BLOB,               -- Encrypted (SQLCipher + AES-GCM)

    -- Additional data
    url TEXT,                      -- NEW: optional (API endpoint)
    config TEXT,                   -- NEW: optional JSON (extra key-value pairs)

    -- Metadata
    metadata TEXT,                 -- JSON (provider-specific data)

    -- Rotation tracking
    key_id TEXT,                   -- Provider's key identifier
    last_rotated INTEGER,          -- NEW: timestamp of last rotation
    rotation_schedule TEXT,        -- NEW: JSON (rotation config)

    -- Timestamps
    created_at INTEGER NOT NULL,
    updated_at INTEGER NOT NULL
);

-- NEW: Rotation history tracking
CREATE TABLE rotations (
    id TEXT PRIMARY KEY,
    credential_name TEXT NOT NULL,
    rotated_fields TEXT NOT NULL,  -- JSON: ["secret_key", "public_key", "url"]
    old_key_id TEXT,
    new_key_id TEXT,
    plugin_name TEXT NOT NULL,
    rotated_at INTEGER NOT NULL,
    rotated_by TEXT NOT NULL,      -- "scheduler", "manual", "mcp"
    metadata TEXT,                  -- JSON
    FOREIGN KEY (credential_name) REFERENCES credentials(name) ON DELETE CASCADE
);

CREATE INDEX idx_rotations_credential ON rotations(credential_name);
CREATE INDEX idx_rotations_date ON rotations(rotated_at);
```

### Go Credential Struct

```go
// core/database.go
type Credential struct {
    ID          string
    Name        string
    APIType     string
    Environment *string  // Optional: "dev", "staging", "prod"

    // Keys (at least one required)
    PublicKey   *string  // Optional: anon keys, publishable keys
    SecretKey   *string  // Optional: service role, API keys (encrypted)

    // Additional data
    URL         *string             // Optional: API endpoint
    Config      map[string]string   // Optional: extra key-value pairs

    // Metadata
    Metadata    map[string]string

    // Rotation
    KeyID       *string
    LastRotated *time.Time

    // Timestamps
    CreatedAt   time.Time
    UpdatedAt   time.Time
}

// Validation
func (c *Credential) Validate() error {
    if c.Name == "" {
        return errors.New("name is required")
    }
    if c.APIType == "" {
        return errors.New("api_type is required")
    }
    if c.PublicKey == nil && c.SecretKey == nil {
        return errors.New("at least one of public_key or secret_key is required")
    }
    return nil
}

// HasSecret returns true if this credential has a secret key
func (c *Credential) HasSecret() bool {
    return c.SecretKey != nil && *c.SecretKey != ""
}

// HasPublic returns true if this credential has a public key
func (c *Credential) HasPublic() bool {
    return c.PublicKey != nil && *c.PublicKey != ""
}
```

### Storage Security

**Encryption levels:**
- `secret_key`: SQLCipher + AES-256-GCM (double encryption) - **HEAVY SECURITY**
- `public_key`: Plain text (it's meant to be public) - **NO ENCRYPTION**
- `url`, `config`: Plain text (not sensitive) - **NO ENCRYPTION**

**Rationale:** Public keys are designed to be shared publicly (frontend code, client apps). No point encrypting them. Focus encryption on secrets.

## Plugin System Architecture

### Plugin Interface

```go
// rotation/plugin.go
package rotation

import (
    "context"
    "time"
    "github.com/busyrockin/api-vault/core"
)

// Plugin defines the interface all rotation plugins must implement
type Plugin interface {
    // Name returns unique plugin identifier (e.g., "supabase", "openai")
    Name() string

    // RotatableFields declares which credential fields this plugin can rotate
    // Returns fields like: [FieldSecretKey, FieldPublicKey, FieldURL]
    RotatableFields() []RotatableField

    // Rotate performs the actual rotation with the provider's API
    Rotate(ctx context.Context, cred *core.Credential, cfg Config) (*Result, error)

    // Validate checks if a credential can be rotated by this plugin
    Validate(cred *core.Credential) error

    // ConfigSchema returns the configuration fields needed for rotation
    ConfigSchema() ConfigSchema
}

// RotatableField represents a credential field that can be rotated
type RotatableField string

const (
    FieldSecretKey RotatableField = "secret_key"
    FieldPublicKey RotatableField = "public_key"
    FieldURL       RotatableField = "url"
)

// Result contains the new credential values after rotation
type Result struct {
    // New values (only set for fields that were rotated)
    NewSecretKey *string
    NewPublicKey *string
    NewURL       *string

    // Metadata
    KeyID       string
    ExpiresAt   time.Time
    Metadata    map[string]string
    OldKeyGrace time.Duration  // How long old credentials remain valid
}

// Config is the configuration passed to rotation plugins
type Config map[string]interface{}

// ConfigSchema defines what configuration a plugin needs
type ConfigSchema struct {
    Fields []ConfigField
}

type ConfigField struct {
    Name        string
    Type        string  // "string", "int", "bool", "secret"
    Description string
    Required    bool
    Default     interface{}
}
```

### Registry

```go
// rotation/registry.go
package rotation

import (
    "fmt"
    "sync"
)

type Registry struct {
    plugins map[string]Plugin
    mu      sync.RWMutex
}

var globalRegistry = NewRegistry()

func NewRegistry() *Registry {
    return &Registry{
        plugins: make(map[string]Plugin),
    }
}

func GetGlobalRegistry() *Registry {
    return globalRegistry
}

func (r *Registry) Register(p Plugin) error {
    r.mu.Lock()
    defer r.mu.Unlock()

    name := p.Name()
    if _, exists := r.plugins[name]; exists {
        return fmt.Errorf("plugin %s already registered", name)
    }

    r.plugins[name] = p
    return nil
}

func (r *Registry) Get(name string) (Plugin, bool) {
    r.mu.RLock()
    defer r.mu.RUnlock()
    p, ok := r.plugins[name]
    return p, ok
}

func (r *Registry) List() []Plugin {
    r.mu.RLock()
    defer r.mu.RUnlock()

    plugins := make([]Plugin, 0, len(r.plugins))
    for _, p := range r.plugins {
        plugins = append(plugins, p)
    }
    return plugins
}
```

## Database Methods

### Updated core/database.go

Add these methods to `Database`:

```go
// AddCredentialV2 stores a new credential with flexible fields
func (d *Database) AddCredentialV2(cred *core.Credential) error {
    if err := cred.Validate(); err != nil {
        return err
    }

    d.mu.Lock()
    defer d.mu.Unlock()

    // Encrypt secret key if present
    var secretBlob []byte
    var err error
    if cred.HasSecret() {
        secretBlob, err = d.encrypt([]byte(*cred.SecretKey))
        if err != nil {
            return err
        }
    }

    now := time.Now().Unix()

    // Convert maps to JSON
    configJSON, _ := json.Marshal(cred.Config)
    metadataJSON, _ := json.Marshal(cred.Metadata)

    _, err = d.db.Exec(`
        INSERT INTO credentials (
            id, name, api_type, environment,
            public_key, secret_key, url, config, metadata,
            key_id, last_rotated, created_at, updated_at
        ) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
    `,
        newID(), cred.Name, cred.APIType, cred.Environment,
        cred.PublicKey, secretBlob, cred.URL, string(configJSON), string(metadataJSON),
        cred.KeyID, nil, now, now,
    )

    if err != nil && isUniqueViolation(err) {
        return ErrDuplicate
    }

    return err
}

// GetCredentialV2 retrieves a credential with all fields
func (d *Database) GetCredentialV2(name string) (*core.Credential, error) {
    d.mu.RLock()
    defer d.mu.RUnlock()

    var cred core.Credential
    var secretBlob []byte
    var environment, publicKey, url, keyID sql.NullString
    var configJSON, metadataJSON string
    var lastRotated sql.NullInt64
    var createdAt, updatedAt int64

    err := d.db.QueryRow(`
        SELECT id, name, api_type, environment,
               public_key, secret_key, url, config, metadata,
               key_id, last_rotated, created_at, updated_at
        FROM credentials
        WHERE name = ?
    `, name).Scan(
        &cred.ID, &cred.Name, &cred.APIType, &environment,
        &publicKey, &secretBlob, &url, &configJSON, &metadataJSON,
        &keyID, &lastRotated, &createdAt, &updatedAt,
    )

    if err == sql.ErrNoRows {
        return nil, ErrNotFound
    }
    if err != nil {
        return nil, err
    }

    // Set optional fields
    if environment.Valid {
        cred.Environment = &environment.String
    }
    if publicKey.Valid {
        cred.PublicKey = &publicKey.String
    }
    if url.Valid {
        cred.URL = &url.String
    }
    if keyID.Valid {
        cred.KeyID = &keyID.String
    }

    // Decrypt secret key if present
    if len(secretBlob) > 0 {
        plaintext, err := d.decrypt(secretBlob)
        if err != nil {
            return nil, err
        }
        secret := string(plaintext)
        cred.SecretKey = &secret
    }

    // Parse JSON fields
    json.Unmarshal([]byte(configJSON), &cred.Config)
    json.Unmarshal([]byte(metadataJSON), &cred.Metadata)

    // Timestamps
    cred.CreatedAt = time.Unix(createdAt, 0)
    cred.UpdatedAt = time.Unix(updatedAt, 0)
    if lastRotated.Valid {
        t := time.Unix(lastRotated.Int64, 0)
        cred.LastRotated = &t
    }

    return &cred, nil
}

// RotateCredential updates credential fields and logs the rotation
func (d *Database) RotateCredential(name string, result *rotation.Result, pluginName, rotatedBy string) error {
    d.mu.Lock()
    defer d.mu.Unlock()

    // Track which fields were rotated
    rotatedFields := []string{}

    tx, err := d.db.Begin()
    if err != nil {
        return err
    }
    defer tx.Rollback()

    now := time.Now().Unix()

    // Build UPDATE query dynamically based on what was rotated
    updates := []string{"updated_at = ?", "last_rotated = ?"}
    args := []interface{}{now, now}

    if result.NewSecretKey != nil {
        blob, err := d.encrypt([]byte(*result.NewSecretKey))
        if err != nil {
            return err
        }
        updates = append(updates, "secret_key = ?")
        args = append(args, blob)
        rotatedFields = append(rotatedFields, "secret_key")
    }

    if result.NewPublicKey != nil {
        updates = append(updates, "public_key = ?")
        args = append(args, *result.NewPublicKey)
        rotatedFields = append(rotatedFields, "public_key")
    }

    if result.NewURL != nil {
        updates = append(updates, "url = ?")
        args = append(args, *result.NewURL)
        rotatedFields = append(rotatedFields, "url")
    }

    if result.KeyID != "" {
        updates = append(updates, "key_id = ?")
        args = append(args, result.KeyID)
    }

    // Update credential
    query := fmt.Sprintf("UPDATE credentials SET %s WHERE name = ?", strings.Join(updates, ", "))
    args = append(args, name)
    _, err = tx.Exec(query, args...)
    if err != nil {
        return fmt.Errorf("update credential: %w", err)
    }

    // Log rotation
    rotatedFieldsJSON, _ := json.Marshal(rotatedFields)
    metadataJSON, _ := json.Marshal(result.Metadata)

    _, err = tx.Exec(`
        INSERT INTO rotations (
            id, credential_name, rotated_fields, new_key_id,
            plugin_name, rotated_at, rotated_by, metadata
        ) VALUES (?, ?, ?, ?, ?, ?, ?, ?)
    `,
        newID(), name, string(rotatedFieldsJSON), result.KeyID,
        pluginName, now, rotatedBy, string(metadataJSON),
    )
    if err != nil {
        return fmt.Errorf("log rotation: %w", err)
    }

    return tx.Commit()
}

// GetRotationHistory returns rotation history for a credential
func (d *Database) GetRotationHistory(name string, limit int) ([]RotationRecord, error) {
    d.mu.RLock()
    defer d.mu.RUnlock()

    rows, err := d.db.Query(`
        SELECT id, rotated_fields, new_key_id, plugin_name,
               rotated_at, rotated_by, metadata
        FROM rotations
        WHERE credential_name = ?
        ORDER BY rotated_at DESC
        LIMIT ?
    `, name, limit)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var records []RotationRecord
    for rows.Next() {
        var r RotationRecord
        var rotatedFieldsJSON, metadataJSON string
        var rotatedAt int64

        err := rows.Scan(
            &r.ID, &rotatedFieldsJSON, &r.NewKeyID, &r.PluginName,
            &rotatedAt, &r.RotatedBy, &metadataJSON,
        )
        if err != nil {
            return nil, err
        }

        json.Unmarshal([]byte(rotatedFieldsJSON), &r.RotatedFields)
        json.Unmarshal([]byte(metadataJSON), &r.Metadata)
        r.RotatedAt = time.Unix(rotatedAt, 0)

        records = append(records, r)
    }

    return records, rows.Err()
}

type RotationRecord struct {
    ID            string
    RotatedFields []string
    NewKeyID      string
    PluginName    string
    RotatedAt     time.Time
    RotatedBy     string
    Metadata      map[string]string
}
```

## Example Plugins

### OpenAI Plugin (Secret-Only)

```go
// rotation/plugins/openai/plugin.go
package openai

import (
    "context"
    "encoding/json"
    "fmt"
    "github.com/busyrockin/api-vault/core"
    "github.com/busyrockin/api-vault/rotation"
)

type Plugin struct{}

func New() *Plugin {
    return &Plugin{}
}

func (p *Plugin) Name() string {
    return "openai"
}

func (p *Plugin) RotatableFields() []rotation.RotatableField {
    // OpenAI only rotates secret keys
    return []rotation.RotatableField{rotation.FieldSecretKey}
}

func (p *Plugin) Validate(cred *core.Credential) error {
    if cred.APIType != "openai" {
        return fmt.Errorf("credential type must be 'openai'")
    }
    if !cred.HasSecret() {
        return fmt.Errorf("openai credentials require secret_key")
    }
    return nil
}

func (p *Plugin) ConfigSchema() rotation.ConfigSchema {
    return rotation.ConfigSchema{
        Fields: []rotation.ConfigField{
            {
                Name:        "organization_id",
                Type:        "string",
                Description: "OpenAI organization ID",
                Required:    true,
            },
            {
                Name:        "admin_key",
                Type:        "secret",
                Description: "OpenAI admin key for key management API",
                Required:    true,
            },
        },
    }
}

func (p *Plugin) Rotate(ctx context.Context, cred *core.Credential, cfg rotation.Config) (*rotation.Result, error) {
    orgID, _ := cfg["organization_id"].(string)
    adminKey, _ := cfg["admin_key"].(string)

    // Create new API key via OpenAI API
    newKey, keyID, err := p.createAPIKey(ctx, orgID, adminKey)
    if err != nil {
        return nil, err
    }

    return &rotation.Result{
        NewSecretKey: &newKey,  // Only rotate secret
        KeyID:        keyID,
        Metadata: map[string]string{
            "organization_id": orgID,
        },
        OldKeyGrace: 1 * time.Minute,
    }, nil
}

func (p *Plugin) createAPIKey(ctx context.Context, orgID, adminKey string) (string, string, error) {
    // Implementation: Call OpenAI API to create new key
    // POST https://api.openai.com/v1/organization/api_keys
    return "sk-new-key", "key_12345", nil
}
```

### Supabase Plugin (Public + Secret + URL)

```go
// rotation/plugins/supabase/plugin.go
package supabase

import (
    "context"
    "fmt"
    "github.com/busyrockin/api-vault/core"
    "github.com/busyrockin/api-vault/rotation"
)

type Plugin struct{}

func New() *Plugin {
    return &Plugin{}
}

func (p *Plugin) Name() string {
    return "supabase"
}

func (p *Plugin) RotatableFields() []rotation.RotatableField {
    // Supabase can rotate both keys (though typically only secret rotates)
    return []rotation.RotatableField{
        rotation.FieldSecretKey,
        rotation.FieldPublicKey,  // Rarely rotates but possible
    }
}

func (p *Plugin) Validate(cred *core.Credential) error {
    if cred.APIType != "supabase" {
        return fmt.Errorf("credential type must be 'supabase'")
    }
    // Supabase needs at least URL and one key
    if cred.URL == nil {
        return fmt.Errorf("supabase credentials require url")
    }
    if !cred.HasSecret() && !cred.HasPublic() {
        return fmt.Errorf("supabase credentials require at least one key")
    }
    return nil
}

func (p *Plugin) ConfigSchema() rotation.ConfigSchema {
    return rotation.ConfigSchema{
        Fields: []rotation.ConfigField{
            {
                Name:        "project_ref",
                Type:        "string",
                Description: "Supabase project reference",
                Required:    true,
            },
            {
                Name:        "access_token",
                Type:        "secret",
                Description: "Supabase management API token",
                Required:    true,
            },
            {
                Name:        "rotate_service_role",
                Type:        "bool",
                Description: "Rotate service_role key (secret)",
                Default:     true,
            },
        },
    }
}

func (p *Plugin) Rotate(ctx context.Context, cred *core.Credential, cfg rotation.Config) (*rotation.Result, error) {
    projectRef, _ := cfg["project_ref"].(string)
    accessToken, _ := cfg["access_token"].(string)
    rotateServiceRole, _ := cfg["rotate_service_role"].(bool)

    result := &rotation.Result{
        Metadata: map[string]string{
            "project_ref": projectRef,
        },
        OldKeyGrace: 5 * time.Minute,
    }

    // Rotate service_role key (secret) if requested
    if rotateServiceRole && cred.HasSecret() {
        newSecret, keyID, err := p.createServiceRoleKey(ctx, projectRef, accessToken)
        if err != nil {
            return nil, err
        }
        result.NewSecretKey = &newSecret
        result.KeyID = keyID
    }

    // Note: anon key (public) rarely changes, typically don't rotate

    return result, nil
}

func (p *Plugin) createServiceRoleKey(ctx context.Context, projectRef, token string) (string, string, error) {
    // Implementation: Call Supabase Management API
    // POST https://api.supabase.com/v1/projects/{ref}/api-keys
    return "eyJh...service_role", "key_abc", nil
}
```

## CLI Updates

### Updated cmd/add.go

```go
var addCmd = &cobra.Command{
    Use:   "add <name>",
    Short: "Add a new credential",
    Args:  cobra.ExactArgs(1),
    RunE: func(cmd *cobra.Command, args []string) error {
        name := args[0]

        apiType, _ := cmd.Flags().GetString("type")
        environment, _ := cmd.Flags().GetString("env")
        publicKey, _ := cmd.Flags().GetString("public")
        secretKey, _ := cmd.Flags().GetString("secret")
        url, _ := cmd.Flags().GetString("url")

        // At least one key required
        if publicKey == "" && secretKey == "" {
            return fmt.Errorf("at least one of --public or --secret is required")
        }

        cred := &core.Credential{
            Name:    name,
            APIType: apiType,
        }

        if environment != "" {
            cred.Environment = &environment
        }
        if publicKey != "" {
            cred.PublicKey = &publicKey
        }
        if secretKey != "" {
            cred.SecretKey = &secretKey
        }
        if url != "" {
            cred.URL = &url
        }

        db, err := openVault()
        if err != nil {
            return err
        }
        defer db.Close()

        if err := db.AddCredentialV2(cred); err != nil {
            return err
        }

        fmt.Printf("✓ Added: %s\n", name)
        return nil
    },
}

func init() {
    rootCmd.AddCommand(addCmd)
    addCmd.Flags().StringP("type", "t", "", "API type (openai, supabase, etc.)")
    addCmd.Flags().StringP("env", "e", "", "Environment (dev, staging, prod)")
    addCmd.Flags().String("public", "", "Public key (anon, publishable)")
    addCmd.Flags().String("secret", "", "Secret key (service_role, API key)")
    addCmd.Flags().String("url", "", "API endpoint URL")
    addCmd.MarkFlagRequired("type")
}
```

### New cmd/rotate.go

```go
var rotateCmd = &cobra.Command{
    Use:   "rotate <name>",
    Short: "Rotate an API credential",
    Args:  cobra.ExactArgs(1),
    RunE: func(cmd *cobra.Command, args []string) error {
        name := args[0]

        db, err := openVault()
        if err != nil {
            return err
        }
        defer db.Close()

        // Get credential
        cred, err := db.GetCredentialV2(name)
        if err != nil {
            return err
        }

        // Get plugin
        registry := rotation.GetGlobalRegistry()
        plugin, ok := registry.Get(cred.APIType)
        if !ok {
            return fmt.Errorf("no rotation plugin for type '%s'", cred.APIType)
        }

        // Validate
        if err := plugin.Validate(cred); err != nil {
            return err
        }

        // Get config (from flags or interactive)
        config := rotation.Config{}  // Simplified

        // Rotate
        fmt.Printf("Rotating %s (%s)...\n", name, cred.APIType)
        ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
        defer cancel()

        result, err := plugin.Rotate(ctx, cred, config)
        if err != nil {
            return fmt.Errorf("rotation failed: %w", err)
        }

        // Update database
        err = db.RotateCredential(name, result, plugin.Name(), "manual")
        if err != nil {
            return err
        }

        // Report what was rotated
        rotatedFields := []string{}
        if result.NewSecretKey != nil {
            rotatedFields = append(rotatedFields, "secret_key")
        }
        if result.NewPublicKey != nil {
            rotatedFields = append(rotatedFields, "public_key")
        }
        if result.NewURL != nil {
            rotatedFields = append(rotatedFields, "url")
        }

        fmt.Printf("✓ Rotated successfully\n")
        fmt.Printf("  Fields: %s\n", strings.Join(rotatedFields, ", "))
        if result.KeyID != "" {
            fmt.Printf("  New Key ID: %s\n", result.KeyID)
        }
        if result.OldKeyGrace > 0 {
            fmt.Printf("  Old key grace: %s\n", result.OldKeyGrace)
        }

        return nil
    },
}

func init() {
    rootCmd.AddCommand(rotateCmd)
}
```

### New cmd/history.go

```go
var historyCmd = &cobra.Command{
    Use:   "history <name>",
    Short: "Show rotation history for a credential",
    Args:  cobra.ExactArgs(1),
    RunE: func(cmd *cobra.Command, args []string) error {
        name := args[0]
        limit, _ := cmd.Flags().GetInt("limit")

        db, err := openVault()
        if err != nil {
            return err
        }
        defer db.Close()

        records, err := db.GetRotationHistory(name, limit)
        if err != nil {
            return err
        }

        if len(records) == 0 {
            fmt.Println("No rotation history")
            return nil
        }

        fmt.Printf("Rotation history for %s:\n\n", name)
        for _, r := range records {
            fmt.Printf("%s\n", r.RotatedAt.Format("2006-01-02 15:04:05"))
            fmt.Printf("  Plugin: %s\n", r.PluginName)
            fmt.Printf("  By: %s\n", r.RotatedBy)
            fmt.Printf("  Fields: %s\n", strings.Join(r.RotatedFields, ", "))
            if r.NewKeyID != "" {
                fmt.Printf("  Key ID: %s\n", r.NewKeyID)
            }
            fmt.Println()
        }

        return nil
    },
}

func init() {
    rootCmd.AddCommand(historyCmd)
    historyCmd.Flags().IntP("limit", "n", 10, "Number of records to show")
}
```

## Migration Strategy

### Database Migration

Create `core/migrations.go`:

```go
package core

// MigrateToV2 migrates database from V1 (single api_key field) to V2 (flexible model)
func (d *Database) MigrateToV2() error {
    d.mu.Lock()
    defer d.mu.Unlock()

    // Check if already migrated
    var count int
    err := d.db.QueryRow("SELECT COUNT(*) FROM pragma_table_info('credentials') WHERE name='public_key'").Scan(&count)
    if err == nil && count > 0 {
        return nil  // Already migrated
    }

    tx, err := d.db.Begin()
    if err != nil {
        return err
    }
    defer tx.Rollback()

    // Add new columns
    _, err = tx.Exec(`
        ALTER TABLE credentials ADD COLUMN environment TEXT;
        ALTER TABLE credentials ADD COLUMN public_key TEXT;
        ALTER TABLE credentials ADD COLUMN url TEXT;
        ALTER TABLE credentials ADD COLUMN config TEXT;
        ALTER TABLE credentials ADD COLUMN last_rotated INTEGER;
        ALTER TABLE credentials ADD COLUMN rotation_schedule TEXT;
        ALTER TABLE credentials ADD COLUMN key_id TEXT;
    `)
    if err != nil {
        return err
    }

    // Rename api_key to secret_key
    _, err = tx.Exec(`
        ALTER TABLE credentials RENAME COLUMN api_key TO secret_key;
        ALTER TABLE credentials RENAME COLUMN api_type TO api_type;
    `)
    if err != nil {
        return err
    }

    // Create rotations table
    _, err = tx.Exec(`
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

        CREATE INDEX idx_rotations_credential ON rotations(credential_name);
        CREATE INDEX idx_rotations_date ON rotations(rotated_at);
    `)
    if err != nil {
        return err
    }

    return tx.Commit()
}
```

Run migration in `main.go` or `cmd/init.go` on vault open.

## Usage Examples

### Example 1: OpenAI (Secret-Only)

```bash
# Add OpenAI credential (secret only)
api-vault add openai-prod \
  --type openai \
  --secret "sk-proj-abc123..."

# Rotate it
api-vault rotate openai-prod

# Check history
api-vault history openai-prod
```

### Example 2: Supabase Frontend (Public + URL)

```bash
# Add Supabase anon key for frontend (no secret)
api-vault add supabase-web \
  --type supabase \
  --public "eyJh...anon" \
  --url "https://xyz.supabase.co" \
  --env prod

# Get it (returns public key)
api-vault get supabase-web
```

### Example 3: Supabase Full-Stack (Public + Secret + URL)

```bash
# Add both anon and service_role
api-vault add supabase-prod \
  --type supabase \
  --public "eyJh...anon" \
  --secret "eyJh...service_role" \
  --url "https://xyz.supabase.co" \
  --env prod

# Rotate service_role (secret)
api-vault rotate supabase-prod

# History shows only secret was rotated
api-vault history supabase-prod
# Output: Fields: secret_key
```

### Example 4: Stripe (Public + Secret)

```bash
# Add Stripe keys
api-vault add stripe-live \
  --type stripe \
  --public "pk_live_..." \
  --secret "sk_live_..." \
  --env prod

# Rotate secret key
api-vault rotate stripe-live
```

## Testing Checklist

- [ ] Can add credential with only secret
- [ ] Can add credential with only public
- [ ] Can add credential with both public + secret
- [ ] Can add credential with URL
- [ ] Can add credential with environment
- [ ] Cannot add credential with neither public nor secret
- [ ] Get credential returns all fields
- [ ] Secret is encrypted, public is plain text
- [ ] Rotation updates only specified fields
- [ ] Rotation history shows which fields rotated
- [ ] OpenAI plugin only rotates secret
- [ ] Supabase plugin can rotate secret (service_role)
- [ ] Migration from V1 to V2 preserves data

## Security Checklist

- [ ] Secret keys double-encrypted (SQLCipher + AES-GCM)
- [ ] Public keys stored plain text (not sensitive)
- [ ] URLs and config plain text (not sensitive)
- [ ] Old keys have grace periods
- [ ] All rotations audited in rotations table
- [ ] Plugin configs don't leak in logs
- [ ] HTTP requests have timeouts
- [ ] Errors don't leak sensitive data

## Success Criteria

Phase 2 complete when:
- ✅ Flexible credential model implemented (public/secret/url optional)
- ✅ Database migrated to V2 schema
- ✅ Plugin system with RotatableFields
- ✅ OpenAI plugin (secret-only rotation)
- ✅ Supabase plugin (secret + optional public rotation)
- ✅ CLI supports --public, --secret, --url flags
- ✅ Rotation command works
- ✅ History command shows rotation audit trail
- ✅ All security measures in place

---

**Polymath Implementation Notes:**

1. **Order of implementation:** Database schema → Credential model → Plugin interface → Plugins → CLI
2. **Test as you go:** Add tests for each component before moving to next
3. **Security first:** Encrypt secrets immediately, never log sensitive data
4. **Backward compatibility:** Migration must preserve existing credentials
5. **Fail safe:** Better to fail rotation than corrupt vault
6. **Grace periods:** Always give old keys time to phase out
7. **Audit everything:** Every rotation logged with timestamp and actor

**This is production-grade credential rotation. Build it solid.**
