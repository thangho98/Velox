/**
 * Auth store
 *
 * - Web: tokens persisted to localStorage
 * - Desktop (Tauri): tokens persisted to OS keychain via platform.secureStorage
 */
import { useEffect, useState } from 'react'
import { createAuthStore } from '@velox/shared/stores'
import { getPlatform } from '@velox/shared/platform'
import { isTauri } from '@/platform'
import type { StateStorage } from 'zustand/middleware'

const desktopStorage: StateStorage = {
  getItem: async (name) => (await getPlatform().secureStorage.getItem(name)) ?? null,
  setItem: async (name, value) => {
    await getPlatform().secureStorage.setItem(name, value)
  },
  removeItem: async (name) => {
    await getPlatform().secureStorage.removeItem(name)
  },
}

const webStorage: StateStorage = {
  getItem: (name) => localStorage.getItem(name),
  setItem: (name, value) => localStorage.setItem(name, value),
  removeItem: (name) => localStorage.removeItem(name),
}

export const useAuthStore = createAuthStore(isTauri() ? desktopStorage : webStorage)

/**
 * Returns true once the persist middleware has finished rehydrating from storage.
 * On web (sync localStorage) this is true on first render. On desktop (async keychain)
 * the first render returns false; gate route guards on this to avoid a flash of
 * /onboarding when a valid token exists in the keychain.
 */
export function useAuthHydrated(): boolean {
  const [hydrated, setHydrated] = useState(() => useAuthStore.persist.hasHydrated())
  useEffect(() => {
    const unsub = useAuthStore.persist.onFinishHydration(() => setHydrated(true))
    setHydrated(useAuthStore.persist.hasHydrated())
    return unsub
  }, [])
  return hydrated
}
