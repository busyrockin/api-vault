# Phase 2: Auto-Rotation System - Implementation Guide

## Context

**API Vault** now has a working encrypted credential store with an elegant TUI. Phase 2 adds the killer feature: **automatic API key rotation** with a plugin architecture.

**What exists (Phase 1 Complete):**
- ✅ Encrypted storage (SQLCipher + AES-256-GCM)
- ✅ CLI commands (init, add, get, list, delete, setup)
- ✅ Interactive TUI (list -i, setup wizard)
- ✅ MCP server (read-only: get_credential, list_credentials)
- ✅ Master password: "butters"

**What Phase 2 adds:**
- Plugin system for API rotation (extensible for any provider)
- Built-in plugins: Supabase, OpenAI, Anthropic, Stripe, GitHub
- Background scheduler daemon
- MCP server write-with-approval (agents can request rotation)
- Rotation history and audit logging

## Architecture Overview

### Plugin System Design

```
rotation/
├── plugin.go              # Plugin interface
├── registry.go            # Plugin registry and discovery
├── scheduler.go           # Background rotation scheduler
└── plugins/
    ├── supabase/
    │   ├── plugin.go      # Supabase rotation implementation
    │   └── config.json    # Plugin metadata
    ├── openai/
    │   ├── plugin.go
    │   └── config.json
    ├── anthropic/
    │   ├── plugin.go
    │   └── config.json
    ├── stripe/
    │   ├── plugin.go
    │   └── config.json
    └── github/
        ├── plugin.go
        └── config.json
```

### Plugin Interface

Every rotation plugin must implement:

```go
type Plugin interface {
    // Name returns the plugin identifier (e.g., "supabase", "openai")
    Name() string

    // Rotate generates a new API key for the given credential
    // Returns the new key and metadata (e.g., key ID, expiry)
    Rotate(ctx context.Context, credential *core.Credential, config Config) (*RotationResult, error)

    // Validate checks if a credential can be rotated by this plugin
    Validate(credential *core.Credential) error

    // Config returns the configuration schema for this plugin
    ConfigSchema() ConfigSchema
}

type RotationResult struct {
    NewKey      string
    KeyID       string
    ExpiresAt   time.Time
    Metadata    map[string]string
    OldKeyGrace time.Duration  // How long old key remains valid
}

type Config map[string]interface{}

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

### Rotation Flow

```
User/Agent Request
    ↓
Scheduler or Manual Trigger
    ↓
Load Credential from Database
    ↓
Find Matching Plugin (by APIType)
    ↓
Plugin.Validate(credential)
    ↓
Plugin.Rotate(credential, config)
    ↓
Get New Key from Provider API
    ↓
Store New Key in Database
    ↓
Log Rotation Event
    ↓
Return Success + Metadata
```

### Database Schema Changes

Add rotation tracking table:

```sql
CREATE TABLE rotations (
    id TEXT PRIMARY KEY,
    credential_name TEXT NOT NULL,
    old_key_id TEXT,
    new_key_id TEXT,
    plugin_name TEXT NOT NULL,
    rotated_at INTEGER NOT NULL,
    rotated_by TEXT NOT NULL,  -- "scheduler", "manual", "mcp"
    metadata TEXT,              -- JSON
    FOREIGN KEY (credential_name) REFERENCES credentials(name)
);
```

Update credentials table with rotation metadata:

```sql
ALTER TABLE credentials ADD COLUMN last_rotated INTEGER;
ALTER TABLE credentials ADD COLUMN rotation_schedule TEXT;  -- JSON config
ALTER TABLE credentials ADD COLUMN key_id TEXT;             -- Provider key ID
```

## Implementation Plan

### Step 1: Core Plugin System (Day 1)

**Files to create:**
1. `rotation/plugin.go` - Plugin interface and types
2. `rotation/registry.go` - Plugin registry
3. `rotation/result.go` - RotationResult and related types

**Key code patterns:**

```go
// rotation/plugin.go
package rotation

