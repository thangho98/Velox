export const DETAIL_PANEL_ANIMATION_MS = 520

export function normalizeLanguageCode(language: string | null | undefined): string {
  const value = (language ?? '').trim().toLowerCase()
  switch (value) {
    case 'en':
    case 'eng':
      return 'eng'
    case 'vi':
    case 'vie':
      return 'vie'
    case 'zh':
    case 'zho':
    case 'chi':
      return 'zho'
    default:
      return value
  }
}

export function languageMatches(
  lhs: string | null | undefined,
  rhs: string | null | undefined,
): boolean {
  if (!lhs || !rhs) return false
  return normalizeLanguageCode(lhs) === normalizeLanguageCode(rhs)
}

export function formatRuntimeMinutes(seconds: number): string {
  if (!seconds || Number.isNaN(seconds)) return ''
  return `${Math.max(1, Math.round(seconds / 60))}m`
}

export function formatChannelLayout(channels: number): string {
  if (!channels || Number.isNaN(channels)) return ''
  if (channels === 6) return '5.1'
  if (channels === 8) return '7.1'
  if (channels >= 6) return `${channels}ch`
  return `${channels}.0`
}

export function formatResolutionLabel(height: number): string {
  if (!height || Number.isNaN(height)) return ''
  if (height >= 2160) return '4K'
  return `${height}p`
}

export function formatLanguageLabel(language: string | null | undefined): string {
  const value = normalizeLanguageCode(language)
  switch (value) {
    case 'eng':
      return 'English'
    case 'vie':
      return 'Vietnamese'
    case 'zho':
      return 'Chinese'
    default:
      return language ? language.charAt(0).toUpperCase() + language.slice(1) : 'Unknown'
  }
}

export function formatTime(seconds: number): string {
  if (!seconds || Number.isNaN(seconds)) return '0:00'
  const mins = Math.floor(seconds / 60)
  const secs = Math.floor(seconds % 60)
  const hours = Math.floor(mins / 60)
  if (hours > 0) {
    return `${hours}:${(mins % 60).toString().padStart(2, '0')}:${secs.toString().padStart(2, '0')}`
  }
  return `${mins}:${secs.toString().padStart(2, '0')}`
}

export function getWallClock(): string {
  return new Date().toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })
}
