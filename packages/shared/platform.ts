/**
 * Platform abstraction — web vs mobile inject their own implementations
 *
 * Web: localStorage for both storage and secureStorage (tokens in memory for security)
 * Mobile: MMKV for regular storage, expo-secure-store for tokens
 */

export interface StorageAdapter {
  getItem(key: string): string | null | Promise<string | null>
  setItem(key: string, value: string): void | Promise<void>
  removeItem(key: string): void | Promise<void>
}

export interface PlatformAdapter {
  /**
   * Regular storage (settings, UI state)
   * - Web: localStorage
   * - Mobile: MMKV
   */
  storage: StorageAdapter

  /**
   * Secure storage (tokens, sensitive data)
   * - Web: localStorage (consider using memory for tokens in production)
   * - Mobile: expo-secure-store
   */
  secureStorage: StorageAdapter

  /** Device name for X-Device-Name header */
  getDeviceName(): string

  /**
   * Base URL for API
   * - Web: '/api' (relative, uses Vite proxy)
   * - Mobile: 'http://192.168.1.x:8098/api' (direct LAN connection)
   */
  getApiBaseUrl(): string
}

// Singleton — set once at app init, used by api client + stores
let _platform: PlatformAdapter | null = null

export function initPlatform(adapter: PlatformAdapter): void {
  if (_platform) {
    console.warn('[Platform] Platform already initialized. Ignoring duplicate init.')
    return
  }
  _platform = adapter
}

export function getPlatform(): PlatformAdapter {
  if (!_platform) {
    throw new Error('[Platform] Platform not initialized. Call initPlatform() first.')
  }
  return _platform
}
