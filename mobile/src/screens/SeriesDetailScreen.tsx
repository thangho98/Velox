import { useState, useMemo, useRef, useEffect, useLayoutEffect } from 'react'
import {
  View,
  ScrollView,
  Image,
  Text,
  TouchableOpacity,
  StyleSheet,
  ActivityIndicator,
  Modal,
  Share,
  Alert,
  StatusBar,
  Platform,
  Pressable,
  Dimensions,
} from 'react-native'
import { LinearGradient } from 'expo-linear-gradient'
import { Tv, Play, Film, Check, Link, Pencil, Lock, ChevronLeft, MoreVertical } from 'lucide-react-native'
import { useNavigation, useRoute, RouteProp } from '@react-navigation/native'
import type { NativeStackNavigationProp } from '@react-navigation/native-stack'
import { useSeriesDetail, useSeasons, useEpisodes, useContinueWatching, useNextUp } from '@velox/shared/hooks'
import {
  useEditSeriesMetadata,
  useEditEpisodeMetadata,
  useRefreshMetadata as useRefreshSeriesMetadata,
} from '@velox/shared/hooks/media/useMetadataOps'
import { useSeriesTrailers } from '@velox/shared/hooks/media/useCinema'
import { useEpisodesProgress } from '@velox/shared/hooks/media/useEpisodesProgress'
import { seriesImage, mediaImage } from '@velox/shared/lib'
import { getServerUrl } from '../platform/mobile-adapter'
import { MetadataEditor } from '../components/MetadataEditor'
import { EpisodeEditDialog } from '../components/EpisodeEditDialog'
import { YouTubePlayer } from '../components/YouTubePlayer'
import { Toast } from '../components/Toast'
import { useResponsiveLayout, scaledFont } from '../lib/responsive'
import type { RootStackParamList } from '../../App'
import type {
  Episode,
  SeriesMetadataEditRequest,
  EpisodeMetadataEditRequest,
} from '@velox/shared/types'

type NavigationProp = NativeStackNavigationProp<RootStackParamList>
type SeriesRouteProp = RouteProp<RootStackParamList, 'SeriesDetail'>

const { width: SCREEN_WIDTH } = Dimensions.get('window')

