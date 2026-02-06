# Next Steps: Building API Vault MVP

## Immediate Decisions (Do This First)

### 1. Choose Your Language

**Option A: Go** (Recommended for MVP speed)
- Pro: Fast to build, single binary, excellent crypto libraries, Cobra for CLI
- Pro: Cross-platform compilation is trivial
- Con: Slightly more verbose than Python

**Option B: Rust**
- Pro: Maximum security, excellent type system, prevents entire classes of bugs
- Con: Steeper learning curve, slower initial development
- Best if: You value correctness over speed of development

**Option C: Python**
- Pro: Fastest to prototype, great for MCP (FastMCP is Python-native)
- Con: Distribution is harder (need Python runtime), slower performance
- Best if: You want to validate idea fastest

**My recommendation: Start with Go for core + Python for MCP server**
- Go handles storage/encryption/CLI (security-critical)
- Python handles MCP server (fast iteration, FastMCP is excellent)
- Best of both worlds

### 2. Choose Your Encryption Library

**For Go:**
- `golang.org/x/crypto` - Standard library, battle-tested
- Or `age` package - Modern, simple, audited

**For Rust:**
- `ring` - AWS-backed crypto library
- `sodiumoxide` - Rust bindings for libsodium

**For Python:**
- `cryptography` - Industry standard
- `PyNaCl` - libsodium bindings

**Decision: Go with established libraries. Don't roll your own crypto.**

---

## Week 1: Core Storage (Days 1-7)

### Day 1-2: Database Setup

**Goal:** Encrypted SQLite database that stores key-value pairs

**Tasks:**
1. Initialize Go module: `go mod init github.com/yourusername/api-vault`
2. Install SQLCipher: https://github.com/sqlcipher/sqlcipher
3. Install Go SQLite driver: `go get github.com/mutecomm/go-sqlcipher/v4`
4. Create database schema:
   ```sql
   CREATE TABLE credentials (
       id TEXT PRIMARY KEY,
       name TEXT UNIQUE NOT NULL,
       api_key TEXT NOT NULL,
       api_type TEXT,
       rotation_enabled BOOLEAN DEFAULT 0,
       last_rotated TIMESTAMP,
       expires_at TIMESTAMP,
       metadata JSON,
       created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
       updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
   );
   ```

**Files to create:**
- `core/database.go` - Database connection and initialization
- `core/models.go` - Credential struct definition
- `core/storage.go` - CRUD operations

**Test:** Can you create a database, insert a credential, and retrieve it?

### Day 3-4: Encryption

**Goal:** Encrypt credentials before storing, decrypt on retrieval

**Tasks:**
1. Implement key derivation from user passphrase (use PBKDF2 or Argon2)
2. Store master key in OS keychain (macOS Keychain, Windows Credential Manager, Linux Secret Service)
3. Encrypt `api_key` field before database insertion
4. Decrypt on retrieval

**Libraries:**
- Go: `github.com/zalando/go-keyring` for OS keychain access
- Go: `golang.org/x/crypto/argon2` for key derivation

**Files to create:**
- `core/encryption.go` - Encryption/decryption functions
- `core/keychain.go` - OS keychain integration

**Test:** Store a credential with passphrase, close program, retrieve it later with same passphrase

### Day 5-7: CLI Tool

**Goal:** Command-line interface for managing credentials

**Tasks:**
1. Install Cobra: `go get github.com/spf13/cobra-cli`
2. Create commands:
   - `api-vault init` - Initialize vault (create database, set master password)
   - `api-vault add <name> <key>` - Add credential
   - `api-vault list` - List all credential names (not keys!)
   - `api-vault get <name>` - Retrieve specific credential
   - `api-vault delete <name>` - Remove credential

**Files to create:**
- `cli/root.go` - Root command
- `cli/init.go` - Init command
- `cli/add.go` - Add command
- `cli/list.go` - List command
- `cli/get.go` - Get command
- `cli/delete.go` - Delete command

**Test:** Full workflow - init vault, add 3 API keys, list them, retrieve one, delete one

**Week 1 Milestone:** You have a working encrypted credential store with CLI. This alone is useful.

---

## Week 2: MCP Server (Days 8-14)

