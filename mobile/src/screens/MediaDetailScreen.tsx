import { useState } from 'react'
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
} from 'react-native'
import { Play, Check, Circle, Link, Pencil, Lock, Search, Languages } from 'lucide-react-native'
import { useNavigation, useRoute, RouteProp } from '@react-navigation/native'
import type { NativeStackNavigationProp } from '@react-navigation/native-stack'
import {
  useMedia,
  useMediaWithFiles,
  useProgress,
  useToggleFavorite,
  useSubtitles,
} from '@velox/shared/hooks'
import { usePlayerStore } from '../stores/player'
import { LANG_NAMES } from '@velox/shared/lib/languages'
import {
  useEditMediaMetadata,
  useRefreshMetadata,
} from '@velox/shared/hooks/media/useMetadataOps'
import { useTrailers } from '@velox/shared/hooks/media/useCinema'
import { mediaImage } from '@velox/shared/lib'
import { getServerUrl } from '../platform/mobile-adapter'
import { PlayOptionsBottomSheet } from '../components/PlayOptionsBottomSheet'
import { MetadataEditor } from '../components/MetadataEditor'
import { YouTubePlayer } from '../components/YouTubePlayer'
import { Toast } from '../components/Toast'
import { SubtitleSearchModal } from '../components/SubtitleSearchModal'
import { SubtitleTranslate } from '../components/SubtitleTranslate'
import { useResponsiveLayout, scaledFont, scaledSpacing } from '../lib/responsive'
import type { RootStackParamList } from '../../App'
import type { MetadataEditRequest } from '@velox/shared/types'

type NavigationProp = NativeStackNavigationProp<RootStackParamList>
type MediaRouteProp = RouteProp<RootStackParamList, 'Media'>

