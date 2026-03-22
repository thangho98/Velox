import { memo, type RefObject } from 'react'
import { LuChevronLeft, LuChevronRight, LuInfo, LuListMusic, LuPlay } from 'react-icons/lu'
import { tmdbImage } from '@/lib/image'
import type { Episode, MediaWithFiles, Season } from '@/types/api'
import { DETAIL_PANEL_ANIMATION_MS, formatTime } from './watchHelpers'
import { useTranslation } from '@/hooks/useTranslation'

type DetailPanel = 'none' | 'info' | 'season'

interface WatchDetailSheetProps {
  activeTab: DetailPanel
  displayPanel: DetailPanel
  isEpisode: boolean
  isPanelVisible: boolean
  infoAudioChannels: string
  infoAudioCodec: string
  infoAudioLanguage: string
  infoCC: boolean
  infoEpisodeLabel: string | null
  infoLogoUrl: string | null
  infoResolution: string
  infoRuntime: string
  infoVideoCodec: string
  infoYear: number | null
  media: MediaWithFiles
  mediaId: number
  onEpisodeSelect: (episodeMediaId: number, isCurrentEpisode: boolean) => void
  onScrollSeasonCarousel: (direction: 'prev' | 'next') => void
  onSeasonSelect: (seasonId: number) => void
  onToggleDetailPanel: (panel: 'info' | 'season') => void
  seasonCarouselRef: RefObject<HTMLDivElement | null>
  seasonPanelEpisodes: Episode[]
  seasonPanelSeasonId: number
  seasons: Season[]
}

