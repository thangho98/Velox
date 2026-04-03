# Phase 03: Extract Platform-Agnostic API Client
Status: ⬜ Pending
Dependencies: Phase 01, 02

## Objective
Tạo shared API client không phụ thuộc browser. Webapp wraps với web adapter.

## Context — Vấn đề cần giải quyết
File `webapp/src/lib/fetch.ts` (222 LOC) có 2 chỗ phụ thuộc browser:
1. **Line 96-98:** `localStorage.getItem('velox_device_name')` — cần adapter
2. **Line 4:** `API_BASE_URL = '/api'` — web dùng relative path (Vite proxy), mobile cần absolute URL

Logic còn lại (token refresh, queue, error handling) 100% platform-agnostic.

## Implementation Steps

### 1. Tạo `packages/shared/api/client.ts`
- [ ] Copy logic từ `webapp/src/lib/fetch.ts` nhưng thay thế browser-specific bằng `getPlatform()`:
  ```typescript
  import type { ApiErrorResponse, ApiResponse } from '../types'
  import { getPlatform } from '../platform'

  // Token refresh state — module-level singleton
  let isRefreshing = false
  let refreshSubscribers: Array<{
    resolve: (token: string) => void
    reject: (error: Error) => void
  }> = []

  export class ApiError extends Error {
    status: number
    constructor(message: string, status: number) {
      super(message)
      this.name = 'ApiError'
      this.status = status
    }
  }

  // Token callbacks — injected by auth store
  let getAccessToken: () => string | null = () => null
  let getRefreshToken: () => string | null = () => null
  let setTokensCallback: (
    accessToken: string,
    refreshToken: string,
    expiresIn: number,
  ) => void = () => {}
  let onSessionExpiredCallback: () => void = () => {}

  export function setTokenCallbacks(
    getAccess: () => string | null,
    getRefresh: () => string | null,
    setTokens: (accessToken: string, refreshToken: string, expiresIn: number) => void,
  ) {
    getAccessToken = getAccess
    getRefreshToken = getRefresh
    setTokensCallback = setTokens
  }

  export function setSessionExpiredCallback(callback: () => void) {
    onSessionExpiredCallback = callback
  }
  ```
  
  **Key changes trong `fetchWithAuth`:**
  ```typescript
  async function fetchWithAuth<T>(endpoint: string, options: RequestInit = {}): Promise<T> {
    const platform = getPlatform()
    const url = `${platform.getApiBaseUrl()}${endpoint}`  // ← thay API_BASE_URL
    // ...
    const deviceName = platform.getDeviceName()           // ← thay localStorage
    if (deviceName) {
      headers['X-Device-Name'] = deviceName
    }
    // ... rest giữ nguyên
  }
  ```
  
  **Key changes trong `refreshAccessToken`:**
  ```typescript
  async function refreshAccessToken(): Promise<string | null> {
    const refreshToken = getRefreshToken()
    if (!refreshToken) return null
    try {
      const response = await fetch(`${getPlatform().getApiBaseUrl()}/auth/refresh`, {
        // ← thay API_BASE_URL
  ```
  
  **Stream URL helpers:**
  ```typescript
  export function getDirectStreamUrl(mediaId: number, token?: string): string {
    const base = `${getPlatform().getApiBaseUrl()}/stream/${mediaId}`
    return token ? `${base}?token=${encodeURIComponent(token)}` : base
  }

  export function getHlsMasterUrl(mediaId: number, token?: string): string {
    const base = `${getPlatform().getApiBaseUrl()}/stream/${mediaId}/hls/master.m3u8`
    return token ? `${base}?token=${encodeURIComponent(token)}` : base
  }
  ```

### 2. Tạo `packages/shared/api/index.ts`
- [ ] ```typescript
  export {
    api,
    ApiError,
    setTokenCallbacks,
    setSessionExpiredCallback,
    getDirectStreamUrl,
    getHlsMasterUrl,
  } from './client'
  ```

### 3. Update `webapp/src/lib/fetch.ts` → thin re-export
- [ ] ```typescript
  // Re-export everything from shared
  export {
    api,
    ApiError,
    setTokenCallbacks,
    setSessionExpiredCallback,
    getDirectStreamUrl,
    getHlsMasterUrl,
  } from '@velox/shared/api'
  ```

### 4. Tạo web platform adapter
- [ ] Tạo `webapp/src/platform/web-adapter.ts`:
  ```typescript
  import type { PlatformAdapter } from '@velox/shared/platform'

  export const webPlatform: PlatformAdapter = {
    storage: {
      getItem: (key) => localStorage.getItem(key),
      setItem: (key, value) => localStorage.setItem(key, value),
      removeItem: (key) => localStorage.removeItem(key),
    },
    secureStorage: {
      // Web doesn't have secure storage — use localStorage
      getItem: (key) => localStorage.getItem(key),
      setItem: (key, value) => localStorage.setItem(key, value),
      removeItem: (key) => localStorage.removeItem(key),
    },
    getDeviceName: () => localStorage.getItem('velox_device_name') || '',
    getApiBaseUrl: () => '/api',  // Vite proxy handles this
  }
  ```

### 5. Init platform ở webapp entry point
- [ ] Thêm vào đầu `webapp/src/main.tsx` (TRƯỚC React render):
  ```typescript
  import { initPlatform } from '@velox/shared/platform'
  import { webPlatform } from './platform/web-adapter'
  initPlatform(webPlatform)
  ```
  ⚠️ Phải đặt TRƯỚC mọi import khác sử dụng `api` hoặc `getPlatform()`.

### 6. Verify
- [ ] `cd webapp && pnpm build && pnpm lint`
- [ ] Manual test: login → browse → play — tất cả API calls phải hoạt động

## Files to Create
- `packages/shared/api/client.ts` — platform-agnostic API client
- `packages/shared/api/index.ts` — barrel export
- `webapp/src/platform/web-adapter.ts` — web-specific PlatformAdapter

## Files to Modify
- `webapp/src/lib/fetch.ts` — replace with thin re-export
- `webapp/src/main.tsx` — add `initPlatform(webPlatform)` at top

## Test Criteria
- [ ] `packages/shared/api/client.ts` — no `localStorage`, no hardcoded `/api`
- [ ] Webapp build pass
- [ ] Manual test: login, browse media, play video — all API calls work

---
Next Phase: [phase-04-extract-stores.md](phase-04-extract-stores.md)