export function MediaDetailScreen() {
  const navigation = useNavigation<NavigationProp>()
  const route = useRoute<MediaRouteProp>()
  const { id } = route.params
  const layout = useResponsiveLayout()

  const { data: media, isLoading: loadingMedia } = useMedia(id)
  const { data: mediaFiles, isLoading: loadingFiles } = useMediaWithFiles(id)
  const { data: progress } = useProgress(id)
  const toggleFavorite = useToggleFavorite()
  const { youtubeKey } = useTrailers(id)
  const { data: subtitles = [] } = useSubtitles(id)
  const { subtitleLanguage, setSubtitleLanguage } = usePlayerStore()

  // Metadata editing
  const [showMetadataEditor, setShowMetadataEditor] = useState(false)
  const editMetadata = useEditMediaMetadata(id)
  const refreshMetadata = useRefreshMetadata(id)

  const [isFavorited, setIsFavorited] = useState(false)
  const [isWatched, setIsWatched] = useState(false)
  const [showMenu, setShowMenu] = useState(false)
  const [showPlayOptions, setShowPlayOptions] = useState(false)
  const [showTrailer, setShowTrailer] = useState(false)
  const [showSubtitleSelector, setShowSubtitleSelector] = useState(false)
  const [showSubtitleSearch, setShowSubtitleSearch] = useState(false)
  const [showSubtitleTranslate, setShowSubtitleTranslate] = useState(false)
  const [playOptions, setPlayOptions] = useState<{
    startFromBeginning?: boolean
  }>({})

  // Responsive dimensions
  const posterWidth = layout.sideBySideDetail ? 200 : 150
  const posterHeight = layout.sideBySideDetail ? 300 : 225
  const backdropHeight = layout.sideBySideDetail ? layout.height * 0.4 : layout.height * 0.5
  const actionButtonSize = layout.largeControls ? 60 : 50

  const isLoading = loadingMedia || loadingFiles

  if (isLoading) {
    return (
      <View style={styles.loading}>
        <ActivityIndicator size="large" color="#e50914" />
      </View>
    )
  }

  if (!media) {
    return (
      <View style={styles.loading}>
        <Text style={styles.errorText}>Media not found</Text>
      </View>
    )
  }

  const backdropUrl = media.backdrop_path ? mediaImage(media.backdrop_path, 'w1280') : null
  const posterUrl = media.poster_path ? mediaImage(media.poster_path, 'w500') : null

  const year = media.release_date ? media.release_date.split('-')[0] : null

  // Calculate "Ends at" time
  const endsAt = progress?.position
    ? new Date(progress.position * 1000).toLocaleTimeString([], {
        hour: '2-digit',
        minute: '2-digit',
      })
    : null

  // Calculate progress percentage
  const progressPercent =
    progress && media.duration && media.duration > 0
      ? (progress.position / media.duration) * 100
      : 0

  // Format runtime
  const runtimeFormatted = media.duration
    ? `${Math.floor(media.duration / 60)}h ${media.duration % 60}m`
    : null

  const hasProgress = progressPercent > 0

  const handlePlay = () => {
    setPlayOptions({})
    setShowPlayOptions(true)
  }

  const handlePlayFromBeginning = () => {
    setPlayOptions({ startFromBeginning: true })
    setShowPlayOptions(true)
  }

  const handlePlayWithOptions = () => {
    navigation.navigate('Episode', {
      id,
      seriesId: 0,
      startFromBeginning: playOptions.startFromBeginning,
    })
  }

  const handleToggleFavorite = () => {
    setIsFavorited(!isFavorited)
    toggleFavorite.mutate(id)
  }

  const handleToggleWatched = () => {
    setIsWatched(!isWatched)
  }

  const handleCopyStreamUrl = async () => {
    setShowMenu(false)
    try {
      const serverUrl = getServerUrl() ?? 'http://localhost:8080'
      await Share.share({
        message: `Stream URL for ${media.title}: ${serverUrl}/play/${id}`,
      })
    } catch {
      Alert.alert('Error', 'Could not share stream URL')
    }
  }

  const handleRefreshMetadata = () => {
    setShowMenu(false)
    refreshMetadata.mutate(undefined, {
      onSuccess: () => {
        Toast.info('Metadata refreshed successfully')
      },
      onError: () => {
        Toast.error('Failed to refresh metadata')
      },
    })
  }

  const handleSaveMetadata = (req: MetadataEditRequest) => {
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

  const handleEditMetadata = () => {
    setShowMenu(false)
    setShowMetadataEditor(true)
  }

  // Get primary file for technical specs
  const primaryFile = mediaFiles?.files?.find((f) => f.is_primary) || mediaFiles?.files?.[0]

  // ── Shared info content (used in both layouts) ────────────────────────────

  const renderInfoContent = () => (
    <>
      <Text style={[styles.title, { fontSize: scaledFont(24, layout.fontScale) }]}>
        {media.title}
      </Text>

      {/* Metadata line */}
      <View style={styles.metadataLine}>
        {year && (
          <Text style={[styles.year, { fontSize: scaledFont(14, layout.fontScale) }]}>{year}</Text>
        )}
        {year && runtimeFormatted && (
          <Text style={[styles.separator, { fontSize: scaledFont(14, layout.fontScale) }]}>
            ·
          </Text>
        )}
        {runtimeFormatted && (
          <Text style={[styles.runtime, { fontSize: scaledFont(14, layout.fontScale) }]}>
            {runtimeFormatted}
          </Text>
        )}
        {endsAt && (
          <>
            <Text style={[styles.separator, { fontSize: scaledFont(14, layout.fontScale) }]}>
              ·
            </Text>
            <Text style={[styles.endsAt, { fontSize: scaledFont(14, layout.fontScale) }]}>
              Ends at {endsAt}
            </Text>
          </>
        )}
      </View>

      {/* Ratings badges */}
      <View style={styles.ratingsRow}>
        {media.rating && media.rating > 0 && (
          <View style={[styles.ratingBadge, styles.ratingBadgeTmdb]}>
            <Text
              style={[styles.ratingBadgeText, { fontSize: scaledFont(12, layout.fontScale) }]}
            >
              TMDB {media.rating.toFixed(1)}
            </Text>
          </View>
        )}
        {media.imdb_rating && media.imdb_rating > 0 && (
          <View style={[styles.ratingBadge, styles.ratingBadgeImdb]}>
            <Text
              style={[styles.ratingBadgeText, { fontSize: scaledFont(12, layout.fontScale) }]}
            >
              IMDb {media.imdb_rating.toFixed(1)}
            </Text>
          </View>
        )}
      </View>
    </>
  )

  return (
    <ScrollView style={styles.container}>
      {/* Backdrop with gradient */}
      <View style={[styles.backdropContainer, { height: backdropHeight }]}>
        {backdropUrl ? (
          <Image
            source={{ uri: backdropUrl }}
            style={styles.backdrop}
            resizeMode="cover"
          />
        ) : (
          <View style={[styles.backdrop, styles.backdropPlaceholder]} />
        )}
        {/* Gradient overlay */}
        <View style={styles.gradientOverlay} />

        {/* Lock badge */}
        {media.metadata_locked && (
          <View style={styles.lockBadge}>
            <Lock size={12} color="#fff" />
            <Text style={styles.lockText}>Locked</Text>
          </View>
        )}
      </View>

      {/* Content */}
      <View style={[styles.content, { padding: layout.screenPadding, paddingTop: 0 }]}>
        {/* Poster and Info Row */}
        {layout.sideBySideDetail ? (
          // Tablet landscape / TV: poster left, info right
          <View style={styles.headerSideBySide}>
            <View style={styles.posterContainer}>
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
                  <Text style={styles.posterText}>{media.title?.charAt(0)}</Text>
                </View>
              )}
            </View>

            <View style={styles.infoSideBySide}>
              {renderInfoContent()}

              {/* Action Buttons inline in side-by-side mode */}
              <View style={[styles.actionsRow, { marginTop: 16 }]}>
                <View style={styles.playButtonContainer}>
                  <TouchableOpacity
                    style={[
                      styles.playButton,
                      { paddingVertical: scaledSpacing(14, layout.fontScale) },
                    ]}
                    onPress={handlePlay}
                  >
                    <View style={{ flexDirection: 'row', alignItems: 'center', gap: 8 }}>
                      <Play size={scaledFont(18, layout.fontScale)} color="#fff" />
                      <Text
                        style={[
                          styles.playButtonText,
                          { fontSize: scaledFont(17, layout.fontScale) },
                        ]}
                      >
                        {hasProgress ? 'Resume' : 'Play'}
                      </Text>
                    </View>
                  </TouchableOpacity>
                  {hasProgress && (
                    <TouchableOpacity
                      style={styles.fromBeginningButton}
                      onPress={handlePlayFromBeginning}
                    >
                      <Text
                        style={[
                          styles.fromBeginningText,
                          { fontSize: scaledFont(13, layout.fontScale) },
                        ]}
                      >
                        From Beginning
                      </Text>
                    </TouchableOpacity>
                  )}
                </View>

                <TouchableOpacity
                  style={[
                    styles.actionButton,
                    { width: actionButtonSize, height: actionButtonSize, borderRadius: actionButtonSize / 2 },
                    isWatched && styles.actionButtonActive,
                  ]}
                  onPress={handleToggleWatched}
                >
                  {isWatched ? (
                    <Check size={scaledFont(22, layout.fontScale)} color="#fff" />
                  ) : (
                    <Circle size={scaledFont(22, layout.fontScale)} color="#fff" />
                  )}
                </TouchableOpacity>

                <TouchableOpacity
                  style={[
                    styles.actionButton,
                    { width: actionButtonSize, height: actionButtonSize, borderRadius: actionButtonSize / 2 },
                    isFavorited && styles.actionButtonFavorite,
                  ]}
                  onPress={handleToggleFavorite}
                >
                  <Text style={[styles.actionIcon, { fontSize: scaledFont(22, layout.fontScale) }]}>
                    {isFavorited ? '\u2665' : '\u2661'}
                  </Text>
                </TouchableOpacity>

                <TouchableOpacity
                  style={[
                    styles.menuButton,
                    { width: actionButtonSize, height: actionButtonSize, borderRadius: actionButtonSize / 2 },
                  ]}
                  onPress={() => setShowMenu(true)}
                >
                  <Text style={[styles.menuIcon, { fontSize: scaledFont(24, layout.fontScale) }]}>
                    ⋮
                  </Text>
                </TouchableOpacity>
              </View>
            </View>
          </View>
        ) : (
          // Phone: stacked layout (existing code)
          <>
            <View style={styles.header}>
              <View style={styles.posterContainer}>
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
                    <Text style={styles.posterText}>{media.title?.charAt(0)}</Text>
                  </View>
                )}
              </View>

              <View style={styles.info}>{renderInfoContent()}</View>
            </View>

            {/* Action Buttons (phone layout) */}
            <View style={styles.actionsRow}>
              <View style={styles.playButtonContainer}>
                <TouchableOpacity style={styles.playButton} onPress={handlePlay}>
                  <View style={{ flexDirection: 'row', alignItems: 'center', gap: 8 }}>
                    <Play size={18} color="#fff" />
                    <Text style={styles.playButtonText}>
                      {hasProgress ? 'Resume' : 'Play'}
                    </Text>
                  </View>
                </TouchableOpacity>
                {hasProgress && (
                  <TouchableOpacity
                    style={styles.fromBeginningButton}
                    onPress={handlePlayFromBeginning}
                  >
                    <Text style={styles.fromBeginningText}>From Beginning</Text>
                  </TouchableOpacity>
                )}
              </View>

              <TouchableOpacity
                style={[styles.actionButton, isWatched && styles.actionButtonActive]}
                onPress={handleToggleWatched}
              >
                {isWatched ? <Check size={22} color="#fff" /> : <Circle size={22} color="#fff" />}
              </TouchableOpacity>

              <TouchableOpacity
                style={[styles.actionButton, isFavorited && styles.actionButtonFavorite]}
                onPress={handleToggleFavorite}
              >
                <Text style={styles.actionIcon}>{isFavorited ? '\u2665' : '\u2661'}</Text>
              </TouchableOpacity>

              <TouchableOpacity style={styles.menuButton} onPress={() => setShowMenu(true)}>
                <Text style={styles.menuIcon}>⋮</Text>
              </TouchableOpacity>
            </View>
          </>
        )}

        {/* Trailer section */}
        {youtubeKey && (
          <TouchableOpacity style={styles.trailerSection} onPress={() => setShowTrailer(true)}>
            <View style={styles.trailerThumbnailContainer}>
              <Image
                source={{ uri: `https://img.youtube.com/vi/${youtubeKey}/mqdefault.jpg` }}
                style={styles.trailerThumbnail}
                resizeMode="cover"
              />
              <View style={styles.trailerPlayOverlay}>
                <View style={styles.trailerPlayCircle}>
                  <Play size={20} color="#fff" fill="#fff" />
                </View>
              </View>
            </View>
            <View style={styles.trailerInfo}>
              <Text style={[styles.trailerLabel, { fontSize: scaledFont(10, layout.fontScale) }]}>
                TRAILER
              </Text>
              <Text
                style={[styles.trailerTitle, { fontSize: scaledFont(13, layout.fontScale) }]}
                numberOfLines={1}
              >
                {media.title}
              </Text>
              <Text style={[styles.trailerHint, { fontSize: scaledFont(11, layout.fontScale) }]}>
                Tap to watch on YouTube
              </Text>
            </View>
          </TouchableOpacity>
        )}

        {/* Subtitle management section */}
        <View style={styles.subtitleSection}>
          {subtitles.length > 0 && (
            <TouchableOpacity
              style={styles.subtitleButton}
              onPress={() => setShowSubtitleSelector(true)}
            >
              <View style={styles.subtitleButtonContent}>
                <Text
                  style={[
                    styles.subtitleButtonIcon,
                    { fontSize: scaledFont(14, layout.fontScale) },
                  ]}
                >
                  CC
                </Text>
                <View style={{ flex: 1 }}>
                  <Text
                    style={[
                      styles.subtitleButtonText,
                      { fontSize: scaledFont(14, layout.fontScale) },
                    ]}
                  >
                    Subtitles
                  </Text>
                  <Text
                    style={[
                      styles.subtitleButtonValue,
                      { fontSize: scaledFont(12, layout.fontScale) },
                    ]}
                  >
                    {subtitleLanguage
                      ? LANG_NAMES[subtitleLanguage] || subtitleLanguage
                      : 'Off'}
                  </Text>
                </View>
              </View>
            </TouchableOpacity>
          )}

          <View style={styles.subtitleActions}>
            <TouchableOpacity
              style={styles.subtitleActionButton}
              onPress={() => setShowSubtitleSearch(true)}
            >
              <Search size={scaledFont(16, layout.fontScale)} color="rgba(255, 255, 255, 0.6)" />
              <Text
                style={[
                  styles.subtitleActionText,
                  { fontSize: scaledFont(13, layout.fontScale) },
                ]}
              >
                Search Subtitles
              </Text>
            </TouchableOpacity>

            {subtitles.filter((s) => !s.is_image).length > 0 && (
              <TouchableOpacity
                style={styles.subtitleActionButton}
                onPress={() => setShowSubtitleTranslate(true)}
              >
                <Languages
                  size={scaledFont(16, layout.fontScale)}
                  color="rgba(255, 255, 255, 0.6)"
                />
                <Text
                  style={[
                    styles.subtitleActionText,
                    { fontSize: scaledFont(13, layout.fontScale) },
                  ]}
                >
                  Translate
                </Text>
              </TouchableOpacity>
            )}
          </View>
        </View>

        {/* Progress bar */}
        {progressPercent > 0 && (
          <View style={styles.progressSection}>
            <View style={styles.progressBar}>
              <View style={[styles.progressFill, { width: `${progressPercent}%` }]} />
            </View>
            <Text
              style={[styles.progressText, { fontSize: scaledFont(12, layout.fontScale) }]}
            >
              {Math.floor((progress?.position || 0) / 60)}m watched
            </Text>
          </View>
        )}

        {/* Overview */}
        {media.overview && (
          <View style={styles.section}>
            <Text
              style={[styles.sectionTitle, { fontSize: scaledFont(18, layout.fontScale) }]}
            >
              Overview
            </Text>
            <Text
              style={[
                styles.overview,
                {
                  fontSize: scaledFont(14, layout.fontScale),
                  lineHeight: scaledFont(22, layout.fontScale),
                },
              ]}
            >
              {media.overview}
            </Text>
          </View>
        )}

        {/* Technical Specs */}
        {primaryFile && (
          <View style={styles.section}>
            <Text
              style={[styles.sectionTitle, { fontSize: scaledFont(18, layout.fontScale) }]}
            >
              Technical Info
            </Text>
            <View style={styles.specsGrid}>
              {primaryFile.width && primaryFile.height && (
                <View style={styles.specItem}>
                  <Text
                    style={[styles.specLabel, { fontSize: scaledFont(12, layout.fontScale) }]}
                  >
                    Resolution
                  </Text>
                  <Text
                    style={[styles.specValue, { fontSize: scaledFont(14, layout.fontScale) }]}
                  >
                    {primaryFile.width}x{primaryFile.height}
                  </Text>
                </View>
              )}
              {primaryFile.video_codec && (
                <View style={styles.specItem}>
                  <Text
                    style={[styles.specLabel, { fontSize: scaledFont(12, layout.fontScale) }]}
                  >
                    Video
                  </Text>
                  <Text
                    style={[styles.specValue, { fontSize: scaledFont(14, layout.fontScale) }]}
                  >
                    {primaryFile.video_codec}
                  </Text>
                </View>
              )}
              {primaryFile.audio_codec && (
                <View style={styles.specItem}>
                  <Text
                    style={[styles.specLabel, { fontSize: scaledFont(12, layout.fontScale) }]}
                  >
                    Audio
                  </Text>
                  <Text
                    style={[styles.specValue, { fontSize: scaledFont(14, layout.fontScale) }]}
                  >
                    {primaryFile.audio_codec}
                  </Text>
                </View>
              )}
              {primaryFile.container && (
                <View style={styles.specItem}>
                  <Text
                    style={[styles.specLabel, { fontSize: scaledFont(12, layout.fontScale) }]}
                  >
                    Container
                  </Text>
                  <Text
                    style={[styles.specValue, { fontSize: scaledFont(14, layout.fontScale) }]}
                  >
                    {primaryFile.container}
                  </Text>
                </View>
              )}
              {primaryFile.file_size && primaryFile.file_size > 0 && (
                <View style={styles.specItem}>
                  <Text
                    style={[styles.specLabel, { fontSize: scaledFont(12, layout.fontScale) }]}
                  >
                    Size
                  </Text>
                  <Text
                    style={[styles.specValue, { fontSize: scaledFont(14, layout.fontScale) }]}
                  >
                    {formatFileSize(primaryFile.file_size)}
                  </Text>
                </View>
              )}
            </View>
          </View>
        )}

        {/* Media Files */}
        {mediaFiles?.files && mediaFiles.files.length > 0 && (
          <View style={styles.section}>
            <Text
              style={[styles.sectionTitle, { fontSize: scaledFont(18, layout.fontScale) }]}
            >
              Files
            </Text>
            {mediaFiles.files.map((file) => (
              <View key={file.id} style={styles.fileItem}>
                <Text
                  style={[styles.fileName, { fontSize: scaledFont(14, layout.fontScale) }]}
                >
                  {file.file_path.split('/').pop()}
                </Text>
                <Text
                  style={[styles.fileInfo, { fontSize: scaledFont(12, layout.fontScale) }]}
                >
                  {file.width}x{file.height} · {formatFileSize(file.file_size)} ·{' '}
                  {file.container}
                </Text>
              </View>
            ))}
          </View>
        )}
      </View>

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

      {/* Play Options Bottom Sheet */}
      <PlayOptionsBottomSheet
        visible={showPlayOptions}
        title={media.title || 'Movie'}
        onPlay={handlePlayWithOptions}
        onClose={() => setShowPlayOptions(false)}
      />

      {/* Metadata Editor */}
      {showMetadataEditor && media && (
        <MetadataEditor
          type="media"
          item={media}
          genres={(media as any).genres}
          onSave={handleSaveMetadata}
          onRefresh={() => refreshMetadata.mutate()}
          isSaving={editMetadata.isPending}
          isRefreshing={refreshMetadata.isPending}
          onClose={() => setShowMetadataEditor(false)}
        />
      )}

      {/* Subtitle Language Selector Modal */}
      <Modal
        visible={showSubtitleSelector}
        transparent
        animationType="fade"
        onRequestClose={() => setShowSubtitleSelector(false)}
      >
        <TouchableOpacity
          style={styles.menuOverlay}
          activeOpacity={1}
          onPress={() => setShowSubtitleSelector(false)}
        >
          <View style={styles.menuContainer}>
            <Text style={styles.menuTitle}>Subtitles</Text>

            <TouchableOpacity
              style={[styles.menuItem, !subtitleLanguage && styles.menuItemActive]}
              onPress={() => {
                setSubtitleLanguage(null, null)
                setShowSubtitleSelector(false)
              }}
            >
              <Text
                style={[styles.menuItemText, !subtitleLanguage && styles.menuItemTextActive]}
              >
                Off
              </Text>
            </TouchableOpacity>

            {subtitles
              .filter((s) => !s.is_image)
              .map((subtitle) => (
                <TouchableOpacity
                  key={subtitle.id}
                  style={[
                    styles.menuItem,
                    subtitleLanguage === subtitle.language && styles.menuItemActive,
                  ]}
                  onPress={() => {
                    setSubtitleLanguage(subtitle.language, subtitle.id)
                    setShowSubtitleSelector(false)
                  }}
                >
                  <Text
                    style={[
                      styles.menuItemText,
                      subtitleLanguage === subtitle.language && styles.menuItemTextActive,
                    ]}
                  >
                    {LANG_NAMES[subtitle.language] || subtitle.language} (
                    {subtitle.format.toUpperCase()})
                  </Text>
                </TouchableOpacity>
              ))}

            <TouchableOpacity
              style={styles.menuItem}
              onPress={() => setShowSubtitleSelector(false)}
            >
              <Text style={styles.menuItemText}>Cancel</Text>
            </TouchableOpacity>
          </View>
        </TouchableOpacity>
      </Modal>

      {/* YouTube Trailer Player */}
      <YouTubePlayer
        videoId={youtubeKey || ''}
        visible={showTrailer}
        onClose={() => setShowTrailer(false)}
      />

      {/* Subtitle Search Modal */}
      <SubtitleSearchModal
        mediaId={id}
        defaultLang={subtitleLanguage}
        visible={showSubtitleSearch}
        onClose={() => setShowSubtitleSearch(false)}
        onSubtitleDownloaded={() => {
          // Subtitles list will auto-refresh via query invalidation
        }}
      />

      {/* Subtitle Translate Modal */}
      <SubtitleTranslate
        subtitles={subtitles}
        visible={showSubtitleTranslate}
        onClose={() => setShowSubtitleTranslate(false)}
        onTranslated={() => {
          // Subtitles list will auto-refresh via query invalidation
        }}
      />
    </ScrollView>
  )
}

