import { useState } from 'react'
import { useMarkerStats, useBackfillMarkers } from '@/hooks/stores/useAdmin'
import { useLibraries } from '@/hooks/stores/useMedia'
import { useMarkerProgress } from '@/hooks/useNotifications'
import { useTranslation } from '@/hooks/useTranslation'
import { Select } from '@/components/ui/Select'
import { SectionHeader, Field } from './shared'

export function MarkersSection() {
  const { t } = useTranslation('settings')
  const { data: stats, isLoading } = useMarkerStats()
  const { data: libraries } = useLibraries()
  const backfill = useBackfillMarkers()
  const progress = useMarkerProgress()
  const [selectedLibrary, setSelectedLibrary] = useState(0)

  // Auto-select first library
  if (selectedLibrary === 0 && libraries && libraries.length > 0) {
    setSelectedLibrary(libraries[0].id)
  }

  const isRunning = progress?.status === 'running'
  const progressPercent =
    progress && progress.total && progress.total > 0
      ? Math.round(((progress.current ?? 0) / progress.total) * 100)
      : 0

  const introCoverage =
    stats && stats.total_files > 0
      ? Math.round((stats.files_with_intro / stats.total_files) * 100)
      : 0
  const creditsCoverage =
    stats && stats.total_files > 0
      ? Math.round((stats.files_with_credits / stats.total_files) * 100)
      : 0

  // Extract just the filename from full path
  const currentFileName = progress?.file_name?.split('/').pop() ?? ''

  return (
    <div className="space-y-8">
      <SectionHeader
        title={t('sections.markers.title')}
        description={t('sections.markers.description')}
      />

      {/* Stats Overview */}
      <div className="rounded-lg bg-netflix-black p-6 ring-1 ring-white/10">
        <h3 className="mb-4 text-lg font-semibold text-white">{t('markers.overview')}</h3>
        {isLoading ? (
          <div className="text-sm text-gray-400">Loading...</div>
        ) : stats && stats.total_markers > 0 ? (
          <>
            <div className="grid grid-cols-2 gap-4 sm:grid-cols-4">
              <div className="rounded-lg bg-netflix-gray p-3">
                <div className="text-2xl font-bold text-white">{stats.total_markers}</div>
                <div className="text-xs text-gray-400">{t('markers.totalMarkers')}</div>
              </div>
              <div className="rounded-lg bg-netflix-gray p-3">
                <div className="text-2xl font-bold text-blue-400">{stats.intro_markers}</div>
                <div className="text-xs text-gray-400">{t('markers.introMarkers')}</div>
              </div>
              <div className="rounded-lg bg-netflix-gray p-3">
                <div className="text-2xl font-bold text-purple-400">{stats.credits_markers}</div>
                <div className="text-xs text-gray-400">{t('markers.creditsMarkers')}</div>
              </div>
              <div className="rounded-lg bg-netflix-gray p-3">
                <div className="text-2xl font-bold text-gray-300">{stats.total_files}</div>
                <div className="text-xs text-gray-400">{t('markers.totalFiles')}</div>
              </div>
            </div>

            {/* Coverage bars */}
            <div className="mt-4 space-y-3">
              <div>
                <div className="mb-1 flex items-center justify-between text-sm">
                  <span className="text-gray-400">
                    {t('markers.filesWithIntro')} ({stats.files_with_intro}/{stats.total_files})
                  </span>
                  <span className="text-white font-medium">{introCoverage}%</span>
                </div>
                <div className="h-2 w-full overflow-hidden rounded-full bg-gray-700">
                  <div
                    className="h-full rounded-full bg-blue-500 transition-all duration-500"
                    style={{ width: `${introCoverage}%` }}
                  />
                </div>
              </div>
              <div>
                <div className="mb-1 flex items-center justify-between text-sm">
                  <span className="text-gray-400">
                    {t('markers.filesWithCredits')} ({stats.files_with_credits}/{stats.total_files})
                  </span>
                  <span className="text-white font-medium">{creditsCoverage}%</span>
                </div>
                <div className="h-2 w-full overflow-hidden rounded-full bg-gray-700">
                  <div
                    className="h-full rounded-full bg-purple-500 transition-all duration-500"
                    style={{ width: `${creditsCoverage}%` }}
                  />
                </div>
              </div>
            </div>

            {/* Source breakdown */}
            <div className="mt-4 border-t border-white/10 pt-4">
              <h4 className="mb-2 text-sm font-medium text-gray-300">
                {t('markers.sourceBreakdown')}
              </h4>
              <div className="flex gap-6 text-sm">
                {stats.chapter_source > 0 && (
                  <span className="text-green-400">
                    {t('markers.chapter')}: {stats.chapter_source}
                  </span>
                )}
                {stats.fingerprint_source > 0 && (
                  <span className="text-blue-400">
                    {t('markers.fingerprint')}: {stats.fingerprint_source}
                  </span>
                )}
                {stats.manual_source > 0 && (
                  <span className="text-yellow-400">
                    {t('markers.manual')}: {stats.manual_source}
                  </span>
                )}
              </div>
            </div>
          </>
        ) : (
          <p className="text-sm text-gray-500">{t('markers.noMarkers')}</p>
        )}
      </div>

      {/* Run Detection */}
      <div className="rounded-lg bg-netflix-black p-6 ring-1 ring-white/10">
        <h3 className="mb-2 text-lg font-semibold text-white">{t('markers.detection')}</h3>
        <p className="mb-4 text-sm text-gray-400">{t('markers.detectionDesc')}</p>

        <div className="flex flex-wrap items-end gap-4">
          <Field label={t('markers.selectLibrary')}>
            <Select
              className="w-full"
              value={selectedLibrary}
              onChange={(e) => setSelectedLibrary(Number(e.target.value))}
              disabled={isRunning}
            >
              {libraries?.map((lib) => (
                <option key={lib.id} value={lib.id}>
                  {lib.name}
                </option>
              ))}
            </Select>
          </Field>
          <button
            onClick={() => backfill.mutate({ library_id: selectedLibrary })}
            disabled={isRunning || backfill.isPending || selectedLibrary === 0}
            className="rounded-lg bg-netflix-red px-6 py-2.5 text-sm font-semibold text-white transition-colors hover:bg-netflix-red-hover disabled:opacity-50"
          >
            {isRunning ? t('markers.running') : t('markers.runDetection')}
          </button>
        </div>

        {/* Real-time progress */}
        {progress && progress.status === 'running' && (
          <div className="mt-4 space-y-3">
            <div>
              <div className="mb-1 flex items-center justify-between text-sm text-gray-400">
                <span>
                  {progress.current}/{progress.total} files
                </span>
                <span>{progressPercent}%</span>
              </div>
              <div className="h-3 w-full overflow-hidden rounded-full bg-gray-700">
                <div
                  className="h-full rounded-full bg-netflix-red transition-all duration-300"
                  style={{ width: `${progressPercent}%` }}
                />
              </div>
            </div>
            {currentFileName && (
              <div className="truncate text-sm text-gray-500">Analyzing: {currentFileName}</div>
            )}
            <div className="flex gap-6 text-sm">
              <span className="text-green-400">Processed: {progress.processed}</span>
              <span className="text-gray-400">Skipped: {progress.skipped}</span>
              {(progress.failed ?? 0) > 0 && (
                <span className="text-red-400">Failed: {progress.failed}</span>
              )}
            </div>
          </div>
        )}

        {/* Complete result */}
        {progress && progress.status === 'complete' && (
          <div className="mt-4 rounded-lg bg-netflix-gray p-4 text-sm">
            <div className="text-green-400">
              {t('markers.resultProcessed', { processed: progress.processed })}
            </div>
            {(progress.skipped ?? 0) > 0 && (
              <div className="text-gray-400">
                {t('markers.resultSkipped', { skipped: progress.skipped })}
              </div>
            )}
            {(progress.failed ?? 0) > 0 && (
              <div className="text-red-400">Failed: {progress.failed} files</div>
            )}
          </div>
        )}

        {progress && progress.status === 'error' && (
          <div className="mt-4 rounded-lg bg-red-900/30 p-4 text-sm text-red-400">
            Detection failed. Check server logs for details.
          </div>
        )}
      </div>

      {/* How it works */}
      <div className="rounded-lg bg-netflix-black p-6 ring-1 ring-white/10">
        <h3 className="mb-3 text-lg font-semibold text-white">{t('markers.howItWorks')}</h3>
        <ul className="space-y-2 text-sm text-gray-400">
          <li className="flex gap-2">
            <span className="mt-0.5 text-green-400">●</span>
            {t('markers.howChapter')}
          </li>
          <li className="flex gap-2">
            <span className="mt-0.5 text-blue-400">●</span>
            {t('markers.howFingerprint')}
          </li>
          <li className="flex gap-2">
            <span className="mt-0.5 text-yellow-400">●</span>
            {t('markers.howManual')}
          </li>
        </ul>
      </div>
    </div>
  )
}
