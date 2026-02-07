# Phase 1: Terminal UI Implementation - Build & Test Guide

## Context

You're working on **API Vault** - an encrypted credential manager for AI agents and vibe coders. The core CLI and MCP server are complete and working. Phase 1 adds an interactive terminal UI (TUI) using BubbleTea.

**What exists:**
- ✅ Core encrypted storage (SQLCipher + AES-256-GCM)
- ✅ CLI commands (init, add, get, list, delete)
- ✅ MCP server for Claude Code integration
- ✅ Master password: "butters" (set in `API_VAULT_PASSWORD` env var)

**What was just added (Phase 1):**
- `ui/styles.go` - Lipgloss styling constants
- `cmd/interactive.go` - Interactive credential selector (navigation, filtering, copy-to-clipboard)
- `cmd/setup.go` - Setup wizard for adding credentials
- Modified `cmd/list.go` - Added `--interactive` flag
- Updated `go.mod` - Added BubbleTea, Lipgloss, clipboard dependencies

## Project Structure

```
api-vault/
├── core/
│   └── database.go          # Encrypted storage (SQLCipher + AES-GCM)
├── cmd/
│   ├── root.go             # Cobra root
│   ├── helpers.go          # Shared utilities (readPassword, openVault)
│   ├── init.go             # Initialize vault
│   ├── add.go              # Add credential
│   ├── get.go              # Get credential
│   ├── list.go             # List credentials (NOW WITH --interactive flag)
│   ├── delete.go           # Delete credential
│   ├── interactive.go      # NEW: Interactive list with BubbleTea
│   └── setup.go            # NEW: Setup wizard
├── ui/
│   └── styles.go           # NEW: Lipgloss styling
├── server/                  # MCP server (Python)
├── main.go                 # Entry point
├── go.mod                  # Dependencies
└── Makefile                # Build with SQLCipher flags
```

## Build Instructions

### Step 1: Download Dependencies

```bash
cd ~/Documents/Projects/api-vault
go mod tidy
```

**Expected output:**
```
go: downloading github.com/charmbracelet/bubbletea v1.2.4
go: downloading github.com/charmbracelet/lipgloss v1.0.0
go: downloading github.com/atotto/clipboard v0.1.4
```

### Step 2: Build Binary

```bash
make build
```

**Expected output:**
```
go build -tags="libsqlite3 sqlite_cgo" -ldflags="-s -w" -o api-vault .
```

**If build fails:**
- Check SQLCipher is installed: `brew list | grep sqlcipher`
- If not: `brew install sqlcipher`
- Verify Go version: `go version` (should be 1.21+)

## Testing Phase 1 Features

### Test 1: Interactive List

```bash
# Launch interactive mode
./api-vault list --interactive

# Or use short flag
./api-vault list -i
```

**Expected UI:**
```
┌─ 🔑 API Vault ────────────────────────────────────┐
│                                                    │
│ [✓]  openai-production  openai                    │
│ ❯ [✓]  supabase-prod  supabase                    │
│ [⚠]  stripe-test  stripe                          │
│                                                    │
│ [↑↓/jk] Navigate  [Enter] Copy  [d] Delete  [q] Quit │
└────────────────────────────────────────────────────┘
```

**Test these controls:**
1. ↑/↓ or j/k - Navigate between credentials
2. Enter - Copy credential to clipboard (shows preview screen)
3. Type letters - Filter credentials (e.g., type "open" to show only OpenAI)
4. Backspace - Remove filter characters
5. d - Delete selected credential (careful!)
6. q or Ctrl+C - Quit

**Status Indicators:**
- `[✓]` Green - Recent (< 7 days old)
- `[✓]` Normal - OK (7-30 days)
- `[⚠]` Yellow - Warning (30-90 days)
- `[✗]` Red - Old (90+ days)

### Test 2: Setup Wizard

```bash
./api-vault setup
```

**Expected flow:**

**Screen 1: Service Selection**
```
┌─ 🔑 Add Credential ───────────────────────────────┐
│                                                    │
│ Select Service:                                    │
│                                                    │
│ ❯ OpenAI                                          │
│   Anthropic                                        │
│   Supabase                                         │
│   Stripe                                           │
│   GitHub                                           │
│   Custom                                           │
│                                                    │
│ [↑↓] Navigate  [Enter] Select  [Esc] Cancel       │
└────────────────────────────────────────────────────┘
```

**Screen 2: Account Name**
```
Account Name: production_
```

**Screen 3: API Key** (masked)
```
API Key: ********************_
```

**Screen 4: Success**
```
✓ Credential Saved
Successfully saved: openai-production

You can now access this credential with:
  api-vault get openai-production
```

**Test workflow:**
1. Use ↑/↓ to select "Custom"
2. Press Enter
3. Type "test" for account name
4. Press Enter
5. Type "sk-test-key-12345" for API key
6. Press Enter
7. Verify success message shows "custom-test"

### Test 3: End-to-End Workflow

```bash
# 1. Add credential via wizard
./api-vault setup
# Select "Custom" → "test" → "sk-test-key-12345"

# 2. Verify it appears in list
./api-vault list
# Should show: custom-test    custom    2026-02-06

# 3. View in interactive mode
./api-vault list -i
# Navigate to "custom-test" → Press Enter → Verify clipboard

# 4. Get via CLI
./api-vault get custom-test
# Should output: sk-test-key-12345

# 5. Delete via interactive mode
./api-vault list -i
# Navigate to "custom-test" → Press 'd' → Confirm deletion

# 6. Verify deletion
./api-vault list
# Should not show "custom-test"
```

