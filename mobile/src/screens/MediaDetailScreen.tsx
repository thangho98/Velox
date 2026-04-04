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
import { LinearGradient } from 'expo-linear-gradient'
import {
  Play,
  Check,
  Link,
  Pencil,
  Lock,
  Star,
  Heart,
} from 'lucide-react-native'
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
import { MetadataEditor } from '../components/MetadataEditor'
import { YouTubePlayer } from '../components/YouTubePlayer'
import { Toast } from '../components/Toast'
import { useResponsiveLayout, scaledFont } from '../lib/responsive'
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
  const { subtitleLanguage, subtitleTrackId, setSubtitleLanguage } = usePlayerStore()

  // Metadata editing
  const [showMetadataEditor, setShowMetadataEditor] = useState(false)
  const editMetadata = useEditMediaMetadata(id)
  const refreshMetadata = useRefreshMetadata(id)

  const [isFavorited, setIsFavorited] = useState(false)
  const [isWatched, setIsWatched] = useState(false)
  const [showMenu, setShowMenu] = useState(false)
  const [showTrailer, setShowTrailer] = useState(false)
  const [showSubtitleSelector, setShowSubtitleSelector] = useState(false)

  // Responsive dimensions
  const posterWidth = layout.sideBySideDetail ? 200 : 200
  const posterHeight = layout.sideBySideDetail ? 300 : 300
  const backdropHeight = layout.height

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

  // Get primary file for technical specs and duration
  const primaryFile = mediaFiles?.files?.find((f) => f.is_primary) || mediaFiles?.files?.[0]
  const durationSeconds = primaryFile?.duration || (media.duration ? media.duration : 0)

  // Calculate "Ends at" time (matching web logic)
  const remainingSeconds = durationSeconds > 0 ? durationSeconds - (progress?.position || 0) : 0
  const endsAt =
    remainingSeconds > 0
      ? new Date(Date.now() + remainingSeconds * 1000).toLocaleTimeString([], {
          hour: 'numeric',
          minute: '2-digit',
        })
      : null

  // Calculate progress percentage
  const progressPercent =
    progress && durationSeconds > 0
      ? Math.min(100, (progress.position / durationSeconds) * 100)
      : 0
  const hasProgress = progressPercent > 0

  const handlePlay = () => {
    navigation.navigate('Episode', { id, seriesId: 0 })
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

  return (
    <View style={styles.container}>
      {/* Fixed backdrop — stays behind scroll */}
      <View style={styles.fixedBackdrop}>
        {backdropUrl ? (
          <Image source={{ uri: backdropUrl }} style={styles.backdrop} resizeMode="cover" />
        ) : (
          <View style={[styles.backdrop, styles.backdropPlaceholder]} />
        )}
        {/* Heavy overlay matching web (from-netflix-black via-netflix-black/80 to-netflix-black/30) */}
        <LinearGradient
          colors={[
            'rgba(20,20,20,0.45)',
            'rgba(20,20,20,0.65)',
            'rgba(20,20,20,0.88)',
            '#141414',
          ]}
          locations={[0, 0.25, 0.55, 0.8]}
          style={StyleSheet.absoluteFillObject}
        />
        {/* Top vignette */}
        <LinearGradient
          colors={['rgba(0,0,0,0.7)', 'transparent']}
          locations={[0, 0.5]}
          style={{ position: 'absolute', top: 0, left: 0, right: 0, height: '25%' }}
        />

        {/* Lock badge */}
        {media.metadata_locked && (
          <View style={styles.lockBadge}>
            <Lock size={12} color="#fff" />
            <Text style={styles.lockText}>Locked</Text>
          </View>
        )}
      </View>

      {/* Scrollable content over backdrop */}
      <ScrollView
        style={styles.scrollView}
        bounces={false}
        showsVerticalScrollIndicator={false}
      >
        {/* Spacer — poster sits in upper portion of backdrop (like web py-24) */}
        <View style={{ height: layout.height * 0.15 }} />

        {/* Content */}
        <View style={[styles.contentArea, { paddingHorizontal: layout.screenPadding }]}>
        {layout.sideBySideDetail ? (
          // ── Tablet: poster left, info right ──
          <View style={styles.sideBySideRow}>
            <View style={styles.posterShadow}>
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
                  <Text style={styles.posterInitial}>{media.title?.charAt(0)}</Text>
                </View>
              )}
            </View>

            <View style={styles.sideBySideInfo}>
              {renderTitle()}
              {renderMetadata()}
              {renderTechSpecs()}
              {renderOverview()}
              {renderActions()}
            </View>
          </View>
        ) : (
          // ── Phone: centered layout (matching web) ──
          <>
            {/* Poster centered */}
            <View style={styles.posterCenter}>
              <View style={styles.posterShadow}>
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
                    <Text style={styles.posterInitial}>{media.title?.charAt(0)}</Text>
                  </View>
                )}
              </View>
            </View>

            {renderTitle()}
            {renderMetadata()}
            {renderTechSpecs()}
            {renderOverview()}
            {renderActions()}
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

        {/* Progress bar */}
        {hasProgress && (
          <View style={styles.progressSection}>
            <View style={styles.progressBar}>
              <View style={[styles.progressFill, { width: `${progressPercent}%` }]} />
            </View>
            <Text
              style={[styles.progressText, { fontSize: scaledFont(12, layout.fontScale) }]}
            >
              {progress?.completed
                ? 'Watched'
                : `${formatDuration(Math.max(0, remainingSeconds))} remaining`}
            </Text>
          </View>
        )}

        {/* Bottom spacing */}
        <View style={{ height: 40 }} />
      </View>

      {/* ── Modals ── */}

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
              <Text style={styles.menuItemIcon}>{'\u{1F504}'}</Text>
              <Text style={styles.menuItemText}>Refresh Metadata</Text>
            </TouchableOpacity>

            <TouchableOpacity style={styles.menuItem} onPress={() => setShowMenu(false)}>
              <Text style={styles.menuItemText}>Cancel</Text>
            </TouchableOpacity>
          </View>
        </TouchableOpacity>
      </Modal>

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
              .map((subtitle) => {
                const isActive = subtitleTrackId === subtitle.id
                return (
                  <TouchableOpacity
                    key={subtitle.id}
                    style={[styles.menuItem, isActive && styles.menuItemActive]}
                    onPress={() => {
                      setSubtitleLanguage(subtitle.language, subtitle.id)
                      setShowSubtitleSelector(false)
                    }}
                  >
                    <Text
                      style={[styles.menuItemText, isActive && styles.menuItemTextActive]}
                    >
                      {subtitle.label ||
                        `${LANG_NAMES[subtitle.language] || subtitle.language} (${subtitle.format.toUpperCase()})`}
                    </Text>
                  </TouchableOpacity>
                )
              })}

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

      </ScrollView>
    </View>
  )

  // ── Render helpers (inline, use closure) ──

  function renderTitle() {
    return (
      <Text
        style={[styles.title, { fontSize: scaledFont(28, layout.fontScale) }]}
      >
        {media!.title}
      </Text>
    )
  }

  function renderMetadata() {
    const items: string[] = []
    if (year) items.push(year)
    if (durationSeconds > 0) items.push(formatDuration(durationSeconds))

    return (
      <View style={styles.metadataRow}>
        {items.map((item, i) => (
          <View key={i} style={styles.metadataItemRow}>
            {i > 0 && <Text style={styles.metadataSep}>{' \u00b7 '}</Text>}
            <Text style={[styles.metadataText, { fontSize: scaledFont(14, layout.fontScale) }]}>
              {item}
            </Text>
          </View>
        ))}
        {endsAt && (
          <View style={styles.metadataItemRow}>
            {items.length > 0 && <Text style={styles.metadataSep}>{' \u00b7 '}</Text>}
            <Text
              style={[
                styles.metadataText,
                styles.endsAtText,
                { fontSize: scaledFont(14, layout.fontScale) },
              ]}
            >
              {'Ends at '}
              {endsAt}
            </Text>
          </View>
        )}
        {media!.rating != null && media!.rating > 0 && (
          <View style={styles.metadataItemRow}>
            <Text style={styles.metadataSep}>{' '}</Text>
            <Star size={14} color="#eab308" fill="#eab308" />
            <Text
              style={[
                styles.metadataText,
                { fontSize: scaledFont(14, layout.fontScale), marginLeft: 3 },
              ]}
            >
              {media!.rating.toFixed(1)}
            </Text>
          </View>
        )}
      </View>
    )
  }

  function renderTechSpecs() {
    if (!primaryFile) return null
    const specs: Array<{ label: string; value: string }> = []

    if (primaryFile.video_codec) {
      const res = primaryFile.height > 0 ? `${primaryFile.height}p ` : ''
      specs.push({ label: 'Video', value: `${res}${primaryFile.video_codec.toUpperCase()}` })
    }
    if (primaryFile.audio_codec) {
      specs.push({ label: 'Audio', value: primaryFile.audio_codec.toUpperCase() })
    }
    if (primaryFile.container) {
      specs.push({ label: 'Container', value: primaryFile.container.toUpperCase() })
    }
    if (primaryFile.file_size && primaryFile.file_size > 0) {
      specs.push({ label: 'Size', value: formatFileSize(primaryFile.file_size) })
    }

    if (specs.length === 0) return null

    return (
      <View style={styles.techSpecsRow}>
        {specs.map((spec, i) => (
          <View key={i} style={styles.techSpecItem}>
            <Text style={styles.techSpecLabel}>{spec.label}</Text>
            <Text style={styles.techSpecValue}>{' '}{spec.value}</Text>
          </View>
        ))}
      </View>
    )
  }

  function renderOverview() {
    if (!media!.overview) return null
    return (
      <Text
        style={[
          styles.overview,
          {
            fontSize: scaledFont(14, layout.fontScale),
            lineHeight: scaledFont(22, layout.fontScale),
          },
        ]}
      >
        {media!.overview}
      </Text>
    )
  }

  function renderActions() {
    return (
      <View style={styles.actionsRow}>
        {/* Small Play button (matching web) */}
        <TouchableOpacity style={styles.playButton} onPress={handlePlay}>
          <Play size={16} color="#fff" fill="#fff" />
          <Text style={styles.playButtonText}>
            {hasProgress ? 'Resume' : 'Play'}
          </Text>
        </TouchableOpacity>

        {/* Icon actions */}
        <TouchableOpacity
          style={[styles.iconButton, isWatched && styles.actionButtonActive]}
          onPress={handleToggleWatched}
        >
          <Check size={20} color={isWatched ? '#fff' : '#aaa'} />
        </TouchableOpacity>

        <TouchableOpacity
          style={[styles.iconButton, isFavorited && styles.actionButtonFavorite]}
          onPress={handleToggleFavorite}
        >
          <Heart
            size={20}
            color={isFavorited ? '#ec4899' : '#aaa'}
            fill={isFavorited ? '#ec4899' : 'none'}
          />
        </TouchableOpacity>

        <TouchableOpacity style={styles.iconButton} onPress={() => setShowMenu(true)}>
          <Text style={styles.menuDots}>{'\u22EE'}</Text>
        </TouchableOpacity>

        {/* Inline subtitle selector (matching web) */}
        {subtitles.length > 0 && (
          <View style={styles.inlineSubtitleRow}>
            <Text style={styles.inlineSubtitleLabel}>Subtitles</Text>
            <TouchableOpacity
              style={styles.inlineSubtitlePicker}
              onPress={() => setShowSubtitleSelector(true)}
            >
              <Text style={styles.inlineSubtitleValue} numberOfLines={1}>
                {(() => {
                  if (!subtitleLanguage) return 'Off'
                  const active = subtitles.find((s) => s.id === subtitleTrackId)
                  return active?.label || LANG_NAMES[subtitleLanguage] || subtitleLanguage
                })()}
              </Text>
              <Text style={styles.inlineSubtitleArrow}>{'\u25BE'}</Text>
            </TouchableOpacity>
          </View>
        )}
      </View>
    )
  }
}

