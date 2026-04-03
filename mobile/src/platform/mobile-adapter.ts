/**
 * Mobile-specific PlatformAdapter implementation
 *
 * Uses expo-secure-store for secure storage (tokens).
 * API base URL is configurable — user enters server URL on first launch.
 */

import * as SecureStore from 'expo-secure-store'
import type { PlatformAdapter } from '@velox/shared/platform'

const SERVER_URL_KEY = 'velox_server_url'

let _serverUrl: string | null = null

/** Get current server URL (null if not configured) */
export function getServerUrl(): string | null {
  return _serverUrl
}

/** Set server URL and persist to SecureStore */
export async function setServerUrl(url: string): Promise<void> {
  // Normalize: remove trailing slash
  const normalized = url.replace(/\/+$/, '')
  _serverUrl = normalized
  await SecureStore.setItemAsync(SERVER_URL_KEY, normalized)
}

/** Load saved server URL from SecureStore (call on app start) */
export async function loadServerUrl(): Promise<string | null> {
  try {
    const saved = await SecureStore.getItemAsync(SERVER_URL_KEY)
    if (saved) {
      _serverUrl = saved
    }
    return _serverUrl
  } catch {
    return null
  }
}

/** Clear server URL (for switching servers) */
export async function clearServerUrl(): Promise<void> {
  _serverUrl = null
  await SecureStore.deleteItemAsync(SERVER_URL_KEY)
}

const storageAdapter = {
  getItem: async (key: string) => {
    try {
      return await SecureStore.getItemAsync(key)
    } catch {
      return null
    }
  },
  setItem: async (key: string, value: string) => {
    await SecureStore.setItemAsync(key, value)
  },
  removeItem: async (key: string) => {
    await SecureStore.deleteItemAsync(key)
  },
}

export const mobilePlatform: PlatformAdapter = {
  storage: storageAdapter,
  secureStorage: storageAdapter,
  getDeviceName: () => 'Mobile Device',
  getApiBaseUrl: () => {
    if (!_serverUrl) {
      throw new Error('Server URL not configured. Call setServerUrl() first.')
    }
    return `${_serverUrl}/api`
  },
}