import (
    "context"
    "time"
    "github.com/busyrockin/api-vault/core"
)

type Plugin interface {
    Name() string
    Rotate(ctx context.Context, cred *core.Credential, cfg Config) (*Result, error)
    Validate(cred *core.Credential) error
    ConfigSchema() ConfigSchema
}

type Result struct {
    NewKey      string
    KeyID       string
    ExpiresAt   time.Time
    Metadata    map[string]string
    OldKeyGrace time.Duration
}

type Config map[string]interface{}

type ConfigSchema struct {
    Fields []ConfigField
}

type ConfigField struct {
    Name        string
    Type        string
    Description string
    Required    bool
    Default     interface{}
}
```

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

func NewRegistry() *Registry {
    return &Registry{
        plugins: make(map[string]Plugin),
    }
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

func (r *Registry) List() []string {
    r.mu.RLock()
    defer r.mu.RUnlock()

    names := make([]string, 0, len(r.plugins))
    for name := range r.plugins {
        names = append(names, name)
    }
    return names
}
```

### Step 2: Database Rotation Support (Day 1)

**Files to modify:**
1. `core/database.go` - Add rotation methods and table

**Add to database.go:**

```go
// AddRotationTable creates the rotations tracking table
func (d *Database) AddRotationTable() error {
    _, err := d.db.Exec(`
        CREATE TABLE IF NOT EXISTS rotations (
            id TEXT PRIMARY KEY,
            credential_name TEXT NOT NULL,
            old_key_id TEXT,
            new_key_id TEXT,
            plugin_name TEXT NOT NULL,
            rotated_at INTEGER NOT NULL,
            rotated_by TEXT NOT NULL,
            metadata TEXT,
            FOREIGN KEY (credential_name) REFERENCES credentials(name) ON DELETE CASCADE
        );

        CREATE INDEX IF NOT EXISTS idx_rotations_credential
            ON rotations(credential_name);
        CREATE INDEX IF NOT EXISTS idx_rotations_date
            ON rotations(rotated_at);
    `)
    return err
}

// RotateCredential updates a credential with a new key and logs the rotation
func (d *Database) RotateCredential(name, newKey, keyID, pluginName, rotatedBy string, metadata map[string]string) error {
    d.mu.Lock()
    defer d.mu.Unlock()

    // Get old key ID
    var oldKeyID sql.NullString
    err := d.db.QueryRow("SELECT key_id FROM credentials WHERE name = ?", name).Scan(&oldKeyID)
    if err != nil {
        return fmt.Errorf("get old key id: %w", err)
    }

    // Encrypt new key
    blob, err := d.encrypt([]byte(newKey))
    if err != nil {
        return err
    }

    now := time.Now().Unix()

    // Start transaction
    tx, err := d.db.Begin()
    if err != nil {
        return err
    }
    defer tx.Rollback()

    // Update credential
    _, err = tx.Exec(
        `UPDATE credentials
         SET api_key = ?, key_id = ?, last_rotated = ?, updated_at = ?
         WHERE name = ?`,
        blob, keyID, now, now, name,
    )
    if err != nil {
        return fmt.Errorf("update credential: %w", err)
    }

    // Log rotation
    metaJSON, _ := json.Marshal(metadata)
    _, err = tx.Exec(
        `INSERT INTO rotations (id, credential_name, old_key_id, new_key_id, plugin_name, rotated_at, rotated_by, metadata)
         VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
        newID(), name, oldKeyID.String, keyID, pluginName, now, rotatedBy, string(metaJSON),
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

    rows, err := d.db.Query(
        `SELECT id, old_key_id, new_key_id, plugin_name, rotated_at, rotated_by, metadata
         FROM rotations
         WHERE credential_name = ?
         ORDER BY rotated_at DESC
         LIMIT ?`,
        name, limit,
    )
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var records []RotationRecord
    for rows.Next() {
        var r RotationRecord
        var metaJSON string
        err := rows.Scan(&r.ID, &r.OldKeyID, &r.NewKeyID, &r.PluginName, &r.RotatedAt, &r.RotatedBy, &metaJSON)
        if err != nil {
            return nil, err
        }
        json.Unmarshal([]byte(metaJSON), &r.Metadata)
        records = append(records, r)
    }

    return records, rows.Err()
}

