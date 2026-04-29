/**
 * Desktop (Tauri) PlatformAdapter implementation.
 *
 * - Regular storage: tauri-plugin-store (JSON file under app config dir)
 * - Secure storage: OS keychain via Rust `keyring` crate (Tauri commands)
 * - API base URL: full http://NAS/api URL persisted in store (no Vite proxy)
 */

import type { PlatformAdapter } from '@velox/shared/platform'
import { invoke } from '@tauri-apps/api/core'
import { LazyStore } from '@tauri-apps/plugin-store'

const SETTINGS_FILE = 'velox-settings.json'
const STORE = new LazyStore(SETTINGS_FILE)

const STORAGE_PREFIX = 'storage:'
const SERVER_URL_KEY = 'server_url'
const DEVICE_NAME_KEY = 'device_name'

let cachedApiBaseUrl: string = ''
let cachedDeviceName: string = ''

const adapter: PlatformAdapter = {
  storage: {
    getItem: (key) => STORE.get<string>(STORAGE_PREFIX + key).then((v) => v ?? null),
    setItem: async (key, value) => {
      await STORE.set(STORAGE_PREFIX + key, value)
      await STORE.save()
    },
    removeItem: async (key) => {
      await STORE.delete(STORAGE_PREFIX + key)
      await STORE.save()
    },
  },
  secureStorage: {
    getItem: (key) => invoke<string | null>('secure_get', { key }),
    setItem: (key, value) => invoke<void>('secure_set', { key, value }),
    removeItem: (key) => invoke<void>('secure_remove', { key }),
  },
  getDeviceName: () => cachedDeviceName,
  getApiBaseUrl: () => {
    // Fail loudly if not paired yet — silent fallback to '/api' makes every
    // request hit the Vite dev proxy in Tauri, masking the real bug.
    if (!cachedApiBaseUrl) {
      console.error(
        '[desktop-adapter] getApiBaseUrl called before server URL is set — ' +
          'user likely has stale auth token without paired server. ' +
          'Logout via Settings → Desktop App and pair again.',
      )
      return '/api'
    }
    return cachedApiBaseUrl.replace(/\/+$/, '') + '/api'
  },
}

/** Initialize the adapter — loads cached server URL + device name from disk. */
export async function initDesktopAdapter(): Promise<PlatformAdapter> {
  const [url, name] = await Promise.all([
    STORE.get<string>(SERVER_URL_KEY),
    STORE.get<string>(DEVICE_NAME_KEY),
  ])
  cachedApiBaseUrl = url ?? ''
  cachedDeviceName = name ?? ''
  return adapter
}

export async function setServerUrl(url: string) {
  cachedApiBaseUrl = url
  await STORE.set(SERVER_URL_KEY, url)
  await STORE.save()
}

export async function setDeviceName(name: string) {
  cachedDeviceName = name
  await STORE.set(DEVICE_NAME_KEY, name)
  await STORE.save()
}

export async function getServerUrl(): Promise<string> {
  return (await STORE.get<string>(SERVER_URL_KEY)) ?? ''
}

export async function clearServerUrl() {
  cachedApiBaseUrl = ''
  await STORE.delete(SERVER_URL_KEY)
  await STORE.save()
}

/** Sync check usable from React render — true once the adapter has a server URL. */
export function isPaired(): boolean {
  return !!cachedApiBaseUrl
}
