# Phase 04: Extract Zustand Store Factories
Status: ⬜ Pending
Dependencies: Phase 03

## Objective
Tạo store factories trong shared. Webapp instantiates với localStorage adapter.

## Context
- `webapp/src/stores/auth.ts` (91 LOC): Zustand + `persist(localStorage)` + token callbacks registration
- `webapp/src/stores/player.ts` (173 LOC): Zustand + `persist(localStorage)` — subtitle prefs, volume, quality
- `webapp/src/stores/ui.ts` (91 LOC): Web-specific (sidebar, search modal, toasts) — **KHÔNG shared**

## Implementation Steps

### 1. Tạo `packages/shared/stores/createAuthStore.ts`
- [ ] Extract state interface + logic, nhận `StateStorage` adapter:
  ```typescript
  import { create } from 'zustand'
  import { persist, createJSONStorage, type StateStorage } from 'zustand/middleware'
  import type { UserInfo } from '../types'
  import { setTokenCallbacks, setSessionExpiredCallback } from '../api'

  export interface AuthState {
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

  export function createAuthStore(storage: StateStorage) {
    const store = create<AuthState>()(
      persist(
        (set, get) => ({
          // ... exact same logic as webapp/src/stores/auth.ts lines 27-63
          accessToken: null,
          refreshToken: null,
          tokenExpiresAt: null,
          user: null,
          isAuthenticated: false,

          setTokens: (accessToken, refreshToken, expiresIn) => {
            const expiresAt = Date.now() + expiresIn * 1000
            set({ accessToken, refreshToken, tokenExpiresAt: expiresAt, isAuthenticated: true })
          },

          setUser: (user) => set({ user }),

          logout: () => set({
            accessToken: null, refreshToken: null, tokenExpiresAt: null,
            user: null, isAuthenticated: false,
          }),

          isTokenExpired: () => {
            const { tokenExpiresAt } = get()
            if (!tokenExpiresAt) return true
            return Date.now() >= tokenExpiresAt - 60000
          },
        }),
        {
          name: 'velox-auth',
          storage: createJSONStorage(() => storage),
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
      () => store.getState().accessToken,
      () => store.getState().refreshToken,
      (accessToken, refreshToken, expiresIn) => {
        store.getState().setTokens(accessToken, refreshToken, expiresIn)
      },
    )
    setSessionExpiredCallback(() => {
      store.getState().logout()
    })

    return store
  }
  ```

### 2. Tạo `packages/shared/stores/createPlayerStore.ts`
- [ ] Extract PlayerState interface + logic:
  ```typescript
  import { create } from 'zustand'
  import { persist, createJSONStorage, type StateStorage } from 'zustand/middleware'

  export interface PlayerState {
    // ... exact same interface as webapp/src/stores/player.ts lines 4-59
    volume: number
    isMuted: boolean
    playbackRate: number
    subtitleLanguage: string | null
    subtitleTrackId: number | null
    secondarySubtitleLanguage: string | null
    secondarySubtitleTrackId: number | null
    subtitleSize: 'small' | 'medium' | 'large'
    subtitleColor: string
    subtitleBackground: 'solid' | 'semi' | 'none'
    subtitleOffsets: Record<number, number>
    audioLanguage: string | null
    audioTrackId: number | null
    maxQuality: number | 'auto'
    aspectRatio: 'contain' | 'cover' | 'fill'
    repeatMode: 'none' | 'one' | 'all'
    lastPositions: Record<number, number>
    // + all action methods
    setVolume: (volume: number) => void
    toggleMute: () => void
    // ... etc — copy all methods from player.ts
  }

  export function createPlayerStore(storage: StateStorage) {
    return create<PlayerState>()(
      persist(
        (set, get) => ({
          // ... exact same logic as webapp/src/stores/player.ts lines 64-148
        }),
        {
          name: 'velox-player',
          storage: createJSONStorage(() => storage),
          partialize: (state) => ({
            // ... exact same as lines 153-170
          }),
        },
      ),
    )
  }
  ```

### 3. Tạo `packages/shared/stores/index.ts`
- [ ] ```typescript
  export { createAuthStore, type AuthState } from './createAuthStore'
  export { createPlayerStore, type PlayerState } from './createPlayerStore'
  ```

### 4. Update `webapp/src/stores/auth.ts` — use factory
- [ ] ```typescript
  import { createAuthStore } from '@velox/shared/stores'

  const localStorageAdapter = {
    getItem: (name: string) => localStorage.getItem(name),
    setItem: (name: string, value: string) => localStorage.setItem(name, value),
    removeItem: (name: string) => localStorage.removeItem(name),
  }

  export const useAuthStore = createAuthStore(localStorageAdapter)
  ```

### 5. Update `webapp/src/stores/player.ts` — use factory
- [ ] ```typescript
  import { createPlayerStore } from '@velox/shared/stores'

  const localStorageAdapter = {
    getItem: (name: string) => localStorage.getItem(name),
    setItem: (name: string, value: string) => localStorage.setItem(name, value),
    removeItem: (name: string) => localStorage.removeItem(name),
  }

  export const usePlayerStore = createPlayerStore(localStorageAdapter)
  ```

### 6. Verify
- [ ] `cd webapp && pnpm build && pnpm lint`
- [ ] Manual test: login → token persist → refresh page → still logged in
- [ ] Manual test: change volume/subtitle prefs → refresh → settings preserved

## Files to Create
- `packages/shared/stores/createAuthStore.ts`
- `packages/shared/stores/createPlayerStore.ts`
- `packages/shared/stores/index.ts` — barrel

## Files to Modify
- `webapp/src/stores/auth.ts` — replace with factory call (~5 lines)
- `webapp/src/stores/player.ts` — replace with factory call (~5 lines)

## Files NOT Modified
- `webapp/src/stores/ui.ts` — web-only, stays as-is
- `webapp/src/stores/index.ts` — still re-exports auth, player, ui

## Test Criteria
- [ ] Token persist + auto-login after refresh
- [ ] Player prefs persist (volume, subtitle language, quality)
- [ ] `stores/ui.ts` unchanged and working

---
Next Phase: [phase-05-extract-libs.md](phase-05-extract-libs.md)