### Day 8-9: MCP Server Setup

**Goal:** Build MCP server that exposes credentials to AI agents

**Tasks:**
1. Create new Python project: `mkdir server && cd server`
2. Install FastMCP: `pip install fastmcp --break-system-packages`
3. Create basic MCP server that reads from your Go database
4. Expose one tool: `get_credential(name: str) -> str`

**Files to create:**
- `server/main.py` - MCP server entry point
- `server/tools.py` - Tool definitions
- `server/database.py` - Read-only access to SQLite database

**Challenge:** Python needs to read Go's encrypted database
**Solution:** Either:
- Option A: Python calls Go CLI as subprocess: `subprocess.run(['api-vault', 'get', name])`
- Option B: Python reimplements decryption (more work, but faster)

**My recommendation: Option A for MVP. It's quick and works.**

**Test:** Start MCP server, connect with Claude Code, request a credential

### Day 10-11: Agent Permission System

**Goal:** User approves credential access once, server remembers

**Tasks:**
1. Add `approvals` table to database:
   ```sql
   CREATE TABLE approvals (
       credential_name TEXT,
       project_path TEXT,
       approved_at TIMESTAMP,
       PRIMARY KEY (credential_name, project_path)
   );
   ```
2. When agent requests credential:
   - Check if approval exists for (credential_name, project_path)
   - If yes: return credential
   - If no: prompt user for approval, store decision
3. Implement approval prompt (CLI prompt or notification)

**Files to create:**
- `server/approvals.py` - Approval logic
- `cli/approve.go` - Approval prompt (called by Python server)

**Test:** Request same credential twice - should only prompt once

### Day 12-14: MCP Integration & Testing

**Goal:** Polish MCP server, test with multiple AI tools

**Tasks:**
1. Add more tools:
   - `list_credentials()` - Show available credentials (names only)
   - `add_credential(name, key)` - Add new credential from agent
   - `rotate_credential(name)` - Trigger manual rotation
2. Write MCP server configuration for Claude Code
3. Test with multiple projects to verify approval isolation

**Files to create:**
- `server/config.json` - MCP server configuration
- `docs/CLAUDE_CODE_SETUP.md` - Instructions for connecting to Claude Code

**Test:** Use API Vault in an actual project with Claude Code. Can the agent access credentials without you manually providing them?

**Week 2 Milestone:** MCP server works. AI agents can request credentials with user approval. This is the core value proposition.

---

## Week 3: Auto-Rotation (Days 15-21)

### Day 15-16: Rotation Framework

**Goal:** Generic system for rotating API keys

**Tasks:**
1. Define interface for rotation plugins:
   ```go
   type RotationPlugin interface {
       CanRotate(apiType string) bool
       Rotate(currentKey string) (newKey string, err error)
   }
   ```
2. Create plugin registry
3. Add rotation scheduler (check every hour, rotate if needed)

**Files to create:**
- `rotation/plugin.go` - Plugin interface
- `rotation/scheduler.go` - Rotation scheduler
- `rotation/registry.go` - Plugin registry

### Day 17-18: Supabase Rotation Plugin

**Goal:** Automatically rotate Supabase service role keys

**Tasks:**
1. Research Supabase Management API
2. Implement plugin that:
   - Takes Supabase project URL + current key
   - Uses Management API to generate new service role key
   - Returns new key
3. Test rotation manually

**Files to create:**
- `rotation/plugins/supabase.go` - Supabase rotation plugin

**Note:** Some APIs don't support programmatic rotation. Start with ones that do (Supabase, Stripe, OpenAI).

### Day 19-20: OpenAI Rotation Plugin

**Goal:** Automatically rotate OpenAI API keys

**Tasks:**
1. Research OpenAI API key management
2. Implement rotation plugin
3. Add to registry

**Files to create:**
- `rotation/plugins/openai.go` - OpenAI rotation plugin

### Day 21: Background Service

**Goal:** Run rotation checks automatically

**Tasks:**
1. Create background service that runs rotation scheduler
2. Add `api-vault daemon start` command
3. Run rotation checks every hour
4. Notify user when keys are rotated

