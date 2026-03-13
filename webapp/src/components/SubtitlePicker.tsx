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
  if (label) {
    const match = label.match(/^(.*?)\s*\(([^)]+)\)$/)
    if (match) return { name: match[1].trim(), fmt: match[2].trim() }
    return { name: label, fmt: '' }
  }
  // Fallback to language name when label is empty
  const name = (language && LANG_NAMES[language]) || language || 'Unknown'
  return { name, fmt: '' }
}

export function SubtitlePicker({
  subtitles,
  primaryLanguage,
  secondaryLanguage = null,
  onSelectPrimary,
  onSelectSecondary,
  dualMode = false,
  mediaId,
  onSubtitleAdded,
}: SubtitlePickerProps) {
  const [showSearch, setShowSearch] = useState(false)
  const allSubs = subtitles

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
              selected={primaryLanguage === sub.language}
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
                    selected={secondaryLanguage === sub.language}
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
