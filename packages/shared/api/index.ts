/**
 * @velox/shared — API client exports
 */

export {
  api,
  ApiError,
  setTokenCallbacks,
  setSessionExpiredCallback,
  getDirectStreamUrl,
  getHlsMasterUrl,
} from './client'

export type { ApiErrorResponse, ApiResponse } from '../types/index'
