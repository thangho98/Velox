import type { PlaybackSubtitleTrack } from '@/types/api'

interface SubtitlePickerProps {
  subtitles: PlaybackSubtitleTrack[]
  /** Currently selected primary subtitle language, or null for Off */
  primaryLanguage: string | null
  /** Currently selected secondary subtitle language, or null for Off */
  secondaryLanguage?: string | null
  onSelectPrimary: (language: string | null) => void
  onSelectSecondary?: (language: string | null) => void
  /** If true, show dual-sub selection (primary + secondary columns) */
  dualMode?: boolean
}

export function SubtitlePicker({
  subtitles,
  primaryLanguage,
  secondaryLanguage = null,
  onSelectPrimary,
  onSelectSecondary,
  dualMode = false,
}: SubtitlePickerProps) {
  // Separate text subs from image subs (PGS/VobSub)
  const textSubs = subtitles.filter((s) => !s.is_image)
  const imageSubs = subtitles.filter((s) => s.is_image)

  return (
    <div className="min-w-[200px] rounded-lg bg-black/90 p-2 shadow-xl">
      {dualMode ? (
        <div className="grid grid-cols-2 gap-1">
          {/* Primary column */}
          <div>
            <p className="mb-1 px-2 text-xs font-semibold text-gray-400">Primary</p>
            <OptionList items={textSubs} selected={primaryLanguage} onSelect={onSelectPrimary} />
          </div>
          {/* Secondary column */}
          <div>
            <p className="mb-1 px-2 text-xs font-semibold text-gray-400">Secondary</p>
            <OptionList
              items={textSubs}
              selected={secondaryLanguage}
              onSelect={onSelectSecondary ?? (() => {})}
            />
          </div>
        </div>
      ) : (
        <OptionList items={textSubs} selected={primaryLanguage} onSelect={onSelectPrimary} />
      )}

      {/* Image subtitle section */}
      {imageSubs.length > 0 && (
        <div className="mt-1 border-t border-white/10 pt-1">
          <p className="mb-1 px-2 text-xs text-gray-500">Image subtitles (server burn-in)</p>
          {imageSubs.map((sub) => (
            <button
              key={sub.id}
              onClick={() => onSelectPrimary(sub.language)}
              className={`w-full rounded px-3 py-2 text-left text-sm ${
                primaryLanguage === sub.language
                  ? 'bg-netflix-red text-white'
                  : 'text-gray-400 hover:bg-white/10 hover:text-white'
              }`}
            >
              {sub.label}
              <span className="ml-1 text-xs opacity-60">(PGS)</span>
            </button>
          ))}
        </div>
      )}
    </div>
  )
}

interface OptionListProps {
  items: PlaybackSubtitleTrack[]
  selected: string | null
  onSelect: (language: string | null) => void
}

function OptionList({ items, selected, onSelect }: OptionListProps) {
  return (
    <>
      <button
        onClick={() => onSelect(null)}
        className={`w-full rounded px-3 py-2 text-left text-sm ${
          selected === null ? 'bg-netflix-red text-white' : 'text-white hover:bg-white/10'
        }`}
      >
        Off
      </button>
      {items.map((sub) => (
        <button
          key={sub.id}
          onClick={() => onSelect(sub.language)}
          className={`w-full rounded px-3 py-2 text-left text-sm ${
            selected === sub.language ? 'bg-netflix-red text-white' : 'text-white hover:bg-white/10'
          }`}
        >
          {sub.label}
        </button>
      ))}
    </>
  )
}
