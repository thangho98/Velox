import { useParams, Link } from 'react-router'
import { useState } from 'react'
import {
  useSeriesDetail,
  useSeasons,
  useEpisodes,
  useContinueWatching,
  useNextUp,
  useEditSeriesMetadata,
  useUploadSeriesImage,
  useSeriesGenres,
  useSeriesCredits,
  useEditEpisodeMetadata,
  useDownloadSeriesToNas,
  useRemoveSeriesDownloadFromNas,
} from '@/hooks/stores/useMedia'
import { useToast } from '@/components/Toast'
import { useAuthStore } from '@/stores/auth'
import { EpisodeCard } from '@/components/EpisodeCard'
import { ResponsiveImage } from '@/components/ResponsiveImage'
import { MetadataEditor } from '@/components/metadata/MetadataEditor'
import { EpisodeEditDialog } from '@/components/metadata/EpisodeEditDialog'
import { useSeriesTrailers } from '@/hooks/useCinemaMode'
import { YouTubeBackground } from '@/components/YouTubeBackground'
import { useTranslation } from '@/hooks/useTranslation'
import type { Episode } from '@/types/api'
import {
  LuChevronLeft,
  LuFilm,
  LuTv,
  LuLock,
  LuPencil,
  LuPlay,
  LuDownload,
  LuLoaderCircle,
  LuTrash,
} from 'react-icons/lu'

