import { useState } from 'react'
import { Shield, Plus, Settings as SettingsIcon, Lock } from 'lucide-react'
import { useStore } from './store'
import Dashboard from './Dashboard'
import Detail from './Detail'
import SettingsView from './Settings'
import Modal from './Modal'

function TrustIndicator({ open, onToggle }: { open: boolean; onToggle: () => void }) {
  return (
    <div className="relative">
      <button onClick={onToggle} className="p-2 rounded-xl hover:bg-white/40 transition-colors" title="Encryption status">
        <Lock className="w-5 h-5 text-apple-green" />
      </button>
      {open && (
        <div className="glass absolute right-0 top-12 w-72 p-4 z-50 animate-scale-in">
          <h3 className="font-semibold text-sm mb-3">Encryption Status</h3>
          <div className="space-y-2 text-sm">
            <div className="flex items-center gap-2">
              <div className="w-2 h-2 rounded-full bg-apple-green" />
              <span>SQLCipher — AES-256 at rest</span>
            </div>
            <div className="flex items-center gap-2">
              <div className="w-2 h-2 rounded-full bg-apple-green" />
              <span>AES-256-GCM per-field encryption</span>
            </div>
            <div className="flex items-center gap-2">
              <div className="w-2 h-2 rounded-full bg-apple-green" />
              <span>Keys never leave local vault</span>
            </div>
            <div className="flex items-center gap-2">
              <div className="w-2 h-2 rounded-full bg-apple-blue" />
              <span>MCP server — read-only access</span>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}

export default function App() {
  const { view, navigate, toggleAddModal, showAddModal } = useStore()
  const [trustOpen, setTrustOpen] = useState(false)

  return (
    <div className="max-w-2xl mx-auto px-4 py-6 min-h-screen" onClick={() => trustOpen && setTrustOpen(false)}>
      <header className="flex items-center justify-between mb-8">
        <button onClick={() => navigate('dashboard')} className="flex items-center gap-2.5">
          <div className="w-9 h-9 rounded-xl bg-apple-blue flex items-center justify-center">
            <Shield className="w-5 h-5 text-white" />
          </div>
          <span className="text-lg font-bold tracking-tight">API Vault</span>
        </button>

        <div className="flex items-center gap-1" onClick={(e) => e.stopPropagation()}>
          <button onClick={toggleAddModal} className="p-2 rounded-xl hover:bg-white/40 transition-colors" title="Add credential">
            <Plus className="w-5 h-5" />
          </button>
          <button onClick={() => navigate('settings')} className="p-2 rounded-xl hover:bg-white/40 transition-colors" title="Settings">
            <SettingsIcon className="w-5 h-5" />
          </button>
          <TrustIndicator open={trustOpen} onToggle={() => setTrustOpen(!trustOpen)} />
        </div>
      </header>

      {view === 'dashboard' && <Dashboard />}
      {view === 'detail' && <Detail />}
      {view === 'settings' && <SettingsView />}

      {showAddModal && <Modal />}
    </div>
  )
}
