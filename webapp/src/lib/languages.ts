import type { PlaybackSubtitleTrack } from '@/types/api'

export const LANG_NAMES: Record<string, string> = {
  eng: 'English',
  en: 'English',
  vie: 'Vietnamese',
  vi: 'Vietnamese',
  jpn: 'Japanese',
  ja: 'Japanese',
  kor: 'Korean',
  ko: 'Korean',
  zho: 'Chinese',
  zh: 'Chinese',
  fra: 'French',
  fr: 'French',
  deu: 'German',
  de: 'German',
  spa: 'Spanish',
  es: 'Spanish',
  ita: 'Italian',
  it: 'Italian',
  por: 'Portuguese',
  pt: 'Portuguese',
  rus: 'Russian',
  ru: 'Russian',
  tha: 'Thai',
  th: 'Thai',
  ara: 'Arabic',
  ar: 'Arabic',
  hin: 'Hindi',
  hi: 'Hindi',
  ind: 'Indonesian',
  id: 'Indonesian',
  msa: 'Malay',
  ms: 'Malay',
}

export function parseSubtitleLabel(
  label: string,
  language?: string,
): { name: string; fmt: string } {
  const languageName = (language && LANG_NAMES[language]) || language || 'Unknown'
  const normalizedLabel = label.trim()

  if (/^(sdh|cc|forced)$/i.test(normalizedLabel)) {
    return { name: languageName, fmt: normalizedLabel.toUpperCase() }
  }

  if (label) {
    const match = label.match(/^(.*?)\s*\(([^)]+)\)$/)
    if (match) return { name: match[1].trim(), fmt: match[2].trim() }
    return { name: label, fmt: '' }
  }
  return { name: languageName, fmt: '' }
}

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

export function buildVisibleSubtitles(
  subtitles: PlaybackSubtitleTrack[],
  allowImageSubtitles: boolean,
): PlaybackSubtitleTrack[] {
  const byLanguage = new Map<string, PlaybackSubtitleTrack>()

  for (const subtitle of subtitles) {
    if (!allowImageSubtitles && subtitle.is_image) continue

    const key = normalizeLanguageCode(subtitle.language || String(subtitle.id))
    const current = byLanguage.get(key)
    if (!current) {
      byLanguage.set(key, subtitle)
      continue
    }

    if (current.is_image && !subtitle.is_image) {
      byLanguage.set(key, subtitle)
      continue
    }

    if (!current.is_default && subtitle.is_default) {
      byLanguage.set(key, subtitle)
    }
  }

  return Array.from(byLanguage.values())
}
