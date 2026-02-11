# Next Steps: MCPI (MCP + API)

## Current State Assessment

**What's working:**
- Core encrypted storage with SQLCipher + AES-256-GCM (production-quality)
- Full CLI: init, add, get, list, delete, setup wizard, interactive TUI
- Rotation framework with plugin interface and history tracking
- MCP server scaffold (Python/FastMCP) with get/list tools
- React frontend prototype with glassmorphism design
- 7 passing unit tests for core encryption layer

**What's incomplete:**
- Rotation plugins are stubs (return fake data)
- Frontend uses mock data (no backend connection)
- MCP server shells out to Go binary via subprocess (fragile)
- No approval flow UI
- Build only works on macOS with Homebrew SQLCipher
- Minimal test coverage

---

## Priority 1: Rebrand to MCPI

Rename the project from `api-vault` to `mcpi`. The name captures exactly what this is: an MCP server that securely manages all your APIs.

**Tasks:**
- Rename Go module from `api-vault` to `mcpi`
- Rename binary output from `api-vault` to `mcpi`
- Update all CLI references (`api-vault init` → `mcpi init`)
- Update README, docs, and comments
- Update `.gitignore` entries
- Update MCP server binary path references
- Update frontend references

---

## Priority 2: Go-Native MCP Server

**Why:** The current Python MCP server calls the Go binary via subprocess, parsing tabular stdout. This is fragile, slow, and passes the vault password via environment variable (visible in `ps`).

**Solution:** Replace the Python server with a Go-native MCP implementation using `mark3labs/mcp-go` or similar. The `mcpi` binary becomes both CLI and MCP server.

**Tasks:**
- Add `mcpi serve` command that starts an MCP server (stdio transport)
- Implement MCP tools: `get_credential`, `list_credentials`, `add_credential`, `rotate_credential`
- Remove Python server directory
- Update MCP configuration docs for the new single-binary approach

**Result:** One binary, one process, no subprocess overhead, no password in env vars.

---

## Priority 3: Cross-Platform Build

**Why:** The Makefile hardcodes `/opt/homebrew/opt/sqlcipher` (macOS Homebrew only). No one on Linux can build this.

**Tasks:**
- Add OS detection to Makefile (or switch to a build script)
- Support Linux SQLCipher paths (`/usr/lib`, `/usr/local/lib`)
- Support macOS Intel (`/usr/local/opt/sqlcipher`) and Apple Silicon (`/opt/homebrew/opt/sqlcipher`)
- Add a `Dockerfile` for reproducible builds
- Document build prerequisites per platform

---

## Priority 4: Real Rotation Plugins

**Why:** The OpenAI and Supabase plugins return fake data (`sk-rotated-stub-...`). The rotation architecture is proven but unvalidated against real APIs.

**Tasks:**
- Implement OpenAI rotation plugin (Admin API: create new key, delete old key)
- Implement Supabase rotation plugin (Management API)
- Add plugin configuration storage (secure storage for admin tokens needed by plugins)
- Add ConfigSchema validation (the schema definitions exist but aren't enforced)
- Test end-to-end: rotate a real key, verify the new key works

---

## Priority 5: Approval Flow

**Why:** This is the key security differentiator. When an AI agent requests a credential via MCP, the user should explicitly approve it. The database schema and logging infrastructure exist, but there's no actual prompt.

**Tasks:**
- Implement approval prompt in MCP server (terminal notification or system notification)
- Store approvals per (credential_name, project_path) pair
- Add `mcpi approvals list` and `mcpi approvals revoke` commands
- Add TTL for approvals (auto-expire after N days)

---

## Priority 6: Connect Frontend to Backend

**Why:** The React app is a visual prototype with hardcoded mock data. It needs a real backend.

**Tasks:**
- Add `mcpi api` command that starts a local HTTP server
- Implement REST endpoints: GET/POST/DELETE `/credentials`, GET `/rotations`
- Authenticate the HTTP API with the master password (session token)
- Wire React app to call real endpoints instead of mock data
- Add error handling and loading states

---

## Priority 7: Expand Test Coverage

**Current:** 7 tests in `core/database_test.go` only.

**Tasks:**
- Add CLI integration tests (full add → get → rotate → delete cycle)
- Add MCP server tests (tool invocation, approval flow)
- Add rotation plugin tests (mock HTTP for API calls)
- Add frontend component tests
- Set up CI pipeline (GitHub Actions) to run tests on push

---

## Priority 8: Background Scheduler

**Why:** Manual rotation works, but the value of auto-rotation is that it happens automatically.

**Tasks:**
- Implement `mcpi daemon` command (background process)
- Check rotation schedules on configurable interval
- Execute rotation plugins when credentials are due
- Log rotation results and notify user
- Add `mcpi daemon status` to check if daemon is running

---

## Architecture Decision: Single Binary

The long-term architecture should be a single Go binary (`mcpi`) that handles everything:

```
mcpi init          # Initialize vault
mcpi add <name>    # Add credential
mcpi get <name>    # Retrieve credential
mcpi list          # List credentials (--interactive for TUI)
mcpi rotate <name> # Manual rotation
mcpi serve         # Start MCP server (stdio transport)
mcpi api           # Start local HTTP API (for frontend)
mcpi daemon        # Background rotation scheduler
mcpi approvals     # Manage agent approvals
```

No Python. No subprocess. No parsing stdout. One binary that does everything.

---

## Success Criteria

- [ ] `mcpi serve` works as a native MCP server with Claude Code
- [ ] At least one rotation plugin makes real API calls
- [ ] Frontend displays real credentials from the vault
- [ ] Builds on both macOS and Linux
- [ ] 30+ tests covering core, CLI, MCP, and rotation
- [ ] Approval flow prompts user and remembers decisions
