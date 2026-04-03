# Phase 07: Migrate Webapp Imports to @velox/shared
Status: ⬜ Pending
Dependencies: Phase 02-06

## Objective
Webapp imports tất cả shared code từ `@velox/shared`. Xóa duplicate files.

## Implementation Steps

### 1. Update webapp hook barrels → re-export từ shared
- [ ] `webapp/src/hooks/stores/useMedia.ts`:
  ```typescript
  export * from '@velox/shared/hooks/media'
  ```
- [ ] `webapp/src/hooks/stores/useSettings.ts`:
  ```typescript
  export * from '@velox/shared/hooks/settings'
  ```

### 2. Update useAdmin + useUsers → re-export
- [ ] `webapp/src/hooks/stores/useAdmin.ts`:
  ```typescript
  export * from '@velox/shared/hooks/useAdmin'
  ```
- [ ] `webapp/src/hooks/stores/useUsers.ts`:
  ```typescript
  export * from '@velox/shared/hooks/useUsers'
  ```

### 3. Update useAuth.ts — keep web-only, re-export shared
- [ ] ```typescript
  // Re-export shared auth hooks
  export * from '@velox/shared/hooks/useAuth'

  // Web-only hooks that use react-router
  import { useEffect } from 'react'
  import { useNavigate } from 'react-router'
  import { useAuthStore } from '@/stores/auth'
  import { useRefreshToken } from '@velox/shared/hooks/useAuth'

  // Token Refresh Logic Hook (web-specific: uses setInterval)
  export function useTokenRefresh() {
    const { isTokenExpired, refreshToken } = useAuthStore()
    const { mutate: refresh } = useRefreshToken()

    useEffect(() => {
      if (!refreshToken) return
      const interval = setInterval(() => {
        if (isTokenExpired()) refresh()
      }, 60000)
      return () => clearInterval(interval)
    }, [refreshToken, isTokenExpired, refresh])
  }

  // Auth Guard Hook (web-specific: uses react-router navigate)
  export function useRequireAuth() {
    const { isAuthenticated } = useAuthStore()
    const navigate = useNavigate()

    useEffect(() => {
      if (!isAuthenticated) navigate('/login')
    }, [isAuthenticated, navigate])

    return isAuthenticated
  }
  ```

### 4. Xóa duplicate source files
- [ ] Xóa 9 media hook files: `webapp/src/hooks/stores/media/useLibrary.ts`, `useMediaQuery.ts`, `usePlayback.ts`, `useProgress.ts`, `useContinueWatching.ts`, `useSeries.ts`, `useSubtitleOps.ts`, `useMetadataOps.ts`, `useGenres.ts`
- [ ] Xóa media barrel: `webapp/src/hooks/stores/media/index.ts` (nếu có — `useMedia.ts` giữ lại là barrel)
- [ ] Xóa 5 settings hook files: `webapp/src/hooks/stores/settings/factory.ts`, `useMetadataSettings.ts`, `useSubtitleSettings.ts`, `usePlaybackSettings.ts`, `usePretranscodeSettings.ts`
- [ ] Xóa settings barrel: `webapp/src/hooks/stores/settings/index.ts` (nếu có)

### 5. Verify tất cả imports resolve
- [ ] `cd webapp && pnpm build && pnpm lint`
- [ ] Check: no broken imports, no duplicate exports

## Files to Delete
- `webapp/src/hooks/stores/media/useLibrary.ts`
- `webapp/src/hooks/stores/media/useMediaQuery.ts`
- `webapp/src/hooks/stores/media/usePlayback.ts`
- `webapp/src/hooks/stores/media/useProgress.ts`
- `webapp/src/hooks/stores/media/useContinueWatching.ts`
- `webapp/src/hooks/stores/media/useSeries.ts`
- `webapp/src/hooks/stores/media/useSubtitleOps.ts`
- `webapp/src/hooks/stores/media/useMetadataOps.ts`
- `webapp/src/hooks/stores/media/useGenres.ts`
- `webapp/src/hooks/stores/settings/factory.ts`
- `webapp/src/hooks/stores/settings/useMetadataSettings.ts`
- `webapp/src/hooks/stores/settings/useSubtitleSettings.ts`
- `webapp/src/hooks/stores/settings/usePlaybackSettings.ts`
- `webapp/src/hooks/stores/settings/usePretranscodeSettings.ts`

## Files to Modify
- `webapp/src/hooks/stores/useMedia.ts` — re-export from shared
- `webapp/src/hooks/stores/useSettings.ts` — re-export from shared
- `webapp/src/hooks/stores/useAdmin.ts` — re-export from shared
- `webapp/src/hooks/stores/useUsers.ts` — re-export from shared
- `webapp/src/hooks/stores/useAuth.ts` — re-export shared + keep web-only hooks

## Test Criteria
- [ ] `pnpm build` — zero errors
- [ ] `pnpm lint` — zero errors
- [ ] Webapp imports dạng `import { useLibraries } from '@/hooks/stores/useMedia'` vẫn hoạt động
- [ ] No circular dependency warnings

---
Next Phase: [phase-08-verify-webapp.md](phase-08-verify-webapp.md)
