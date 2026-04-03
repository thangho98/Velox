/**
 * Player store — web implementation using shared factory
 */
import { createPlayerStore } from '@velox/shared/stores'

// Web uses localStorage for persistence
const localStorageAdapter = {
  getItem: (name: string) => localStorage.getItem(name),
  setItem: (name: string, value: string) => localStorage.setItem(name, value),
  removeItem: (name: string) => localStorage.removeItem(name),
}

export const usePlayerStore = createPlayerStore(localStorageAdapter)
