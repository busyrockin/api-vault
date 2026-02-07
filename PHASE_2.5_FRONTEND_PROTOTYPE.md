# Phase 2.5: Frontend Prototype - Apple Aesthetic Design

## Mission

Build a **visual prototype** of the desktop app to test UX flows, validate the Apple-inspired aesthetic, and design the trust/security experience BEFORE committing to Tauri. Use mock data - no real credentials.

## Why This Phase?

**You're a visual person, not a terminal person.** The TUI is elegant but limited. Before building the full desktop app (Phase 4), we need to:

1. **Validate the UX** - Does the flow feel seamless for humans AND agents?
2. **Test the aesthetic** - Does it feel Apple-quality (liquid glass, minimalist, friendly)?
3. **Design trust indicators** - Do users feel secure?
4. **Iterate quickly** - React prototypes are faster than Tauri apps
5. **Document the vision** - Create a reference for Phase 4 implementation

## Current Codebase Analysis

### What We Have (Phase 1 & 2)

**Core Storage (`core/database.go`):**
```go
type Credential struct {
    ID, Name, APIType string
    Environment       *string  // "dev", "staging", "prod"
    PublicKey         *string  // Anon keys, publishable keys
    SecretKey         *string  // Service role, API keys (encrypted)
    URL               *string  // API endpoint
    Config            map[string]string
    KeyID             *string
    LastRotated       *time.Time
    CreatedAt, UpdatedAt time.Time
}
```

**Rotation System (`rotation/rotation.go`):**
```go
type Plugin interface {
    Name() string
    RotatableFields() []RotatableField
    Rotate(ctx, cred, cfg) (*Result, error)
    Validate(cred) error
    ConfigSchema() ConfigSchema
}
```

**Operations Available:**
- Add credential (flexible: public/secret/url/env)
- Get credential (decrypt and return)
- List credentials (all with metadata)
- Delete credential
- Rotate credential (via plugin)
- View rotation history

### What We Need to Design

**5 Core Screens:**
1. **Dashboard** - Overview of all credentials
2. **Add/Edit** - Credential creation/modification flow
3. **Detail View** - Single credential with rotation history
4. **Settings** - Vault password, preferences, plugins
5. **Trust Indicator** - Always-visible security status

## Apple Aesthetic Design Language

### Visual Philosophy

**Inspired by:**
- macOS Ventura/Sonoma glassmorphism
- iOS translucency and depth
- Apple's SF Pro font system
- Minimal, purposeful animations
- Generous whitespace
- Clear visual hierarchy

### Design System

**Colors:**
```css
/* Primary Palette */
--glass-bg: rgba(255, 255, 255, 0.7);
--glass-border: rgba(255, 255, 255, 0.2);
--glass-shadow: rgba(0, 0, 0, 0.1);

/* Accent Colors */
--primary: #007AFF;        /* Apple Blue */
--success: #34C759;        /* Apple Green */
--warning: #FF9500;        /* Apple Orange */
--destructive: #FF3B30;    /* Apple Red */
--secondary: #5856D6;      /* Apple Purple */

/* Neutral Tones */
--text-primary: rgba(0, 0, 0, 0.85);
--text-secondary: rgba(0, 0, 0, 0.55);
--text-tertiary: rgba(0, 0, 0, 0.35);

/* Dark Mode (optional) */
--glass-bg-dark: rgba(30, 30, 30, 0.7);
--text-primary-dark: rgba(255, 255, 255, 0.85);
```

**Typography (SF Pro):**
```css
/* Headings */
--font-display: 'SF Pro Display', -apple-system, BlinkMacSystemFont, sans-serif;
--font-text: 'SF Pro Text', -apple-system, BlinkMacSystemFont, sans-serif;
--font-mono: 'SF Mono', 'Monaco', monospace;

/* Scale */
--text-xs: 11px;
--text-sm: 13px;
--text-base: 15px;
--text-lg: 17px;
--text-xl: 20px;
--text-2xl: 24px;
--text-3xl: 28px;
```

**Spacing (8px base unit):**
```css
--space-1: 4px;
--space-2: 8px;
--space-3: 12px;
--space-4: 16px;
--space-6: 24px;
--space-8: 32px;
--space-12: 48px;
```

