/**
 * Auth store — mobile implementation using shared factory + SecureStore
 */
import { createAuthStore } from '@velox/shared/stores'
import { setAuthStateGetter } from '@velox/shared/hooks/auth'
import * as SecureStore from 'expo-secure-store'

// SecureStore adapter for Zustand persist
const secureStorageAdapter = {
  getItem: async (name: string): Promise<string | null> => {
    try {
      return await SecureStore.getItemAsync(name)
    } catch {
      return null
    }
  },
  setItem: async (name: string, value: string): Promise<void> => {
    await SecureStore.setItemAsync(name, value)
  },
  removeItem: async (name: string): Promise<void> => {
    await SecureStore.deleteItemAsync(name)
  },
}

export const useAuthStore = createAuthStore(secureStorageAdapter)

// Inject auth state getter into shared hooks
setAuthStateGetter(() => useAuthStore.getState())