function formatFileSize(bytes: number): string {
  if (bytes < 1024) return `${bytes} B`
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`
  if (bytes < 1024 * 1024 * 1024) return `${(bytes / (1024 * 1024)).toFixed(1)} MB`
  return `${(bytes / (1024 * 1024 * 1024)).toFixed(2)} GB`
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
  backdropContainer: {
    position: 'relative',
  },
  backdrop: {
    width: '100%',
    height: '100%',
  },
  backdropPlaceholder: {
    backgroundColor: '#1f1f1f',
  },
  gradientOverlay: {
    position: 'absolute',
    left: 0,
    right: 0,
    bottom: 0,
    height: '70%',
    backgroundColor: 'transparent',
  },
  lockBadge: {
    position: 'absolute',
    top: 16,
    right: 16,
    flexDirection: 'row',
    alignItems: 'center',
    backgroundColor: 'rgba(0, 0, 0, 0.7)',
    paddingHorizontal: 10,
    paddingVertical: 6,
    borderRadius: 6,
    gap: 4,
  },
  lockText: {
    color: '#fff',
    fontSize: 12,
    fontWeight: '600',
  },
  content: {
    marginTop: -80,
    paddingTop: 0,
  },
  // Phone: stacked layout
  header: {
    flexDirection: 'row',
    marginBottom: 20,
  },
  // Tablet: side-by-side layout
  headerSideBySide: {
    flexDirection: 'row',
    marginBottom: 20,
    gap: 20,
  },
  posterContainer: {
    shadowColor: '#000',
    shadowOffset: { width: 0, height: 4 },
    shadowOpacity: 0.4,
    shadowRadius: 10,
    elevation: 10,
  },
  poster: {
    borderRadius: 10,
    backgroundColor: '#1f1f1f',
  },
  posterPlaceholder: {
    justifyContent: 'center',
    alignItems: 'center',
  },
  posterText: {
    fontSize: 56,
    fontWeight: 'bold',
    color: '#444',
  },
  info: {
    flex: 1,
    marginLeft: 16,
    justifyContent: 'flex-end',
    paddingBottom: 8,
  },
  infoSideBySide: {
    flex: 1,
    justifyContent: 'flex-end',
    paddingBottom: 8,
  },
  title: {
    fontSize: 24,
    fontWeight: 'bold',
    color: '#fff',
    marginBottom: 8,
  },
  metadataLine: {
    flexDirection: 'row',
    alignItems: 'center',
    marginBottom: 8,
    flexWrap: 'wrap',
  },
  year: {
    fontSize: 14,
    color: '#aaa',
  },
  separator: {
    fontSize: 14,
    color: '#666',
    marginHorizontal: 6,
  },
  runtime: {
    fontSize: 14,
    color: '#aaa',
  },
  endsAt: {
    fontSize: 14,
    color: '#e50914',
  },
  ratingsRow: {
    flexDirection: 'row',
    gap: 8,
    flexWrap: 'wrap',
  },
  ratingBadge: {
    paddingHorizontal: 8,
    paddingVertical: 4,
    borderRadius: 4,
  },
  ratingBadgeTmdb: {
    backgroundColor: '#01b4e4',
  },
  ratingBadgeImdb: {
    backgroundColor: '#f5c518',
  },
  ratingBadgeText: {
    fontSize: 12,
    fontWeight: '600',
    color: '#000',
  },
  actionsRow: {
    flexDirection: 'row',
    gap: 12,
    marginBottom: 4,
    alignItems: 'center',
  },
  playButtonContainer: {
    flex: 1,
    flexDirection: 'row',
    gap: 8,
  },
  playButton: {
    flex: 1,
    backgroundColor: '#e50914',
    borderRadius: 8,
    paddingVertical: 14,
    alignItems: 'center',
  },
  playButtonText: {
    color: '#fff',
    fontSize: 17,
    fontWeight: '600',
  },
  fromBeginningButton: {
    paddingHorizontal: 12,
    paddingVertical: 8,
    borderRadius: 6,
    backgroundColor: '#2a2a2a',
    justifyContent: 'center',
  },
  fromBeginningText: {
    color: '#aaa',
    fontSize: 13,
    fontWeight: '500',
  },
  actionButton: {
    width: 50,
    height: 50,
    borderRadius: 25,
    backgroundColor: '#2a2a2a',
    justifyContent: 'center',
    alignItems: 'center',
    borderWidth: 1,
    borderColor: '#444',
  },
  actionButtonActive: {
    backgroundColor: '#1a5c1a',
    borderColor: '#2a8c2a',
  },
  actionButtonFavorite: {
    backgroundColor: '#5c1a4a',
    borderColor: '#e50914',
  },
  actionIcon: {
    fontSize: 22,
    color: '#fff',
  },
  menuButton: {
    width: 50,
    height: 50,
    borderRadius: 25,
    backgroundColor: '#2a2a2a',
    justifyContent: 'center',
    alignItems: 'center',
    borderWidth: 1,
    borderColor: '#444',
  },
  menuIcon: {
    fontSize: 24,
    color: '#fff',
  },
  trailerSection: {
    flexDirection: 'row',
    backgroundColor: '#1a1a1a',
    borderRadius: 12,
    overflow: 'hidden',
    marginTop: 12,
    borderWidth: 1,
    borderColor: 'rgba(255, 255, 255, 0.1)',
  },
  trailerThumbnailContainer: {
    width: 140,
    height: 80,
    position: 'relative',
  },
  trailerThumbnail: {
    width: '100%',
    height: '100%',
  },
  trailerPlayOverlay: {
    ...StyleSheet.absoluteFillObject,
    justifyContent: 'center',
    alignItems: 'center',
    backgroundColor: 'rgba(0, 0, 0, 0.3)',
  },
  trailerPlayCircle: {
    width: 36,
    height: 36,
    borderRadius: 18,
    backgroundColor: 'rgba(229, 9, 20, 0.9)',
    justifyContent: 'center',
    alignItems: 'center',
    paddingLeft: 2,
  },
  trailerInfo: {
    flex: 1,
    padding: 10,
    justifyContent: 'center',
  },
  trailerLabel: {
    fontSize: 10,
    fontWeight: '700',
    color: '#e50914',
    letterSpacing: 1,
    marginBottom: 2,
  },
  trailerTitle: {
    fontSize: 13,
    fontWeight: '600',
    color: '#fff',
  },
  trailerHint: {
    fontSize: 11,
    color: 'rgba(255, 255, 255, 0.35)',
    marginTop: 2,
  },
  subtitleSection: {
    marginTop: 16,
    gap: 8,
  },
  subtitleButton: {
    backgroundColor: '#1a1a1a',
    borderRadius: 10,
    paddingHorizontal: 16,
    paddingVertical: 12,
    borderWidth: 1,
    borderColor: 'rgba(255, 255, 255, 0.1)',
  },
  subtitleButtonContent: {
    flexDirection: 'row',
    alignItems: 'center',
    gap: 12,
  },
  subtitleButtonIcon: {
    fontSize: 14,
    fontWeight: '800',
    color: '#fff',
    backgroundColor: 'rgba(255, 255, 255, 0.15)',
    paddingHorizontal: 6,
    paddingVertical: 2,
    borderRadius: 4,
    overflow: 'hidden',
  },
  subtitleButtonText: {
    fontSize: 14,
    fontWeight: '600',
    color: '#fff',
  },
  subtitleButtonValue: {
    fontSize: 12,
    color: 'rgba(255, 255, 255, 0.4)',
    marginTop: 1,
  },
  subtitleActions: {
    flexDirection: 'row',
    gap: 8,
  },
  subtitleActionButton: {
    flex: 1,
    flexDirection: 'row',
    alignItems: 'center',
    justifyContent: 'center',
    gap: 6,
    backgroundColor: 'rgba(255, 255, 255, 0.08)',
    paddingVertical: 10,
    borderRadius: 8,
    borderWidth: 1,
    borderColor: 'rgba(255, 255, 255, 0.1)',
  },
  subtitleActionText: {
    fontSize: 13,
    fontWeight: '500',
    color: 'rgba(255, 255, 255, 0.6)',
  },
  progressSection: {
    marginTop: 16,
    marginBottom: 20,
  },
  progressBar: {
    height: 4,
    backgroundColor: '#333',
    borderRadius: 2,
    overflow: 'hidden',
    marginBottom: 8,
  },
  progressFill: {
    height: '100%',
    backgroundColor: '#e50914',
    borderRadius: 2,
  },
  progressText: {
    fontSize: 12,
    color: '#888',
    textAlign: 'center',
  },
  section: {
    marginBottom: 24,
  },
  sectionTitle: {
    fontSize: 18,
    fontWeight: '600',
    color: '#fff',
    marginBottom: 12,
  },
  overview: {
    fontSize: 14,
    color: '#ccc',
    lineHeight: 22,
  },
  specsGrid: {
    flexDirection: 'row',
    flexWrap: 'wrap',
    gap: 16,
  },
  specItem: {
    minWidth: 100,
  },
  specLabel: {
    fontSize: 12,
    color: '#666',
    marginBottom: 4,
  },
  specValue: {
    fontSize: 14,
    color: '#fff',
    fontWeight: '500',
  },
  fileItem: {
    backgroundColor: '#1f1f1f',
    borderRadius: 8,
    padding: 12,
    marginBottom: 8,
  },
  fileName: {
    fontSize: 14,
    color: '#fff',
    marginBottom: 4,
  },
  fileInfo: {
    fontSize: 12,
    color: '#888',
  },
  menuOverlay: {
    flex: 1,
    backgroundColor: 'rgba(0, 0, 0, 0.7)',
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
  menuItemActive: {
    backgroundColor: '#e50914',
  },
  menuItemTextActive: {
    color: '#fff',
    fontWeight: '600',
  },
})
