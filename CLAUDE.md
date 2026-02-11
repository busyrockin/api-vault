# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## What This Is

API Vault — secure credential manager for AI agents. "1Password for APIs" with MCP-native integration. Local-first, dual-encrypted storage, CLI + TUI + MCP server.

## Build Commands

SQLCipher CGO flags are required for all Go operations. The Makefile handles this:

```bash
make build          # compile binary → ./api-vault
make test           # run all Go tests
make run ARGS="get openai"  # run without building
```

Single test: `CGO_CFLAGS="-I/opt/homebrew/opt/sqlcipher/include" CGO_LDFLAGS="-L/opt/homebrew/opt/sqlcipher/lib -lsqlcipher" CGO_ENABLED=1 go test ./core -run TestRoundTrip`

Prerequisite: `brew install sqlcipher`

### MCP Server (Python)

```bash
cd server && source .venv/bin/activate && python main.py
```

Requires `API_VAULT_PASSWORD` env var. The MCP server shells out to the Go binary for all vault operations.

### Frontend (React — scaffolding only)

```bash
cd frontend && npm install && npm run dev
```

## Architecture

Three layers, two languages:

**Go binary** (`main.go` → `cmd/` → `core/`) — all security-critical operations. Single binary, no daemon.

**Python MCP server** (`server/`) — thin bridge. FastMCP wraps the Go binary via subprocess. Two tools exposed: `get_credential`, `list_credentials`. Access logging in `approvals.py`.

**React frontend** (`frontend/`) — Vite + React + Zustand + Tailwind. Currently mock data only, not connected to the Go backend.

### Encryption Model

Dual-layer in `core/database.go`:
1. **SQLCipher** — full-disk encryption of the SQLite file via `_pragma_key`
2. **AES-256-GCM** — per-field encryption of secret/public keys. Key derived from master password via Argon2id. Nonce prepended to ciphertext blob.

Salt lives in the `config` table inside the encrypted DB.

### Data Model

Two schema versions coexist. V1 methods (`AddCredential`, `GetCredential`) use simple name/apiKey/apiType. V2 methods (`AddCredentialV2`, `GetCredentialV2`) support the full `Credential` struct with environment, public/secret keys, URL, config map, key ID, and rotation tracking. Migration in `migrateV2()` is idempotent — adds columns to existing V1 tables.

### Rotation Framework

`rotation/` package. Plugin interface: `Name()`, `RotatableFields()`, `Rotate()`, `Validate()`, `ConfigSchema()`. Global registry pattern. Implementations exist for OpenAI and Supabase. Rotation is transactional — updates credential + writes audit log in one tx.

### CLI Structure

Cobra-based in `cmd/`. Commands: `init`, `add`, `get`, `list`, `delete`, `setup` (TUI wizard), `rotate`, `history`. Vault stored at `~/.api-vault/vault.db`. Password read from `API_VAULT_PASSWORD` env var or terminal prompt.

### TUI

Bubble Tea interactive mode via `list -i`. Lipgloss styles in `ui/styles.go`. Setup wizard in `cmd/setup.go`.

## Key Patterns

- Module: `github.com/busyrockin/api-vault`
- Error sentinels: `ErrNotFound`, `ErrDuplicate`, `ErrDecryptFail`
- IDs: 16 random bytes → hex string (no UUID dependency)
- Timestamps: Unix int64 in SQLite, `time.Time` in Go structs
- Vault access: `cmd/helpers.go:openVault()` handles path resolution + password prompt
- `get` outputs raw key to stdout (no newline) for piping; all other output goes to stderr
