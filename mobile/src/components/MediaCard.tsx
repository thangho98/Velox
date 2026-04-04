import { View, Text, Image, TouchableOpacity, StyleSheet } from 'react-native'
import { Star } from 'lucide-react-native'
import type { MediaListItem, ContinueWatchingItem } from '@velox/shared/types'
import { mediaImage } from '@velox/shared/lib'
import { cardWidth as calcCardWidth, useResponsiveLayout, scaledFont } from '../lib/responsive'

interface MediaCardProps {
  item: MediaListItem
  onPress?: () => void
  onLongPress?: () => void
  size?: 'small' | 'medium' | 'large'
  /** Optional progress data for continue watching cards */
  progress?: {
    position: number
    duration: number
    completed?: boolean
  }
  /** Show badge (Movie/Series) on poster */
  showBadge?: boolean
  /** Show rating badge on poster */
  showRating?: boolean
  /** Continue watching item data */
  continueWatching?: ContinueWatchingItem
  /** Grid columns (from responsive layout) */
  columns?: number
  /** Screen width override */
  containerWidth?: number
  /** Card gap override */
  gap?: number
  /** Screen padding override */
  padding?: number
}

// Fallback dimensions when no responsive props are given
const FALLBACK_WIDTHS = { small: 110, medium: 160, large: 340 }

export function MediaCard({
  item,
  onPress,
  onLongPress,
  size = 'medium',
  progress,
  showBadge = false,
  showRating = false,
  continueWatching,
  columns,
  containerWidth,
  gap,
  padding,
}: MediaCardProps) {
  const layout = useResponsiveLayout()

  // Calculate card width dynamically from responsive data
  const cols = columns ?? layout.gridColumns
  const sw = containerWidth ?? layout.width
  const g = gap ?? layout.cardGap
  const p = padding ?? layout.screenPadding

  let cw: number
  if (size === 'large') {
    cw = sw - p * 2
  } else {
    cw = calcCardWidth(cols, sw, g, p)
  }

  const cardWidth = cw || FALLBACK_WIDTHS[size]
  const cardHeight =
    size === 'large' ? cardWidth * 0.6 : size === 'small' ? cardWidth * 1.5 : cardWidth * 1.4

  // Use continueWatching data if available, otherwise use item data
  const imagePath = continueWatching?.poster_path || item.poster_path
  const imageUrl = mediaImage(imagePath, 'w500')

  // Calculate progress percentage
  const progressPercent = progress
    ? (progress.position / progress.duration) * 100
    : continueWatching && continueWatching.duration > 0
    ? (continueWatching.position / continueWatching.duration) * 100
    : 0

  // Calculate remaining time in minutes
  const remainingMinutes =
    continueWatching && continueWatching.duration > 0
      ? Math.ceil((continueWatching.duration - continueWatching.position) / 60)
      : progress && progress.duration > 0
      ? Math.ceil((progress.duration - progress.position) / 60)
      : 0

  // Format display title: for episodes show "S1E1 · Series Name" format
  const displayTitle = continueWatching?.media_type === 'episode' && continueWatching.series_title
    ? `S${continueWatching.season_number}E${continueWatching.episode_number} · ${continueWatching.series_title}`
    : continueWatching?.title || item.title

  // Determine badge text and color
  const isSeries = item.media_type === 'episode' || item.type === 'series'
  const badgeText = isSeries ? 'Series' : 'Movie'
  const badgeColor = isSeries ? '#7c3aed' : '#2563eb'

  return (
    <TouchableOpacity
      style={[styles.card, { width: cardWidth, height: cardHeight }]}
      onPress={onPress}
      onLongPress={onLongPress}
      delayLongPress={500}
      activeOpacity={0.8}
    >
      {/* Poster Image */}
      {imageUrl ? (
        <Image
          source={{ uri: imageUrl }}
          style={styles.poster}
          resizeMode="cover"
        />
      ) : (
        <View style={[styles.poster, styles.placeholder]}>
          <Text style={styles.placeholderText}>{item.title?.charAt(0)}</Text>
        </View>
      )}

      {/* Badge (top-left corner) */}
      {showBadge && (
        <View style={[styles.badge, { backgroundColor: badgeColor }]}>
          <Text style={[styles.badgeText, { fontSize: scaledFont(10, layout.fontScale) }]}>{badgeText}</Text>
        </View>
      )}

      {/* Rating Badge (top-right corner) */}
      {showRating && item.rating !== undefined && item.rating > 0 && (
        <View style={styles.ratingBadge}>
          <View style={styles.ratingInner}>
            <Star size={scaledFont(10, layout.fontScale)} color="#ffd700" fill="#ffd700" />
            <Text style={[styles.ratingText, { fontSize: scaledFont(10, layout.fontScale) }]}>{item.rating.toFixed(1)}</Text>
          </View>
        </View>
      )}

      {/* Gradient Overlay (bottom) */}
      <View style={styles.gradientOverlay}>
        {/* Text on top of image */}
        <View style={styles.textOverlay}>
          <Text style={[styles.title, { fontSize: scaledFont(13, layout.fontScale) }]} numberOfLines={2}>
            {displayTitle}
          </Text>
          {remainingMinutes > 0 && (
            <Text style={[styles.remaining, { fontSize: scaledFont(11, layout.fontScale) }]}>{remainingMinutes}m remaining</Text>
          )}
        </View>

        {/* Progress bar at very bottom */}
        {progressPercent > 0 && (
          <View style={styles.progressBar}>
            <View style={[styles.progressFill, { width: `${progressPercent}%` }]} />
          </View>
        )}
      </View>
    </TouchableOpacity>
  )
}

