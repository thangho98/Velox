/**
 * Re-export API client from shared package
 * Webapp uses this for backward compatibility
 */
export {
  api,
  ApiError,
  setTokenCallbacks,
  setSessionExpiredCallback,
  getDirectStreamUrl,
  getHlsMasterUrl,
} from '@velox/shared/api'

export type { ApiErrorResponse, ApiResponse } from '@velox/shared/types'
