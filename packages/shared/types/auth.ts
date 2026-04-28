// Auth, User, Session types

export interface LoginRequest {
  username: string
  password: string
}

export interface LoginResponse {
  access_token: string
  refresh_token: string
  expires_in: number
  user: UserInfo
}

export interface RefreshRequest {
  refresh_token: string
}

export interface RefreshResponse {
  access_token: string
  refresh_token: string
  expires_in: number
}

export interface ChangePasswordRequest {
  old_password: string
  new_password: string
}

export interface UserInfo {
  id: number
  username: string
  display_name: string
  is_admin: boolean
}

export interface User {
  id: number
  username: string
  display_name: string
  is_admin: boolean
  avatar_path: string
  created_at: string
  updated_at: string
}

export interface UserPreferences {
  user_id: number
  subtitle_language: string
  audio_language: string
  max_streaming_quality: string
  theme: string
  language: string
  last_live_channel_id: number | null
}

export interface UpdateProfileRequest {
  display_name: string
}

export interface Session {
  id: number
  user_id: number
  refresh_token_id?: number
  device_name: string
  ip_address: string
  user_agent: string
  expires_at: string
  last_active_at: string
  created_at: string
}
