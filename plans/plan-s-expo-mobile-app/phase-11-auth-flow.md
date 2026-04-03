# Phase 11: Auth Flow (Login + Token Refresh + Guard)
Status: ⬜ Pending
Dependencies: Phase 10

## Objective
Login screen + server URL config + auth guard (with hydration gate) + token refresh.

## Implementation Steps

### 1. Server Config Screen
- [ ] `mobile/app/server-config.tsx`:
  - Text input cho server URL (e.g. `http://192.168.1.100:8098`)
  - "Test Connection" button → `GET {url}/api/setup/status`
    - Success → show server name/version
    - Fail → show error (network error, timeout, invalid URL)
  - "Continue" button (disabled until test passes)
  - Save URL to MMKV → navigate to login
  - Layout: centered card, Velox logo/text at top

### 2. Login Screen
- [ ] `mobile/app/login.tsx`:
  - Server URL display (with "Change" link → back to server config)
  - Username + password text inputs
  - "Log In" button → `useLogin()` from `@velox/shared/hooks/useAuth`
  - Loading state on button
  - Error message display (wrong password, server error)
  - Navigate to `/(tabs)` on success
  - Keyboard avoiding view

### 3. Auth Guard with Hydration Gate
- [ ] Update `mobile/app/_layout.tsx`:

  ⚠️ **Auth store dùng SecureStore (async).** Zustand persist hydrate async → `isAuthenticated` sẽ là `false` trong vài ms đầu tiên khi cold start, DÙ user đã login. Nếu check `isAuthenticated` ngay lập tức → redirect nhầm về `/login`.

  **Giải pháp:** Thêm `hasHydrated` state, show splash/loading cho tới khi store đã hydrate xong.

  ```typescript
  import { useAuthStore } from '../src/stores/auth'
  import { hasServerUrl } from '../src/platform/storage'
  import { useRouter, useSegments } from 'expo-router'
  import { useEffect, useState } from 'react'

  function useProtectedRoute() {
    const { isAuthenticated } = useAuthStore()
    const segments = useSegments()
    const router = useRouter()

    // Wait for Zustand persist to hydrate from SecureStore
    const [hasHydrated, setHasHydrated] = useState(false)
    useEffect(() => {
      // Zustand persist v5: onFinishHydration callback
      const unsub = useAuthStore.persist.onFinishHydration(() => {
        setHasHydrated(true)
      })
      // If already hydrated (e.g. hot reload), set immediately
      if (useAuthStore.persist.hasHydrated()) {
        setHasHydrated(true)
      }
      return unsub
    }, [])

    useEffect(() => {
      // Don't redirect until hydration complete — prevents flash to login
      if (!hasHydrated) return

      const inAuthGroup = segments[0] === '(tabs)'

      if (!hasServerUrl()) {
        router.replace('/server-config')
      } else if (!isAuthenticated && inAuthGroup) {
        router.replace('/login')
      } else if (isAuthenticated && !inAuthGroup) {
        router.replace('/(tabs)')
      }
    }, [isAuthenticated, segments, hasHydrated])

    return hasHydrated
  }

  export default function RootLayout() {
    const hasHydrated = useProtectedRoute()

    if (!hasHydrated) {
      // Show splash screen until SecureStore hydration completes
      return <SplashScreen />  // Or <View className="flex-1 bg-black" />
    }

    return (
      <QueryClientProvider client={queryClient}>
        <Stack screenOptions={{ headerShown: false }} />
      </QueryClientProvider>
    )
  }
  ```

  Route structure:
  ```
  mobile/app/
  ├── _layout.tsx         # Root: providers + auth guard + hydration gate
  ├── server-config.tsx   # Step 1: Enter server URL
  ├── login.tsx           # Step 2: Login
  └── (tabs)/             # Protected routes
      ├── _layout.tsx     # Tab bar
      ├── index.tsx       # Home (placeholder)
      └── profile.tsx     # Profile (placeholder)
  ```

### 4. Token Refresh (mobile-specific)
- [ ] Tạo `mobile/src/hooks/useTokenRefresh.ts`:
  ```typescript
  import { useEffect } from 'react'
  import { AppState, type AppStateStatus } from 'react-native'
  import { useAuthStore } from '../stores/auth'

  export function useTokenRefresh() {
    const { isTokenExpired, refreshToken } = useAuthStore()

    useEffect(() => {
      if (!refreshToken) return

      // Check on interval (same as web — every 60s)
      const interval = setInterval(() => {
        if (isTokenExpired()) {
          // api client handles refresh automatically on 401
          // this is a proactive refresh before expiry
        }
      }, 60000)

      // Also refresh when app comes to foreground
      const handleAppState = (state: AppStateStatus) => {
        if (state === 'active' && isTokenExpired()) {
          // Proactive refresh
        }
      }
      const sub = AppState.addEventListener('change', handleAppState)

      return () => {
        clearInterval(interval)
        sub.remove()
      }
    }, [refreshToken, isTokenExpired])
  }
  ```
- [ ] Call `useTokenRefresh()` in root layout (after hydration)

## Files to Create
- `mobile/app/server-config.tsx`
- `mobile/app/login.tsx`
- `mobile/src/hooks/useTokenRefresh.ts`

## Files to Modify
- `mobile/app/_layout.tsx` — add auth guard (with hydration gate) + token refresh

## Test Criteria
- [ ] Server URL input → test connection → success/fail feedback
- [ ] Login → token stored in **SecureStore** (encrypted) → navigate to tabs
- [ ] Close app → reopen → still authenticated (**SecureStore** persists across restart)
- [ ] **Cold start: NO flash to login screen** — splash shown until hydration completes
- [ ] Auth guard redirects to login when not authenticated (after hydration)
- [ ] Auth guard redirects to server-config when no URL set
- [ ] Token refresh works when app comes to foreground
- [ ] Logout → clears auth (SecureStore) → redirects to login

---
Next Phase: [phase-12-home-screen.md](phase-12-home-screen.md)
