import { useState, useMemo, useRef, useEffect } from 'react'
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
} from 'react-native'
import { Tv, Play, Film, Check, Link, Pencil, Lock, ChevronLeft } from 'lucide-react-native'
import { useNavigation, useRoute, RouteProp } from '@react-navigation/native'
import type { NativeStackNavigationProp } from '@react-navigation/native-stack'
import { useSeriesDetail, useSeasons, useEpisodes } from '@velox/shared/hooks'
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
type SeriesRouteProp = RouteProp<RootStackParamList, 'Series'>

export function SeriesDetailScreen() {
  const navigation = useNavigation<NavigationProp>()
  const route = useRoute<SeriesRouteProp>()
  const { id } = route.params
  const layout = useResponsiveLayout()
  const { data: series, isLoading: loadingSeries } = useSeriesDetail(id)
  const { data: seasons, isLoading: loadingSeasons } = useSeasons(id)
  const { youtubeKey } = useSeriesTrailers(id)
  const [selectedSeasonId, setSelectedSeasonId] = useState<number | null>(null)
  const [showFullOverview, setShowFullOverview] = useState(false)
  const [showMenu, setShowMenu] = useState(false)
  const [showMetadataEditor, setShowMetadataEditor] = useState(false)
  const [showTrailer, setShowTrailer] = useState(false)
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

  // Responsive dimensions — web-like side-by-side layout
  const posterWidth = layout.sideBySideDetail ? 170 : 110
  const posterHeight = layout.sideBySideDetail ? 255 : 165
  const episodeThumbWidth = layout.sideBySideDetail ? 180 : 130
  const episodeThumbHeight = layout.sideBySideDetail ? 100 : 75
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

  const handlePlay = () => {
    if (episodes && episodes.length > 0) {
      navigation.navigate('Episode', {
        id: episodes[0].id,
        seriesId: id,
      })
    }
  }

  const handleEpisodePress = (episode: Episode) => {
    navigation.navigate('Episode', {
      id: episode.id,
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

      {/* ── Full-screen Backdrop ── */}
      {backdropUrl ? (
        <Image source={{ uri: backdropUrl }} style={styles.fullBackdrop} resizeMode="cover" />
      ) : (
        <View style={[styles.fullBackdrop, styles.fullBackdropPlaceholder]} />
      )}

      {/* ── Vertical gradient bands: top → bottom ── */}
      <View style={styles.vGrad0} />
      <View style={styles.vGrad1} />
      <View style={styles.vGrad2} />
      <View style={styles.vGrad3} />
      <View style={styles.vGrad4} />

      {/* ── Horizontal gradient bands: left → right ── */}
      <View style={styles.hGrad0} />
      <View style={styles.hGrad1} />
      <View style={styles.hGrad2} />
      <View style={styles.hGrad3} />

      {/* ── Bottom solid fade ── */}
      <View style={styles.bottomSolid} />

      {/* ── Floating overlay header (rendered LAST → zIndex above all gradients) ── */}
      <View style={[styles.overlayHeader, { paddingTop: statusBarHeight + 8 }]}>
        <TouchableOpacity
          style={styles.backButton}
          onPress={() => navigation.goBack()}
          hitSlop={{ top: 10, bottom: 10, left: 10, right: 10 }}
        >
          <ChevronLeft size={26} color="#fff" />
        </TouchableOpacity>
        <Text style={[styles.overlayTitle, { fontSize: scaledFont(16, layout.fontScale) }]} numberOfLines={1}>
          {series.title}
        </Text>
        <TouchableOpacity
          style={styles.overlayMenuButton}
          onPress={() => setShowMenu(true)}
          hitSlop={{ top: 10, bottom: 10, left: 10, right: 10 }}
        >
          <Text style={styles.menuIcon}>⋮</Text>
        </TouchableOpacity>
      </View>

      {/* ── Floating overlay header ── */}
      <View style={[styles.overlayHeader, { paddingTop: statusBarHeight + 8 }]}>
        <TouchableOpacity
          style={styles.backButton}
          onPress={() => navigation.goBack()}
          hitSlop={{ top: 10, bottom: 10, left: 10, right: 10 }}
        >
          <ChevronLeft size={26} color="#fff" />
        </TouchableOpacity>
        <Text style={[styles.overlayTitle, { fontSize: scaledFont(16, layout.fontScale) }]} numberOfLines={1}>
          {series.title}
        </Text>
        <TouchableOpacity
          style={styles.overlayMenuButton}
          onPress={() => setShowMenu(true)}
          hitSlop={{ top: 10, bottom: 10, left: 10, right: 10 }}
        >
          <Text style={styles.menuIcon}>⋮</Text>
        </TouchableOpacity>
      </View>

      {/* ── Scrollable content ── */}
      <ScrollView
        style={styles.scrollContent}
        contentContainerStyle={[
          styles.scrollContentContainer,
          { paddingBottom: layout.screenPadding * 2 },
        ]}
      >
        {/* Poster + meta floating over hero bottom */}
        <View style={styles.heroMeta}>
          {/* Poster */}
          <View style={styles.posterWrapper}>
            {posterUrl ? (
              <Image
                source={{ uri: posterUrl }}
                style={[styles.poster, { width: posterWidth, height: posterHeight }]}
                resizeMode="cover"
              />
            ) : (
              <View
                style={[
                  styles.poster,
                  styles.posterPlaceholder,
                  { width: posterWidth, height: posterHeight },
                ]}
              >
                <Text style={styles.posterText}>{series.title?.charAt(0)}</Text>
              </View>
            )}
          </View>

          {/* Info column — title + meta + actions to the right of poster */}
          <View style={styles.heroMetaInfo}>
            {/* Title row */}
            <View style={styles.titleRow}>
              <Text
                style={[styles.title, { fontSize: scaledFont(28, layout.fontScale) }]}
                numberOfLines={2}
              >
                {series.title}
              </Text>
              {series.metadata_locked && (
                <View style={styles.lockBadge}>
                  <Lock size={scaledFont(11, layout.fontScale)} color="#fff" />
                </View>
              )}
            </View>

            {/* Year · Status · Network inline */}
            <View style={styles.metaRow}>
              {series.first_air_date && (
                <Text style={[styles.year, { fontSize: scaledFont(13, layout.fontScale) }]}>
                  {series.first_air_date.split('-')[0]}
                </Text>
              )}
              {series.status && (
                <View style={styles.statusBadge}>
                  <Text style={[styles.statusText, { fontSize: scaledFont(10, layout.fontScale) }]}>
                    {series.status}
                  </Text>
                </View>
              )}
              {series.network && (
                <View style={styles.networkRow}>
                  <Tv size={scaledFont(11, layout.fontScale)} color="#888" />
                  <Text style={[styles.networkText, { fontSize: scaledFont(12, layout.fontScale) }]}>
                    {series.network}
                  </Text>
                </View>
              )}
            </View>

            {/* Play + Trailer buttons */}
            <View style={styles.actionRow}>
              <TouchableOpacity style={styles.playButton} onPress={handlePlay}>
                <View style={{ flexDirection: 'row', alignItems: 'center', gap: 6 }}>
                  <Play size={scaledFont(16, layout.fontScale)} color="#fff" />
                  <Text style={[styles.playButtonText, { fontSize: scaledFont(14, layout.fontScale) }]}>
                    Play
                  </Text>
                </View>
              </TouchableOpacity>

              {youtubeKey && (
                <TouchableOpacity style={styles.trailerButton} onPress={() => setShowTrailer(true)}>
                  <View style={{ flexDirection: 'row', alignItems: 'center', gap: 5 }}>
                    <Film size={scaledFont(13, layout.fontScale)} color="#fff" />
                    <Text
                      style={[styles.trailerButtonText, { fontSize: scaledFont(12, layout.fontScale) }]}
                    >
                      Trailer
                    </Text>
                  </View>
                </TouchableOpacity>
              )}
            </View>
          </View>
        </View>

        {/* Overview */}
        {series.overview && (
          <TouchableOpacity onPress={() => setShowFullOverview(!showFullOverview)}>
            <Text
              style={[
                styles.overview,
                {
                  fontSize: scaledFont(13, layout.fontScale),
                  lineHeight: scaledFont(20, layout.fontScale),
                },
              ]}
              numberOfLines={showFullOverview ? undefined : 2}
            >
              {series.overview}
            </Text>
            {!showFullOverview && series.overview.length > 80 && (
              <Text style={[styles.readMore, { fontSize: scaledFont(13, layout.fontScale) }]}>
                Read more
              </Text>
            )}
          </TouchableOpacity>
        )}

        {/* Seasons */}
        {sortedSeasons.length > 0 && (
          <View style={styles.section}>
            <ScrollView
              ref={seasonScrollRef}
              horizontal
              showsHorizontalScrollIndicator={false}
              style={styles.seasonTabs}
              contentContainerStyle={{ gap: 8, paddingRight: layout.screenPadding }}
            >
              {sortedSeasons.map((season) => (
                <TouchableOpacity
                  key={season.id}
                  style={[
                    styles.seasonTab,
                    currentSeason?.id === season.id && styles.seasonTabActive,
                  ]}
                  onPress={() => setSelectedSeasonId(season.id)}
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

            {/* Episode List */}
            {loadingEpisodes ? (
              <ActivityIndicator size="small" color="#e50914" style={styles.episodeLoader} />
            ) : episodes && episodes.length > 0 ? (
              <View style={styles.episodeList}>
                {episodes.map((episode) => {
                  const episodeProgress = progressMap.get(episode.media_id)
                  const progressPercent =
                    episodeProgress && episode.duration
                      ? Math.min((episodeProgress.position / episode.duration) * 100, 100)
                      : 0
                  const isCompleted = episodeProgress?.completed === true
                  return (
                    <Pressable
                      key={episode.id}
                      style={({ pressed }) => [
                        styles.episodeItem,
                        pressed && styles.episodeItemPressed,
                      ]}
                      onPress={() => handleEpisodePress(episode)}
                    >
                      {/* Still */}
                      <View
                        style={[
                          styles.episodeThumbContainer,
                          { width: episodeThumbWidth, height: episodeThumbHeight },
                        ]}
                      >
                        {episode.still_path ? (
                          <Image
                            source={{ uri: mediaImage(episode.still_path, 'w400') || '' }}
                            style={[
                              styles.episodeThumb,
                              { width: episodeThumbWidth, height: episodeThumbHeight },
                            ]}
                            resizeMode="cover"
                          />
                        ) : (
                          <View
                            style={[
                              styles.episodeThumb,
                              styles.episodeThumbPlaceholder,
                              { width: episodeThumbWidth, height: episodeThumbHeight },
                            ]}
                          >
                            <Text
                              style={[
                                styles.episodeThumbText,
                                { fontSize: scaledFont(22, layout.fontScale) },
                              ]}
                            >
                              {episode.episode_number}
                            </Text>
                          </View>
                        )}
                        {episodeProgress && episodeProgress.position > 0 && !isCompleted && (
                          <View style={styles.episodeProgressOverlay}>
                            <View style={styles.episodeProgressBar}>
                              <View
                                style={[
                                  styles.episodeProgressFill,
                                  { width: `${progressPercent}%` },
                                ]}
                              />
                            </View>
                          </View>
                        )}
                        {/* Play icon overlay */}
                        <View style={styles.episodePlayIcon}>
                          <Play size={16} color="#fff" />
                        </View>
                      </View>

                      {/* Info */}
                      <View style={styles.episodeInfo}>
                        <Text
                          style={[
                            styles.episodeTitle,
                            { fontSize: scaledFont(14, layout.fontScale) },
                          ]}
                          numberOfLines={1}
                        >
                          {episode.episode_number}. {episode.title}
                        </Text>
                        {episode.duration && (
                          <Text
                            style={[
                              styles.episodeRuntime,
                              { fontSize: scaledFont(12, layout.fontScale) },
                            ]}
                          >
                            {Math.floor(episode.duration / 60)}m
                          </Text>
                        )}
                        {episode.overview && (
                          <Text
                            style={[
                              styles.episodeOverview,
                              {
                                fontSize: scaledFont(12, layout.fontScale),
                                lineHeight: scaledFont(18, layout.fontScale),
                              },
                            ]}
                            numberOfLines={2}
                          >
                            {episode.overview}
                          </Text>
                        )}
                      </View>

                      {isCompleted && (
                        <View style={styles.watchedIndicator}>
                          <Check size={14} color="#fff" />
                        </View>
                      )}
                    </Pressable>
                  )
                })}
              </View>
            ) : (
              <Text style={[styles.noEpisodes, { fontSize: scaledFont(14, layout.fontScale) }]}>
                No episodes available
              </Text>
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

  // ── Full-screen Backdrop ──────────────────────────────────────────────────────
  fullBackdrop: {
    position: 'absolute',
    top: 0,
    left: 0,
    right: 0,
    bottom: 0,
    zIndex: 0,
  },
  fullBackdropPlaceholder: {
    backgroundColor: '#1f1f1f',
  },

  // ── Vertical gradient bands: top → bottom (dark → transparent) ─────────────────
  vGrad0: { position: 'absolute', top: 0, left: 0, right: 0, height: '20%', zIndex: 1, backgroundColor: 'rgba(20,20,20,0.95)' },
  vGrad1: { position: 'absolute', top: '20%', left: 0, right: 0, height: '20%', zIndex: 1, backgroundColor: 'rgba(20,20,20,0.78)' },
  vGrad2: { position: 'absolute', top: '40%', left: 0, right: 0, height: '20%', zIndex: 1, backgroundColor: 'rgba(20,20,20,0.55)' },
  vGrad3: { position: 'absolute', top: '60%', left: 0, right: 0, height: '20%', zIndex: 1, backgroundColor: 'rgba(20,20,20,0.32)' },
  vGrad4: { position: 'absolute', top: '80%', left: 0, right: 0, bottom: 0, zIndex: 1, backgroundColor: 'rgba(20,20,20,0.10)' },

  // ── Horizontal gradient bands: left → right (dark → transparent) ─────────────────
  hGrad0: { position: 'absolute', top: 0, left: 0, bottom: 0, width: '35%', zIndex: 2, backgroundColor: 'rgba(20,20,20,0.85)' },
  hGrad1: { position: 'absolute', top: 0, bottom: 0, left: '25%', width: '25%', zIndex: 2, backgroundColor: 'rgba(20,20,20,0.50)' },
  hGrad2: { position: 'absolute', top: 0, bottom: 0, left: '45%', width: '20%', zIndex: 2, backgroundColor: 'rgba(20,20,20,0.20)' },
  hGrad3: { position: 'absolute', top: 0, bottom: 0, left: '60%', width: '15%', zIndex: 2, backgroundColor: 'rgba(20,20,20,0.05)' },

  // ── Bottom solid fade ─────────────────────────────────────────────────────────
  bottomSolid: { position: 'absolute', left: 0, right: 0, bottom: 0, height: '25%', zIndex: 3, backgroundColor: '#141414' },

  // ── Overlay Header ───────────────────────────────────────────────────────────
  overlayHeader: {
    position: 'absolute',
    top: 0,
    left: 0,
    right: 0,
    flexDirection: 'row',
    alignItems: 'center',
    paddingHorizontal: 12,
    zIndex: 100,
  },
  backButton: {
    width: 40,
    height: 40,
    borderRadius: 20,
    backgroundColor: 'rgba(0,0,0,0.6)',
    justifyContent: 'center',
    alignItems: 'center',
  },
  overlayTitle: {
    flex: 1,
    color: '#fff',
    fontWeight: '600',
    textAlign: 'center',
    marginHorizontal: 8,
    textShadowColor: 'rgba(0,0,0,0.9)',
    textShadowOffset: { width: 0, height: 1 },
    textShadowRadius: 4,
  },
  overlayMenuButton: {
    width: 40,
    height: 40,
    borderRadius: 20,
    backgroundColor: 'rgba(0,0,0,0.6)',
    justifyContent: 'center',
    alignItems: 'center',
  },

  // ── Scroll content ──────────────────────────────────────────────────────────
  scrollContent: {
    position: 'absolute',
    top: 0,
    left: 0,
    right: 0,
    bottom: 0,
    zIndex: 10,
  },
  scrollContentContainer: {
    flexGrow: 1,
  },

  // ── Hero meta (poster + info side-by-side) ────────────────────────────────
  heroMeta: {
    flexDirection: 'row',
    paddingHorizontal: 16,
    alignItems: 'flex-end',
    paddingTop: 80, // Space for overlay header
    gap: 14,
  },
  posterWrapper: {
    shadowColor: '#000',
    shadowOffset: { width: 0, height: 6 },
    shadowOpacity: 0.65,
    shadowRadius: 14,
    elevation: 14,
    flexShrink: 0,
  },
  poster: {
    borderRadius: 8,
    backgroundColor: '#1f1f1f',
  },
  posterPlaceholder: {
    justifyContent: 'center',
    alignItems: 'center',
  },
  posterText: {
    fontSize: 36,
    fontWeight: 'bold',
    color: '#444',
  },
  heroMetaInfo: {
    flex: 1,
    justifyContent: 'flex-end',
    paddingBottom: 2,
  },
  titleRow: {
    flexDirection: 'row',
    alignItems: 'flex-start',
    gap: 8,
    marginBottom: 6,
  },
  title: {
    fontSize: 28,
    fontWeight: 'bold',
    color: '#fff',
    flex: 1,
    textShadowColor: 'rgba(0,0,0,0.9)',
    textShadowOffset: { width: 0, height: 2 },
    textShadowRadius: 6,
  },
  lockBadge: {
    backgroundColor: 'rgba(0, 0, 0, 0.6)',
    paddingHorizontal: 6,
    paddingVertical: 4,
    borderRadius: 4,
  },
  metaRow: {
    flexDirection: 'row',
    alignItems: 'center',
    gap: 8,
    marginBottom: 12,
  },
  year: {
    fontSize: 13,
    color: '#bbb',
  },
  statusBadge: {
    backgroundColor: 'rgba(128, 90, 213, 0.25)',
    paddingHorizontal: 6,
    paddingVertical: 2,
    borderRadius: 4,
  },
  statusText: {
    fontSize: 10,
    fontWeight: '700',
    color: '#c4a0ff',
  },
  networkRow: {
    flexDirection: 'row',
    alignItems: 'center',
    gap: 3,
  },
  networkText: {
    fontSize: 12,
    color: '#888',
  },
  actionRow: {
    flexDirection: 'row',
    alignItems: 'center',
    gap: 10,
    marginBottom: 12,
  },
  playButton: {
    backgroundColor: '#e50914',
    borderRadius: 6,
    paddingVertical: 10,
    paddingHorizontal: 18,
    flexDirection: 'row',
    alignItems: 'center',
  },
  playButtonText: {
    color: '#fff',
    fontWeight: '700',
  },
  trailerButton: {
    backgroundColor: 'rgba(255,255,255,0.15)',
    borderRadius: 6,
    paddingVertical: 10,
    paddingHorizontal: 14,
    borderWidth: 1,
    borderColor: 'rgba(255,255,255,0.2)',
    flexDirection: 'row',
    alignItems: 'center',
  },
  trailerButtonText: {
    color: '#fff',
    fontWeight: '600',
  },
  overview: {
    fontSize: 13,
    color: '#bbb',
    lineHeight: 20,
    paddingHorizontal: 16,
    marginBottom: 20,
  },
  readMore: {
    fontSize: 13,
    color: '#e50914',
    marginTop: 4,
    paddingHorizontal: 16,
  },

  // ── Sections ─────────────────────────────────────────────────────────────────
  section: {
    marginBottom: 24,
  },
  seasonTabs: {
    marginBottom: 16,
    paddingHorizontal: 16,
  },
  seasonTab: {
    paddingHorizontal: 14,
    paddingVertical: 8,
    borderRadius: 20,
    backgroundColor: 'rgba(255,255,255,0.08)',
    borderWidth: 1,
    borderColor: 'rgba(255,255,255,0.1)',
  },
  seasonTabActive: {
    backgroundColor: '#e50914',
    borderColor: '#e50914',
  },
  seasonTabText: {
    fontSize: 13,
    color: '#888',
  },
  seasonTabTextActive: {
    color: '#fff',
    fontWeight: '600',
  },
  episodeLoader: {
    marginVertical: 20,
  },
  noEpisodes: {
    color: '#666',
    fontSize: 14,
    textAlign: 'center',
    marginVertical: 20,
  },

  // ── Episode cards ────────────────────────────────────────────────────────────
  episodeList: {
    gap: 10,
    paddingHorizontal: 16,
  },
  episodeItem: {
    flexDirection: 'row',
    alignItems: 'center',
    backgroundColor: 'rgba(255,255,255,0.04)',
    borderRadius: 10,
    borderWidth: 1,
    borderColor: 'rgba(255,255,255,0.06)',
    overflow: 'hidden',
  },
  episodeItemPressed: {
    backgroundColor: 'rgba(255,255,255,0.08)',
    borderColor: 'rgba(255,255,255,0.12)',
  },
  episodeThumbContainer: {
    position: 'relative',
  },
  episodeThumb: {
    backgroundColor: '#2a2a2a',
  },
  episodeThumbPlaceholder: {
    justifyContent: 'center',
    alignItems: 'center',
  },
  episodeThumbText: {
    fontSize: 22,
    fontWeight: 'bold',
    color: '#444',
  },
  episodeProgressOverlay: {
    position: 'absolute',
    bottom: 0,
    left: 0,
    right: 0,
    height: 3,
    backgroundColor: 'rgba(255,255,255,0.15)',
  },
  episodeProgressBar: {
    height: '100%',
    backgroundColor: 'rgba(229,9,20,0.9)',
  },
  episodeProgressFill: {
    height: '100%',
    backgroundColor: '#e50914',
  },
  episodePlayIcon: {
    position: 'absolute',
    top: '50%',
    left: '50%',
    marginTop: -12,
    marginLeft: -12,
    width: 24,
    height: 24,
    borderRadius: 12,
    backgroundColor: 'rgba(0,0,0,0.55)',
    justifyContent: 'center',
    alignItems: 'center',
  },
  episodeInfo: {
    flex: 1,
    paddingHorizontal: 12,
    paddingVertical: 10,
  },
  episodeTitle: {
    fontSize: 14,
    fontWeight: '600',
    color: '#fff',
    marginBottom: 3,
  },
  episodeRuntime: {
    fontSize: 12,
    color: '#777',
    marginBottom: 3,
  },
  episodeOverview: {
    fontSize: 12,
    color: '#666',
    lineHeight: 17,
  },
  watchedIndicator: {
    width: 24,
    height: 24,
    borderRadius: 12,
    backgroundColor: 'rgba(26,92,26,0.8)',
    justifyContent: 'center',
    alignItems: 'center',
    marginRight: 10,
  },

  // ── Menu ──────────────────────────────────────────────────────────────────────
  menuIcon: {
    fontSize: 20,
    color: '#fff',
  },
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
