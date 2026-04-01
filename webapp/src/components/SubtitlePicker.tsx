import { useState } from 'react'
import { LuCaptions, LuSearch } from 'react-icons/lu'
import { SubtitleSearchModal } from '@/components/SubtitleSearchModal'
import { SubtitleTranslate } from '@/components/SubtitleTranslate'
import { parseSubtitleLabel, languageMatches, buildVisibleSubtitles } from '@/lib/languages'
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
    <div className="w-full rounded-xl bg-[#242424] shadow-2xl ring-1 ring-white/10 overflow-hidden sm:w-72">
      {/* Header */}
      <div className="border-b border-white/10 px-4 py-3 text-center">
        <p className="text-sm font-semibold text-white">Subtitles</p>
      </div>

      {/* Primary list */}
      <div className="max-h-[50vh] overflow-y-auto">
        <SubRow
          icon={<LuCaptions size={18} />}
          name="Off"
          fmt=""
          selected={primaryLanguage === null}
          onClick={() => onSelectPrimary(null, null)}
        />

        {allSubs.map((sub) => {
          const { name, fmt } = parseSubtitleLabel(sub.label, sub.language)
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
                const { name, fmt } = parseSubtitleLabel(sub.label, sub.language)
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
        <SubtitleTranslate subtitles={allSubs} onTranslated={() => onSubtitleAdded?.()} />
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

// ── Private helpers ──────────────────────────────────────────────────────────

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
  const { name, fmt } = parseSubtitleLabel(subtitle.label, subtitle.language)
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

function SourceSelector({
  title,
  sources,
  selectedTrackId,
  onSelect,
}: {
  title: string
  sources: PlaybackSubtitleTrack[]
  selectedTrackId: number | null
  onSelect: (trackId: number | null) => void
}) {
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

function SubRow({
  icon,
  name,
  fmt,
  selected,
  onClick,
}: {
  icon: React.ReactNode
  name: string
  fmt: string
  selected: boolean
  onClick: () => void
}) {
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
