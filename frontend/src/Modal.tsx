import { useEffect, useRef, useState } from 'react'
import { X } from 'lucide-react'
import { useStore } from './store'
import { SERVICE_META, SERVICE_CATEGORIES } from './data'
import type { Credential, Environment } from './types'

export default function Modal() {
  const { toggleAddModal, addCredential } = useStore()
  const ref = useRef<HTMLDialogElement>(null)
  const [service, setService] = useState('openai')
  const [customService, setCustomService] = useState('')
  const [name, setName] = useState('')
  const [env, setEnv] = useState<Environment>('production')
  const [fieldValues, setFieldValues] = useState<Record<string, string>>({})
  const [showSecrets, setShowSecrets] = useState<Record<string, boolean>>({})

  useEffect(() => {
    ref.current?.showModal()
    ref.current?.focus()
  }, [])

  const meta = SERVICE_META[service] ?? SERVICE_META.other

  // Reset field values when service changes
  useEffect(() => {
    setFieldValues({})
    setShowSecrets({})
  }, [service])

  const setField = (key: string, value: string) =>
    setFieldValues((prev) => ({ ...prev, [key]: value }))

  const canSubmit = meta.fields
    .filter((f) => f.required)
    .every((f) => fieldValues[f.key]?.trim())
    && (service !== 'other' || customService.trim())

  const submit = (e: React.FormEvent) => {
    e.preventDefault()
    const now = Date.now()
    const displayName = service === 'other' ? customService.trim() : meta.name
    const cred: Credential = {
      id: `cred-${crypto.randomUUID().slice(0, 8)}`,
      name: name || `${displayName.toLowerCase().replace(/\s+/g, '-')}-${env}`,
      service,
      ...(service === 'other' && { customService: customService.trim() }),
      environment: env,
      fields: Object.fromEntries(
        Object.entries(fieldValues).filter(([, v]) => v.trim()),
      ),
      status: 'active',
      createdAt: now,
      lastRotated: now,
      rotations: [],
    }
    addCredential(cred)
  }

  const close = () => {
    ref.current?.close()
    toggleAddModal()
  }

  const envs: Environment[] = ['production', 'staging', 'development']

  return (
    <dialog ref={ref} onClose={close} onClick={(e) => e.target === ref.current && close()}>
      <form onSubmit={submit} className="glass p-6 animate-scale-in">
        <div className="flex items-center justify-between mb-5">
          <h2 className="text-lg font-bold">Add Credential</h2>
          <button type="button" onClick={close} className="p-1 rounded-lg hover:bg-white/50 transition-colors">
            <X className="w-5 h-5" />
          </button>
        </div>

        <div className="space-y-4">
          {/* Service */}
          <div>
            <label className="text-xs font-medium text-apple-gray block mb-1.5">Service</label>
            <select
              value={service}
              onChange={(e) => setService(e.target.value)}
              className="w-full glass-subtle px-3 py-2 text-sm rounded-xl outline-none"
            >
              {SERVICE_CATEGORIES.map(({ category, services }) => (
                <optgroup key={category} label={category}>
                  {services.map(([key, m]) => (
                    <option key={key} value={key}>{m.icon} {m.name}</option>
                  ))}
                </optgroup>
              ))}
            </select>
          </div>

          {/* Custom service name */}
          {service === 'other' && (
            <div>
              <label className="text-xs font-medium text-apple-gray block mb-1.5">Service Name</label>
              <input
                type="text"
                value={customService}
                onChange={(e) => setCustomService(e.target.value)}
                required
                placeholder="e.g. RevenueCat, Plaid, Mapbox"
                className="w-full glass-subtle px-3 py-2 text-sm rounded-xl outline-none placeholder:text-apple-gray/50"
              />
            </div>
          )}

          {/* Name */}
          <div>
            <label className="text-xs font-medium text-apple-gray block mb-1.5">Name</label>
            <input
              type="text"
              value={name}
              onChange={(e) => setName(e.target.value)}
              placeholder={`${service === 'other' ? (customService || 'custom').toLowerCase().replace(/\s+/g, '-') : service}-${env}`}
              className="w-full glass-subtle px-3 py-2 text-sm rounded-xl outline-none placeholder:text-apple-gray/50"
            />
          </div>

          {/* Environment */}
          <div>
            <label className="text-xs font-medium text-apple-gray block mb-1.5">Environment</label>
            <div className="flex gap-2">
              {envs.map((e) => (
                <button
                  key={e}
                  type="button"
                  onClick={() => setEnv(e)}
                  className={`flex-1 py-1.5 text-xs font-medium rounded-lg transition-colors ${
                    env === e ? 'bg-apple-blue text-white' : 'glass-subtle hover:bg-white/60'
                  }`}
                >
                  {e}
                </button>
              ))}
            </div>
          </div>

          {/* Dynamic fields from schema */}
          {meta.fields.map((field) => (
            <div key={field.key}>
              <label className="text-xs font-medium text-apple-gray block mb-1.5">
                {field.label}
                {!field.required && <span className="text-apple-gray/50"> (optional)</span>}
              </label>
              {field.type === 'secret' ? (
                <div className="relative">
                  <input
                    type={showSecrets[field.key] ? 'text' : 'password'}
                    value={fieldValues[field.key] ?? ''}
                    onChange={(e) => setField(field.key, e.target.value)}
                    required={field.required}
                    placeholder={field.hint}
                    className="w-full glass-subtle px-3 py-2 text-sm rounded-xl outline-none pr-16 font-mono placeholder:font-sans placeholder:text-apple-gray/50"
                  />
                  <button
                    type="button"
                    onClick={() => setShowSecrets((prev) => ({ ...prev, [field.key]: !prev[field.key] }))}
                    className="absolute right-2 top-1/2 -translate-y-1/2 text-xs text-apple-blue font-medium"
                  >
                    {showSecrets[field.key] ? 'Hide' : 'Show'}
                  </button>
                </div>
              ) : (
                <input
                  type={field.type === 'url' ? 'url' : 'text'}
                  value={fieldValues[field.key] ?? ''}
                  onChange={(e) => setField(field.key, e.target.value)}
                  required={field.required}
                  placeholder={field.hint || (field.type === 'url' ? 'https://...' : '')}
                  className="w-full glass-subtle px-3 py-2 text-sm rounded-xl outline-none font-mono placeholder:font-sans placeholder:text-apple-gray/50"
                />
              )}
            </div>
          ))}
        </div>

        <button
          type="submit"
          disabled={!canSubmit}
          className="w-full mt-6 py-2.5 bg-apple-blue text-white text-sm font-semibold rounded-xl hover:bg-apple-blue/90 transition-colors disabled:opacity-40 disabled:cursor-not-allowed"
        >
          Add to Vault
        </button>
      </form>
    </dialog>
  )
}
