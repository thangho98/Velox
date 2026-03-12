import { create } from 'zustand'
import { persist, createJSONStorage } from 'zustand/middleware'
import type { UserInfo } from '@/types/api'
import { setTokenCallbacks, setSessionExpiredCallback } from '@/lib/fetch'

interface AuthState {
  // Tokens
  accessToken: string | null
  refreshToken: string | null
  tokenExpiresAt: number | null // timestamp

  // User
  user: UserInfo | null
  isAuthenticated: boolean

  // Actions
  setTokens: (accessToken: string, refreshToken: string, expiresIn: number) => void
  setUser: (user: UserInfo | null) => void
  logout: () => void

  // Computed
  isTokenExpired: () => boolean
}

export const useAuthStore = create<AuthState>()(
  persist(
    (set, get) => ({
      accessToken: null,
      refreshToken: null,
      tokenExpiresAt: null,
      user: null,
      isAuthenticated: false,

      setTokens: (accessToken, refreshToken, expiresIn) => {
        const expiresAt = Date.now() + expiresIn * 1000
        set({
          accessToken,
          refreshToken,
          tokenExpiresAt: expiresAt,
          isAuthenticated: true,
        })
      },

      setUser: (user) => {
        set({ user })
      },

      logout: () => {
        set({
          accessToken: null,
          refreshToken: null,
          tokenExpiresAt: null,
          user: null,
          isAuthenticated: false,
        })
      },

      isTokenExpired: () => {
        const { tokenExpiresAt } = get()
        if (!tokenExpiresAt) return true
        // Consider token expired 60 seconds before actual expiry
        return Date.now() >= tokenExpiresAt - 60000
      },
    }),
    {
      name: 'velox-auth',
      storage: createJSONStorage(() => localStorage),
      partialize: (state) => ({
        accessToken: state.accessToken,
        refreshToken: state.refreshToken,
        tokenExpiresAt: state.tokenExpiresAt,
        user: state.user,
        isAuthenticated: state.isAuthenticated,
      }),
    },
  ),
)

// Register token callbacks for API client
setTokenCallbacks(
  () => useAuthStore.getState().accessToken,
  () => useAuthStore.getState().refreshToken,
  (accessToken, refreshToken, expiresIn) => {
    useAuthStore.getState().setTokens(accessToken, refreshToken, expiresIn)
  },
)

// Clear auth state when session expires (fetch layer calls this on unrecoverable 401)
setSessionExpiredCallback(() => {
  useAuthStore.getState().logout()
})
