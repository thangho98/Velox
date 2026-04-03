# Phase 06: Extract Query Hooks → @velox/shared
Status: ⬜ Pending
Dependencies: Phase 02, 03

## Objective
Move tất cả React Query hooks (API data fetching) sang shared. Tách web-specific auth hooks.

## Context — Dependency Analysis
| Hook file | Imports | Shareable? |
|-----------|---------|-----------|
| `media/useLibrary.ts` (66 LOC) | `api`, types | ✅ 100% |
| `media/useMediaQuery.ts` (59 LOC) | `api`, types | ✅ 100% |
| `media/usePlayback.ts` (100 LOC) | `api`, types | ✅ 100% |
| `media/useProgress.ts` (102 LOC) | `api`, types | ✅ 100% |
| `media/useContinueWatching.ts` (71 LOC) | `api`, types, `useProgress` keys | ✅ 100% |
| `media/useSeries.ts` (93 LOC) | `api`, types | ✅ 100% |
| `media/useSubtitleOps.ts` (44 LOC) | `api`, types | ✅ 100% |
| `media/useMetadataOps.ts` (101 LOC) | `api`, types | ✅ 100% |
| `media/useGenres.ts` (106 LOC) | `api`, types | ✅ 100% |
| `settings/factory.ts` (39 LOC) | `api` | ✅ 100% |
| `settings/useMetadataSettings.ts` (72 LOC) | factory | ✅ 100% |
| `settings/useSubtitleSettings.ts` (29 LOC) | factory | ✅ 100% |
| `settings/usePlaybackSettings.ts` (57 LOC) | factory | ✅ 100% |
| `settings/usePretranscodeSettings.ts` (158 LOC) | factory, `api` | ✅ 100% |
| `useAdmin.ts` (162 LOC) | `api`, types | ✅ 100% |
| `useUsers.ts` (88 LOC) | `api`, types | ✅ 100% |
| `useAuth.ts` — API + query hooks | `api`, types, `useAuthStore` | ✅ 90% |
| `useAuth.ts` → `useRequireAuth` (~12 LOC) | `useNavigate` (react-router) | ❌ Web-only |
| `useAuth.ts` → `useTokenRefresh` (~15 LOC) | `useEffect` + `setInterval` | ⚠️ Tách riêng |

## Implementation Steps

### 1. Copy media hooks (9 files)
- [ ] Copy tất cả files từ `webapp/src/hooks/stores/media/` → `packages/shared/hooks/media/`
- [ ] Fix imports trong từng file:
  ```
  BEFORE: import { api } from '@/lib/fetch'
  AFTER:  import { api } from '../../api'

  BEFORE: import type { Media, ... } from '@/types/api'
  AFTER:  import type { Media, ... } from '../../types'
  ```
- [ ] Tạo barrel `packages/shared/hooks/media/index.ts`:
  ```typescript
  export * from './useLibrary'
  export * from './useMediaQuery'
  export * from './usePlayback'
  export * from './useProgress'
  export * from './useContinueWatching'
  export * from './useSeries'
  export * from './useSubtitleOps'
  export * from './useMetadataOps'
  export * from './useGenres'
  ```

### 2. Copy settings hooks (5 files)
- [ ] Copy `webapp/src/hooks/stores/settings/` → `packages/shared/hooks/settings/`
- [ ] Fix imports (same pattern)
- [ ] Tạo barrel `packages/shared/hooks/settings/index.ts`:
  ```typescript
  export * from './factory'
  export * from './useMetadataSettings'
  export * from './useSubtitleSettings'
  export * from './usePlaybackSettings'
  export * from './usePretranscodeSettings'
  ```

### 3. Copy admin + users hooks
- [ ] Copy `webapp/src/hooks/stores/useAdmin.ts` → `packages/shared/hooks/useAdmin.ts`
- [ ] Copy `webapp/src/hooks/stores/useUsers.ts` → `packages/shared/hooks/useUsers.ts`
- [ ] Fix imports

### 4. Tách useAuth.ts → shared + web-only
- [ ] Tạo `packages/shared/hooks/useAuth.ts` — CHỈ chứa:
  - `authApi` object (login, refresh, logout, me, changePassword, listSessions, revokeSession)
  - `authKeys` query key factory
  - `profileApi` object (getProfile, updateProfile, getPreferences, updatePreferences)
  - `profileKeys` query key factory
  - `useMe`, `useLogin`, `useLogout`, `useRefreshToken`, `useChangePassword`
  - `useSessions`, `useRevokeSession`
  - `useProfile`, `useUpdateProfile`, `usePreferences`, `useUpdatePreferences`
  - `useSetupStatus`, `useSetup`, `useWizardStatus`, `useCompleteWizard`

  ⚠️ **KHÔNG copy:** `useTokenRefresh`, `useRequireAuth` — dùng `useNavigate` (react-router)

  ⚠️ **useLogin, useLogout cần authStore** — dùng factory pattern:
  ```typescript
  import type { AuthState } from '../stores'
  
  // Module-level store reference — set by each platform
  let _getAuthState: (() => AuthState) | null = null
  
  export function setAuthStateGetter(getter: () => AuthState) {
    _getAuthState = getter
  }
  
  function getAuthState(): AuthState {
    if (!_getAuthState) throw new Error('Auth state getter not set')
    return _getAuthState()
  }
  
  export function useLogin() {
    const queryClient = useQueryClient()
    // Use getter instead of direct store import
    return useMutation({
      mutationFn: authApi.login,
      onSuccess: (data) => {
        getAuthState().setTokens(data.access_token, data.refresh_token, data.expires_in)
        getAuthState().setUser(data.user)
        queryClient.setQueryData(authKeys.me(), data.user)
      },
    })
  }
  ```

  Webapp init: `setAuthStateGetter(() => useAuthStore.getState())`

### 5. Tạo main barrel
- [ ] `packages/shared/hooks/index.ts`:
  ```typescript
  export * from './media'
  export * from './settings'
  export * from './useAuth'
  export * from './useAdmin'
  export * from './useUsers'
  ```

### 6. Verify shared package compiles
- [ ] `cd packages/shared && npx tsc --noEmit`

### 7. Init auth state getter in webapp
- [ ] In `webapp/src/main.tsx` (after `initPlatform`):
  ```typescript
  import { setAuthStateGetter } from '@velox/shared/hooks/useAuth'
  import { useAuthStore } from './stores/auth'
  setAuthStateGetter(() => useAuthStore.getState())
  ```

## Files to Create
- `packages/shared/hooks/media/` — 9 hook files + `index.ts`
- `packages/shared/hooks/settings/` — 5 hook files + `index.ts`
- `packages/shared/hooks/useAuth.ts` — shared auth hooks (without web-only)
- `packages/shared/hooks/useAdmin.ts`
- `packages/shared/hooks/useUsers.ts`
- `packages/shared/hooks/index.ts` — main barrel

## Files to Modify
- `webapp/src/main.tsx` — add `setAuthStateGetter()`

## Test Criteria
- [ ] `cd packages/shared && npx tsc --noEmit` — no type errors
- [ ] All exports resolve
- [ ] No `@/` imports in shared (all relative)
- [ ] No `localStorage`, `document`, `window`, `navigator` refs in shared
- [ ] No `react-router` imports in shared

---
Next Phase: [phase-07-migrate-webapp.md](phase-07-migrate-webapp.md)