const styles = StyleSheet.create({
  card: {
    borderRadius: 8,
    overflow: 'hidden',
    backgroundColor: '#1f1f1f',
    marginRight: 8,
  },
  poster: {
    width: '100%',
    height: '100%',
    backgroundColor: '#2a2a2a',
  },
  placeholder: {
    justifyContent: 'center',
    alignItems: 'center',
  },
  placeholderText: {
    fontSize: 32,
    fontWeight: 'bold',
    color: '#444',
  },
  badge: {
    position: 'absolute',
    top: 8,
    left: 8,
    backgroundColor: 'rgba(0, 0, 0, 0.7)',
    paddingHorizontal: 8,
    paddingVertical: 4,
    borderRadius: 4,
    zIndex: 2,
  },
  badgeText: {
    color: '#fff',
    fontSize: 10,
    fontWeight: '600',
    textTransform: 'uppercase',
  },
  ratingBadge: {
    position: 'absolute',
    top: 8,
    right: 8,
    backgroundColor: 'rgba(0, 0, 0, 0.7)',
    paddingHorizontal: 6,
    paddingVertical: 3,
    borderRadius: 4,
    zIndex: 2,
  },
  ratingInner: {
    flexDirection: 'row',
    alignItems: 'center',
    gap: 3,
  },
  ratingText: {
    color: '#ffd700',
    fontSize: 10,
    fontWeight: '600',
  },
  gradientOverlay: {
    ...StyleSheet.absoluteFillObject,
    justifyContent: 'flex-end',
    // Gradient from transparent to black at bottom
    backgroundColor: 'rgba(0, 0, 0, 0.3)',
  },
  textOverlay: {
    position: 'absolute',
    bottom: 0,
    left: 0,
    right: 0,
    padding: 10,
    paddingBottom: 20, // Extra space for progress bar
  },
  title: {
    fontSize: 13,
    fontWeight: '600',
    color: '#fff',
    textShadowColor: 'rgba(0, 0, 0, 0.8)',
    textShadowOffset: { width: 0, height: 1 },
    textShadowRadius: 2,
  },
  remaining: {
    fontSize: 11,
    color: '#aaa',
    marginTop: 4,
  },
  progressBar: {
    position: 'absolute',
    bottom: 0,
    left: 0,
    right: 0,
    height: 3,
    backgroundColor: 'rgba(255, 255, 255, 0.2)',
  },
  progressFill: {
    height: '100%',
    backgroundColor: '#e50914',
  },
})