export const WatchDetailSheet = memo(function WatchDetailSheet({
  activeTab,
  displayPanel,
  isEpisode,
  isPanelVisible,
  infoAudioChannels,
  infoAudioCodec,
  infoAudioLanguage,
  infoCC,
  infoEpisodeLabel,
  infoLogoUrl,
  infoResolution,
  infoRuntime,
  infoVideoCodec,
  infoYear,
  media,
  mediaId,
  onEpisodeSelect,
  onScrollSeasonCarousel,
  onSeasonSelect,
  onToggleDetailPanel,
  seasonCarouselRef,
  seasonPanelEpisodes,
  seasonPanelSeasonId,
  seasons,
}: WatchDetailSheetProps) {
  const { t } = useTranslation('watch')
  return (
    <div
      className={`absolute inset-x-0 bottom-0 z-20 px-6 pb-4 will-change-transform transition-[opacity,transform,filter] duration-[520ms] ease-[cubic-bezier(0.22,1,0.36,1)] ${
        displayPanel === 'none'
          ? 'pointer-events-none translate-y-10 opacity-0 blur-[2px]'
          : 'translate-y-0 opacity-100 blur-0'
      }`}
      style={{ transitionDuration: `${DETAIL_PANEL_ANIMATION_MS}ms` }}
      aria-hidden={displayPanel === 'none'}
    >
      <div
        className={`relative left-1/2 w-[calc(100vw-2.5rem)] max-w-none -translate-x-1/2 pt-4 will-change-transform transition-[opacity,transform] duration-[560ms] ease-[cubic-bezier(0.22,1,0.36,1)] ${
          isPanelVisible ? 'translate-y-0 opacity-100' : 'translate-y-6 opacity-0'
        }`}
      >
        <div className="mb-4 flex w-full items-center gap-6">
          <button
            onClick={() => onToggleDetailPanel('info')}
            className={`flex items-center gap-1.5 text-sm font-semibold transition-colors ${
              activeTab === 'info' ? 'text-white' : 'text-white/45 hover:text-white/80'
            }`}
          >
            <LuInfo size={14} />
            {t('detail.info')}
          </button>
          {isEpisode && (
            <button
              onClick={() => onToggleDetailPanel('season')}
              className={`flex items-center gap-1.5 text-sm font-semibold transition-colors ${
                activeTab === 'season' ? 'text-white' : 'text-white/45 hover:text-white/80'
              }`}
            >
              <LuListMusic size={12} />
              {t('detail.season')}
            </button>
          )}
        </div>

        {displayPanel === 'info' && (
          <div className="w-full overflow-hidden rounded-[32px] border border-white/8 bg-[linear-gradient(135deg,rgba(255,255,255,0.12),rgba(255,255,255,0.03))] shadow-[0_24px_80px_rgba(0,0,0,0.34)] backdrop-blur-2xl">
            <div className="flex flex-col gap-5 px-6 py-6 md:flex-row md:items-start md:gap-7 md:px-8">
              <div className="hidden h-[126px] w-[224px] shrink-0 overflow-hidden rounded-[22px] border border-white/8 bg-black/45 shadow-[0_16px_30px_rgba(0,0,0,0.28)] md:block">
                {media.media.thumb_path ? (
                  <img
                    src={tmdbImage(media.media.thumb_path, 'w500')!}
                    alt={media.media.title}
                    className="h-full w-full object-cover"
                  />
                ) : media.media.backdrop_path ? (
                  <img
                    src={tmdbImage(media.media.backdrop_path, 'w780')!}
                    alt={media.media.title}
                    className="h-full w-full object-cover"
                  />
                ) : (
                  <div className="flex h-full w-full items-center justify-center bg-black/55">
                    <LuPlay size={26} className="text-white/20" />
                  </div>
                )}
              </div>

              <div className="min-w-0 flex-1">
                <div className="mb-4">
                  {infoLogoUrl ? (
                    <img
                      src={infoLogoUrl}
                      alt={media.media.title}
                      className="mb-3 h-10 max-w-[320px] object-contain object-left brightness-110 drop-shadow-[0_8px_20px_rgba(0,0,0,0.35)] md:h-14 md:max-w-[420px]"
                    />
                  ) : (
                    <h2 className="mb-2 text-2xl font-bold tracking-tight text-white md:text-4xl">
                      {media.media.title}
                    </h2>
                  )}

                  <div className="flex flex-wrap items-center gap-x-3 gap-y-2 text-[15px] text-white/88">
                    <span className="font-semibold text-white">{media.media.title}</span>
                    {infoEpisodeLabel ? (
                      <span className="rounded-full border border-white/10 bg-white/6 px-2.5 py-1 text-xs font-semibold uppercase tracking-[0.14em] text-white/78">
                        {infoEpisodeLabel}
                      </span>
                    ) : null}
                  </div>
                </div>

                <div className="mb-3 flex flex-wrap items-center gap-x-3 gap-y-2 text-[15px] text-white/78">
                  {infoYear ? <span>{infoYear}</span> : null}
                  {infoRuntime ? <span>{infoRuntime}</span> : null}
                  {infoCC ? (
                    <span className="rounded-md border border-white/18 px-1.5 py-0.5 text-xs font-bold tracking-[0.12em] text-white/82">
                      CC
                    </span>
                  ) : null}
                </div>

                <div className="mb-5 flex flex-wrap items-center gap-x-4 gap-y-2 text-[15px] text-white/88">
                  {infoResolution ? <span>{infoResolution}</span> : null}
                  {infoVideoCodec ? <span>{infoVideoCodec}</span> : null}
                  {infoAudioCodec ? (
                    <span>
                      {infoAudioLanguage} {infoAudioCodec} {infoAudioChannels}
                    </span>
                  ) : null}
                </div>

                {media.media.overview && (
                  <p className="max-w-3xl text-sm leading-7 text-white/66 md:text-[15px]">
                    {media.media.overview}
                  </p>
                )}
              </div>
            </div>
          </div>
        )}

        {displayPanel === 'season' && isEpisode && (
          <div className="w-full">
            <div className="mb-4 flex flex-wrap items-center gap-3">
              {seasons.map((season) => (
                <button
                  key={season.id}
                  onClick={() => onSeasonSelect(season.id)}
                  className={`rounded-full border px-4 py-2 text-sm font-semibold transition-[background-color,color,border-color,transform] duration-200 ${
                    seasonPanelSeasonId === season.id
                      ? 'border-white bg-white text-black shadow-[0_10px_30px_rgba(255,255,255,0.16)]'
                      : 'border-white/8 bg-white/8 text-white/70 hover:border-white/15 hover:bg-white/14 hover:text-white'
                  }`}
                >
                  {t('detail.season')} {season.season_number}
                </button>
              ))}
            </div>

            <div className="relative rounded-[30px] border border-white/8 bg-black/36 px-4 py-3 shadow-[0_24px_70px_rgba(0,0,0,0.35)] backdrop-blur-xl">
              {seasonPanelEpisodes.length > 1 && (
                <>
                  <button
                    type="button"
                    onClick={() => onScrollSeasonCarousel('prev')}
                    className="absolute left-3 top-1/2 z-20 flex h-11 w-11 -translate-y-1/2 items-center justify-center rounded-full border border-white/10 bg-black/62 text-white/78 backdrop-blur-md transition-all duration-200 hover:scale-105 hover:border-white/20 hover:bg-black/78 hover:text-white"
                    aria-label={t('detail.previousEpisodes')}
                  >
                    <LuChevronLeft size={18} />
                  </button>

                  <button
                    type="button"
                    onClick={() => onScrollSeasonCarousel('next')}
                    className="absolute right-3 top-1/2 z-20 flex h-11 w-11 -translate-y-1/2 items-center justify-center rounded-full border border-white/10 bg-black/62 text-white/78 backdrop-blur-md transition-all duration-200 hover:scale-105 hover:border-white/20 hover:bg-black/78 hover:text-white"
                    aria-label={t('detail.nextEpisodes')}
                  >
                    <LuChevronRight size={18} />
                  </button>
                </>
              )}

              <div
                ref={seasonCarouselRef}
                className="flex snap-x snap-mandatory gap-4 overflow-x-auto px-12 pb-2 pr-12 [scrollbar-width:none] [&::-webkit-scrollbar]:hidden"
              >
                {seasonPanelEpisodes.map((episode) => {
                  const isCurrentEpisode = episode.media_id === mediaId
                  const stillImage = tmdbImage(episode.still_path, 'w300')

                  return (
                    <button
                      key={episode.id}
                      onClick={() => onEpisodeSelect(episode.media_id, isCurrentEpisode)}
                      className={`group relative flex h-[190px] w-[240px] shrink-0 snap-start flex-col justify-end overflow-hidden rounded-[22px] border text-left transition-[transform,background-color,border-color,box-shadow] duration-200 ${
                        isCurrentEpisode
                          ? 'border-white/20 bg-white/[0.14] text-white shadow-[0_18px_45px_rgba(0,0,0,0.32)]'
                          : 'border-transparent bg-black/62 text-white/82 hover:-translate-y-1 hover:border-white/10 hover:text-white'
                      }`}
                    >
                      <div className="absolute inset-0">
                        {stillImage ? (
                          <img
                            src={stillImage}
                            alt={episode.title}
                            className="h-full w-full object-cover transition-transform duration-500 group-hover:scale-[1.04]"
                          />
                        ) : (
                          <div className="flex h-full w-full items-center justify-center bg-black/55">
                            <LuPlay size={24} className="text-white/25" />
                          </div>
                        )}
                      </div>

                      <div className="absolute inset-0 bg-gradient-to-t from-black via-black/45 to-black/10" />

                      <div className="absolute left-3 top-3 flex items-center gap-2">
                        <span className="rounded-full bg-black/55 px-2 py-1 text-[10px] font-bold uppercase tracking-[0.16em] text-white/72 backdrop-blur-sm">
                          E{episode.episode_number}
                        </span>
                        {isCurrentEpisode && (
                          <span className="rounded-full bg-netflix-red px-2.5 py-1 text-[10px] font-black uppercase tracking-[0.12em] text-white shadow-[0_8px_18px_rgba(229,9,20,0.35)]">
                            {t('detail.nowPlaying')}
                          </span>
                        )}
                      </div>

                      <div className="relative z-10 flex min-h-[74px] flex-col justify-end px-4 pb-4">
                        <p className="line-clamp-2 text-[0.95rem] font-semibold tracking-tight text-white drop-shadow">
                          {episode.title}
                        </p>
                        <div className="mt-1.5 flex items-center gap-2 text-[11px] font-medium text-white/65">
                          <span>{t('detail.episode', { number: episode.episode_number })}</span>
                          {episode.duration ? <span>{formatTime(episode.duration)}</span> : null}
                        </div>
                      </div>
                    </button>
                  )
                })}

                {seasonPanelEpisodes.length === 0 && (
                  <div className="w-full rounded-[24px] border border-white/8 bg-black/55 px-5 py-6 text-sm text-white/48">
                    {t('detail.noEpisodesFound')}
                  </div>
                )}
              </div>
            </div>
          </div>
        )}
      </div>
    </div>
  )
})
