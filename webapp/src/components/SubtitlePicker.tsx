import { useState } from 'react'
import { LuCaptions, LuSearch } from 'react-icons/lu'
import { SubtitleSearchModal } from '@/components/SubtitleSearchModal'
import type { PlaybackSubtitleTrack } from '@/types/api'

interface SubtitlePickerProps {
  subtitles: PlaybackSubtitleTrack[]
  primaryLanguage: string | null
  secondaryLanguage?: string | null
  onSelectPrimary: (language: string | null) => void
  onSelectSecondary?: (language: string | null) => void
  dualMode?: boolean
  allowImageSubtitles?: boolean
  mediaId: number
  onSubtitleAdded?: () => void
}

const LANG_NAMES: Record<string, string> = {
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

function parseLabel(label: string, language?: string): { name: string; fmt: string } {
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
  // Fallback to language name when label is empty
  return { name: languageName, fmt: '' }
}

function normalizeLanguageCode(language: string | null | undefined): string {
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

function languageMatches(lhs: string | null | undefined, rhs: string | null | undefined): boolean {
  if (!lhs || !rhs) return false
  return normalizeLanguageCode(lhs) === normalizeLanguageCode(rhs)
}

function buildVisibleSubtitles(
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

export function SubtitlePicker({
  subtitles,
  primaryLanguage,
  secondaryLanguage = null,
  onSelectPrimary,
  onSelectSecondary,
  dualMode = false,
  allowImageSubtitles = false,
  mediaId,
  onSubtitleAdded,
}: SubtitlePickerProps) {
  const [showSearch, setShowSearch] = useState(false)
  const allSubs = buildVisibleSubtitles(subtitles, allowImageSubtitles)

  return (
    <div className="w-72 rounded-xl bg-[#242424] shadow-2xl ring-1 ring-white/10 overflow-hidden">
      {/* Header */}
      <div className="border-b border-white/10 px-4 py-3 text-center">
        <p className="text-sm font-semibold text-white">Subtitles</p>
      </div>

      {/* Primary list */}
      <div className="max-h-[50vh] overflow-y-auto">
        {/* Off */}
        <SubRow
          icon={<LuCaptions size={18} />}
          name="Off"
          fmt=""
          selected={primaryLanguage === null}
          onClick={() => onSelectPrimary(null)}
        />

        {allSubs.map((sub) => {
          const { name, fmt } = parseLabel(sub.label, sub.language)
          return (
            <SubRow
              key={sub.id}
              icon={<LuCaptions size={18} />}
              name={name}
              fmt={fmt || sub.format}
              selected={languageMatches(primaryLanguage, sub.language)}
              onClick={() => onSelectPrimary(sub.language)}
            />
          )
        })}
      </div>

      {/* Secondary subtitle section (dual mode) */}
      {dualMode && onSelectSecondary && (
        <>
          <div className="border-t border-white/10 px-4 py-2">
            <p className="text-[10px] font-semibold uppercase tracking-wider text-white/40">
              Secondary subtitle
            </p>
          </div>
          <div className="max-h-[25vh] overflow-y-auto border-b border-white/10">
            <SubRow
              icon={<LuCaptions size={18} />}
              name="Off"
              fmt=""
              selected={secondaryLanguage === null}
              onClick={() => onSelectSecondary(null)}
            />
            {allSubs
              .filter((s) => !s.is_image)
              .map((sub) => {
                const { name, fmt } = parseLabel(sub.label, sub.language)
                return (
                  <SubRow
                    key={sub.id}
                    icon={<LuCaptions size={18} />}
                    name={name}
                    fmt={fmt || sub.format}
                    selected={languageMatches(secondaryLanguage, sub.language)}
                    onClick={() => onSelectSecondary(sub.language)}
                  />
                )
              })}
          </div>
        </>
      )}

      {/* Search for Subtitles */}
      <button
        onClick={() => setShowSearch(true)}
        className="flex w-full items-center gap-3 px-4 py-3 text-sm text-white/50 transition-colors hover:bg-white/8 hover:text-white/80"
      >
        <LuSearch size={18} className="shrink-0" />
        <span>Search for Subtitles</span>
      </button>

      {/* Search modal */}
      {showSearch && (
        <SubtitleSearchModal
          mediaId={mediaId}
          defaultLang={primaryLanguage}
          onClose={() => setShowSearch(false)}
          onSubtitleDownloaded={() => {
            onSubtitleAdded?.()
            setShowSearch(false)
          }}
        />
      )}
    </div>
  )
}

interface SubRowProps {
  icon: React.ReactNode
  name: string
  fmt: string
  selected: boolean
  onClick: () => void
}

function SubRow({ icon, name, fmt, selected, onClick }: SubRowProps) {
  return (
    <button
      onClick={onClick}
      className={`flex w-full items-center gap-3 px-4 py-2.5 text-left transition-colors hover:bg-white/8 ${
        selected ? 'text-white' : 'text-white/70'
      }`}
    >
      <span className={`shrink-0 ${selected ? 'text-white' : 'text-white/40'}`}>{icon}</span>
      <span className="flex-1 min-w-0">
        <span className="block text-sm font-medium leading-tight">{name}</span>
        {fmt && <span className="block text-xs text-white/40 leading-tight mt-0.5">{fmt}</span>}
      </span>
      <span className="shrink-0 w-4 text-center text-sm">{selected ? '✓' : ''}</span>
    </button>
  )
}
