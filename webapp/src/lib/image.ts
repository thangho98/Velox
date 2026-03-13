/**
 * Convert a TMDb image path to the local proxy URL.
 * Returns undefined if the path is empty/null so callers can use it directly
 * in optional props: `posterPath={tmdbImage(media.poster_path, 'w500')}`.
 */
export function tmdbImage(
  path: string | null | undefined,
  size: string = 'w500',
): string | undefined {
  if (!path) return undefined
  // Already a full URL or local API path — pass through
  if (path.startsWith('http') || path.startsWith('/api/')) return path
  // Strip leading slash for the proxy route
  const cleaned = path.startsWith('/') ? path.slice(1) : path
  return `/api/images/tmdb/${size}/${cleaned}`
}
