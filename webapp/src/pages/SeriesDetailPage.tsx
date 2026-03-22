import { useParams, Link } from 'react-router'
import { useEffect, useState } from 'react'
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
} from '@/hooks/stores/useMedia'
import { useAuthStore } from '@/stores/auth'
import { EpisodeCard } from '@/components/EpisodeCard'
import { LuChevronLeft, LuFilm, LuPlay, LuTv, LuPencil, LuLock } from 'react-icons/lu'
import { seriesImage } from '@/lib/image'
import { MetadataEditor } from '@/components/metadata/MetadataEditor'
import { EpisodeEditDialog } from '@/components/metadata/EpisodeEditDialog'
import { useSeriesTrailers } from '@/hooks/useCinemaMode'
import { YouTubeBackground } from '@/components/YouTubeBackground'
import { useTranslation } from '@/hooks/useTranslation'
import type { Episode } from '@/types/api'

export function SeriesDetailPage() {
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
  const [showEditor, setShowEditor] = useState(false)
  const { youtubeKey } = useSeriesTrailers(id)

  const [editingEpisode, setEditingEpisode] = useState<Episode | null>(null)
  const currentSeasonId = selectedSeasonId || seasons?.[0]?.id || 0
  const { mutate: editEpisode, isPending: isEpisodeSaving } = useEditEpisodeMetadata(
    id,
    currentSeasonId,
  )
  const { data: continueWatchingData } = useContinueWatching({ limit: 100 })
  const { data: nextUpData } = useNextUp({ limit: 100 })
  const { data: episodes, isLoading: episodesLoading } = useEpisodes(
    id,
    selectedSeasonId || seasons?.[0]?.id || 0,
  )

  useEffect(() => {
    if (!seasons?.length) return
    if (selectedSeasonId && seasons.some((season) => season.id === selectedSeasonId)) return
    setSelectedSeasonId(seasons[0].id)
  }, [selectedSeasonId, seasons])

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

  return (
    <div className="min-h-screen bg-netflix-black">
      {/* Backdrop — YouTube trailer or static image */}
      {(youtubeKey || series.backdrop_path) && (
        <div className="fixed inset-0 h-screen">
          {youtubeKey ? (
            <YouTubeBackground videoId={youtubeKey} muted className="absolute inset-0" />
          ) : (
            <img
              src={seriesImage(series.backdrop_path, 'w1280')!}
              alt={series.title}
              className="h-full w-full object-cover"
            />
          )}
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
              {series.poster_path ? (
                <img
                  src={seriesImage(series.poster_path, 'w500')!}
                  alt={series.title}
                  className="w-64 rounded-lg shadow-2xl lg:w-80"
                />
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

              {/* Play Button */}
              {playTargetMediaId && (
                <div className="mb-6 flex flex-wrap items-center gap-3">
                  <Link
                    to={`/watch/${playTargetMediaId}`}
                    className="flex items-center gap-2 rounded bg-netflix-red px-6 py-2.5 text-sm font-semibold text-white transition-colors hover:bg-red-700"
                  >
                    <LuPlay size={18} className="fill-current" />
                    {playLabel}
                  </Link>
                  {playSubtitle && <p className="text-sm text-gray-400">{playSubtitle}</p>}
                </div>
              )}
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
                        selectedSeasonId === season.id
                          ? 'bg-netflix-red text-white'
                          : 'bg-netflix-dark text-gray-300 hover:bg-netflix-gray'
                      }`}
                    >
                      {t('detail.season')} {season.season_number}
                    </button>
                  ))}
                </div>
                {selectedSeasonId && seasons.find((s) => s.id === selectedSeasonId)?.title && (
                  <p className="mt-2 text-sm text-gray-400">
                    {seasons.find((s) => s.id === selectedSeasonId)?.title}
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
