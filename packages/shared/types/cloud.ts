// Cloud storage types — mirrors backend/internal/handler/storage_provider.go

export type ProviderHealth =
  | 'healthy'
  | 'expiring_soon'
  | 'expired'
  | 'error'
  | 'unknown'

export interface CloudDriverMeta {
  type: string
  display_name: string
  auth_flow: 'password' | 'oauth'
}

export interface StorageProvider {
  id: number
  provider_type: string
  display_name: string
  account_email: string
  account_type?: string
  quota_bytes?: number
  used_bytes?: number
  account_expires_at?: string
  token_expires_at?: string
  last_refresh_at?: string
  last_error?: string
  health: ProviderHealth
  created_at: string
}

export interface CreateProviderPasswordRequest {
  provider_type: string
  display_name: string
  auth: {
    flow: 'password'
    email: string
    password: string
  }
}

export interface ValidateFolderURLResponse {
  folder_id: string
  item_count: number
  sample_names: string[]
}
