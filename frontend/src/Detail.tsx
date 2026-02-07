import { useState } from 'react'
import { ArrowLeft, Eye, EyeOff, Copy, Check, RotateCw } from 'lucide-react'
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

const statusStyles = {
  active: 'bg-apple-green/15 text-apple-green',
  warning: 'bg-apple-orange/15 text-apple-orange',
  expired: 'bg-apple-red/15 text-apple-red',
} as const

function SecretField({ label, value }: { label: string; value: string }) {
  const [visible, setVisible] = useState(false)
  const [copied, setCopied] = useState(false)

  const copy = () => {
    navigator.clipboard.writeText(value)
    setCopied(true)
    setTimeout(() => setCopied(false), 1500)
  }

  const masked = value.slice(0, 8) + '\u2022'.repeat(Math.min(value.length - 8, 24))

  return (
    <div className="glass-subtle p-4">
      <div className="text-xs text-apple-gray mb-2 font-medium">{label}</div>
      <div className="flex items-center gap-2">
        <code className="flex-1 text-sm font-mono break-all select-all">
          {visible ? value : masked}
        </code>
        <button onClick={() => setVisible(!visible)} className="p-1.5 rounded-lg hover:bg-white/50 transition-colors">
          {visible ? <EyeOff className="w-4 h-4" /> : <Eye className="w-4 h-4" />}
        </button>
        <button onClick={copy} className="p-1.5 rounded-lg hover:bg-white/50 transition-colors">
          {copied ? <Check className="w-4 h-4 text-apple-green" /> : <Copy className="w-4 h-4" />}
        </button>
      </div>
    </div>
  )
}

function PlainField({ label, value }: { label: string; value: string }) {
  const [copied, setCopied] = useState(false)

  const copy = () => {
    navigator.clipboard.writeText(value)
    setCopied(true)
    setTimeout(() => setCopied(false), 1500)
  }

  return (
    <div className="glass-subtle p-4">
      <div className="text-xs text-apple-gray mb-2 font-medium">{label}</div>
      <div className="flex items-center gap-2">
        <code className="flex-1 text-sm font-mono break-all select-all">{value}</code>
        <button onClick={copy} className="p-1.5 rounded-lg hover:bg-white/50 transition-colors">
          {copied ? <Check className="w-4 h-4 text-apple-green" /> : <Copy className="w-4 h-4" />}
        </button>
      </div>
    </div>
  )
}

export default function Detail() {
  const { credentials, selectedId, navigate } = useStore()
  const cred = credentials.find((c) => c.id === selectedId)
  if (!cred) return null

  const meta = SERVICE_META[cred.service] ?? SERVICE_META.other
  const displayName = cred.customService ?? meta.name

  return (
    <div className="animate-fade-in">
      <button onClick={() => navigate('dashboard')} className="flex items-center gap-1.5 text-apple-blue text-sm mb-6 hover:underline">
        <ArrowLeft className="w-4 h-4" /> Back
      </button>

      {/* Header */}
      <div className="glass p-6 mb-4">
        <div className="flex items-center gap-4 mb-4">
          <span className="text-4xl">{meta.icon}</span>
          <div className="flex-1">
            <h2 className="text-xl font-bold">{cred.name}</h2>
            <div className="text-sm text-apple-gray mt-0.5">{displayName} · {cred.environment}</div>
          </div>
          <span className={`badge ${statusStyles[cred.status]}`}>{cred.status}</span>
        </div>

        <div className="grid grid-cols-2 gap-4 text-sm">
          <div>
            <span className="text-apple-gray">Created</span>
            <div className="font-medium mt-0.5">{new Date(cred.createdAt).toLocaleDateString()}</div>
          </div>
          <div>
            <span className="text-apple-gray">Last rotated</span>
            <div className="font-medium mt-0.5">{relativeTime(cred.lastRotated)}</div>
          </div>
        </div>
      </div>

      {/* Fields — driven by schema */}
      <div className="space-y-3 mb-4">
        {meta.fields.map((field) => {
          const value = cred.fields[field.key]
          if (!value) return null
          return field.type === 'secret'
            ? <SecretField key={field.key} label={field.label} value={value} />
            : <PlainField key={field.key} label={field.label} value={value} />
        })}
      </div>

      {/* Rotation History */}
      <div className="glass p-5">
        <h3 className="font-semibold text-sm mb-4 flex items-center gap-2">
          <RotateCw className="w-4 h-4" /> Rotation History
        </h3>
        {cred.rotations.length === 0 ? (
          <p className="text-sm text-apple-gray">No rotations yet</p>
        ) : (
          <div className="space-y-3">
            {cred.rotations.map((rot) => (
              <div key={rot.id} className="flex items-start gap-3 text-sm">
                <div className="w-2 h-2 rounded-full bg-apple-blue mt-1.5 shrink-0" />
                <div>
                  <div className="font-medium">{rot.reason}</div>
                  <div className="text-apple-gray text-xs mt-0.5">
                    {relativeTime(rot.timestamp)} · {rot.automatic ? 'Automatic' : 'Manual'}
                  </div>
                </div>
              </div>
            ))}
          </div>
        )}
      </div>
    </div>
  )
}
