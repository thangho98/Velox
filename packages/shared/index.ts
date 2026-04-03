/**
 * @velox/shared — Platform-agnostic shared package
 *
 * Exports types, API client, hooks, stores, and platform utilities
 * for both webapp and mobile apps.
 */

// Platform utilities
export {
  initPlatform,
  getPlatform,
  type PlatformAdapter,
  type StorageAdapter,
} from './platform'

// Types
export * from './types/index'

// API client
export * from './api/index'

// Stores
export * from './stores/index'

// Hooks
export * from './hooks/index'

// Libs
export * from './lib/index'
