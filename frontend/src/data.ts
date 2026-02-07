import type { Credential, ServiceMeta, FieldSchema } from './types'

const now = Date.now()
const day = 86_400_000

// Helper to reduce boilerplate in field definitions
const f = (key: string, label: string, type: FieldSchema['type'], required: boolean, hint: string): FieldSchema =>
  ({ key, label, type, required, hint })

export const SERVICE_META: Record<string, ServiceMeta> = {
  // ── AI / ML ──
  openai:      { name: 'OpenAI',       color: '#10a37f', icon: '🤖', category: 'AI / ML', fields: [f('secret', 'API Key', 'secret', true, 'sk-...')] },
  anthropic:   { name: 'Anthropic',    color: '#d4a574', icon: '🧠', category: 'AI / ML', fields: [f('secret', 'API Key', 'secret', true, 'sk-ant-...')] },
  google_ai:   { name: 'Google AI',    color: '#4285f4', icon: '✨', category: 'AI / ML', fields: [f('secret', 'API Key', 'secret', true, 'AIza...')] },
  replicate:   { name: 'Replicate',    color: '#3d3d3d', icon: '🔁', category: 'AI / ML', fields: [f('secret', 'API Token', 'secret', true, 'r8_...')] },
  huggingface: { name: 'Hugging Face', color: '#ffbd45', icon: '🤗', category: 'AI / ML', fields: [f('secret', 'Access Token', 'secret', true, 'hf_...')] },
  cohere:      { name: 'Cohere',       color: '#39594d', icon: '🔤', category: 'AI / ML', fields: [f('secret', 'API Key', 'secret', true, '')] },
  mistral:     { name: 'Mistral',      color: '#f54e42', icon: '🌬️', category: 'AI / ML', fields: [f('secret', 'API Key', 'secret', true, '')] },
  groq:        { name: 'Groq',         color: '#f55036', icon: '⚡', category: 'AI / ML', fields: [f('secret', 'API Key', 'secret', true, 'gsk_...')] },
  together:    { name: 'Together AI',  color: '#0f6fff', icon: '🤝', category: 'AI / ML', fields: [f('secret', 'API Key', 'secret', true, '')] },

  // ── Cloud ──
  aws:          { name: 'AWS',          color: '#ff9900', icon: '☁️', category: 'Cloud', fields: [f('access_key', 'Access Key ID', 'secret', true, 'AKIA...'), f('secret', 'Secret Access Key', 'secret', true, ''), f('region', 'Region', 'text', false, 'us-east-1')] },
  gcp:          { name: 'GCP',          color: '#4285f4', icon: '🌐', category: 'Cloud', fields: [f('secret', 'API Key or Service Account JSON', 'secret', true, ''), f('project_id', 'Project ID', 'text', true, 'my-project-123')] },
  azure:        { name: 'Azure',        color: '#0078d4', icon: '🔷', category: 'Cloud', fields: [f('secret', 'Client Secret', 'secret', true, ''), f('client_id', 'Client ID', 'text', true, ''), f('tenant_id', 'Tenant ID', 'text', true, '')] },
  vercel:       { name: 'Vercel',       color: '#000000', icon: '▲',  category: 'Cloud', fields: [f('secret', 'API Token', 'secret', true, '')] },
  netlify:      { name: 'Netlify',      color: '#00c7b7', icon: '🌿', category: 'Cloud', fields: [f('secret', 'Personal Access Token', 'secret', true, '')] },
  flyio:        { name: 'Fly.io',       color: '#7b3fe4', icon: '🪰', category: 'Cloud', fields: [f('secret', 'API Token', 'secret', true, 'fo1_...')] },
  railway:      { name: 'Railway',      color: '#0b0d0e', icon: '🚂', category: 'Cloud', fields: [f('secret', 'API Token', 'secret', true, '')] },
  cloudflare:   { name: 'Cloudflare',   color: '#f38020', icon: '🔶', category: 'Cloud', fields: [f('secret', 'API Token', 'secret', true, ''), f('account_id', 'Account ID', 'text', false, ''), f('zone_id', 'Zone ID', 'text', false, '')] },
  digitalocean: { name: 'DigitalOcean', color: '#0080ff', icon: '🐳', category: 'Cloud', fields: [f('secret', 'API Token', 'secret', true, 'dop_v1_...')] },

  // ── Database ──
  supabase:  { name: 'Supabase',      color: '#3ecf8e', icon: '⚡', category: 'Database', fields: [f('secret', 'Service Role Key', 'secret', true, 'eyJ...'), f('anon_key', 'Anon Key', 'text', false, 'eyJ...'), f('url', 'Project URL', 'url', true, 'https://xyz.supabase.co')] },
  firebase:  { name: 'Firebase',      color: '#ffca28', icon: '🔥', category: 'Database', fields: [f('secret', 'Service Account Key (JSON)', 'secret', true, ''), f('project_id', 'Project ID', 'text', true, '')] },
  planetscale:{ name: 'PlanetScale',  color: '#000000', icon: '🪐', category: 'Database', fields: [f('secret', 'Service Token', 'secret', true, 'pscale_tkn_...'), f('url', 'Connection String', 'url', false, '')] },
  turso:     { name: 'Turso',         color: '#4ff8d2', icon: '🐢', category: 'Database', fields: [f('secret', 'Auth Token', 'secret', true, ''), f('url', 'Database URL', 'url', true, 'libsql://...')] },
  neon:      { name: 'Neon',          color: '#00e599', icon: '🟢', category: 'Database', fields: [f('secret', 'API Key', 'secret', true, ''), f('url', 'Connection String', 'url', false, 'postgres://...')] },
  upstash:   { name: 'Upstash',       color: '#00e9a3', icon: '🔺', category: 'Database', fields: [f('secret', 'REST Token', 'secret', true, ''), f('url', 'REST URL', 'url', true, 'https://...')] },
  mongodb:   { name: 'MongoDB Atlas', color: '#00ed64', icon: '🍃', category: 'Database', fields: [f('secret', 'API Key', 'secret', true, ''), f('url', 'Connection String', 'url', false, 'mongodb+srv://...')] },
  pinecone:  { name: 'Pinecone',      color: '#000000', icon: '🌲', category: 'Database', fields: [f('secret', 'API Key', 'secret', true, ''), f('url', 'Environment URL', 'url', false, '')] },

  // ── Payments ──
  stripe:       { name: 'Stripe',        color: '#635bff', icon: '💳', category: 'Payments', fields: [f('secret', 'Secret Key', 'secret', true, 'sk_live_...'), f('public_key', 'Publishable Key', 'text', false, 'pk_live_...'), f('webhook_secret', 'Webhook Secret', 'secret', false, 'whsec_...')] },
  revenuecat:   { name: 'RevenueCat',    color: '#f25a5a', icon: '😺', category: 'Payments', fields: [f('public_key', 'Public API Key', 'text', true, ''), f('secret', 'Secret API Key', 'secret', true, '')] },
  paddle:       { name: 'Paddle',        color: '#ffdd35', icon: '🏓', category: 'Payments', fields: [f('secret', 'API Key', 'secret', true, ''), f('vendor_id', 'Vendor ID', 'text', true, '')] },
  lemonsqueezy: { name: 'LemonSqueezy',  color: '#ffc233', icon: '🍋', category: 'Payments', fields: [f('secret', 'API Key', 'secret', true, '')] },

  // ── Communication ──
  twilio:   { name: 'Twilio',   color: '#f22f46', icon: '📱', category: 'Communication', fields: [f('account_sid', 'Account SID', 'text', true, 'AC...'), f('secret', 'Auth Token', 'secret', true, '')] },
  sendgrid: { name: 'SendGrid', color: '#1a82e2', icon: '📧', category: 'Communication', fields: [f('secret', 'API Key', 'secret', true, 'SG....')] },
  resend:   { name: 'Resend',   color: '#000000', icon: '📮', category: 'Communication', fields: [f('secret', 'API Key', 'secret', true, 're_...')] },
  postmark: { name: 'Postmark', color: '#ffde00', icon: '📬', category: 'Communication', fields: [f('secret', 'Server Token', 'secret', true, '')] },
  slack:    { name: 'Slack',    color: '#4a154b', icon: '💬', category: 'Communication', fields: [f('secret', 'Bot Token', 'secret', true, 'xoxb-...'), f('signing_secret', 'Signing Secret', 'secret', false, '')] },
  discord:  { name: 'Discord',  color: '#5865f2', icon: '🎮', category: 'Communication', fields: [f('secret', 'Bot Token', 'secret', true, ''), f('client_id', 'Application ID', 'text', false, '')] },

  // ── Dev Tools ──
  github:  { name: 'GitHub',   color: '#24292e', icon: '🐙', category: 'Dev Tools', fields: [f('secret', 'Personal Access Token', 'secret', true, 'ghp_...')] },
  gitlab:  { name: 'GitLab',   color: '#fc6d26', icon: '🦊', category: 'Dev Tools', fields: [f('secret', 'Access Token', 'secret', true, 'glpat-...')] },
  linear:  { name: 'Linear',   color: '#5e6ad2', icon: '📐', category: 'Dev Tools', fields: [f('secret', 'API Key', 'secret', true, 'lin_api_...')] },
  sentry:  { name: 'Sentry',   color: '#362d59', icon: '🐛', category: 'Dev Tools', fields: [f('secret', 'Auth Token', 'secret', true, 'sntrys_...'), f('dsn', 'DSN', 'url', false, 'https://...@sentry.io/...')] },
  datadog: { name: 'Datadog',  color: '#632ca6', icon: '🐶', category: 'Dev Tools', fields: [f('secret', 'API Key', 'secret', true, ''), f('app_key', 'Application Key', 'secret', false, '')] },
  algolia: { name: 'Algolia',  color: '#003dff', icon: '🔎', category: 'Dev Tools', fields: [f('secret', 'Admin API Key', 'secret', true, ''), f('app_id', 'Application ID', 'text', true, '')] },

  // ── Media ──
  youtube:    { name: 'YouTube',    color: '#ff0000', icon: '▶️', category: 'Media', fields: [f('secret', 'API Key', 'secret', true, 'AIza...')] },
  cloudinary: { name: 'Cloudinary', color: '#3448c5', icon: '🖼️', category: 'Media', fields: [f('secret', 'API Secret', 'secret', true, ''), f('api_key', 'API Key', 'text', true, ''), f('cloud_name', 'Cloud Name', 'text', true, '')] },
  mux:        { name: 'Mux',        color: '#fb3465', icon: '🎬', category: 'Media', fields: [f('token_id', 'Token ID', 'text', true, ''), f('secret', 'Token Secret', 'secret', true, '')] },

  // ── Auth ──
  auth0: { name: 'Auth0', color: '#eb5424', icon: '🔐', category: 'Auth', fields: [f('secret', 'Client Secret', 'secret', true, ''), f('client_id', 'Client ID', 'text', true, ''), f('domain', 'Domain', 'url', true, 'https://your-tenant.auth0.com')] },
  clerk: { name: 'Clerk', color: '#6c47ff', icon: '🔑', category: 'Auth', fields: [f('secret', 'Secret Key', 'secret', true, 'sk_live_...'), f('public_key', 'Publishable Key', 'text', false, 'pk_live_...')] },

  // ── Other ──
  other: { name: 'Other', color: '#8e8e93', icon: '🔑', category: 'Other', fields: [f('secret', 'API Key', 'secret', true, 'Your API key or token')] },
}