export function SeriesDetailScreen() {
  const navigation = useNavigation<NavigationProp>()
  const route = useRoute<SeriesRouteProp>()
  const { id } = route.params
  const layout = useResponsiveLayout()
  const { data: series, isLoading: loadingSeries } = useSeriesDetail(id)
  const { data: seasons, isLoading: loadingSeasons } = useSeasons(id)
  const { youtubeKey } = useSeriesTrailers(id)
  const { data: continueWatchingData } = useContinueWatching({ limit: 100 })
  const { data: nextUpData } = useNextUp({ limit: 100 })
  const [selectedSeasonId, setSelectedSeasonId] = useState<number | null>(null)
  const [showFullOverview, setShowFullOverview] = useState(false)
  const [showMenu, setShowMenu] = useState(false)
  const [showMetadataEditor, setShowMetadataEditor] = useState(false)
  const [showTrailer, setShowTrailer] = useState(false)

  // Put menu button in navigation header
  useLayoutEffect(() => {
    navigation.setOptions({
      headerRight: () => (
        <TouchableOpacity
          onPress={() => setShowMenu(true)}
          hitSlop={{ top: 10, bottom: 10, left: 10, right: 10 }}
          style={{ padding: 4 }}
        >
          <MoreVertical size={22} color="#fff" />
        </TouchableOpacity>
      ),
    })
  }, [navigation])
  const [episodeToEdit, setEpisodeToEdit] = useState<Episode | null>(null)
  const seasonScrollRef = useRef<ScrollView>(null)

  // Metadata editing hooks
  const editMetadata = useEditSeriesMetadata(id)
  const refreshMetadata = useRefreshSeriesMetadata(id)
  const editEpisodeMetadata = useEditEpisodeMetadata(id, selectedSeasonId ?? 0)

  // Get episodes for selected season
  const { data: episodes, isLoading: loadingEpisodes } = useEpisodes(id, selectedSeasonId ?? 0)

  // Fetch progress for all episodes
  const episodeMediaIds = useMemo(() => episodes?.map((ep) => ep.media_id).filter(Boolean) ?? [], [episodes])
  const { data: episodesProgress } = useEpisodesProgress(episodeMediaIds)

  // Create a map of mediaId -> progress for quick lookup
  const progressMap = useMemo(() => {
    const map = new Map<number, NonNullable<typeof episodesProgress>[number]>()
    if (episodesProgress) {
      episodesProgress.forEach((progress, index) => {
        if (progress && episodeMediaIds[index]) {
          map.set(episodeMediaIds[index], progress)
        }
      })
    }
    return map
  }, [episodesProgress, episodeMediaIds])

  const sortedSeasons = useMemo(() => {
    return [...(seasons || [])].sort((a, b) => a.season_number - b.season_number)
  }, [seasons])

  // Auto-select first season when seasons load
  useEffect(() => {
    if (sortedSeasons.length > 0 && !selectedSeasonId) {
      setSelectedSeasonId(sortedSeasons[0].id)
    }
  }, [sortedSeasons, selectedSeasonId])

  const posterWidth = layout.sideBySideDetail ? 220 : 200
  const posterHeight = layout.sideBySideDetail ? 330 : 300
  const statusBarHeight = Platform.OS === 'ios' ? (StatusBar.currentHeight ?? 44) : (StatusBar.currentHeight ?? 24)

  const isLoading = loadingSeries || loadingSeasons

  if (isLoading) {
    return (
      <View style={styles.loading}>
        <ActivityIndicator size="large" color="#e50914" />
      </View>
    )
  }

  if (!series) {
    return (
      <View style={styles.loading}>
        <Text style={styles.errorText}>Series not found</Text>
      </View>
    )
  }

  const backdropUrl = series.backdrop_path ? seriesImage(series.backdrop_path, 'w1280') : null
  const posterUrl = series.poster_path ? seriesImage(series.poster_path, 'w500') : null

  const currentSeason = selectedSeasonId
    ? sortedSeasons.find((s) => s.id === selectedSeasonId)
    : sortedSeasons[0]

  // Continue watching / next up logic (matching web)
  const continueWatching = continueWatchingData ?? []
  const nextUp = nextUpData ?? []
  const resumeItem = continueWatching.find((item) => item.series_id === id)
  const nextUpItem = nextUp.find((item) => item.series_id === id)
  const playTargetMediaId = resumeItem?.media_id ?? nextUpItem?.media_id ?? episodes?.[0]?.media_id
  const isResume = !!resumeItem
  const playLabel = resumeItem ? 'Resume' : 'Play'
  const playSubtitle = resumeItem
    ? `Continue ${series.title} - ${resumeItem.title}`
    : nextUpItem
      ? nextUpItem.episode_title
      : null

  const handlePlay = () => {
    if (playTargetMediaId) {
      navigation.navigate('Episode', { id: playTargetMediaId, seriesId: id })
      return
    }
    if (episodes && episodes.length > 0) {
      navigation.navigate('Episode', {
        id: episodes[0].media_id,
        seriesId: id,
      })
    }
  }

  const handleEpisodePress = (episode: Episode) => {
    navigation.navigate('Episode', {
      id: episode.media_id,
      seriesId: id,
    })
  }

  const handleCopyStreamUrl = async () => {
    setShowMenu(false)
    try {
      const serverUrl = getServerUrl() ?? 'http://localhost:8080'
      await Share.share({
        message: `Stream URL for ${series.title}: ${serverUrl}/play/series/${id}`,
      })
    } catch {
      Alert.alert('Error', 'Could not share stream URL')
    }
  }

  const handleRefreshMetadata = () => {
    setShowMenu(false)
    refreshMetadata.mutate(undefined as any, {
      onSuccess: () => {
        Toast.info('Metadata refreshed successfully')
      },
      onError: () => {
        Toast.error('Failed to refresh metadata')
      },
    })
  }

  const handleEditMetadata = () => {
    setShowMenu(false)
    setShowMetadataEditor(true)
  }

  const handleSaveMetadata = (req: SeriesMetadataEditRequest) => {
    editMetadata.mutate(req, {
      onSuccess: () => {
        Toast.success('Metadata saved successfully')
        setShowMetadataEditor(false)
      },
      onError: () => {
        Toast.error('Failed to save metadata')
      },
    })
  }

  const handleSaveEpisodeMetadata = (
    episodeId: number,
    req: EpisodeMetadataEditRequest,
  ) => {
    editEpisodeMetadata.mutate(
      { episodeId, req },
      {
        onSuccess: () => {
          Toast.success('Episode metadata saved')
          setEpisodeToEdit(null)
        },
        onError: () => {
          Toast.error('Failed to save episode metadata')
        },
      },
    )
  }

  return (
    <View style={styles.container}>
      <StatusBar translucent backgroundColor="transparent" barStyle="light-content" />

      {/* ── Fixed backdrop (doesn't scroll) ── */}
      {backdropUrl ? (
        <Image source={{ uri: backdropUrl }} style={styles.fixedBackdrop} resizeMode="cover" />
      ) : (
        <View style={[styles.fixedBackdrop, styles.backdropPlaceholder]} />
      )}
      <LinearGradient
        colors={['transparent', 'rgba(20,20,20,0.3)', 'rgba(20,20,20,0.8)', '#141414']}
        locations={[0, 0.3, 0.65, 1]}
        style={styles.fixedBackdrop}
      />
      <LinearGradient
        colors={['rgba(20,20,20,0.7)', 'rgba(20,20,20,0.4)', 'transparent']}
        start={{ x: 0, y: 0.5 }}
        end={{ x: 1, y: 0.5 }}
        style={styles.fixedBackdrop}
      />


      <ScrollView
        style={styles.scroll}
        contentContainerStyle={{ paddingBottom: 40 }}
        showsVerticalScrollIndicator={false}
      >
        {/* ══════════════════════════════════════════════════════════════════
           HERO — poster + info over fixed backdrop
           ══════════════════════════════════════════════════════════════════ */}
        {/* Spacer so poster sits in upper portion of backdrop */}
        <View style={{ height: layout.height * 0.15 }} />

        <View style={[styles.heroSection, { paddingHorizontal: layout.screenPadding }]}>
          {/* Poster centered (matching web + MediaDetailScreen) */}
          <View style={styles.posterCenter}>
            <View style={styles.posterShadow}>
              {posterUrl ? (
                <Image
                  source={{ uri: posterUrl }}
                  style={[styles.poster, { width: posterWidth, height: posterHeight }]}
                  resizeMode="cover"
                />
              ) : (
                <View style={[styles.poster, styles.posterPlaceholder, { width: posterWidth, height: posterHeight }]}>
                  <Film size={48} color="#444" />
                </View>
              )}
            </View>
          </View>

          {/* Title */}
          <Text
            style={[styles.title, { fontSize: scaledFont(28, layout.fontScale) }]}
            numberOfLines={2}
          >
            {series.title}
          </Text>

          {/* Meta row: Year · Status · Network · Edit */}
          <View style={styles.metaRow}>
            {series.first_air_date && (
              <Text style={[styles.metaText, { fontSize: scaledFont(13, layout.fontScale) }]}>
                {series.first_air_date.split('-')[0]}
              </Text>
            )}
            {series.status && (
              <View style={styles.statusBadge}>
                <Text style={[styles.statusText, { fontSize: scaledFont(11, layout.fontScale) }]}>
                  {series.status}
                </Text>
              </View>
            )}
            {series.network && (
              <View style={styles.networkRow}>
                <Tv size={scaledFont(12, layout.fontScale)} color="#999" />
                <Text style={[styles.metaText, { fontSize: scaledFont(12, layout.fontScale) }]}>
                  {series.network}
                </Text>
              </View>
            )}
            <TouchableOpacity
              style={styles.editBadge}
              onPress={handleEditMetadata}
              hitSlop={{ top: 8, bottom: 8, left: 8, right: 8 }}
            >
              <Pencil size={11} color="#ccc" />
              <Text style={styles.editBadgeText}>Edit</Text>
            </TouchableOpacity>
            {series.metadata_locked && (
              <View style={styles.lockBadge}>
                <Lock size={11} color="#fbbf24" />
              </View>
            )}
          </View>

          {/* Overview */}
          {series.overview && (
            <TouchableOpacity onPress={() => setShowFullOverview(!showFullOverview)} activeOpacity={0.7}>
              <Text
                style={[styles.overview, {
                  fontSize: scaledFont(13, layout.fontScale),
                  lineHeight: scaledFont(20, layout.fontScale),
                }]}
                numberOfLines={showFullOverview ? undefined : 3}
              >
                {series.overview}
              </Text>
              {!showFullOverview && series.overview.length > 120 && (
                <Text style={[styles.readMore, { fontSize: scaledFont(12, layout.fontScale) }]}>
                  Read more
                </Text>
              )}
            </TouchableOpacity>
          )}

          {/* Play/Resume button */}
          <View style={styles.actionRow}>
            <TouchableOpacity style={styles.playButton} onPress={handlePlay} activeOpacity={0.8}>
              <Play size={scaledFont(16, layout.fontScale)} color="#fff" fill="#fff" />
              <Text style={[styles.playButtonText, { fontSize: scaledFont(14, layout.fontScale) }]}>
                {playLabel}
              </Text>
            </TouchableOpacity>

            {youtubeKey && (
              <TouchableOpacity style={styles.trailerButton} onPress={() => setShowTrailer(true)} activeOpacity={0.8}>
                <Film size={scaledFont(14, layout.fontScale)} color="#fff" />
                <Text style={[styles.trailerButtonText, { fontSize: scaledFont(13, layout.fontScale) }]}>
                  Trailer
                </Text>
              </TouchableOpacity>
            )}
          </View>

          {playSubtitle && (
            <Text style={[styles.playSubtitle, { fontSize: scaledFont(12, layout.fontScale) }]}>
              {playSubtitle}
            </Text>
          )}
        </View>

        {/* ══════════════════════════════════════════════════════════════════
           EPISODES SECTION
           ══════════════════════════════════════════════════════════════════ */}
        {sortedSeasons.length > 0 && (
          <View style={{ paddingHorizontal: layout.screenPadding }}>
            <Text style={[styles.sectionTitle, { fontSize: scaledFont(22, layout.fontScale) }]}>
              Episodes
            </Text>

            {/* Season tabs */}
            <ScrollView
              ref={seasonScrollRef}
              horizontal
              showsHorizontalScrollIndicator={false}
              style={styles.seasonTabs}
              contentContainerStyle={{ gap: 8 }}
            >
              {sortedSeasons.map((season) => (
                <TouchableOpacity
                  key={season.id}
                  style={[
                    styles.seasonTab,
                    currentSeason?.id === season.id && styles.seasonTabActive,
                  ]}
                  onPress={() => setSelectedSeasonId(season.id)}
                  activeOpacity={0.7}
                >
                  <Text
                    style={[
                      styles.seasonTabText,
                      { fontSize: scaledFont(13, layout.fontScale) },
                      currentSeason?.id === season.id && styles.seasonTabTextActive,
                    ]}
                  >
                    Season {season.season_number}
                  </Text>
                </TouchableOpacity>
              ))}
            </ScrollView>

            {/* Episode list */}
            {loadingEpisodes ? (
              <ActivityIndicator size="small" color="#e50914" style={{ marginVertical: 24 }} />
            ) : episodes && episodes.length > 0 ? (
              <View style={styles.episodeList}>
                {episodes.map((episode) => {
                  const episodeProgress = progressMap.get(episode.media_id)
                  const progressPercent =
                    episodeProgress && episode.duration
                      ? Math.min((episodeProgress.position / episode.duration) * 100, 100)
                      : 0
                  const isCompleted = episodeProgress?.completed === true
                  const hasProgress = !!episodeProgress && episodeProgress.position > 0 && !isCompleted && (episode.duration ?? 0) > 0
                  const thumbWidth = layout.sideBySideDetail ? 160 : 120
                  const thumbHeight = thumbWidth * 0.625 // 16:10 ratio

                  return (
                    <Pressable
                      key={episode.id}
                      style={({ pressed }) => [
                        styles.episodeCard,
                        pressed && styles.episodeCardPressed,
                      ]}
                      onPress={() => handleEpisodePress(episode)}
                    >
                      {/* Thumbnail */}
                      <View style={[styles.thumbContainer, { width: thumbWidth, height: thumbHeight }]}>
                        {episode.still_path ? (
                          <Image
                            source={{ uri: mediaImage(episode.still_path, 'w300') || '' }}
                            style={[styles.thumbImage, { width: thumbWidth, height: thumbHeight }]}
                            resizeMode="cover"
                          />
                        ) : (
                          <View style={[styles.thumbImage, styles.thumbPlaceholder, { width: thumbWidth, height: thumbHeight }]}>
                            <Film size={24} color="#555" />
                          </View>
                        )}
                        {/* Play overlay */}
                        <View style={styles.thumbPlayOverlay}>
                          <View style={styles.thumbPlayCircle}>
                            <Play size={14} color="#fff" fill="#fff" />
                          </View>
                        </View>
                        {/* Progress bar on thumbnail */}
                        {hasProgress && (
                          <View style={styles.thumbProgress}>
                            <View style={[styles.thumbProgressFill, { width: `${progressPercent}%` }]} />
                          </View>
                        )}
                      </View>

                      {/* Episode info */}
                      <View style={styles.episodeInfo}>
                        <View style={styles.episodeTitleRow}>
                          <Text
                            style={[styles.episodeNumber, { fontSize: scaledFont(15, layout.fontScale) }]}
                          >
                            {episode.episode_number}
                          </Text>
                          <Text
                            style={[styles.episodeTitle, { fontSize: scaledFont(14, layout.fontScale) }]}
                            numberOfLines={1}
                          >
                            {episode.title}
                          </Text>
                          {isCompleted && (
                            <Check size={16} color="#22c55e" />
                          )}
                        </View>
                        {episode.overview && (
                          <Text
                            style={[styles.episodeOverview, {
                              fontSize: scaledFont(12, layout.fontScale),
                              lineHeight: scaledFont(18, layout.fontScale),
                            }]}
                            numberOfLines={2}
                          >
                            {episode.overview}
                          </Text>
                        )}
                        {hasProgress && (
                          <View style={styles.episodeProgressRow}>
                            <View style={styles.episodeProgressTrack}>
                              <View style={[styles.episodeProgressFill, { width: `${progressPercent}%` }]} />
                            </View>
                            <Text style={[styles.episodeProgressText, { fontSize: scaledFont(11, layout.fontScale) }]}>
                              {Math.round(progressPercent)}%
                            </Text>
                          </View>
                        )}
                        {episode.duration && !hasProgress && (
                          <Text style={[styles.episodeDuration, { fontSize: scaledFont(11, layout.fontScale) }]}>
                            {formatDuration(episode.duration)}
                          </Text>
                        )}
                      </View>

                      {/* Three-dot menu */}
                      <TouchableOpacity
                        style={styles.episodeMenuBtn}
                        onPress={() => setEpisodeToEdit(episode)}
                        hitSlop={{ top: 8, bottom: 8, left: 8, right: 8 }}
                      >
                        <MoreVertical size={18} color="#666" />
                      </TouchableOpacity>
                    </Pressable>
                  )
                })}
              </View>
            ) : (
              <View style={styles.noEpisodesContainer}>
                <Text style={[styles.noEpisodes, { fontSize: scaledFont(14, layout.fontScale) }]}>
                  No episodes available
                </Text>
              </View>
            )}
          </View>
        )}
      </ScrollView>

      {/* Action Menu Modal */}
      <Modal
        visible={showMenu}
        transparent
        animationType="fade"
        onRequestClose={() => setShowMenu(false)}
      >
        <TouchableOpacity
          style={styles.menuOverlay}
          activeOpacity={1}
          onPress={() => setShowMenu(false)}
        >
          <View style={styles.menuContainer}>
            <Text style={styles.menuTitle}>Options</Text>

            <TouchableOpacity style={styles.menuItem} onPress={handleCopyStreamUrl}>
              <Link size={18} color="#fff" />
              <Text style={styles.menuItemText}>Copy Stream URL</Text>
            </TouchableOpacity>

            <TouchableOpacity style={styles.menuItem} onPress={handleEditMetadata}>
              <Pencil size={18} color="#fff" />
              <Text style={styles.menuItemText}>Edit Metadata</Text>
            </TouchableOpacity>

            <TouchableOpacity style={styles.menuItem} onPress={handleRefreshMetadata}>
              <Text style={styles.menuItemIcon}>🔄</Text>
              <Text style={styles.menuItemText}>Refresh Metadata</Text>
            </TouchableOpacity>

            <TouchableOpacity style={styles.menuItem} onPress={() => setShowMenu(false)}>
              <Text style={styles.menuItemText}>Cancel</Text>
            </TouchableOpacity>
          </View>
        </TouchableOpacity>
      </Modal>

      {/* Metadata Editor */}
      {showMetadataEditor && series && (
        <MetadataEditor
          type="series"
          item={series}
          genres={(series as any).genres}
          onSave={handleSaveMetadata}
          onRefresh={() => refreshMetadata.mutate(undefined as any)}
          isSaving={editMetadata.isPending}
          isRefreshing={refreshMetadata.isPending}
          onClose={() => setShowMetadataEditor(false)}
        />
      )}

      {/* YouTube Trailer Player */}
      <YouTubePlayer
        videoId={youtubeKey || ''}
        visible={showTrailer}
        onClose={() => setShowTrailer(false)}
      />

      {/* Episode Edit Dialog */}
      {episodeToEdit && (
        <EpisodeEditDialog
          episode={episodeToEdit}
          onSave={handleSaveEpisodeMetadata}
          isSaving={editEpisodeMetadata.isPending}
          onClose={() => setEpisodeToEdit(null)}
        />
      )}
    </View>
  )
}

