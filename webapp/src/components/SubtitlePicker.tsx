import { useState } from 'react'
import { LuCaptions, LuLanguages, LuSearch } from 'react-icons/lu'
import { SubtitleSearchModal } from '@/components/SubtitleSearchModal'
import { useTranslateSubtitle } from '@/hooks/stores/useMedia'
import type { PlaybackSubtitleTrack } from '@/types/api'

interface SubtitlePickerProps {
  subtitles: PlaybackSubtitleTrack[]
  primaryLanguage: string | null
  primaryTrackId?: number | null
  secondaryLanguage?: string | null
  secondaryTrackId?: number | null
  onSelectPrimary: (language: string | null, trackId?: number | null) => void
  onSelectPrimarySource?: (trackId: number | null) => void
  onSelectSecondary?: (language: string | null, trackId?: number | null) => void
  onSelectSecondarySource?: (trackId: number | null) => void
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
  primaryTrackId = null,
  secondaryLanguage = null,
  secondaryTrackId = null,
  onSelectPrimary,
  onSelectPrimarySource,
  onSelectSecondary,
  onSelectSecondarySource,
  dualMode = false,
  allowImageSubtitles = false,
  mediaId,
  onSubtitleAdded,
}: SubtitlePickerProps) {
  const [showSearch, setShowSearch] = useState(false)
  const allSubs = buildVisibleSubtitles(subtitles, allowImageSubtitles)
  const primarySources = buildSubtitleSources(subtitles, primaryLanguage, allowImageSubtitles)
  const secondarySources = buildSubtitleSources(subtitles, secondaryLanguage, false)
  const effectivePrimaryTrackId = primaryTrackId ?? primarySources[0]?.id ?? null
  const effectiveSecondaryTrackId = secondaryTrackId ?? secondarySources[0]?.id ?? null

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
          onClick={() => onSelectPrimary(null, null)}
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
              onClick={() => onSelectPrimary(sub.language, sub.id)}
            />
          )
        })}
      </div>

      {primarySources.length > 1 && onSelectPrimarySource && (
        <SourceSelector
          title="Subtitle Source"
          sources={primarySources}
          selectedTrackId={effectivePrimaryTrackId}
          onSelect={onSelectPrimarySource}
        />
      )}

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
              onClick={() => onSelectSecondary(null, null)}
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
                    onClick={() => onSelectSecondary(sub.language, sub.id)}
                  />
                )
              })}
          </div>
          {secondarySources.length > 1 && onSelectSecondarySource && (
            <SourceSelector
              title="Secondary Source"
              sources={secondarySources}
              selectedTrackId={effectiveSecondaryTrackId}
              onSelect={onSelectSecondarySource}
            />
          )}
        </>
      )}

      {/* Translate existing subtitle */}
      {allSubs.length > 0 && (
        <TranslateRow subtitles={allSubs} onTranslated={() => onSubtitleAdded?.()} />
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

function buildSubtitleSources(
  subtitles: PlaybackSubtitleTrack[],
  language: string | null,
  allowImageSubtitles: boolean,
): PlaybackSubtitleTrack[] {
  if (!language) return []
  return subtitles.filter((subtitle) => {
    if (!allowImageSubtitles && subtitle.is_image) return false
    return languageMatches(subtitle.language, language)
  })
}

function buildSourceLabel(subtitle: PlaybackSubtitleTrack): string {
  const { name, fmt } = parseLabel(subtitle.label, subtitle.language)
  const sourceName = name || subtitle.language || `Track ${subtitle.id}`
  const meta = [
    `#${subtitle.id}`,
    fmt || subtitle.format?.toUpperCase(),
    subtitle.is_default ? 'Default' : null,
  ]
    .filter(Boolean)
    .join(' • ')
  return meta ? `${sourceName} (${meta})` : sourceName
}

interface SourceSelectorProps {
  title: string
  sources: PlaybackSubtitleTrack[]
  selectedTrackId: number | null
  onSelect: (trackId: number | null) => void
}

function SourceSelector({ title, sources, selectedTrackId, onSelect }: SourceSelectorProps) {
  return (
    <div className="border-t border-white/10 px-4 py-3">
      <p className="mb-2 text-[10px] font-semibold uppercase tracking-wider text-white/40">
        {title}
      </p>
      <select
        value={selectedTrackId ?? ''}
        onChange={(e) => onSelect(e.target.value ? Number(e.target.value) : null)}
        className="w-full rounded-lg bg-white/6 px-3 py-2 text-sm text-white outline-none ring-1 ring-white/10 transition-colors hover:bg-white/10 focus:ring-white/20"
      >
        {sources.map((source) => (
          <option key={source.id} value={source.id} className="bg-[#242424] text-white">
            {buildSourceLabel(source)}
          </option>
        ))}
      </select>
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

const TRANSLATE_LANGS = [
  { code: 'vi', label: 'Vietnamese' },
  { code: 'en', label: 'English' },
  { code: 'fr', label: 'French' },
  { code: 'de', label: 'German' },
  { code: 'es', label: 'Spanish' },
  { code: 'ja', label: 'Japanese' },
  { code: 'ko', label: 'Korean' },
  { code: 'zh', label: 'Chinese' },
  { code: 'pt', label: 'Portuguese' },
  { code: 'ru', label: 'Russian' },
  { code: 'th', label: 'Thai' },
]

function TranslateRow({
  subtitles,
  onTranslated,
}: {
  subtitles: PlaybackSubtitleTrack[]
  onTranslated: () => void
}) {
  const [open, setOpen] = useState(false)
  const [sourceId, setSourceId] = useState<number | null>(null)
  const [targetLang, setTargetLang] = useState('')
  const translateMutation = useTranslateSubtitle()

  // Only show text-based subtitles as source
  const textSubs = subtitles.filter((s) => !s.is_image)
  if (textSubs.length === 0) return null

  const handleTranslate = () => {
    const subId = sourceId ?? textSubs[0]?.id
    if (!subId || !targetLang) return
    translateMutation.mutate(
      { subtitleId: subId, targetLanguage: targetLang },
      {
        onSuccess: () => {
          onTranslated()
          setOpen(false)
          setTargetLang('')
        },
      },
    )
  }

  if (!open) {
    return (
      <button
        onClick={() => setOpen(true)}
        className="flex w-full items-center gap-3 px-4 py-3 text-sm text-white/50 transition-colors hover:bg-white/8 hover:text-white/80"
      >
        <LuLanguages size={18} className="shrink-0" />
        <span>Translate Subtitle</span>
      </button>
    )
  }

  return (
    <div className="border-t border-white/10 px-4 py-3 space-y-3">
      <p className="text-[10px] font-semibold uppercase tracking-wider text-white/40">
        Translate subtitle
      </p>

      {textSubs.length > 1 && (
        <select
          value={sourceId ?? textSubs[0]?.id ?? ''}
          onChange={(e) => setSourceId(Number(e.target.value))}
          className="w-full rounded-lg bg-white/6 px-3 py-2 text-sm text-white outline-none ring-1 ring-white/10"
        >
          {textSubs.map((s) => (
            <option key={s.id} value={s.id} className="bg-[#242424] text-white">
              {s.label || s.language} ({s.format})
            </option>
          ))}
        </select>
      )}

      <select
        value={targetLang}
        onChange={(e) => setTargetLang(e.target.value)}
        className="w-full rounded-lg bg-white/6 px-3 py-2 text-sm text-white outline-none ring-1 ring-white/10"
      >
        <option value="" className="bg-[#242424] text-white">
          Translate to...
        </option>
        {TRANSLATE_LANGS.map((l) => (
          <option key={l.code} value={l.code} className="bg-[#242424] text-white">
            {l.label}
          </option>
        ))}
      </select>

      <div className="flex gap-2">
        <button
          onClick={handleTranslate}
          disabled={!targetLang || translateMutation.isPending}
          className="flex-1 rounded-lg bg-blue-600 px-3 py-2 text-sm font-medium text-white transition-colors hover:bg-blue-500 disabled:opacity-50"
        >
          {translateMutation.isPending ? 'Translating...' : 'Translate'}
        </button>
        <button
          onClick={() => setOpen(false)}
          className="rounded-lg bg-white/10 px-3 py-2 text-sm text-white/70 hover:bg-white/15"
        >
          Cancel
        </button>
      </div>

      {translateMutation.isError && (
        <p className="text-xs text-red-400">
          {(translateMutation.error as Error)?.message || 'Translation failed'}
        </p>
      )}
    </div>
  )
}