**Glassmorphism Effect:**
```css
.glass-card {
    background: rgba(255, 255, 255, 0.7);
    backdrop-filter: blur(20px) saturate(180%);
    border: 1px solid rgba(255, 255, 255, 0.2);
    border-radius: 16px;
    box-shadow:
        0 8px 32px rgba(0, 0, 0, 0.08),
        inset 0 1px 0 rgba(255, 255, 255, 0.5);
}

.glass-card:hover {
    background: rgba(255, 255, 255, 0.8);
    box-shadow:
        0 12px 48px rgba(0, 0, 0, 0.12),
        inset 0 1px 0 rgba(255, 255, 255, 0.6);
    transition: all 0.2s ease;
}
```

**Animations (subtle, purposeful):**
```css
/* Micro-interactions */
@keyframes fadeIn {
    from { opacity: 0; transform: translateY(8px); }
    to { opacity: 1; transform: translateY(0); }
}

@keyframes scaleIn {
    from { opacity: 0; transform: scale(0.95); }
    to { opacity: 1; transform: scale(1); }
}

/* Spring physics */
.spring-transition {
    transition: transform 0.3s cubic-bezier(0.34, 1.56, 0.64, 1);
}
```

## Screen Designs

### 1. Dashboard (Main Screen)

```
┌─────────────────────────────────────────────────────────────────┐
│  🔐 Agent Vault                                  [+] [⚙️]  [🔒]  │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  ┌─────────────────────────────────────────────────────────┐   │
│  │ 🔍 Search credentials...                                │   │
│  └─────────────────────────────────────────────────────────┘   │
│                                                                 │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐         │
│  │ 📊 All       │  │ ✓ Active     │  │ ⚠ Expiring   │         │
│  │ 12           │  │ 10           │  │ 2            │         │
│  └──────────────┘  └──────────────┘  └──────────────┘         │
│                                                                 │
│  ┌───────────────────────────────────────────────────────┐     │
│  │ [✓] openai-production              OpenAI  prod       │     │
│  │     sk-proj-abc...xyz                                 │     │
│  │     Last rotated: 2 days ago             [Rotate →]   │     │
│  └───────────────────────────────────────────────────────┘     │
│                                                                 │
│  ┌───────────────────────────────────────────────────────┐     │
│  │ [⚠] supabase-prod                  Supabase  prod     │     │
│  │     eyJh...service_role                               │     │
│  │     Last rotated: 87 days ago            [Rotate →]   │     │
│  └───────────────────────────────────────────────────────┘     │
│                                                                 │
│  ┌───────────────────────────────────────────────────────┐     │
│  │ [✓] stripe-live                    Stripe   prod      │     │
│  │     sk_live_...                                       │     │
│  │     Last rotated: 15 days ago            [Rotate →]   │     │
│  └───────────────────────────────────────────────────────┘     │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

**Glass Card Styling:**
- Each credential is a glassmorphic card
- Hover state: subtle lift and brightness increase
- Status indicators: ✓ (green), ⚠ (yellow), ✗ (red)
- Smooth fade-in animation on load

**Interactions:**
- Click card → Detail view
- Click "Rotate" → Rotation modal
- Search → Real-time filter
- [+] button → Add credential modal
- [🔒] indicator → Security status (always locked)

### 2. Add Credential Modal

```
┌─────────────────────────────────────────────────────────┐
│  Add Credential                                    [×]  │
├─────────────────────────────────────────────────────────┤
│                                                         │
│  Service Type                                           │
│  ┌───────────────────────────────────────────────────┐ │
│  │ 🤖 OpenAI                                      ▼  │ │
│  └───────────────────────────────────────────────────┘ │
│                                                         │
│  Name                                                   │
│  ┌───────────────────────────────────────────────────┐ │
│  │ production                                        │ │
│  └───────────────────────────────────────────────────┘ │
│  💡 Examples: production, development, personal         │
│                                                         │
│  Environment (Optional)                                 │
│  ┌─────────┐  ┌─────────┐  ┌─────────┐                │
│  │ ○ Dev   │  │ ● Prod  │  │ ○ Stage │                │
│  └─────────┘  └─────────┘  └─────────┘                │
│                                                         │
│  Secret Key                                             │
│  ┌───────────────────────────────────────────────────┐ │
│  │ ••••••••••••••••••••••                 [👁]  [📋] │ │
│  └───────────────────────────────────────────────────┘ │
│  💡 Starts with sk-proj-...                            │
│                                                         │
│  Public Key (Optional)                                  │
│  ┌───────────────────────────────────────────────────┐ │
│  │ pk_...                                 [👁]  [📋] │ │
│  └───────────────────────────────────────────────────┘ │
│                                                         │
│  URL (Optional)                                         │
│  ┌───────────────────────────────────────────────────┐ │
│  │ https://api.openai.com/v1                         │ │
│  └───────────────────────────────────────────────────┘ │
│                                                         │
│              [Cancel]           [Save Credential]       │
│                                                         │
└─────────────────────────────────────────────────────────┘
```

**Features:**
- Service type dropdown with icons
- Smart hints based on selected service
- Toggle password visibility (👁 icon)
- Copy to clipboard (📋 icon)
- Real-time validation
- Smooth modal animation (scale + fade)

**Service Templates:**
- Pre-filled hints per service
- URL auto-populated for known services
- Field visibility based on service (OpenAI doesn't need public key)

### 3. Credential Detail View

```
┌─────────────────────────────────────────────────────────────────┐
│  ← Back                    openai-production              [⋮]  │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  ┌─────────────────────────────────────────────────────────┐   │
│  │ 🤖 OpenAI • Production                                  │   │
│  │                                                         │   │
│  │ Secret Key                             [👁 Show] [📋]   │   │
│  │ sk-proj-abc...xyz                                       │   │
│  │                                                         │   │
│  │ Status: ✓ Active                                        │   │
│  │ Last rotated: 2 days ago                                │   │
│  │ Created: Jan 15, 2026                                   │   │
│  │                                                         │   │
│  │                    [Rotate Now]                         │   │
│  └─────────────────────────────────────────────────────────┘   │
│                                                                 │
│  Rotation History                                               │
│  ┌─────────────────────────────────────────────────────────┐   │
│  │ Feb 4, 2026 • 10:32 AM                                  │   │
│  │ Rotated by: manual                                      │   │
│  │ Plugin: openai                                          │   │
│  │ Fields: secret_key                                      │   │
│  └─────────────────────────────────────────────────────────┘   │
│                                                                 │
│  ┌─────────────────────────────────────────────────────────┐   │
│  │ Feb 1, 2026 • 3:15 PM                                   │   │
│  │ Rotated by: scheduler                                   │   │
│  │ Plugin: openai                                          │   │
│  │ Fields: secret_key                                      │   │
│  └─────────────────────────────────────────────────────────┘   │
│                                                                 │
│  ┌─────────────────────────────────────────────────────────┐   │
│  │ Jan 15, 2026 • 9:00 AM                                  │   │
│  │ Created                                                 │   │
│  └─────────────────────────────────────────────────────────┘   │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