function formatDuration(seconds: number): string {
  const hours = Math.floor(seconds / 3600)
  const mins = Math.floor((seconds % 3600) / 60)
  if (hours > 0) return `${hours}h ${mins}m`
  return `${mins}m`
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
    backgroundColor: '#141414',
  },
  loading: {
    flex: 1,
    justifyContent: 'center',
    alignItems: 'center',
    backgroundColor: '#141414',
  },
  errorText: {
    color: '#888',
    fontSize: 16,
  },

  // ── Scroll ────────────────────────────────────────────────────────────────
  scroll: {
    flex: 1,
  },

  // ── Fixed backdrop (behind scroll) ─────────────────────────────────────
  fixedBackdrop: {
    position: 'absolute',
    top: 0,
    left: 0,
    right: 0,
    height: SCREEN_WIDTH * 1.2,
  },
  backdropPlaceholder: {
    backgroundColor: '#1a1a1a',
  },

  // ── Hero section ──────────────────────────────────────────────────────────
  heroSection: {
    marginBottom: 24,
  },
  posterCenter: {
    alignItems: 'center',
    marginBottom: 20,
  },
  posterShadow: {
    shadowColor: '#000',
    shadowOffset: { width: 0, height: 8 },
    shadowOpacity: 0.7,
    shadowRadius: 16,
    elevation: 16,
  },
  poster: {
    borderRadius: 10,
    backgroundColor: '#1f1f1f',
  },
  posterPlaceholder: {
    justifyContent: 'center',
    alignItems: 'center',
  },

  title: {
    fontWeight: 'bold',
    color: '#fff',
    marginBottom: 8,
    textShadowColor: 'rgba(0,0,0,0.8)',
    textShadowOffset: { width: 0, height: 2 },
    textShadowRadius: 8,
  },
  metaRow: {
    flexDirection: 'row',
    alignItems: 'center',
    flexWrap: 'wrap',
    gap: 8,
    marginBottom: 14,
  },
  metaText: {
    color: '#aaa',
  },
  statusBadge: {
    backgroundColor: 'rgba(168,85,247,0.2)',
    paddingHorizontal: 8,
    paddingVertical: 3,
    borderRadius: 4,
  },
  statusText: {
    fontWeight: '600',
    color: '#c084fc',
  },
  networkRow: {
    flexDirection: 'row',
    alignItems: 'center',
    gap: 4,
  },
  editBadge: {
    flexDirection: 'row',
    alignItems: 'center',
    gap: 4,
    backgroundColor: 'rgba(255,255,255,0.1)',
    paddingHorizontal: 8,
    paddingVertical: 3,
    borderRadius: 4,
  },
  editBadgeText: {
    fontSize: 11,
    fontWeight: '500',
    color: '#ccc',
  },
  lockBadge: {
    backgroundColor: 'rgba(251,191,36,0.15)',
    paddingHorizontal: 6,
    paddingVertical: 3,
    borderRadius: 10,
  },
  actionRow: {
    flexDirection: 'row',
    alignItems: 'center',
    gap: 10,
  },
  playButton: {
    backgroundColor: '#e50914',
    borderRadius: 6,
    paddingVertical: 10,
    paddingHorizontal: 22,
    flexDirection: 'row',
    alignItems: 'center',
    gap: 8,
  },
  playButtonText: {
    color: '#fff',
    fontWeight: '700',
  },
  trailerButton: {
    backgroundColor: 'rgba(255,255,255,0.12)',
    borderRadius: 6,
    paddingVertical: 10,
    paddingHorizontal: 16,
    borderWidth: 1,
    borderColor: 'rgba(255,255,255,0.15)',
    flexDirection: 'row',
    alignItems: 'center',
    gap: 6,
  },
  trailerButtonText: {
    color: '#fff',
    fontWeight: '600',
  },

  // ── Overview ──────────────────────────────────────────────────────────────
  overview: {
    color: '#ccc',
    marginBottom: 14,
  },
  readMore: {
    color: '#e50914',
    marginTop: -10,
    marginBottom: 14,
    fontWeight: '500',
  },
  playSubtitle: {
    color: '#999',
    marginTop: 6,
  },

  // ── Episodes section ──────────────────────────────────────────────────────
  sectionTitle: {
    fontWeight: 'bold',
    color: '#fff',
    marginBottom: 16,
  },
  seasonTabs: {
    marginBottom: 16,
  },
  seasonTab: {
    paddingHorizontal: 16,
    paddingVertical: 8,
    borderRadius: 8,
    backgroundColor: 'rgba(255,255,255,0.06)',
  },
  seasonTabActive: {
    backgroundColor: '#e50914',
  },
  seasonTabText: {
    color: '#888',
    fontWeight: '500',
  },
  seasonTabTextActive: {
    color: '#fff',
    fontWeight: '600',
  },

  // ── Episode cards ─────────────────────────────────────────────────────────
  episodeList: {
    gap: 10,
  },
  episodeCard: {
    flexDirection: 'row',
    alignItems: 'center',
    backgroundColor: 'rgba(30,30,30,0.9)',
    borderRadius: 10,
    padding: 10,
    gap: 12,
  },
  episodeCardPressed: {
    backgroundColor: 'rgba(50,50,50,0.9)',
  },
  thumbContainer: {
    borderRadius: 8,
    overflow: 'hidden',
    position: 'relative',
  },
  thumbImage: {
    backgroundColor: '#1a1a1a',
    borderRadius: 8,
  },
  thumbPlaceholder: {
    justifyContent: 'center',
    alignItems: 'center',
  },
  thumbPlayOverlay: {
    position: 'absolute',
    top: 0,
    left: 0,
    right: 0,
    bottom: 0,
    justifyContent: 'center',
    alignItems: 'center',
    backgroundColor: 'rgba(0,0,0,0.3)',
  },
  thumbPlayCircle: {
    width: 32,
    height: 32,
    borderRadius: 16,
    backgroundColor: 'rgba(229,9,20,0.85)',
    justifyContent: 'center',
    alignItems: 'center',
  },
  thumbProgress: {
    position: 'absolute',
    bottom: 0,
    left: 0,
    right: 0,
    height: 3,
    backgroundColor: 'rgba(255,255,255,0.2)',
  },
  thumbProgressFill: {
    height: '100%',
    backgroundColor: '#e50914',
    borderRadius: 1.5,
  },

  episodeInfo: {
    flex: 1,
    justifyContent: 'center',
  },
  episodeTitleRow: {
    flexDirection: 'row',
    alignItems: 'center',
    gap: 8,
    marginBottom: 4,
  },
  episodeNumber: {
    fontWeight: 'bold',
    color: '#666',
  },
  episodeTitle: {
    fontWeight: '600',
    color: '#fff',
    flex: 1,
  },
  episodeOverview: {
    color: '#888',
    marginTop: 2,
  },
  episodeProgressRow: {
    flexDirection: 'row',
    alignItems: 'center',
    gap: 8,
    marginTop: 6,
    maxWidth: 200,
  },
  episodeProgressTrack: {
    flex: 1,
    height: 3,
    borderRadius: 1.5,
    backgroundColor: 'rgba(255,255,255,0.15)',
  },
  episodeProgressFill: {
    height: '100%',
    borderRadius: 1.5,
    backgroundColor: '#e50914',
  },
  episodeProgressText: {
    color: '#666',
  },
  episodeDuration: {
    color: '#666',
    marginTop: 4,
  },
  episodeMenuBtn: {
    padding: 8,
    alignSelf: 'center',
  },

  noEpisodesContainer: {
    height: 100,
    borderRadius: 10,
    backgroundColor: 'rgba(30,30,30,0.9)',
    justifyContent: 'center',
    alignItems: 'center',
  },
  noEpisodes: {
    color: '#666',
  },

  // ── Menu ──────────────────────────────────────────────────────────────────
  menuOverlay: {
    flex: 1,
    backgroundColor: 'rgba(0, 0, 0, 0.75)',
    justifyContent: 'center',
    alignItems: 'center',
  },
  menuContainer: {
    backgroundColor: '#1f1f1f',
    borderRadius: 12,
    width: '80%',
    maxWidth: 320,
    overflow: 'hidden',
  },
  menuTitle: {
    fontSize: 18,
    fontWeight: 'bold',
    color: '#fff',
    textAlign: 'center',
    paddingVertical: 16,
    borderBottomWidth: 1,
    borderBottomColor: '#333',
  },
  menuItem: {
    flexDirection: 'row',
    alignItems: 'center',
    paddingVertical: 14,
    paddingHorizontal: 20,
    gap: 12,
    borderBottomWidth: 1,
    borderBottomColor: '#2a2a2a',
  },
  menuItemIcon: {
    fontSize: 18,
  },
  menuItemText: {
    fontSize: 16,
    color: '#fff',
  },
})
