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
  refreshAccessToken,
} from './client'

export type { ApiErrorResponse, ApiResponse } from '../types/index'
