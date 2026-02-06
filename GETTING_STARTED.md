# Getting Started: First 30 Minutes

## Quick Setup (Right Now)

### 1. Initialize Go Module

```bash
cd /Users/brianjohnson/Documents/Projects/api-vault
go mod init github.com/yourusername/api-vault
```

### 2. Create Basic Structure

Already done! Your project has:
```
api-vault/
├── cli/              # CLI commands will go here
├── core/             # Storage, encryption, database
├── server/           # MCP server (Python)
├── rotation/         # Auto-rotation plugins
├── docs/             # Documentation
├── examples/         # Usage examples
├── README.md         # Project overview
├── NEXT_STEPS.md     # Detailed build plan
└── .gitignore        # Git ignore rules
```

### 3. First Commit

```bash
git add .
git commit -m "Initial commit: API Vault project structure"
```

---

## Your First Task: Hello World CLI

Let's build the absolute minimum to prove this works.

### Create `main.go`

```go
package main

import (
    "fmt"
    "os"
)

func main() {
    if len(os.Args) < 2 {
        fmt.Println("API Vault - Secure credential management for AI agents")
        fmt.Println("\nUsage:")
        fmt.Println("  api-vault <command>")
        fmt.Println("\nAvailable commands:")
        fmt.Println("  version    Show version")
        os.Exit(0)
    }

    command := os.Args[1]

    switch command {
    case "version":
        fmt.Println("API Vault v0.1.0-dev")
    default:
        fmt.Printf("Unknown command: %s\n", command)
        os.Exit(1)
    }
}
```

### Test It

```bash
go run main.go
go run main.go version
```

**If that works, you're ready to build the real thing.**

---

## What to Build First (Day 1)

Focus on the absolute core: **storing and retrieving one encrypted credential**.

### Minimal Day 1 Goal

Create a program that:
1. Takes a credential name and API key as input
2. Encrypts the API key with a hardcoded password (we'll make this better later)
3. Stores it in a SQLite database
4. Can retrieve and decrypt it later

### Install Dependencies

```bash
# SQLite driver with SQLCipher support
go get github.com/mutecomm/go-sqlcipher/v4

# Crypto utilities
go get golang.org/x/crypto/argon2
```

### Create `core/database.go`

Start with this skeleton:

```go
package core

import (
    "database/sql"
    _ "github.com/mutecomm/go-sqlcipher/v4"
)

type Database struct {
    db *sql.DB
}

func NewDatabase(filepath string, password string) (*Database, error) {
    // TODO: Open SQLCipher database with password
    // TODO: Create credentials table if it doesn't exist
    return &Database{}, nil
}

func (d *Database) AddCredential(name, apiKey string) error {
    // TODO: Encrypt apiKey
    // TODO: Insert into database
    return nil
}

func (d *Database) GetCredential(name string) (string, error) {
    // TODO: Query database
    // TODO: Decrypt apiKey
    return "", nil
}
```

---

## Tools You'll Need

### Go Tools
```bash
# Install Go (if not already)
brew install go  # macOS
# or download from https://go.dev/dl/

# Install Cobra CLI (for building CLI commands)
go install github.com/spf13/cobra-cli@latest

# Verify
go version
cobra-cli --version
```

### Python Tools (for MCP server, Week 2)
```bash
# Install FastMCP
pip install fastmcp --break-system-packages

# Verify
python -c "import fastmcp; print('FastMCP installed')"
```

### Database Tools
```bash
# SQLite command line (useful for debugging)
brew install sqlite  # macOS

# SQLCipher (encrypted SQLite)
brew install sqlcipher  # macOS
```

---

## Working with Claude Code

### Tell Claude Code Your Goal

Open this project in Cursor and use Claude Code. Here's a good first prompt:

> "I'm building an encrypted credential vault in Go. Help me create the core database module that:
> 1. Uses SQLCipher for encrypted SQLite storage
> 2. Has AddCredential and GetCredential functions
> 3. Uses Argon2 for key derivation from a master password
>
> Start with core/database.go. Use the go-sqlcipher/v4 package."

### Iterate with Claude Code

As you build, ask Claude Code to:
- Explain security best practices
- Review your encryption code
- Suggest better error handling
- Help debug SQLite issues
- Write tests for your functions

---

## First Week Checkpoint

By end of Week 1, you should have:

✓ Go module initialized
✓ Database with encrypted storage working
✓ CLI that can add/list/get credentials
✓ At least one credential successfully stored and retrieved

**If you hit this milestone, you're 25% done with the MVP.**

---

## When You Get Stuck

### Common Issues

**SQLCipher won't compile**
- Make sure you have C compiler installed (Xcode Command Line Tools on macOS)
- Try: `xcode-select --install`

**Go dependencies not resolving**
- Run: `go mod tidy`
- Then: `go mod download`

**Can't connect to database**
- Check file permissions
- Verify SQLCipher password is correct
- Try opening with `sqlcipher` CLI to debug

### Ask for Help

- Post in #api-vault Discord (create one!)
- Ask Claude Code to debug
- Check SQLCipher documentation: https://www.zetetic.net/sqlcipher/

---

## Commit Often

Good commit messages:
```bash
git commit -m "feat: add encrypted database module"
git commit -m "feat: implement CLI add command"
git commit -m "fix: handle empty database gracefully"
git commit -m "docs: add database setup instructions"
```

Use conventional commits: https://www.conventionalcommits.org/

---

## Ready to Build?

**Open this folder in Cursor:**
```
/Users/brianjohnson/Documents/Projects/api-vault
```

**Start with:**
1. Create `main.go` (hello world CLI)
2. Create `core/database.go` (encrypted storage)
3. Test it works
4. Expand from there

**You've got the plan. Time to code.**