**Features:**
- Large, clear credential display
- Timeline-style rotation history
- One-click copy
- Secure show/hide toggle
- Rotation button prominent but not alarming

### 4. Trust Indicator (Always Visible)

**Top-right corner lock icon:**

```
┌─────┐
│ 🔒  │  ← Green when vault locked
└─────┘

┌─────┐
│ 🔓  │  ← Yellow when vault unlocked (temp)
└─────┘
```

**On hover/click:**
```
┌───────────────────────────────────┐
│ Vault Status                      │
├───────────────────────────────────┤
│ ✓ Vault encrypted                 │
│ ✓ SQLCipher + AES-256-GCM         │
│ ✓ Argon2id key derivation         │
│ ✓ No network access               │
│ ✓ Local storage only              │
│                                   │
│ Last activity: 2 minutes ago      │
│ Auto-lock: 15 minutes             │
└───────────────────────────────────┘
```

**Security Principles:**
- Always-visible lock status
- Explicit "no network" indicator
- Clear encryption method disclosure
- Auto-lock timer visible
- Feels like iOS keychain trust

### 5. Settings Screen

```
┌─────────────────────────────────────────────────────────────────┐
│  Settings                                              [×]      │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  Security                                                       │
│  ┌─────────────────────────────────────────────────────────┐   │
│  │ Auto-lock after                                         │   │
│  │ ┌──────┐  ┌──────┐  ┌──────┐  ┌──────┐                │   │
│  │ │ 5min │  │●15min│  │ 30min│  │ Never│                │   │
│  │ └──────┘  └──────┘  └──────┘  └──────┘                │   │
│  │                                                         │   │
│  │ [Change Master Password]                               │   │
│  └─────────────────────────────────────────────────────────┘   │
│                                                                 │
│  Rotation Plugins                                               │
│  ┌─────────────────────────────────────────────────────────┐   │
│  │ [✓] OpenAI        Secret-only rotation                  │   │
│  │ [✓] Supabase      Service role rotation                 │   │
│  │ [✓] Anthropic     Secret-only rotation                  │   │
│  │ [○] Stripe        Not configured                        │   │
│  │ [○] GitHub        Not configured                        │   │
│  └─────────────────────────────────────────────────────────┘   │
│                                                                 │
│  Preferences                                                    │
│  ┌─────────────────────────────────────────────────────────┐   │
│  │ [✓] Show credential previews                            │   │
│  │ [✓] Enable clipboard auto-clear (30s)                   │   │
│  │ [○] Dark mode                                           │   │
│  └─────────────────────────────────────────────────────────┘   │
│                                                                 │
│  About                                                          │
│  ┌─────────────────────────────────────────────────────────┐   │
│  │ Agent Vault v1.0.0                                      │   │
│  │ Encrypted credential storage for AI agents              │   │
│  │                                                         │   │
│  │ [View on GitHub] [Documentation] [Report Issue]        │   │
│  └─────────────────────────────────────────────────────────┘   │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

## Technical Implementation

### Tech Stack (Quick Prototype)

```
Frontend:     React 18 + TypeScript
Build Tool:   Vite (fast hot reload)
Styling:      Tailwind CSS + custom glassmorphism
Components:   Radix UI primitives (accessible)
Icons:        SF Symbols (Apple icons) or Lucide
Fonts:        SF Pro (system font fallback)
State:        Zustand (lightweight, no Redux overhead)
Mock Data:    Hardcoded JSON (no real credentials)
```

**Why NOT Tauri yet:**
- Vite is faster for iteration
- No build complexity
- Easy to test in browser
- Can port to Tauri later (same React code)

### Project Structure

```
frontend/
├── src/
│   ├── components/
│   │   ├── Dashboard.tsx
│   │   ├── CredentialCard.tsx
│   │   ├── AddCredentialModal.tsx
│   │   ├── CredentialDetail.tsx
│   │   ├── TrustIndicator.tsx
│   │   ├── SettingsPanel.tsx
│   │   └── GlassCard.tsx          # Reusable glass component
│   ├── lib/
│   │   ├── mockData.ts             # Fake credentials
│   │   └── types.ts                # Credential types
│   ├── styles/
│   │   ├── glass.css               # Glassmorphism effects
│   │   └── animations.css          # Transitions
│   ├── App.tsx
│   └── main.tsx
├── public/
├── index.html
├── package.json
├── tailwind.config.js
├── vite.config.ts
└── tsconfig.json
```

### Mock Data Structure

```typescript
// src/lib/mockData.ts
export const mockCredentials: Credential[] = [
  {
    id: "1",
    name: "openai-production",
    apiType: "openai",
    environment: "prod",
    secretKey: "sk-proj-abc123xyz...789",
    publicKey: null,
    url: null,
    status: "active",
    lastRotated: new Date("2026-02-04"),
    createdAt: new Date("2026-01-15"),
    rotationHistory: [
      {
        id: "r1",
        rotatedAt: new Date("2026-02-04"),
        rotatedBy: "manual",
        plugin: "openai",
        fields: ["secret_key"],
      },
      {
        id: "r2",
        rotatedAt: new Date("2026-02-01"),
        rotatedBy: "scheduler",
        plugin: "openai",
        fields: ["secret_key"],
      },
    ],
  },
  {
    id: "2",
    name: "supabase-prod",
    apiType: "supabase",
    environment: "prod",
    secretKey: "eyJh...service_role",
    publicKey: "eyJh...anon",
    url: "https://xyz.supabase.co",
    status: "warning",  // 87 days old
    lastRotated: new Date("2025-11-09"),
    createdAt: new Date("2025-10-01"),
    rotationHistory: [],
  },
  {
    id: "3",
    name: "stripe-live",
    apiType: "stripe",
    environment: "prod",
    secretKey: "sk_live_abc123...",
    publicKey: "pk_live_xyz789...",
    url: null,
    status: "active",
    lastRotated: new Date("2026-01-20"),
    createdAt: new Date("2025-12-01"),
    rotationHistory: [],
  },
];
```

## Development Phases

### Phase 2.5.1: Setup (30 min)

```bash
cd ~/Documents/Projects/api-vault
mkdir frontend
cd frontend

