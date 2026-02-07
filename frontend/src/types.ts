export type View = 'dashboard' | 'detail' | 'settings'

export type CredentialStatus = 'active' | 'warning' | 'expired'
export type Environment = 'production' | 'staging' | 'development'

export interface RotationRecord {
  id: string
  timestamp: number
  reason: string
  automatic: boolean
}

export interface FieldSchema {
  key: string
  label: string
  type: 'secret' | 'text' | 'url'
  required: boolean
  hint: string
}

export interface ServiceMeta {
  name: string
  color: string
  icon: string
  category: string
  fields: FieldSchema[]
}

export interface Credential {
  id: string
  name: string
  service: string
  customService?: string
  environment: Environment
  fields: Record<string, string>
  status: CredentialStatus
  createdAt: number
  lastRotated: number
  rotations: RotationRecord[]
}

export interface Settings {
  autoLockMinutes: number
  showNotifications: boolean
  compactView: boolean
}
