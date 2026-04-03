/**
 * Web-specific PlatformAdapter implementation
 *
 * Uses localStorage for both regular and secure storage.
 * API base URL is relative (Vite proxy handles CORS).
 */

import type { PlatformAdapter } from '@velox/shared/platform'

export const webPlatform: PlatformAdapter = {
  storage: {
    getItem: (key) => localStorage.getItem(key),
    setItem: (key, value) => localStorage.setItem(key, value),
    removeItem: (key) => localStorage.removeItem(key),
  },
  secureStorage: {
    // Web doesn't have secure storage by default
    // Consider using memory-only for tokens in production
    getItem: (key) => localStorage.getItem(key),
    setItem: (key, value) => localStorage.setItem(key, value),
    removeItem: (key) => localStorage.removeItem(key),
  },
  getDeviceName: () => localStorage.getItem('velox_device_name') || '',
  getApiBaseUrl: () => '/api', // Vite proxy handles this
}
