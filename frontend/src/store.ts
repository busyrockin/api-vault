import { create } from 'zustand'
import type { Credential, View, Settings } from './types'
import { MOCK_CREDENTIALS } from './data'

interface VaultStore {
  credentials: Credential[]
  selectedId: string | null
  view: View
  search: string
  showAddModal: boolean
  settings: Settings

  navigate: (view: View, selectedId?: string | null) => void
  setSearch: (q: string) => void
  toggleAddModal: () => void
  addCredential: (cred: Credential) => void
  updateSettings: (patch: Partial<Settings>) => void
  filtered: () => Credential[]
}

export const useStore = create<VaultStore>((set, get) => ({
  credentials: MOCK_CREDENTIALS,
  selectedId: null,
  view: 'dashboard',
  search: '',
  showAddModal: false,
  settings: { autoLockMinutes: 5, showNotifications: true, compactView: false },

  navigate: (view, selectedId = null) => set({ view, selectedId }),
  setSearch: (search) => set({ search }),
  toggleAddModal: () => set((s) => ({ showAddModal: !s.showAddModal })),
  addCredential: (cred) => set((s) => ({ credentials: [cred, ...s.credentials], showAddModal: false })),
  updateSettings: (patch) => set((s) => ({ settings: { ...s.settings, ...patch } })),
  filtered: () => {
    const { credentials, search } = get()
    if (!search) return credentials
    const q = search.toLowerCase()
    return credentials.filter((c) => c.name.includes(q) || c.service.includes(q))
  },
}))