// Derive ordered categories for <optgroup>
const CATEGORY_ORDER = ['AI / ML', 'Cloud', 'Database', 'Payments', 'Communication', 'Dev Tools', 'Media', 'Auth', 'Other']

export const SERVICE_CATEGORIES: { category: string; services: [string, ServiceMeta][] }[] =
  CATEGORY_ORDER.map((cat) => ({
    category: cat,
    services: Object.entries(SERVICE_META).filter(([, m]) => m.category === cat),
  }))

export const MOCK_CREDENTIALS: Credential[] = [
  {
    id: 'cred-001',
    name: 'openai-production',
    service: 'openai',
    environment: 'production',
    fields: { secret: 'sk-proj-abc123def456ghi789jkl012mno345pqr678stu901vwx234' },
    status: 'active',
    createdAt: now - 90 * day,
    lastRotated: now - 2 * day,
    rotations: [
      { id: 'rot-001', timestamp: now - 2 * day, reason: 'Scheduled rotation', automatic: true },
      { id: 'rot-002', timestamp: now - 32 * day, reason: 'Initial setup', automatic: false },
    ],
  },
  {
    id: 'cred-002',
    name: 'supabase-prod',
    service: 'supabase',
    environment: 'production',
    fields: {
      secret: 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJzdXBhYmFzZSJ9.mock',
      anon_key: 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJhbm9uIjoicHVibGljIn0.anon',
      url: 'https://xyzproject.supabase.co',
    },
    status: 'warning',
    createdAt: now - 87 * day,
    lastRotated: now - 87 * day,
    rotations: [],
  },
  {
    id: 'cred-003',
    name: 'stripe-live',
    service: 'stripe',
    environment: 'production',
    fields: {
      secret: 'sk_test_MOCK_stripe_secret_key_demo_value',
      public_key: 'pk_test_MOCK_stripe_public_key_demo_value',
    },
    status: 'active',
    createdAt: now - 60 * day,
    lastRotated: now - 15 * day,
    rotations: [
      { id: 'rot-003', timestamp: now - 15 * day, reason: 'Manual rotation', automatic: false },
    ],
  },
  {
    id: 'cred-004',
    name: 'github-personal',
    service: 'github',
    environment: 'development',
    fields: { secret: 'ghp_ABC123DEF456GHI789JKL012MNO345PQR678STU' },
    status: 'active',
    createdAt: now - 120 * day,
    lastRotated: now - 30 * day,
    rotations: [
      { id: 'rot-004', timestamp: now - 30 * day, reason: 'Token refresh', automatic: true },
    ],
  },
]