type RotationRecord struct {
    ID         string
    OldKeyID   string
    NewKeyID   string
    PluginName string
    RotatedAt  time.Time
    RotatedBy  string
    Metadata   map[string]string
}
```

### Step 3: Supabase Plugin (Day 2)

**File to create:** `rotation/plugins/supabase/plugin.go`

Supabase rotation uses the Management API to create new service role keys.

```go
package supabase

import (
    "bytes"
    "context"
    "encoding/json"
    "fmt"
    "net/http"
    "time"

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

func (p *Plugin) Validate(cred *core.Credential) error {
    if cred.APIType != "supabase" {
        return fmt.Errorf("credential type must be 'supabase', got '%s'", cred.APIType)
    }
    return nil
}

func (p *Plugin) ConfigSchema() rotation.ConfigSchema {
    return rotation.ConfigSchema{
        Fields: []rotation.ConfigField{
            {
                Name:        "project_ref",
                Type:        "string",
                Description: "Supabase project reference ID",
                Required:    true,
            },
            {
                Name:        "access_token",
                Type:        "secret",
                Description: "Supabase personal access token (for Management API)",
                Required:    true,
            },
        },
    }
}

func (p *Plugin) Rotate(ctx context.Context, cred *core.Credential, cfg rotation.Config) (*rotation.Result, error) {
    projectRef, ok := cfg["project_ref"].(string)
    if !ok {
        return nil, fmt.Errorf("missing or invalid project_ref")
    }

    accessToken, ok := cfg["access_token"].(string)
    if !ok {
        return nil, fmt.Errorf("missing or invalid access_token")
    }

    // Create new service role key via Management API
    url := fmt.Sprintf("https://api.supabase.com/v1/projects/%s/api-keys", projectRef)

    body := map[string]string{
        "name": fmt.Sprintf("rotated-%d", time.Now().Unix()),
    }
    bodyBytes, _ := json.Marshal(body)

    req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(bodyBytes))
    if err != nil {
        return nil, err
    }

    req.Header.Set("Authorization", "Bearer "+accessToken)
    req.Header.Set("Content-Type", "application/json")

    resp, err := http.DefaultClient.Do(req)
    if err != nil {
        return nil, fmt.Errorf("api request failed: %w", err)
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
        return nil, fmt.Errorf("api returned status %d", resp.StatusCode)
    }

    var result struct {
        APIKey string `json:"api_key"`
        ID     string `json:"id"`
    }

    if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
        return nil, fmt.Errorf("decode response: %w", err)
    }

    return &rotation.Result{
        NewKey:      result.APIKey,
        KeyID:       result.ID,
        ExpiresAt:   time.Time{}, // Supabase keys don't expire
        Metadata: map[string]string{
            "project_ref": projectRef,
        },
        OldKeyGrace: 5 * time.Minute, // Keep old key valid for 5 minutes
    }, nil
}
```

### Step 4: OpenAI Plugin (Day 2)

**File to create:** `rotation/plugins/openai/plugin.go`

OpenAI rotation creates a new API key and optionally revokes the old one.

```go
package openai

import (
    "bytes"
    "context"
    "encoding/json"
    "fmt"
    "net/http"
    "time"

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

func (p *Plugin) Validate(cred *core.Credential) error {
    if cred.APIType != "openai" {
        return fmt.Errorf("credential type must be 'openai', got '%s'", cred.APIType)
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
                Description: "OpenAI admin API key (for key management)",
                Required:    true,
            },
            {
                Name:        "revoke_old",
                Type:        "bool",
                Description: "Revoke old key after rotation",
                Required:    false,
                Default:     true,
            },
        },
    }
}