**Files to create:**
- `cli/daemon.go` - Daemon command
- `core/service.go` - Background service

**Test:** Add Supabase key with 7-day rotation policy. Wait for scheduled rotation. Verify new key works.

**Week 3 Milestone:** Auto-rotation works for 2-3 popular APIs. Security is significantly improved.

---

## Week 4: Polish & Launch (Days 22-28)

### Day 22-23: Documentation

**Tasks:**
1. Write comprehensive README
2. Create quickstart guide
3. Document all CLI commands
4. Add MCP setup instructions for Claude Code, Cursor, Windsurf
5. Create security architecture document

**Files to create:**
- `docs/QUICKSTART.md`
- `docs/CLI_REFERENCE.md`
- `docs/MCP_SETUP.md`
- `docs/SECURITY.md`
- `docs/ROTATION.md`

### Day 24: Testing & Debugging

**Tasks:**
1. Test on all platforms (macOS, Linux, Windows)
2. Fix any bugs found
3. Add error handling for edge cases
4. Improve error messages

### Day 25-26: Packaging

**Tasks:**
1. Cross-compile Go binary for all platforms
2. Create install script
3. Set up GitHub releases with binaries
4. Create Homebrew formula (macOS)
5. Write uninstall instructions

**Files to create:**
- `scripts/build.sh` - Cross-compilation script
- `scripts/install.sh` - Installation script
- `api-vault.rb` - Homebrew formula

### Day 27: Soft Launch

**Tasks:**
1. Create GitHub repo (public)
2. Push all code
3. Write launch tweet
4. Post in r/ClaudeAI, r/MachineLearning
5. Share in Discord communities (Claude, Cursor, AI coding tools)

**Launch message template:**
> "I built API Vault - a secure credential manager for AI coding agents. Store your API keys once, access them from Claude Code/Cursor/Windsurf without copy-pasting .env files. Keys auto-rotate for security. Open source MVP is live. [GitHub link]"

### Day 28: Collect Feedback

**Tasks:**
1. Monitor GitHub issues
2. Respond to questions on Twitter/Reddit
3. Join Discord servers and help users
4. Note feature requests
5. Decide what to build next based on feedback

---

## After MVP: What's Next?

Based on user feedback, prioritize:

### If users want easier setup:
- One-command installation
- Auto-detect API keys in .env files and migrate them
- GUI for credential management

### If users want more APIs:
- Rotation plugins for more services
- Community plugin system
- API marketplace integration

### If users want team features:
- Cloud sync (encrypted)
- Team vaults (shared credentials)
- Role-based access control

### If users want better security:
- Hardware token support (YubiKey)
- Biometric authentication
- Audit logs

---

## Resources & References

### MCP Documentation
- FastMCP: https://github.com/jlowin/fastmcp
- MCP Spec: https://modelcontextprotocol.io/
- Claude Code MCP Guide: https://code.claude.com/docs/en/mcp

### Encryption Libraries
- Go crypto: https://pkg.go.dev/golang.org/x/crypto
- SQLCipher: https://www.zetetic.net/sqlcipher/
- age encryption: https://github.com/FiloSottile/age

### CLI Tools
- Cobra: https://cobra.dev/
- Survey (interactive prompts): https://github.com/AlecAivazis/survey

### API Management References
- Supabase Management API: https://supabase.com/docs/reference/api
- OpenAI API Keys: https://platform.openai.com/api-keys
- Stripe API: https://stripe.com/docs/keys

---

## Success Metrics

### Week 1
- ✓ Can store and retrieve encrypted credentials via CLI

### Week 2
- ✓ MCP server working with Claude Code
- ✓ Agent approval system functional

### Week 3
- ✓ Auto-rotation working for 2+ APIs
- ✓ Background service running

### Week 4
- ✓ 100+ GitHub stars (validation)
- ✓ 10+ people using it
- ✓ At least 5 pieces of user feedback

---

## Remember

- **Start small:** Working core is better than perfect everything
- **Talk to users:** Ask for feedback weekly, iterate fast
- **Security first:** Never compromise on encryption or key handling
- **Document everything:** Future you will thank present you
- **Launch early:** 70% done is good enough to get feedback

**You've got this. Time to build.**
