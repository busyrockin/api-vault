# API Vault

A secure credential management system for AI agents and vibe coders. Think "1Password for APIs" with MCP-native integration.

## Vision

Stop hunting for API keys across .env files, notes apps, and browsers. Store credentials once, access them everywhere - from Claude Code, Cursor, Windsurf, or any MCP-compatible tool.

## Core Features (MVP)

- **Encrypted Storage**: SQLite + SQLCipher for secure local credential storage
- **MCP Server**: Native Model Context Protocol integration for AI agent access
- **CLI Tool**: Simple commands to add, list, and retrieve credentials
- **Auto-Rotation**: Automatic API key rotation on schedule (starting with Supabase, OpenAI, Anthropic)
- **Agent Permissions**: One-time approval system - approve once, remember for that project

## Tech Stack

- **Language**: Go (or Rust) for security-critical operations
- **Storage**: SQLite with SQLCipher (encrypted database)
- **MCP**: FastMCP (Python) or MCP SDK (TypeScript)
- **CLI**: Cobra (Go) or Click (Python)
- **Encryption**: libsodium or age

## Project Structure

```
api-vault/
├── cli/           # Command-line interface
├── server/        # MCP server implementation
├── core/          # Core storage and encryption
├── rotation/      # API-specific rotation plugins
├── docs/          # Documentation
└── examples/      # Usage examples
```

## Getting Started

1. Choose your tech stack (Go vs Rust vs Python)
2. Build core encrypted storage
3. Implement MCP server
4. Add CLI commands
5. Test with Claude Code

## Roadmap

### Week 1-2: Core Storage
- Encrypted SQLite database
- Basic CRUD operations
- CLI for add/list/get credentials

### Week 3: MCP Integration
- Build MCP server
- Implement agent request/approval flow
- Test with Claude Code

### Week 4: Auto-Rotation
- Rotation for Supabase API keys
- Rotation for OpenAI API keys
- Background service for scheduling

## Security Principles

- Never store plaintext credentials
- Use established crypto libraries (no custom crypto)
- Local-first (no cloud backend in MVP)
- Open source core (auditable)
- User approval required for agent access

## Research & Analysis

See `api-vault-research.md` for full competitive analysis, market research, and go-to-market strategy.

## License

TBD (Recommend MIT for core, proprietary for cloud features)

---

**Status**: Pre-MVP - Ready to build
