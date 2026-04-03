/**
 * Auth hooks — web implementation
 *
 * Re-exports shared auth hooks and adds web-specific hooks
 * (useTokenRefresh uses setInterval, useRequireAuth uses useNavigate)
 */

// Re-export shared auth hooks
export {
  authKeys,
  profileKeys,
  useMe,
  useLogin,
  useLogout,
  useRefreshToken,
  useChangePassword,
  useSessions,
  useRevokeSession,
  useProfile,
  useUpdateProfile,
  usePreferences,
  useUpdatePreferences,
  useSetupStatus,
  useSetup,
  useWizardStatus,
  useCompleteWizard,
} from '@velox/shared/hooks/auth'

// Web-only imports
import { useEffect } from 'react'
import { useNavigate } from 'react-router'
import { useAuthStore } from '@/stores/auth'
import type {
  LoginRequest,
  LoginResponse,
  RefreshRequest,
  RefreshResponse,
  ChangePasswordRequest,
  UserInfo,
  Session,
} from '@/types/api'
import { api } from '@/lib/fetch'
import { useMutation } from '@tanstack/react-query'

// Local authApi for web-specific hooks
const authApi = {
  login: (data: LoginRequest) => api.post<LoginResponse>('/auth/login', data),
  refresh: (data: RefreshRequest) => api.post<RefreshResponse>('/auth/refresh', data),
  logout: (data: { refresh_token: string }) => api.post<void>('/auth/logout', data),
  me: () => api.get<UserInfo>('/auth/me'),
  changePassword: (data: ChangePasswordRequest) => api.post<void>('/auth/change-password', data),
  listSessions: () => api.get<Session[]>('/auth/sessions'),
  revokeSession: (sessionId: number) => api.delete(`/auth/sessions/${sessionId}`),
}

// Token Refresh Logic Hook (web-specific: uses setInterval)
export function useTokenRefresh() {
  const { refreshToken } = useAuthStore()

  return useMutation({
    mutationFn: async () => {
      if (!refreshToken) throw new Error('No refresh token')
      return authApi.refresh({ refresh_token: refreshToken })
    },
    onSuccess: (data) => {
      useAuthStore.getState().setTokens(data.access_token, data.refresh_token, data.expires_in)
    },
  })
}

// Auth Guard Hook (web-specific: uses react-router navigate)
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