func (p *Plugin) Rotate(ctx context.Context, cred *core.Credential, cfg rotation.Config) (*rotation.Result, error) {
    orgID, ok := cfg["organization_id"].(string)
    if !ok {
        return nil, fmt.Errorf("missing or invalid organization_id")
    }

    adminKey, ok := cfg["admin_key"].(string)
    if !ok {
        return nil, fmt.Errorf("missing or invalid admin_key")
    }

    revokeOld := true
    if v, ok := cfg["revoke_old"].(bool); ok {
        revokeOld = v
    }

    // Create new API key
    url := "https://api.openai.com/v1/organization/api_keys"

    body := map[string]interface{}{
        "name": fmt.Sprintf("api-vault-%d", time.Now().Unix()),
    }
    bodyBytes, _ := json.Marshal(body)

    req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(bodyBytes))
    if err != nil {
        return nil, err
    }

    req.Header.Set("Authorization", "Bearer "+adminKey)
    req.Header.Set("OpenAI-Organization", orgID)
    req.Header.Set("Content-Type", "application/json")

    resp, err := http.DefaultClient.Do(req)
    if err != nil {
        return nil, fmt.Errorf("api request failed: %w", err)
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
        return nil, fmt.Errorf("api returned status %d", resp.StatusCode)
    }

    var result struct {
        Key string `json:"key"`
        ID  string `json:"id"`
    }

    if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
        return nil, fmt.Errorf("decode response: %w", err)
    }

    // Revoke old key if requested
    if revokeOld && cred.Metadata != "" {
        // Extract old key ID from metadata
        var meta map[string]string
        if err := json.Unmarshal([]byte(cred.Metadata), &meta); err == nil {
            if oldKeyID, ok := meta["key_id"]; ok {
                p.revokeKey(ctx, orgID, adminKey, oldKeyID)
            }
        }
    }

    return &rotation.Result{
        NewKey:    result.Key,
        KeyID:     result.ID,
        ExpiresAt: time.Time{},
        Metadata: map[string]string{
            "organization_id": orgID,
            "key_id":          result.ID,
        },
        OldKeyGrace: 1 * time.Minute,
    }, nil
}

func (p *Plugin) revokeKey(ctx context.Context, orgID, adminKey, keyID string) error {
    url := fmt.Sprintf("https://api.openai.com/v1/organization/api_keys/%s", keyID)

    req, err := http.NewRequestWithContext(ctx, "DELETE", url, nil)
    if err != nil {
        return err
    }

    req.Header.Set("Authorization", "Bearer "+adminKey)
    req.Header.Set("OpenAI-Organization", orgID)

    resp, err := http.DefaultClient.Do(req)
    if err != nil {
        return err
    }
    defer resp.Body.Close()

    return nil
}
```

### Step 5: CLI Commands (Day 3)

**Files to create:**
1. `cmd/rotate.go` - Manual rotation command
2. `cmd/plugins.go` - List available plugins

**cmd/rotate.go:**

```go
package cmd

import (
    "context"
    "fmt"
    "os"
    "time"

    "github.com/busyrockin/api-vault/rotation"
    "github.com/spf13/cobra"
)