# Initialize Vite + React + TypeScript
npm create vite@latest . -- --template react-ts

# Install dependencies
npm install
npm install -D tailwindcss postcss autoprefixer
npm install @radix-ui/react-dialog @radix-ui/react-select
npm install zustand
npm install lucide-react

# Setup Tailwind
npx tailwindcss init -p
```

### Phase 2.5.2: Glass Design System (1 hour)

Create reusable glass components and styles:

```tsx
// src/components/GlassCard.tsx
export function GlassCard({
  children,
  hover = true,
  className
}: GlassCardProps) {
  return (
    <div className={cn(
      "glass-card",
      hover && "hover:glass-card-hover",
      className
    )}>
      {children}
    </div>
  );
}
```

```css
/* src/styles/glass.css */
.glass-card {
  background: rgba(255, 255, 255, 0.7);
  backdrop-filter: blur(20px) saturate(180%);
  border: 1px solid rgba(255, 255, 255, 0.2);
  border-radius: 16px;
  box-shadow:
    0 8px 32px rgba(0, 0, 0, 0.08),
    inset 0 1px 0 rgba(255, 255, 255, 0.5);
}

.glass-card-hover {
  transition: all 0.2s cubic-bezier(0.4, 0, 0.2, 1);
}

.glass-card-hover:hover {
  background: rgba(255, 255, 255, 0.8);
  transform: translateY(-2px);
  box-shadow:
    0 12px 48px rgba(0, 0, 0, 0.12),
    inset 0 1px 0 rgba(255, 255, 255, 0.6);
}
```

### Phase 2.5.3: Dashboard (2 hours)

Implement main screen with credential cards.

### Phase 2.5.4: Add Modal (1 hour)

Create add credential flow with service templates.

### Phase 2.5.5: Detail View (1 hour)

Show credential details and rotation history.

### Phase 2.5.6: Trust Indicator (30 min)

Always-visible security status.

### Phase 2.5.7: Settings (1 hour)

Settings panel with plugin status.

**Total: ~7 hours to complete visual prototype**

## Testing Plan

**Visual Testing:**
- [ ] Glass effect looks good on different backgrounds
- [ ] Animations feel Apple-smooth (no jank)
- [ ] Typography scales properly
- [ ] Colors are accessible (WCAG AA)
- [ ] Dark mode (optional but nice)

**UX Testing:**
- [ ] Can add credential in < 30 seconds
- [ ] Search/filter feels instant
- [ ] Trust indicator is always visible
- [ ] Rotation button is obvious
- [ ] No confusing states

**Flow Testing:**
- [ ] Add credential → appears immediately
- [ ] Click card → detail view works
- [ ] Rotate → modal confirms action
- [ ] Settings → changes persist (mock)

## Success Criteria

Phase 2.5 complete when:
- ✅ All 5 screens implemented
- ✅ Glass aesthetic nailed
- ✅ Interactions feel Apple-smooth
- ✅ Mock data flows through all screens
- ✅ Trust indicators clearly visible
- ✅ You (visual person) feel confident in the design
- ✅ Ready to port to Tauri for Phase 4

## What This Gives You

**For Phase 4 (Tauri Desktop App):**
- Exact visual reference
- Tested component structure
- Proven UX flows
- CSS/animation library ready to port
- Design system documented

**For Users:**
- They see the vision before it's built
- Feedback loop is faster
- Trust is designed in from the start
- Apple aesthetic attracts professional users

**For You:**
- Visual validation (you're a visual person!)
- Iterate fast without Tauri complexity
- Test with friends/team easily
- Confidence before full build

---

**Next Step:** Want me to generate the Vite + React scaffold with the glass design system, or should we discuss the design further first?
