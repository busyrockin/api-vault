import { ArrowLeft, Shield, Terminal, Globe, Bell, LayoutGrid } from 'lucide-react'
import { useStore } from './store'

export default function Settings() {
  const { navigate, settings, updateSettings } = useStore()

  const lockOptions = [1, 5, 15, 30] as const

  const plugins = [
    { name: 'MCP Server', icon: Terminal, status: 'active' as const, desc: 'AI agent credential access' },
    { name: 'Auto-Rotation', icon: Shield, status: 'beta' as const, desc: 'Automatic key rotation' },
    { name: 'Cloud Sync', icon: Globe, status: 'coming' as const, desc: 'Cross-device vault sync' },
  ]

  const pluginColors = {
    active: 'bg-apple-green/15 text-apple-green',
    beta: 'bg-apple-purple/15 text-apple-purple',
    coming: 'bg-apple-gray/15 text-apple-gray',
  }

  return (
    <div className="animate-fade-in">
      <button onClick={() => navigate('dashboard')} className="flex items-center gap-1.5 text-apple-blue text-sm mb-6 hover:underline">
        <ArrowLeft className="w-4 h-4" /> Back
      </button>

      <h2 className="text-xl font-bold mb-6">Settings</h2>

      {/* Auto-Lock */}
      <div className="glass p-5 mb-4">
        <h3 className="font-semibold text-sm mb-3">Auto-Lock Timer</h3>
        <div className="flex gap-2">
          {lockOptions.map((min) => (
            <button
              key={min}
              onClick={() => updateSettings({ autoLockMinutes: min })}
              className={`flex-1 py-2 text-sm font-medium rounded-xl transition-colors ${
                settings.autoLockMinutes === min ? 'bg-apple-blue text-white' : 'glass-subtle hover:bg-white/60'
              }`}
            >
              {min}m
            </button>
          ))}
        </div>
      </div>

      {/* Plugins */}
      <div className="glass p-5 mb-4">
        <h3 className="font-semibold text-sm mb-3">Plugins</h3>
        <div className="space-y-3">
          {plugins.map((p) => (
            <div key={p.name} className="flex items-center gap-3">
              <div className="w-8 h-8 rounded-lg bg-apple-blue/10 flex items-center justify-center">
                <p.icon className="w-4 h-4 text-apple-blue" />
              </div>
              <div className="flex-1">
                <div className="text-sm font-medium">{p.name}</div>
                <div className="text-xs text-apple-gray">{p.desc}</div>
              </div>
              <span className={`badge ${pluginColors[p.status]}`}>{p.status}</span>
            </div>
          ))}
        </div>
      </div>

      {/* Preferences */}
      <div className="glass p-5 mb-4">
        <h3 className="font-semibold text-sm mb-3">Preferences</h3>
        <div className="space-y-3">
          <label className="flex items-center justify-between cursor-pointer">
            <div className="flex items-center gap-2.5">
              <Bell className="w-4 h-4 text-apple-gray" />
              <span className="text-sm">Notifications</span>
            </div>
            <input
              type="checkbox"
              checked={settings.showNotifications}
              onChange={(e) => updateSettings({ showNotifications: e.target.checked })}
              className="w-5 h-5 accent-apple-blue"
            />
          </label>
          <label className="flex items-center justify-between cursor-pointer">
            <div className="flex items-center gap-2.5">
              <LayoutGrid className="w-4 h-4 text-apple-gray" />
              <span className="text-sm">Compact view</span>
            </div>
            <input
              type="checkbox"
              checked={settings.compactView}
              onChange={(e) => updateSettings({ compactView: e.target.checked })}
              className="w-5 h-5 accent-apple-blue"
            />
          </label>
        </div>
      </div>

      {/* About */}
      <div className="glass p-5">
        <h3 className="font-semibold text-sm mb-2">About</h3>
        <div className="text-sm text-apple-gray space-y-1">
          <p>API Vault v0.1.0 (prototype)</p>
          <p>Local-first credential management for AI agents</p>
          <p>SQLCipher + AES-256-GCM encryption</p>
        </div>
      </div>
    </div>
  )
}