export default function SeriesDetailPage() {
  const { seriesId } = useParams<{ seriesId: string }>()
  const id = Number(seriesId)
  const [selectedSeasonId, setSelectedSeasonId] = useState<number | null>(null)
  const { t } = useTranslation('media')
  const { t: tCommon } = useTranslation('common')

  const { data: series, isLoading: seriesLoading } = useSeriesDetail(id)
  const { data: seasons, isLoading: seasonsLoading } = useSeasons(id)
  const { mutate: editMetadata, isPending: isSaving } = useEditSeriesMetadata(id)
  const { mutate: uploadImage, isPending: isUploadingImage } = useUploadSeriesImage(id)
  const { data: seriesGenres = [] } = useSeriesGenres(id)
  const { data: seriesCredits = [] } = useSeriesCredits(id)
  const { user } = useAuthStore()
  const { success: showToastSuccess, error: showToastError } = useToast()
  const { mutate: downloadSeriesToNas, isPending: isDownloadingToNas } = useDownloadSeriesToNas(id)
  const { mutate: removeSeriesDownload, isPending: isRemovingSeriesDownload } =
    useRemoveSeriesDownloadFromNas(id)
  const [showEditor, setShowEditor] = useState(false)
  const { youtubeKey } = useSeriesTrailers(id)

  const [editingEpisode, setEditingEpisode] = useState<Episode | null>(null)

  const currentSeasonId =
    selectedSeasonId && seasons?.some((s) => s.id === selectedSeasonId)
      ? selectedSeasonId
      : seasons?.[0]?.id || 0

  const { mutate: editEpisode, isPending: isEpisodeSaving } = useEditEpisodeMetadata(
    id,
    currentSeasonId,
  )
  const { data: continueWatchingData } = useContinueWatching({ limit: 100 })
  const { data: nextUpData } = useNextUp({ limit: 100 })
  const { data: episodes, isLoading: episodesLoading } = useEpisodes(id, currentSeasonId)

  if (seriesLoading) {
    return (
      <div className="flex h-screen items-center justify-center">
        <div className="h-8 w-8 animate-spin rounded-full border-2 border-netflix-red border-t-transparent" />
      </div>
    )
  }

  if (!series) {
    return (
      <div className="flex h-screen flex-col items-center justify-center">
        <h1 className="mb-4 text-4xl font-bold text-white">404</h1>
        <p className="mb-8 text-xl text-gray-400">{t('detail.seriesNotFound')}</p>
        <Link to="/" className="text-netflix-blue hover:underline">
          {tCommon('states.goBackHome')}
        </Link>
      </div>
    )
  }

  const continueWatching = continueWatchingData ?? []
  const nextUp = nextUpData ?? []

  const resumeItem = continueWatching.find((item) => item.series_id === id)
  const nextUpItem = nextUp.find((item) => item.series_id === id)
  const playTargetMediaId = resumeItem?.media_id ?? nextUpItem?.media_id ?? episodes?.[0]?.media_id
  const playLabel = resumeItem
    ? t('actions.resume')
    : nextUpItem
      ? t('actions.playEpisode', {
          season: nextUpItem.season_number,
          episode: nextUpItem.episode_number,
        })
      : t('actions.playFirstEpisode')
  const playSubtitle = resumeItem
    ? t('actions.continue', { title: resumeItem.title })
    : nextUpItem
      ? nextUpItem.episode_title
      : null
  const seriesYear = series.first_air_date ? new Date(series.first_air_date).getFullYear() : null

  const hasCloudMedia = episodes?.some((ep) =>
    ep.media_files?.some(
      (f) =>
        !f.file_path.includes('library/downloads') &&
        (f.file_path.includes('://') || !f.file_path.startsWith('/')),
    ),
  )

  const hasDownloadedFiles = episodes?.some((ep) =>
    ep.media_files?.some(
      (f) => f.file_path.includes('library/downloads') && f.file_path.startsWith('/'),
    ),
  )

  return (
    <div className="min-h-screen bg-netflix-black">
      {/* Backdrop — YouTube trailer or static image */}
      {(youtubeKey || series.backdrop) && (
        <div className="fixed inset-0 h-screen">
          {youtubeKey ? (
            <YouTubeBackground videoId={youtubeKey} muted className="absolute inset-0" />
          ) : series.backdrop ? (
            <ResponsiveImage
              data={series.backdrop}
              sizes="100vw"
              alt={series.title}
              className="h-full w-full"
              loading="eager"
            />
          ) : null}
          <div className="absolute inset-0 bg-gradient-to-t from-netflix-black via-netflix-black/80 to-netflix-black/30" />
          <div className="absolute inset-0 bg-gradient-to-r from-netflix-black via-netflix-black/50 to-transparent" />
        </div>
      )}

      {/* Content */}
      <div className="relative z-10 min-h-screen">
        {/* Back button */}
        <Link
          to="/"
          className="fixed left-4 top-20 z-20 flex items-center gap-2 rounded-full bg-black/50 p-3 text-white backdrop-blur-sm transition-colors hover:bg-black/70"
        >
          <LuChevronLeft size={20} />
        </Link>

        <div className="container mx-auto px-4 py-24 lg:px-8">
          <div className="flex flex-col gap-8 lg:flex-row">
            {/* Poster */}
            <div className="mx-auto flex-shrink-0 lg:mx-0">
              {series.poster ? (
                <div className="w-64 lg:w-80 rounded-lg shadow-2xl overflow-hidden shrink-0">
                  <ResponsiveImage
                    data={series.poster}
                    sizes="(max-width: 768px) 342px, 500px"
                    alt={series.title}
                    loading="eager"
                  />
                </div>
              ) : (
                <div className="flex h-96 w-64 items-center justify-center rounded-lg bg-netflix-dark lg:w-80">
                  <LuFilm size={64} className="text-gray-600" />
                </div>
              )}
            </div>

            {/* Info */}
            <div className="flex-1">
              <h1 className="mb-2 text-3xl font-bold text-white lg:text-5xl">{series.title}</h1>

              {/* Year · Status · Network */}
              <div className="mb-4 flex flex-wrap items-center gap-x-3 gap-y-1 text-sm text-gray-400">
                {seriesYear && <span>{seriesYear}</span>}
                {series.status && (
                  <span className="rounded bg-purple-500/20 px-1.5 py-0.5 text-xs text-purple-400">
                    {series.status}
                  </span>
                )}
                {series.network && (
                  <span className="flex items-center gap-1">
                    <LuTv size={14} />
                    {series.network}
                  </span>
                )}
                {user?.is_admin && series.metadata_locked && (
                  <span
                    className="flex items-center gap-1 rounded-full bg-amber-600/20 px-2 py-0.5 text-xs text-amber-400"
                    title={t('detail.metadataLocked')}
                  >
                    <LuLock size={12} /> {t('actions.editMetadata')}
                  </span>
                )}
                {user?.is_admin && (
                  <button
                    onClick={() => setShowEditor(true)}
                    className="flex items-center gap-1 rounded-full bg-white/10 px-3 py-0.5 text-xs text-gray-300 hover:bg-white/20"
                    title={t('actions.editMetadata')}
                  >
                    <LuPencil size={12} /> {t('actions.edit')}
                  </button>
                )}
              </div>

              {series.overview && (
                <p className="mb-6 max-w-2xl text-base leading-relaxed text-gray-300">
                  {series.overview}
                </p>
              )}

              {/* Play Button and Download to NAS */}
              <div className="mb-6 flex flex-col gap-2">
                <div className="flex flex-wrap items-center gap-3">
                  {playTargetMediaId && (
                    <Link
                      to={`/watch/${playTargetMediaId}`}
                      className="flex items-center gap-2 rounded bg-netflix-red px-8 py-2.5 text-sm font-semibold text-white transition-colors hover:bg-red-700"
                    >
                      <LuPlay size={20} className="fill-current" />
                      {playLabel}
                    </Link>
                  )}

                  {/* Download / Delete Series from NAS */}
                  {user?.is_admin && (
                    <>
                      {hasCloudMedia && (
                        <button
                          onClick={() => {
                            downloadSeriesToNas(undefined, {
                              onSuccess: () => {
                                showToastSuccess(
                                  t(
                                    'detail.downloadSeriesInitiated',
                                    'Đã thêm Series vào hàng đợi tải xuống NAS',
                                  ),
                                )
                              },
                              onError: (error) => {
                                showToastError(
                                  t('detail.downloadError', 'Tải xuống thất bại'),
                                  error instanceof Error ? error.message : 'Unknown error',
                                )
                              },
                            })
                          }}
                          disabled={isDownloadingToNas}
                          className="flex items-center gap-2 rounded bg-[#333]/80 px-6 py-2.5 text-sm font-semibold text-white transition-colors hover:bg-[#444] disabled:opacity-50"
                        >
                          {isDownloadingToNas ? (
                            <LuLoaderCircle size={20} className="animate-spin" />
                          ) : (
                            <LuDownload size={20} />
                          )}
                          {isDownloadingToNas
                            ? t('actions.downloading', 'Đang tải...')
                            : t('actions.downloadSeriesToNas', 'Tải Series về NAS')}
                        </button>
                      )}

                      {hasDownloadedFiles && (
                        <button
                          onClick={() => {
                            if (
                              window.confirm(
                                t(
                                  'actions.confirmDeleteSeries',
                                  'Bạn có chắc muốn xoá toàn bộ bản tải về NAS của Series này không?',
                                ),
                              )
                            ) {
                              removeSeriesDownload(undefined, {
                                onSuccess: () =>
                                  showToastSuccess(
                                    t('actions.deleteSeriesSuccess', 'Đã xoá toàn bộ tập tải về!'),
                                  ),
                                onError: (err) =>
                                  showToastError(
                                    t('actions.deleteSeriesFailed', 'Xoá series thất bại.'),
                                    err instanceof Error ? err.message : '',
                                  ),
                              })
                            }
                          }}
                          disabled={isRemovingSeriesDownload}
                          className="flex items-center gap-2 rounded bg-red-900/30 px-6 py-2.5 text-sm font-semibold text-red-400 transition-colors hover:bg-red-900/60 disabled:opacity-50"
                        >
                          {isRemovingSeriesDownload ? (
                            <LuLoaderCircle size={20} className="animate-spin" />
                          ) : (
                            <LuTrash size={20} />
                          )}
                          {isRemovingSeriesDownload
                            ? t('actions.deleting', 'Đang xoá...')
                            : t('actions.deleteSeries', 'Xoá Series trên NAS')}
                        </button>
                      )}
                    </>
                  )}
                </div>
                {playTargetMediaId && playSubtitle && (
                  <p className="text-sm font-medium text-gray-400">{playSubtitle}</p>
                )}
              </div>
            </div>
          </div>

          {/* Seasons and Episodes */}
          <div className="mt-12">
            <h2 className="mb-6 text-2xl font-bold text-white">{t('detail.episodes')}</h2>

            {/* Season Selector */}
            {seasonsLoading ? (
              <div className="flex h-16 items-center justify-center">
                <div className="h-6 w-6 animate-spin rounded-full border-2 border-netflix-red border-t-transparent" />
              </div>
            ) : seasons && seasons.length > 0 ? (
              <div className="mb-6">
                <div className="flex flex-wrap gap-2">
                  {seasons.map((season) => (
                    <button
                      key={season.id}
                      onClick={() => setSelectedSeasonId(season.id)}
                      className={`rounded-lg px-4 py-2 text-sm font-medium transition-colors ${
                        currentSeasonId === season.id
                          ? 'bg-netflix-red text-white'
                          : 'bg-netflix-dark text-gray-300 hover:bg-netflix-gray'
                      }`}
                    >
                      {t('detail.season')} {season.season_number}
                    </button>
                  ))}
                </div>
                {currentSeasonId && seasons.find((s) => s.id === currentSeasonId)?.title && (
                  <p className="mt-2 text-sm text-gray-400">
                    {seasons.find((s) => s.id === currentSeasonId)?.title}
                  </p>
                )}
              </div>
            ) : null}

            {/* Episodes List */}
            {episodesLoading ? (
              <div className="flex h-32 items-center justify-center">
                <div className="h-8 w-8 animate-spin rounded-full border-2 border-netflix-red border-t-transparent" />
              </div>
            ) : episodes && episodes.length > 0 ? (
              <div className="space-y-3">
                {episodes.map((episode) => (
                  <EpisodeCard
                    key={episode.id}
                    episode={episode}
                    isAdmin={user?.is_admin}
                    onEdit={(ep) => setEditingEpisode(ep)}
                  />
                ))}
              </div>
            ) : (
              <div className="flex h-32 flex-col items-center justify-center rounded-lg bg-netflix-dark">
                <p className="text-gray-400">{t('detail.noEpisodes')}</p>
              </div>
            )}
          </div>
        </div>
      </div>

      {/* Episode Edit Dialog */}
      {editingEpisode && (
        <EpisodeEditDialog
          episode={editingEpisode}
          isSaving={isEpisodeSaving}
          onSave={(req) => {
            editEpisode(
              { episodeId: editingEpisode.id, req },
              { onSuccess: () => setEditingEpisode(null) },
            )
          }}
          onClose={() => setEditingEpisode(null)}
        />
      )}

      {/* Metadata Editor Panel */}
      {showEditor && series && (
        <MetadataEditor
          type="series"
          series={series}
          genres={seriesGenres.map((g) => g.name)}
          credits={seriesCredits}
          onSave={(req) => {
            editMetadata(req, {
              onSuccess: () => setShowEditor(false),
            })
          }}
          onUploadImage={(imageType, file) => {
            uploadImage({ imageType, file })
          }}
          isSaving={isSaving}
          isUploadingImage={isUploadingImage}
          onClose={() => setShowEditor(false)}
        />
      )}
    </div>
  )
}
