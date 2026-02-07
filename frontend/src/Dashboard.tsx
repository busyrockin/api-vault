import { Search } from 'lucide-react'
import { useStore } from './store'
import { SERVICE_META } from './data'

function relativeTime(ts: number): string {
  const diff = Math.round((ts - Date.now()) / 1000)
  const units: [Intl.RelativeTimeFormatUnit, number][] = [
    ['day', 86400], ['hour', 3600], ['minute', 60],
  ]
  const rtf = new Intl.RelativeTimeFormat('en', { numeric: 'auto' })
  for (const [unit, secs] of units) {
    if (Math.abs(diff) >= secs) return rtf.format(Math.round(diff / secs), unit)
  }
  return 'just now'
}

const statusColors = {
  active: 'bg-apple-green/15 text-apple-green',
  warning: 'bg-apple-orange/15 text-apple-orange',
  expired: 'bg-apple-red/15 text-apple-red',
} as const

export default function Dashboard() {
  const { filtered, search, setSearch, navigate, credentials } = useStore()
  const cards = filtered()
  const expiring = credentials.filter((c) => c.status === 'warning').length

  return (
    <div className="animate-fade-in">
      {/* Search */}
      <div className="glass flex items-center gap-3 px-4 py-3 mb-6">
        <Search className="w-4 h-4 text-apple-gray" />
        <input
          type="text"
          value={search}
          onChange={(e) => setSearch(e.target.value)}
          placeholder="Search credentials..."
          className="bg-transparent outline-none w-full text-sm placeholder:text-apple-gray/60"
        />
      </div>

      {/* Stats */}
      <div className="grid grid-cols-3 gap-3 mb-6">
        {[
          { label: 'All', value: credentials.length, color: 'text-apple-blue' },
          { label: 'Active', value: credentials.filter((c) => c.status === 'active').length, color: 'text-apple-green' },
          { label: 'Expiring', value: expiring, color: expiring > 0 ? 'text-apple-orange' : 'text-apple-gray' },
        ].map((stat) => (
          <div key={stat.label} className="glass-subtle text-center py-3 px-2">
            <div className={`text-2xl font-bold ${stat.color}`}>{stat.value}</div>
            <div className="text-xs text-apple-gray mt-0.5">{stat.label}</div>
          </div>
        ))}
      </div>

      {/* Credential Cards */}
      <div className="space-y-3">
        {cards.map((cred, i) => {
          const meta = SERVICE_META[cred.service] ?? SERVICE_META.other
          return (
            <button
              key={cred.id}
              onClick={() => navigate('detail', cred.id)}
              className="glass w-full text-left px-5 py-4 flex items-center gap-4 transition-all hover:scale-[1.01] active:scale-[0.99] animate-fade-in"
              style={{ animationDelay: `${i * 50}ms` }}
            >
              <span className="text-2xl" role="img" aria-label={meta.name}>{meta.icon}</span>
              <div className="flex-1 min-w-0">
                <div className="font-semibold text-sm truncate">{cred.name}</div>
                <div className="text-xs text-apple-gray mt-0.5">
                  {cred.customService ?? meta.name} · {cred.environment} · rotated {relativeTime(cred.lastRotated)}
                </div>
              </div>
              <span className={`badge ${statusColors[cred.status]}`}>{cred.status}</span>
            </button>
          )
        })}
        {cards.length === 0 && (
          <div className="text-center text-apple-gray text-sm py-12">
            No credentials match "{search}"
          </div>
        )}
      </div>
    </div>
  )
}