## Code Quality Checks

### Review New Files

1. **ui/styles.go** - Check styling constants are clean
2. **cmd/interactive.go** - Review BubbleTea model implementation
3. **cmd/setup.go** - Review wizard flow and validation
4. **cmd/list.go** - Verify --interactive flag integration

### Key Code Patterns to Verify

**Database field names (should match core/database.go):**
```go
// Correct field names:
c.Name       // credential name
c.APIType    // API type/service
c.CreatedAt  // creation timestamp
```

**Password handling (should use helper):**
```go
// Always use helper from cmd/helpers.go:
db, err := openVault()  // Reads API_VAULT_PASSWORD env var
```

**Error handling (should be consistent):**
```go
if err != nil {
    return fmt.Errorf("context: %w", err)
}
```

## Troubleshooting

### Build Issues

**Error: "go: command not found"**
- Solution: `brew install go`

**Error: "undefined reference to sqlite3_key"**
- Solution: `brew install sqlcipher`
- Verify Makefile has: `-tags="libsqlite3 sqlite_cgo"`

**Error: "cannot find package"**
- Solution: `go mod tidy`

### Runtime Issues

**Error: "failed to unlock vault (wrong password?)"**
- Check env var: `echo $API_VAULT_PASSWORD`
- Should output: "butters"
- If not: `export API_VAULT_PASSWORD="butters"`
- Add to ~/.zshrc for persistence

**Error: "database is locked"**
- Another process has vault open
- Close other terminals or: `lsof ~/.api-vault/vault.db`

**Colors don't show / UI looks broken**
- Terminal may not support 24-bit color
- Try iTerm2, Alacritty, or modern Terminal.app
- UI still works, just less pretty

**Clipboard doesn't work**
- macOS: Should work out of box
- Linux: Install `xclip` or `xsel`
- Windows: Use Windows Terminal

### Code Issues

**Error: "undefined: runInteractive"**
- `cmd/interactive.go` not compiled
- Verify file exists: `ls -la cmd/interactive.go`
- Clean build: `make clean && make build`

**Error: "undefined: ui.TitleStyle"**
- `ui/styles.go` not found
- Verify: `ls -la ui/styles.go`

## Success Criteria

Phase 1 is complete when all these pass:

- ✅ `make build` completes without errors
- ✅ `./api-vault list -i` launches interactive UI
- ✅ Arrow keys navigate credentials
- ✅ Enter copies to clipboard
- ✅ Typing filters credentials in real-time
- ✅ `./api-vault setup` runs complete wizard flow
- ✅ Credentials added via wizard appear in list
- ✅ Status indicators show correct colors (✓⚠✗)
- ✅ Delete ('d' key) removes credentials

## Next Steps After Phase 1

Once Phase 1 is verified working:
- **Phase 2:** Auto-rotation system (plugin architecture)
- **Phase 3:** Claude Code plugin packaging
- **Phase 4:** Desktop app with Tauri

## Key Files for Cursor Context

When asking Cursor for help, reference these files:
- `cmd/interactive.go` - Interactive list implementation
- `cmd/setup.go` - Setup wizard implementation
- `ui/styles.go` - Styling constants
- `core/database.go` - Database schema and Credential struct
- `Makefile` - Build configuration

## Architecture Notes

**Why BubbleTea?**
- Go-native TUI framework (no external dependencies)
- Clean MVC pattern (Model/View/Update)
- Excellent keyboard handling
- Actively maintained by Charm team

**Why Lipgloss?**
- Companion to BubbleTea for styling
- CSS-like styling in Go
- Handles terminal color detection
- Clean API for borders, colors, spacing

**Design Principles:**
- Minimal keystrokes (j/k vim-style navigation)
- Instant visual feedback (filtering, status colors)
- Safe defaults (d key requires confirmation mindset)
- Clipboard integration (zero friction access)
- Masked credentials (security in shared screens)

## Testing Checklist

Copy this into your testing:

```
[ ] Dependencies downloaded (go mod tidy)
[ ] Binary builds (make build)
[ ] Interactive list launches (./api-vault list -i)
[ ] Navigation works (↑↓ or jk)
[ ] Filtering works (type to filter)
[ ] Copy works (Enter → clipboard)
[ ] Status colors show correctly
[ ] Setup wizard launches (./api-vault setup)
[ ] Service selection works
[ ] Account name accepts input
[ ] API key is masked (shows ***)
[ ] Credential saves successfully
[ ] New credential appears in list
[ ] Delete works (d key)
[ ] Quit works (q or Ctrl+C)
```

## Questions to Consider

1. **Performance:** How does interactive list perform with 50+ credentials?
2. **UX:** Is filtering intuitive enough? Should we add search mode?
3. **Security:** Should delete require confirmation (not just 'd')?
4. **Features:** Do we need sorting (by date, name, type)?
5. **Accessibility:** Can we add mouse support for non-vim users?

## Resources

- BubbleTea docs: https://github.com/charmbracelet/bubbletea
- Lipgloss docs: https://github.com/charmbracelet/lipgloss
- BubbleTea examples: https://github.com/charmbracelet/bubbletea/tree/master/examples
- SQLCipher: https://www.zetetic.net/sqlcipher/

---

**Remember:** This is a vibe-coding tool. The TUI should feel fast, elegant, and invisible. If you find yourself fighting the interface, that's a bug.
