import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { useEffect } from 'react'
import { useNavigate } from 'react-router'
import { api } from '@/lib/fetch'
import { useAuthStore } from '@/stores/auth'
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
} from '@/types/api'

// Auth API Functions
const authApi = {
  login: (data: LoginRequest) => api.post<LoginResponse>('/auth/login', data),
  refresh: (data: RefreshRequest) => api.post<RefreshResponse>('/auth/refresh', data),
  logout: (data: { refresh_token: string }) => api.post<void>('/auth/logout', data),
  me: () => api.get<UserInfo>('/auth/me'),
  changePassword: (data: ChangePasswordRequest) => api.post<void>('/auth/change-password', data),
  listSessions: () => api.get<Session[]>('/auth/sessions'),
  revokeSession: (sessionId: number) => api.delete(`/auth/sessions/${sessionId}`),
}

// Query Keys
export const authKeys = {
  all: ['auth'] as const,
  me: () => [...authKeys.all, 'me'] as const,
  sessions: () => [...authKeys.all, 'sessions'] as const,
}

// React Query Hooks
export function useMe() {
  return useQuery({
    queryKey: authKeys.me(),
    queryFn: authApi.me,
    retry: false,
    staleTime: 5 * 60 * 1000, // 5 minutes
  })
}

export function useLogin() {
  const queryClient = useQueryClient()
  const { setTokens, setUser } = useAuthStore()

  return useMutation({
    mutationFn: authApi.login,
    onSuccess: (data) => {
      setTokens(data.access_token, data.refresh_token, data.expires_in)
      setUser(data.user)
      queryClient.setQueryData(authKeys.me(), data.user)
    },
  })
}

export function useLogout() {
  const queryClient = useQueryClient()
  const { refreshToken, logout: clearAuth } = useAuthStore()

  return useMutation({
    mutationFn: () => {
      if (!refreshToken) return Promise.resolve()
      return authApi.logout({ refresh_token: refreshToken })
    },
    onSuccess: () => {
      clearAuth()
      queryClient.clear()
    },
    onError: () => {
      // Still clear local auth even if server logout fails
      clearAuth()
      queryClient.clear()
    },
  })
}

export function useRefreshToken() {
  const { refreshToken, setTokens } = useAuthStore()

  return useMutation({
    mutationFn: async () => {
      if (!refreshToken) throw new Error('No refresh token')
      return authApi.refresh({ refresh_token: refreshToken })
    },
    onSuccess: (data) => {
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
    staleTime: 30 * 1000, // 30 seconds
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

// Profile API Functions
const profileApi = {
  getProfile: () => api.get<User>('/profile'),
  updateProfile: (data: UpdateProfileRequest) => api.patch<User>('/profile', data),
  getPreferences: () => api.get<UserPreferences>('/profile/preferences'),
  updatePreferences: (data: UserPreferences) =>
    api.put<UserPreferences>('/profile/preferences', data),
}

export const profileKeys = {
  all: ['profile'] as const,
  profile: () => [...profileKeys.all, 'profile'] as const,
  preferences: () => [...profileKeys.all, 'preferences'] as const,
}

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

// Token Refresh Logic Hook
export function useTokenRefresh() {
  const { isTokenExpired, refreshToken } = useAuthStore()
  const { mutate: refresh } = useRefreshToken()

  useEffect(() => {
    if (!refreshToken) return

    // Check every minute
    const interval = setInterval(() => {
      if (isTokenExpired()) {
        refresh()
      }
    }, 60000)

    return () => clearInterval(interval)
  }, [refreshToken, isTokenExpired, refresh])
}

// Auth Guard Hook
export function useRequireAuth() {
  const { isAuthenticated } = useAuthStore()
  const navigate = useNavigate()

  useEffect(() => {
    if (!isAuthenticated) {
      navigate('/login')
    }
  }, [isAuthenticated, navigate])

  return isAuthenticated
}

// Setup check
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
