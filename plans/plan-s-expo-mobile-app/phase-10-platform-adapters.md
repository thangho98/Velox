# Phase 10: Platform Adapters (Storage, API, Capabilities)
Status: ⬜ Pending
Dependencies: Phase 09

## Objective
Implement mobile platform adapter. Auth tokens → SecureStore (encrypted). Prefs/cache → MMKV (fast).

## Context — Storage Strategy
| Data | Storage | Lý do |
|------|---------|-------|
| Access token | SecureStore | Encrypted, Android Keystore |
| Refresh token | SecureStore | Encrypted |
| Player prefs (volume, subtitle) | MMKV | Fast sync read, non-sensitive |
| Server URL | MMKV | Non-sensitive, need sync read |
| Device name | MMKV | Non-sensitive |

⚠️ Zustand persist cần **synchronous** `getItem/setItem` cho hydration. SecureStore là async → auth store cần `createJSONStorage` với async support (Zustand v5 supports this).

## Implementation Steps

### 1. MMKV + SecureStore adapter
- [ ] Tạo `mobile/src/platform/storage.ts`:
  ```typescript
  import { MMKV } from 'react-native-mmkv'
  import * as SecureStore from 'expo-secure-store'
  import type { PlatformAdapter } from '@velox/shared/platform'

  export const mmkv = new MMKV({ id: 'velox-storage' })

  export const mobilePlatform: PlatformAdapter = {
    storage: {
      // MMKV: synchronous, for prefs/cache/server URL
      getItem: (key) => mmkv.getString(key) ?? null,
      setItem: (key, value) => mmkv.set(key, value),
      removeItem: (key) => mmkv.delete(key),
    },
    secureStorage: {
      // SecureStore: async, encrypted, for tokens
      getItem: (key) => SecureStore.getItemAsync(key),
      setItem: (key, value) => SecureStore.setItemAsync(key, value),
      removeItem: (key) => SecureStore.deleteItemAsync(key),
    },
    getDeviceName: () => {
      return mmkv.getString('velox_device_name') || 'Velox Mobile'
    },
    getApiBaseUrl: () => {
      const serverUrl = mmkv.getString('velox_server_url')
      if (!serverUrl) throw new Error('Server URL not configured')
      // e.g. "http://192.168.1.100:8098/api"
      return `${serverUrl}/api`
    },
  }

  // Helpers for server URL management
  export function getServerUrl(): string | null {
    return mmkv.getString('velox_server_url') ?? null
  }
  export function setServerUrl(url: string): void {
    mmkv.set('velox_server_url', url)
  }
  export function hasServerUrl(): boolean {
    return mmkv.contains('velox_server_url')
  }
  ```

### 2. Mobile capabilities (hardcoded ExoPlayer)
- [ ] Tạo `mobile/src/platform/capabilities.ts`:
  ```typescript
  // ExoPlayer supports far more codecs than browser
  // These are sent to backend in PlaybackInfoRequest
  export const mobileCapabilities = {
    videoCodecs: ['h264', 'hevc', 'vp9', 'av1', 'vp8', 'mpeg2', 'mpeg4'],
    audioCodecs: ['aac', 'ac3', 'eac3', 'dts', 'mp3', 'opus', 'flac', 'vorbis', 'truehd'],
    containers: ['mp4', 'mkv', 'webm', 'avi', 'mov', 'ts'],
    supportsHLS: true,
    maxResolution: { width: 3840, height: 2160 },
  }
  ```

### 3. Init platform in app entry
- [ ] Update `mobile/app/_layout.tsx` — add at TOP (before providers):
  ```typescript
  import { initPlatform } from '@velox/shared/platform'
  import { setAuthStateGetter } from '@velox/shared/hooks/useAuth'
  import { mobilePlatform } from '../src/platform/storage'
  import { useAuthStore } from '../src/stores/auth'

  // Init before anything else
  initPlatform(mobilePlatform)
  setAuthStateGetter(() => useAuthStore.getState())
  ```

### 4. Create mobile auth store — SecureStore backend
- [ ] Tạo `mobile/src/stores/auth.ts`:
  ```typescript
  import { createAuthStore } from '@velox/shared/stores'
  import * as SecureStore from 'expo-secure-store'

  // SecureStore adapter (async) — Zustand persist v5 supports async storage
  const secureStorageAdapter = {
    getItem: async (name: string): Promise<string | null> => {
      return SecureStore.getItemAsync(name)
    },
    setItem: async (name: string, value: string): Promise<void> => {
      await SecureStore.setItemAsync(name, value)
    },
    removeItem: async (name: string): Promise<void> => {
      await SecureStore.deleteItemAsync(name)
    },
  }

  export const useAuthStore = createAuthStore(secureStorageAdapter)
  ```
  ⚠️ SecureStore là async → store hydration sẽ async. Cần handle loading state trong auth guard (show splash screen cho tới khi hydrated).

### 5. Create mobile player store — MMKV backend
- [ ] Tạo `mobile/src/stores/player.ts`:
  ```typescript
  import { createPlayerStore } from '@velox/shared/stores'
  import { mmkv } from '../platform/storage'

  // MMKV adapter (sync) — fast for player preferences
  const mmkvStorageAdapter = {
    getItem: (name: string) => mmkv.getString(name) ?? null,
    setItem: (name: string, value: string) => mmkv.set(name, value),
    removeItem: (name: string) => mmkv.delete(name),
  }

  export const usePlayerStore = createPlayerStore(mmkvStorageAdapter)
  ```

## Files to Create
- `mobile/src/platform/storage.ts` — MMKV + SecureStore adapter
- `mobile/src/platform/capabilities.ts` — hardcoded ExoPlayer capabilities
- `mobile/src/stores/auth.ts` — auth store with **SecureStore** backend (encrypted)
- `mobile/src/stores/player.ts` — player store with **MMKV** backend (fast)

## Files to Modify
- `mobile/app/_layout.tsx` — add `initPlatform()` + `setAuthStateGetter()`

## Important Notes
- **Tokens in SecureStore** (encrypted) — NOT in MMKV
- **Player prefs in MMKV** (fast sync) — non-sensitive data
- SecureStore async → auth guard cần wait for hydration before routing
- MMKV sync → player store available immediately

## Test Criteria
- [ ] MMKV read/write works: save server URL → read back
- [ ] SecureStore read/write works: save test token → read back
- [ ] Auth store persists tokens in SecureStore (verify via `SecureStore.getItemAsync`)
- [ ] Player store persists prefs in MMKV
- [ ] `getPlatform().getApiBaseUrl()` returns correct URL after setServerUrl()
- [ ] `getPlatform().getApiBaseUrl()` throws when no URL configured

---
Next Phase: [phase-11-auth-flow.md](phase-11-auth-flow.md)
