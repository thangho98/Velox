import { useState } from 'react'
import { LuLanguages } from 'react-icons/lu'
import { useTranslateSubtitle } from '@/hooks/stores/useMedia'
import type { PlaybackSubtitleTrack } from '@/types/api'

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

export function SubtitleTranslate({
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
