/**
 * Image URL utilities — platform-agnostic.
 *
 * Backend returns fully-routed relative paths like `/api/images/local/...` or
 * `/api/images/tmdb/abc.jpg` (no size segment). The helper prefixes the API
 * base and optionally appends a `width=N` query param. The backend maps that
 * to the nearest TMDb size bucket and, for local uploads, serves the stored
 * asset.
 *
 * External URLs (http/https) pass through untouched.
 */

import { getPlatform } from '../platform'

function getApiBaseUrl(): string {
  return getPlatform().getApiBaseUrl()
}

/**
 * Accepts either a numeric width, a TMDb-style size token (`w500`, `h632`),
 * or `original`. Returns a query-safe width value or `undefined` to skip the
 * query param entirely.
 */
function coerceWidth(size: number | string | undefined): string | undefined {
  if (size === undefined || size === null) return undefined
  if (typeof size === 'number') return size > 0 ? String(size) : undefined
  if (size === 'original') return 'original'
  // `w500`, `h632`, etc. → strip leading letter
  const match = /^[a-z](\d+)$/i.exec(size)
  if (match) return match[1]
  const n = Number(size)
  return Number.isFinite(n) && n > 0 ? String(n) : undefined
}

/**
 * Resolve any image path returned by the backend. Returns `undefined` for
 * empty input so it slots directly into optional props.
 */
export function resolveImageUrl(
  path: string | null | undefined,
  size?: number | string,
): string | undefined {
  if (!path) return undefined
  if (path.startsWith('http://') || path.startsWith('https://')) return path

  const base = getApiBaseUrl()
  // baseUrl ends at `/api`; strip the leading `/api` from the path before
  // joining so we don't get `/api/api/...`.
  const joined = path.startsWith('/api/')
    ? `${base}${path.slice(4)}`
    : `${base}${path.startsWith('/') ? path : '/' + path}`

  const width = coerceWidth(size)
  if (!width) return joined
  const separator = joined.includes('?') ? '&' : '?'
  return `${joined}${separator}width=${width}`
}
