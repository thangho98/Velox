import type { ApiErrorResponse, ApiResponse } from '@/types/api'

const API_BASE_URL = '/api'

// Token refresh state
let isRefreshing = false
let refreshSubscribers: Array<{
  resolve: (token: string) => void
  reject: (error: Error) => void
}> = []

export class ApiError extends Error {
  status: number

  constructor(message: string, status: number) {
    super(message)
    this.name = 'ApiError'
    this.status = status
  }
}

// Callbacks injected by auth store
let getAccessToken: () => string | null = () => null
let getRefreshToken: () => string | null = () => null
let setTokensCallback: (
  accessToken: string,
  refreshToken: string,
  expiresIn: number,
) => void = () => {}
let onSessionExpiredCallback: () => void = () => {}

export function setTokenCallbacks(
  getAccess: () => string | null,
  getRefresh: () => string | null,
  setTokens: (accessToken: string, refreshToken: string, expiresIn: number) => void,
) {
  getAccessToken = getAccess
  getRefreshToken = getRefresh
  setTokensCallback = setTokens
}

export function setSessionExpiredCallback(callback: () => void) {
  onSessionExpiredCallback = callback
}

function onTokenRefreshed(newToken: string) {
  refreshSubscribers.forEach(({ resolve }) => resolve(newToken))
  refreshSubscribers = []
}

function onTokenRefreshFailed(error: Error) {
  refreshSubscribers.forEach(({ reject }) => reject(error))
  refreshSubscribers = []
}

async function refreshAccessToken(): Promise<string | null> {
  const refreshToken = getRefreshToken()
  if (!refreshToken) return null

  try {
    const response = await fetch(`${API_BASE_URL}/auth/refresh`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ refresh_token: refreshToken }),
    })

    if (!response.ok) return null

    const body = await response.json()
    const data = body.data // unwrap {"data": ...} envelope
    setTokensCallback(data.access_token, data.refresh_token, data.expires_in)
    return data.access_token
  } catch {
    return null
  }
}

async function fetchWithAuth<T>(endpoint: string, options: RequestInit = {}): Promise<T> {
  const url = `${API_BASE_URL}${endpoint}`
  const token = getAccessToken()

  const headers: Record<string, string> = {
    'Content-Type': 'application/json',
    ...((options.headers as Record<string, string>) || {}),
  }

  if (token) {
    headers.Authorization = `Bearer ${token}`
  }

  // Add device name header if available
  const deviceName = localStorage.getItem('velox_device_name')
  if (deviceName) {
    headers['X-Device-Name'] = deviceName
  }

  const response = await fetch(url, {
    ...options,
    headers,
  })

  // Handle 401 Unauthorized - try to refresh token
  if (response.status === 401 && token) {
    if (!isRefreshing) {
      isRefreshing = true
      const newToken = await refreshAccessToken()
      isRefreshing = false

      if (newToken) {
        onTokenRefreshed(newToken)
        // Retry the original request with new token
        return fetchWithAuth(endpoint, options)
      }

      // Refresh failed - notify all queued requests and clear auth
      const sessionError = new ApiError('Session expired', 401)
      onTokenRefreshFailed(sessionError)
      onSessionExpiredCallback()
      throw sessionError
    }

    // Another request is already refreshing — queue this one
    return new Promise((resolve, reject) => {
      refreshSubscribers.push({
        resolve: (newToken) => {
          const retryHeaders: Record<string, string> = {
            'Content-Type': 'application/json',
            ...((options.headers as Record<string, string>) || {}),
            Authorization: `Bearer ${newToken}`,
          }
          fetch(url, { ...options, headers: retryHeaders })
            .then(async (res) => {
              if (res.status === 204) {
                resolve(undefined as T)
                return
              }
              const data = await res.json().catch(() => ({}))
              if (!res.ok) {
                reject(
                  new ApiError(
                    (data as ApiErrorResponse).error || `HTTP ${res.status}`,
                    res.status,
                  ),
                )
              } else {
                resolve((data as ApiResponse<T>).data)
              }
            })
            .catch(reject)
        },
        reject,
      })
    })
  }

  // Handle 204 No Content
  if (response.status === 204) {
    return undefined as T
  }

  const data = await response.json().catch(() => ({}))

  if (!response.ok) {
    const errorMessage = (data as ApiErrorResponse).error || `HTTP ${response.status}`
    throw new ApiError(errorMessage, response.status)
  }

  return (data as ApiResponse<T>).data
}

// HTTP Methods
export const api = {
  get: <T>(endpoint: string) => fetchWithAuth<T>(endpoint, { method: 'GET' }),

  post: <T>(endpoint: string, body: unknown) =>
    fetchWithAuth<T>(endpoint, {
      method: 'POST',
      body: JSON.stringify(body),
    }),

  patch: <T>(endpoint: string, body: unknown) =>
    fetchWithAuth<T>(endpoint, {
      method: 'PATCH',
      body: JSON.stringify(body),
    }),

  put: <T>(endpoint: string, body: unknown) =>
    fetchWithAuth<T>(endpoint, {
      method: 'PUT',
      body: JSON.stringify(body),
    }),

  delete: (endpoint: string) => fetchWithAuth<void>(endpoint, { method: 'DELETE' }),
}

// Stream URL helpers (for video streaming)
export function getDirectStreamUrl(mediaId: number, token?: string): string {
  const base = `${API_BASE_URL}/stream/${mediaId}`
  if (token) {
    return `${base}?token=${encodeURIComponent(token)}`
  }
  return base
}

export function getHlsMasterUrl(mediaId: number, token?: string): string {
  const base = `${API_BASE_URL}/stream/${mediaId}/hls/master.m3u8`
  if (token) {
    return `${base}?token=${encodeURIComponent(token)}`
  }
  return base
}