// ── Helpers ──

function formatDuration(seconds: number): string {
  const hours = Math.floor(seconds / 3600)
  const mins = Math.floor((seconds % 3600) / 60)
  if (hours > 0) return `${hours}h ${mins}m`
  return `${mins}m`
}

function formatFileSize(bytes: number): string {
  const gb = bytes / (1024 * 1024 * 1024)
  if (gb >= 1) return `${gb.toFixed(2)} GB`
  const mb = bytes / (1024 * 1024)
  return `${mb.toFixed(1)} MB`
}

// ── Styles ──

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

  // Fixed backdrop (behind scroll)
  fixedBackdrop: {
    position: 'absolute',
    top: 0,
    left: 0,
    right: 0,
    bottom: 0,
  },
  scrollView: {
    flex: 1,
  },

  // Backdrop
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

  // Content area
  contentArea: {
    paddingTop: 0,
  },

  // Poster
  posterCenter: {
    alignItems: 'center',
    marginBottom: 20,
  },
  posterShadow: {
    shadowColor: '#000',
    shadowOffset: { width: 0, height: 8 },
    shadowOpacity: 0.5,
    shadowRadius: 16,
    elevation: 12,
  },
  poster: {
    borderRadius: 10,
    backgroundColor: '#1f1f1f',
  },
  posterPlaceholder: {
    justifyContent: 'center',
    alignItems: 'center',
  },
  posterInitial: {
    fontSize: 56,
    fontWeight: 'bold',
    color: '#444',
  },

  // Side-by-side (tablet)
  sideBySideRow: {
    flexDirection: 'row',
    gap: 24,
    marginBottom: 20,
  },
  sideBySideInfo: {
    flex: 1,
    justifyContent: 'flex-end',
    paddingBottom: 8,
  },

  // Title
  title: {
    fontSize: 28,
    fontWeight: 'bold',
    color: '#fff',
    marginBottom: 8,
  },
  textCenter: {
    textAlign: 'center',
  },

  // Metadata line (year · duration · ends at · rating)
  metadataRow: {
    flexDirection: 'row',
    alignItems: 'center',
    flexWrap: 'wrap',
    marginBottom: 10,
    gap: 2,
  },
  metadataItemRow: {
    flexDirection: 'row',
    alignItems: 'center',
  },
  metadataText: {
    fontSize: 14,
    color: '#aaa',
  },
  metadataSep: {
    fontSize: 14,
    color: '#555',
  },
  endsAtText: {
    color: '#aaa',
  },

  // Tech specs inline (Video · Audio · Container · Size)
  techSpecsRow: {
    flexDirection: 'row',
    flexWrap: 'wrap',
    gap: 16,
    marginBottom: 16,
  },
  techSpecItem: {
    flexDirection: 'row',
  },
  techSpecLabel: {
    fontSize: 13,
    color: '#666',
  },
  techSpecValue: {
    fontSize: 13,
    color: '#ccc',
    fontWeight: '500',
  },

  // Overview
  overview: {
    fontSize: 14,
    color: '#ccc',
    lineHeight: 22,
    marginBottom: 20,
  },

  // Actions (compact row matching web)
  actionsRow: {
    flexDirection: 'row',
    flexWrap: 'wrap',
    gap: 10,
    alignItems: 'center',
    marginBottom: 16,
  },
  playButton: {
    flexDirection: 'row',
    alignItems: 'center',
    gap: 6,
    backgroundColor: '#e50914',
    borderRadius: 6,
    paddingHorizontal: 20,
    paddingVertical: 10,
  },
  playButtonText: {
    color: '#fff',
    fontSize: 14,
    fontWeight: '600',
  },
  iconButton: {
    padding: 6,
  },
  actionButtonActive: {
    opacity: 1,
  },
  actionButtonFavorite: {
    opacity: 1,
  },
  menuDots: {
    fontSize: 22,
    color: '#aaa',
  },
  // Inline subtitle selector (matching web dropdown)
  inlineSubtitleRow: {
    flexDirection: 'row',
    alignItems: 'center',
    gap: 8,
    marginLeft: 4,
  },
  inlineSubtitleLabel: {
    fontSize: 13,
    color: '#888',
  },
  inlineSubtitlePicker: {
    flexDirection: 'row',
    alignItems: 'center',
    backgroundColor: '#2a2a2a',
    borderRadius: 20,
    paddingHorizontal: 14,
    paddingVertical: 8,
    gap: 6,
  },
  inlineSubtitleValue: {
    fontSize: 13,
    color: '#fff',
    maxWidth: 140,
  },
  inlineSubtitleArrow: {
    fontSize: 12,
    color: '#888',
  },

  // Trailer
  trailerSection: {
    flexDirection: 'row',
    backgroundColor: '#1a1a1a',
    borderRadius: 12,
    overflow: 'hidden',
    marginBottom: 12,
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
    position: 'absolute',
    top: 0,
    left: 0,
    right: 0,
    bottom: 0,
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

  // Progress
  progressSection: {
    marginBottom: 16,
  },
  progressBar: {
    height: 3,
    backgroundColor: '#333',
    borderRadius: 2,
    overflow: 'hidden',
    marginBottom: 6,
  },
  progressFill: {
    height: '100%',
    backgroundColor: '#22c55e',
    borderRadius: 2,
  },
  progressText: {
    fontSize: 12,
    color: '#888',
  },

  // Modals
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
