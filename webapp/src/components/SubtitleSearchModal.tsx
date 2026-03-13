import { useState, useEffect } from 'react'
import { LuX, LuDownload, LuCheck, LuLoaderCircle, LuSearch } from 'react-icons/lu'
import { useSubtitleSearch, useDownloadSubtitle } from '@/hooks/stores/useMedia'
import type { SubtitleSearchResult } from '@/types/api'

const LANGUAGES = [
  { code: 'en', name: 'English' },
  { code: 'vi', name: 'Vietnamese' },
  { code: 'fr', name: 'French' },
  { code: 'de', name: 'German' },
  { code: 'es', name: 'Spanish' },
  { code: 'pt', name: 'Portuguese' },
  { code: 'it', name: 'Italian' },
  { code: 'ja', name: 'Japanese' },
  { code: 'ko', name: 'Korean' },
  { code: 'zh', name: 'Chinese' },
  { code: 'nl', name: 'Dutch' },
  { code: 'pl', name: 'Polish' },
  { code: 'ru', name: 'Russian' },
  { code: 'ar', name: 'Arabic' },
  { code: 'tr', name: 'Turkish' },
  { code: 'sv', name: 'Swedish' },
  { code: 'th', name: 'Thai' },
  { code: 'id', name: 'Indonesian' },
]

interface SubtitleSearchModalProps {
  mediaId: number
  defaultLang?: string | null
  onClose: () => void
  onSubtitleDownloaded: () => void
}

export function SubtitleSearchModal({
  mediaId,
  defaultLang,
  onClose,
  onSubtitleDownloaded,
}: SubtitleSearchModalProps) {
  const [lang, setLang] = useState(defaultLang || 'en')
  const [downloaded, setDownloaded] = useState<Set<string>>(new Set())

  const { data: results, isLoading, refetch } = useSubtitleSearch(mediaId, lang)
  const downloadMutation = useDownloadSubtitle(mediaId)

  // Re-search when language changes
  useEffect(() => {
    refetch()
  }, [lang, refetch])

  function handleDownload(result: SubtitleSearchResult) {
    const key = `${result.provider}:${result.external_id}`
    downloadMutation.mutate(
      {
        provider: result.provider,
        external_id: result.external_id,
        language: result.language || lang,
      },
      {
        onSuccess: () => {
          setDownloaded((prev) => new Set(prev).add(key))
          onSubtitleDownloaded()
        },
      },
    )
  }

  const downloadingKey = downloadMutation.isPending
    ? `${downloadMutation.variables?.provider}:${downloadMutation.variables?.external_id}`
    : null

  return (
    <div className="pointer-events-auto absolute inset-0 z-50 flex items-center justify-center bg-black/60 backdrop-blur-sm">
      <div className="w-[480px] max-h-[80vh] rounded-xl bg-[#1a1a1a] shadow-2xl ring-1 ring-white/10 overflow-hidden flex flex-col">
        {/* Header */}
        <div className="flex items-center justify-between border-b border-white/10 px-5 py-3.5">
          <div className="flex items-center gap-2.5">
            <LuSearch size={18} className="text-white/50" />
            <span className="text-sm font-semibold text-white">Search Subtitles</span>
          </div>
          <button
            onClick={onClose}
            className="rounded-lg p-1.5 text-white/40 transition-colors hover:bg-white/10 hover:text-white"
          >
            <LuX size={18} />
          </button>
        </div>

        {/* Language selector */}
        <div className="flex items-center gap-3 border-b border-white/10 px-5 py-3">
          <label className="text-xs font-medium text-white/50">Language</label>
          <select
            value={lang}
            onChange={(e) => setLang(e.target.value)}
            className="rounded-lg bg-white/8 px-3 py-1.5 text-sm text-white outline-none ring-1 ring-white/10 focus:ring-white/30"
          >
            {LANGUAGES.map((l) => (
              <option key={l.code} value={l.code}>
                {l.name}
              </option>
            ))}
          </select>
        </div>

        {/* Results */}
        <div className="flex-1 overflow-y-auto">
          {isLoading && (
            <div className="flex items-center justify-center py-12">
              <LuLoaderCircle size={24} className="animate-spin text-white/30" />
            </div>
          )}

          {!isLoading && results?.length === 0 && (
            <div className="py-12 text-center text-sm text-white/40">
              No subtitles found for this language.
            </div>
          )}

          {!isLoading &&
            results &&
            results.length > 0 &&
            results.map((result) => {
              const key = `${result.provider}:${result.external_id}`
              const isDownloaded = downloaded.has(key)
              const isDownloading = downloadingKey === key

              return (
                <ResultRow
                  key={key}
                  result={result}
                  isDownloaded={isDownloaded}
                  isDownloading={isDownloading}
                  onDownload={() => handleDownload(result)}
                />
              )
            })}
        </div>

        {/* Footer */}
        <div className="border-t border-white/10 px-5 py-2.5">
          <p className="text-[10px] text-white/30">
            {results?.length ?? 0} results from OpenSubtitles + Podnapisi
          </p>
        </div>
      </div>
    </div>
  )
}

function ResultRow({
  result,
  isDownloaded,
  isDownloading,
  onDownload,
}: {
  result: SubtitleSearchResult
  isDownloaded: boolean
  isDownloading: boolean
  onDownload: () => void
}) {
  const providerColor =
    result.provider === 'opensubtitles'
      ? 'bg-green-500/15 text-green-400'
      : 'bg-blue-500/15 text-blue-400'
  const providerLabel = result.provider === 'opensubtitles' ? 'OpenSub' : 'Podnapisi'

  return (
    <div className="flex items-start gap-3 border-b border-white/5 px-5 py-3 transition-colors hover:bg-white/5">
      <div className="flex-1 min-w-0">
        <p className="truncate text-sm text-white/80" title={result.title}>
          {result.title || 'Untitled'}
        </p>
        <div className="mt-1 flex flex-wrap items-center gap-2">
          <span className={`rounded px-1.5 py-0.5 text-[10px] font-medium ${providerColor}`}>
            {providerLabel}
          </span>
          {result.format && (
            <span className="text-[10px] uppercase text-white/30">{result.format}</span>
          )}
          {result.downloads > 0 && (
            <span className="text-[10px] text-white/30">
              {result.downloads > 1000
                ? `${(result.downloads / 1000).toFixed(0)}k`
                : result.downloads}{' '}
              DL
            </span>
          )}
          {result.hearing_impaired && (
            <span className="rounded bg-yellow-500/15 px-1.5 py-0.5 text-[10px] text-yellow-400">
              CC
            </span>
          )}
          {result.forced && (
            <span className="rounded bg-purple-500/15 px-1.5 py-0.5 text-[10px] text-purple-400">
              Forced
            </span>
          )}
          {result.ai_translated && (
            <span className="rounded bg-orange-500/15 px-1.5 py-0.5 text-[10px] text-orange-400">
              AI
            </span>
          )}
        </div>
      </div>

      <button
        onClick={onDownload}
        disabled={isDownloaded || isDownloading}
        className="shrink-0 rounded-lg p-2 text-white/40 transition-colors hover:bg-white/10 hover:text-white disabled:opacity-50 disabled:hover:bg-transparent"
      >
        {isDownloaded ? (
          <LuCheck size={16} className="text-green-400" />
        ) : isDownloading ? (
          <LuLoaderCircle size={16} className="animate-spin" />
        ) : (
          <LuDownload size={16} />
        )}
      </button>
    </div>
  )
}