var rotateCmd = &cobra.Command{
    Use:   "rotate <credential-name>",
    Short: "Rotate an API credential",
    Args:  cobra.ExactArgs(1),
    RunE: func(cmd *cobra.Command, args []string) error {
        name := args[0]

        // Open vault
        db, err := openVault()
        if err != nil {
            return err
        }
        defer db.Close()

        // Get credential
        cred, err := db.GetCredentialMetadata(name)
        if err != nil {
            return fmt.Errorf("get credential: %w", err)
        }

        // Get plugin
        registry := rotation.GetGlobalRegistry()
        plugin, ok := registry.Get(cred.APIType)
        if !ok {
            return fmt.Errorf("no rotation plugin for type '%s'", cred.APIType)
        }

        // Validate
        if err := plugin.Validate(cred); err != nil {
            return fmt.Errorf("validation failed: %w", err)
        }

        // Get config from flags or interactive
        config, err := getRotationConfig(cmd, plugin)
        if err != nil {
            return err
        }

        // Rotate
        fmt.Printf("Rotating %s...\n", name)
        ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
        defer cancel()

        result, err := plugin.Rotate(ctx, cred, config)
        if err != nil {
            return fmt.Errorf("rotation failed: %w", err)
        }

        // Update database
        err = db.RotateCredential(name, result.NewKey, result.KeyID, plugin.Name(), "manual", result.Metadata)
        if err != nil {
            return fmt.Errorf("update database: %w", err)
        }

        fmt.Printf("✓ Rotated successfully\n")
        fmt.Printf("  New Key ID: %s\n", result.KeyID)
        if !result.ExpiresAt.IsZero() {
            fmt.Printf("  Expires: %s\n", result.ExpiresAt.Format(time.RFC3339))
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

### Step 6: Scheduler (Day 3-4)

**File to create:** `rotation/scheduler.go`

Background daemon that rotates credentials on a schedule.

```go
package rotation

import (
    "context"
    "fmt"
    "sync"
    "time"

    "github.com/busyrockin/api-vault/core"
)

type Scheduler struct {
    db       *core.Database
    registry *Registry
    interval time.Duration
    mu       sync.Mutex
    running  bool
    stop     chan struct{}
}

func NewScheduler(db *core.Database, registry *Registry, interval time.Duration) *Scheduler {
    return &Scheduler{
        db:       db,
        registry: registry,
        interval: interval,
        stop:     make(chan struct{}),
    }
}

func (s *Scheduler) Start() error {
    s.mu.Lock()
    defer s.mu.Unlock()

    if s.running {
        return fmt.Errorf("scheduler already running")
    }

    s.running = true
    go s.run()

    return nil
}

func (s *Scheduler) Stop() {
    s.mu.Lock()
    defer s.mu.Unlock()

    if !s.running {
        return
    }

    close(s.stop)
    s.running = false
}

func (s *Scheduler) run() {
    ticker := time.NewTicker(s.interval)
    defer ticker.Stop()

    for {
        select {
        case <-ticker.C:
            s.checkAndRotate()
        case <-s.stop:
            return
        }
    }
}

func (s *Scheduler) checkAndRotate() {
    // Get all credentials with rotation schedules
    creds, err := s.db.GetScheduledCredentials()
    if err != nil {
        fmt.Fprintf(os.Stderr, "scheduler: get credentials: %v\n", err)
        return
    }

    for _, cred := range creds {
        if s.shouldRotate(cred) {
            go s.rotateCredential(cred)
        }
    }
}

func (s *Scheduler) shouldRotate(cred *core.Credential) bool {
    // Parse rotation schedule from credential metadata
    // Example: {"rotation_days": 30}
    // Check if last_rotated + rotation_days < now
    return true // Simplified
}

func (s *Scheduler) rotateCredential(cred *core.Credential) {
    plugin, ok := s.registry.Get(cred.APIType)
    if !ok {
        return
    }

    ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
    defer cancel()

    result, err := plugin.Rotate(ctx, cred, rotation.Config{})
    if err != nil {
        fmt.Fprintf(os.Stderr, "scheduler: rotate %s: %v\n", cred.Name, err)
        return
    }

    err = s.db.RotateCredential(cred.Name, result.NewKey, result.KeyID, plugin.Name(), "scheduler", result.Metadata)
    if err != nil {
        fmt.Fprintf(os.Stderr, "scheduler: update db: %v\n", err)
    }
}
```

### Step 7: MCP Write-with-Approval (Day 4)

**File to modify:** `server/main.py`

Add rotation request tool with approval workflow.

```python
@mcp.tool()
def request_rotation(name: str, reason: str = "") -> str:
    """Request rotation of an API credential. Requires user approval."""

    # Check if credential exists
    try:
        vault.get(name)
    except Exception as e:
        return f"Error: Credential '{name}' not found: {e}"

    # Log approval request
    request_id = approvals.log_rotation_request(name, reason)

    # Return approval message
    return f"""
Rotation requested for: {name}
Reason: {reason or "Not specified"}
Request ID: {request_id}

User must approve this rotation by running:
  api-vault approve-rotation {request_id}
    """.strip()

@mcp.tool()
def check_rotation_status(request_id: str) -> str:
    """Check the status of a rotation request."""
    status = approvals.get_rotation_status(request_id)
    return f"Status: {status}"
```

## Testing Plan

### Test 1: Plugin Registration

```bash
# Build with new rotation code
make build

# List available plugins
./api-vault plugins list

# Expected output:
# Available rotation plugins:
#   supabase - Supabase Management API rotation
#   openai   - OpenAI API key rotation
#   anthropic - Anthropic API key rotation
#   stripe   - Stripe secret key rotation
#   github   - GitHub token rotation
```

### Test 2: Manual Rotation (Supabase)

```bash
# Add a test credential
./api-vault add supabase-test "sbp_test_key_12345" supabase

# Rotate it (you'll need actual Supabase project_ref and access_token)
./api-vault rotate supabase-test \
  --config project_ref=your-project-ref \
  --config access_token=your-access-token

# Expected output:
# Rotating supabase-test...
# ✓ Rotated successfully
#   New Key ID: key_abc123
#   Old key grace: 5m0s

# Verify new key
./api-vault get supabase-test
# Should show new key

# Check rotation history
./api-vault history supabase-test
# Should show rotation record
```

### Test 3: Scheduler

```bash
# Start scheduler daemon
./api-vault scheduler start --interval 1h

# Check status
./api-vault scheduler status
# Expected: Scheduler running, next check in 59m32s

# Stop scheduler
./api-vault scheduler stop
```

### Test 4: MCP Integration

```python
# In Claude Code or Python REPL with MCP client:
from mcp import ClientSession

# Request rotation
result = client.call_tool("request_rotation", {
    "name": "supabase-test",
    "reason": "Automated 30-day rotation"
})
print(result)

# User approves via CLI:
# ./api-vault approve-rotation <request-id>

# Check status
status = client.call_tool("check_rotation_status", {
    "request_id": "<request-id>"
})
print(status)  # Status: approved, rotated
```

## Code Quality Checklist

- [ ] All plugins implement the Plugin interface
- [ ] Error handling includes context (fmt.Errorf with %w)
- [ ] HTTP requests have timeouts (context.WithTimeout)
- [ ] Database operations are transactional where needed
- [ ] Sensitive data (tokens) never logged
- [ ] Plugin configs validate required fields
- [ ] Old keys have grace periods before revocation
- [ ] Rotation events are audited
- [ ] Thread-safe (proper mutex usage)
- [ ] Tests for each plugin

## Security Considerations

1. **Plugin Configs:** Admin keys stored separately (not in vault)
2. **Grace Periods:** Old keys remain valid briefly for zero-downtime
3. **Audit Trail:** All rotations logged with timestamp and actor
4. **MCP Approval:** AI agents can't rotate without user approval
5. **Timeouts:** All API calls have reasonable timeouts
6. **Error Messages:** Don't leak sensitive details in errors

## Next Steps (Phase 3 & 4)

After Phase 2 is complete:
- **Phase 3:** Package as Claude Code plugin
- **Phase 4:** Desktop app with Tauri

## Resources

- Supabase Management API: https://supabase.com/docs/reference/api
- OpenAI API Keys: https://platform.openai.com/docs/api-reference/api-keys
- Go context package: https://pkg.go.dev/context
- HTTP client best practices: https://pkg.go.dev/net/http

## Troubleshooting

**Plugin not found:**
- Check plugin is registered in main.go
- Verify Name() matches credential APIType

**Rotation fails:**
- Check API credentials in config
- Verify network connectivity
- Check provider API status
- Review error message context

**Database errors:**
- Ensure migrations ran (rotations table exists)
- Check foreign key constraints
- Verify transaction handling

---

**Remember:** Rotation is a sensitive operation. Test thoroughly with non-production credentials first. Always maintain grace periods for old keys to avoid service disruptions.
