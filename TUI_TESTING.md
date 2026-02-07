# Terminal UI (TUI) Testing Guide

## Phase 1 Complete: Terminal UI with BubbleTea

The interactive terminal interface has been implemented with:
- Interactive credential selector
- Setup wizard for adding credentials
- Real-time filtering and navigation
- Status indicators for credential age

## Building with New Dependencies

First, download the new Go dependencies:

```bash
cd ~/Documents/Projects/api-vault
go mod tidy
```

This will download:
- `github.com/charmbracelet/bubbletea` - TUI framework
- `github.com/charmbracelet/lipgloss` - Styling
- `github.com/atotto/clipboard` - Clipboard integration

Then rebuild the binary:

```bash
make build
```

## Testing Interactive List

### Basic Navigation

```bash
# Start interactive list (if you have credentials)
./api-vault list --interactive

# Or use the short flag
./api-vault list -i
```

**Controls:**
- `↑/↓` or `j/k` - Navigate up/down
- `Enter` - Copy selected credential to clipboard
- `d` - Delete selected credential (be careful!)
- `q` or `Ctrl+C` - Quit
- Type any letter - Filter credentials by name/type

**What to test:**
1. Arrow key navigation works
2. Credential status indicators appear (✓ for recent, ⚠ for old)
3. Enter copies to clipboard and shows preview
4. Filtering works (type "open" to filter for OpenAI)
5. Backspace removes filter characters
6. Delete removes credentials (test with a test credential)

## Testing Setup Wizard

```bash
# Start interactive setup
./api-vault setup
```

**Workflow:**
1. **Service Selection**
   - Use ↑/↓ to select service (OpenAI, Anthropic, Supabase, etc.)
   - Press Enter to continue

2. **Account Name**
   - Type an account name (e.g., "production", "development")
   - Press Enter to continue

3. **API Key**
   - Paste your API key (it will be masked with asterisks)
   - Press Enter to save

4. **Success Screen**
   - Shows the credential name that was saved
   - Shows the command to retrieve it

**What to test:**
1. Service selection navigation works
2. Account name accepts typing
3. API key is masked with asterisks
4. Credential is saved correctly
5. Success message shows correct credential name
6. Can retrieve the credential with `./api-vault get <name>`

## Testing Flow Example

```bash
# 1. Add a test credential interactively
./api-vault setup
# Select "Custom"
# Enter name: "test"
# Enter key: "sk-test-key-12345"

# 2. View it in interactive list
./api-vault list -i
# Navigate to "custom-test"
# Press Enter to copy
# Check clipboard contains the key

# 3. Test filtering
./api-vault list -i
# Type "test" to filter
# Should only show credentials with "test" in name

# 4. Clean up
./api-vault list -i
# Navigate to "custom-test"
# Press 'd' to delete

# 5. Verify deletion
./api-vault list
# Should not show "custom-test"
```

## Expected Visual Output

### Interactive List
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

### Setup Wizard (Service Selection)
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

### Setup Wizard (API Key Entry)
```
┌─ 🔑 Add Credential ───────────────────────────────┐
│                                                    │
│ Credential: openai-production                      │
│                                                    │
│ API Key: ********************_                     │
│                                                    │
│ Paste your OpenAI API key                         │
│                                                    │
│ [Type] Enter key  [Enter] Save  [Esc] Cancel      │
└────────────────────────────────────────────────────┘
```

## Troubleshooting

### Colors Don't Show
If colors don't render properly, your terminal may not support 24-bit color. The TUI will still work but may look less polished.

**Solution:** Use a modern terminal like:
- iTerm2 (macOS)
- Terminal.app (macOS) - recent versions
- Alacritty
- Kitty
- Windows Terminal (Windows)

### Can't Copy to Clipboard
The clipboard integration uses system clipboard APIs.

**macOS:** Should work out of the box
**Linux:** Requires `xclip` or `xsel` installed
```bash
# Ubuntu/Debian
sudo apt-get install xclip

# Fedora
sudo dnf install xclip
```

**Windows:** Should work with Windows Terminal

### Arrow Keys Don't Work
Some terminals may not properly send arrow key escape sequences.

**Workaround:** Use `j/k` instead of `↓/↑`

## Success Criteria

Phase 1 is complete when:
- ✅ Interactive list renders with styled boxes
- ✅ Navigation works (arrow keys or j/k)
- ✅ Enter copies credential to clipboard
- ✅ Filtering works by typing
- ✅ Setup wizard completes full flow
- ✅ Credentials added via wizard appear in list
- ✅ Status indicators show credential age

## Next Steps

Once Phase 1 testing is complete:
- **Phase 2:** Auto-rotation system (plugin architecture)
- **Phase 3:** Claude Code plugin packaging
- **Phase 4:** Desktop app with Tauri

## Feedback

Please test and report:
1. Which terminal you tested with
2. Any visual rendering issues
3. Any control/navigation issues
4. Feature requests or UX improvements
