/**
 * @velox/shared/hooks — Auth hooks
 *
 * Platform-agnostic auth & profile React Query hooks.
 * The platform layer must call setAuthStateGetter(getAuthState) once at app startup
 * so these hooks can read/write auth state without coupling to a specific store.
 */

import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { api } from '../api'
import type {
  LoginRequest,
  LoginResponse,
  RefreshRequest,
  RefreshResponse,
  ChangePasswordRequest,
  UserInfo,
  Session,
  User,
  UpdateProfileRequest,
  UserPreferences,
} from '../types'

// ── Auth state getter injection ──────────────────────────────────────────────

type AuthState = {
  accessToken: string | null
  refreshToken: string | null
  tokenExpiresAt: number | null
  user: UserInfo | null
  isAuthenticated: boolean
  setTokens: (accessToken: string, refreshToken: string, expiresIn: number) => void
  setUser: (user: UserInfo | null) => void
  logout: () => void
  isTokenExpired: () => boolean
}

let getAuthState: () => AuthState = () => ({
  accessToken: null,
  refreshToken: null,
  tokenExpiresAt: null,
  user: null,
  isAuthenticated: false,
  setTokens: () => {},
  setUser: () => {},
  logout: () => {},
  isTokenExpired: () => true,
})

export function setAuthStateGetter(fn: () => AuthState) {
  getAuthState = fn
}

// ── API Functions ────────────────────────────────────────────────────────────

const authApi = {
  login: (data: LoginRequest) => api.post<LoginResponse>('/auth/login', data),
  refresh: (data: RefreshRequest) => api.post<RefreshResponse>('/auth/refresh', data),
  logout: (data: { refresh_token: string }) => api.post<void>('/auth/logout', data),
  me: () => api.get<UserInfo>('/auth/me'),
  changePassword: (data: ChangePasswordRequest) => api.post<void>('/auth/change-password', data),
  listSessions: () => api.get<Session[]>('/auth/sessions'),
  revokeSession: (sessionId: number) => api.delete(`/auth/sessions/${sessionId}`),
}

const profileApi = {
  getProfile: () => api.get<User>('/profile'),
  updateProfile: (data: UpdateProfileRequest) => api.patch<User>('/profile', data),
  getPreferences: () => api.get<UserPreferences>('/profile/preferences'),
  updatePreferences: (data: UserPreferences) =>
    api.put<UserPreferences>('/profile/preferences', data),
}

// ── Query Keys ───────────────────────────────────────────────────────────────

export const authKeys = {
  all: ['auth'] as const,
  me: () => [...authKeys.all, 'me'] as const,
  sessions: () => [...authKeys.all, 'sessions'] as const,
}

export const profileKeys = {
  all: ['profile'] as const,
  profile: () => [...profileKeys.all, 'profile'] as const,
  preferences: () => [...profileKeys.all, 'preferences'] as const,
}

// ── Auth Hooks ───────────────────────────────────────────────────────────────

export function useMe() {
  return useQuery({
    queryKey: authKeys.me(),
    queryFn: authApi.me,
    retry: false,
    staleTime: 5 * 60 * 1000,
  })
}

export function useLogin() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: authApi.login,
    onSuccess: (data) => {
      const { setTokens, setUser } = getAuthState()
      setTokens(data.access_token, data.refresh_token, data.expires_in)
      setUser(data.user)
      queryClient.setQueryData(authKeys.me(), data.user)
    },
  })
}

export function useLogout() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: () => {
      const { refreshToken } = getAuthState()
      if (!refreshToken) return Promise.resolve()
      return authApi.logout({ refresh_token: refreshToken })
    },
    onSuccess: () => {
      const { logout: clearAuth } = getAuthState()
      clearAuth()
      queryClient.clear()
    },
    onError: () => {
      const { logout: clearAuth } = getAuthState()
      clearAuth()
      queryClient.clear()
    },
  })
}

export function useRefreshToken() {
  return useMutation({
    mutationFn: async () => {
      const { refreshToken } = getAuthState()
      if (!refreshToken) throw new Error('No refresh token')
      return authApi.refresh({ refresh_token: refreshToken })
    },
    onSuccess: (data) => {
      const { setTokens } = getAuthState()
      setTokens(data.access_token, data.refresh_token, data.expires_in)
    },
  })
}

export function useChangePassword() {
  return useMutation({
    mutationFn: authApi.changePassword,
  })
}

export function useSessions() {
  return useQuery({
    queryKey: authKeys.sessions(),
    queryFn: authApi.listSessions,
    staleTime: 30 * 1000,
  })
}

export function useRevokeSession() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: authApi.revokeSession,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: authKeys.sessions() })
    },
  })
}

// ── Profile Hooks ────────────────────────────────────────────────────────────

export function useProfile() {
  return useQuery({
    queryKey: profileKeys.profile(),
    queryFn: profileApi.getProfile,
    staleTime: 5 * 60 * 1000,
  })
}

export function useUpdateProfile() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: profileApi.updateProfile,
    onSuccess: (data) => {
      queryClient.setQueryData(profileKeys.profile(), data)
    },
  })
}

export function usePreferences() {
  return useQuery({
    queryKey: profileKeys.preferences(),
    queryFn: profileApi.getPreferences,
    staleTime: 5 * 60 * 1000,
  })
}

export function useUpdatePreferences() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: profileApi.updatePreferences,
    onSuccess: (data) => {
      queryClient.setQueryData(profileKeys.preferences(), data)
    },
  })
}

// ── Setup / Wizard Hooks ─────────────────────────────────────────────────────

export function useSetupStatus() {
  return useQuery({
    queryKey: ['setup', 'status'],
    queryFn: () => api.get<{ configured: boolean }>('/setup/status'),
    retry: false,
  })
}

export function useSetup() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (data: { username: string; password: string; display_name: string }) =>
      api.post<void>('/setup', data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['setup', 'status'] })
    },
  })
}

export function useWizardStatus() {
  return useQuery({
    queryKey: ['setup', 'wizard'],
    queryFn: () => api.get<{ completed: boolean }>('/setup/wizard'),
    staleTime: 5 * 60 * 1000,
  })
}

export function useCompleteWizard() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: () => api.post<{ completed: boolean }>('/setup/wizard/complete', {}),
    onSuccess: () => {
      queryClient.setQueryData(['setup', 'wizard'], { completed: true })
    },
  })
}
