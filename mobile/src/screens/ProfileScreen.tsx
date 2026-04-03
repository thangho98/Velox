import { useState, useRef } from 'react'
import {
  View,
  Text,
  StyleSheet,
  FlatList,
  TouchableOpacity,
  RefreshControl,
  Image,
  Animated,
  PanResponder,
} from 'react-native'
import { Film, Tv, Trash2 } from 'lucide-react-native'
import { useNavigation } from '@react-navigation/native'
import { NativeStackNavigationProp } from '@react-navigation/native-stack'
import {
  useFavorites,
  useToggleFavorite,
  useRecentlyWatched,
} from '@velox/shared/hooks/media/useProgress'
import { useDismissProgress } from '@velox/shared/hooks/media/useContinueWatching'
import type { UserData } from '@velox/shared/types'
import type { RootStackParamList } from '../../App'
import { GridSkeleton } from '../components/SkeletonLoader'
import { useResponsiveLayout, scaledFont, cardWidth } from '../lib/responsive'

type NavigationProp = NativeStackNavigationProp<RootStackParamList>

interface ProfileMediaItem {
  id: number
  title: string
  poster_path: string | null
  backdrop_path: string | null
  release_date?: string
  media_type: 'movie' | 'episode'
  // Progress data
  position?: number
  duration?: number
  completed?: boolean
}

function MediaCard({
  item,
  onPress,
  showProgress = false,
  posterWidth,
  posterHeight,
  fontScale,
}: {
  item: ProfileMediaItem
  onPress: () => void
  showProgress?: boolean
  posterWidth: number
  posterHeight: number
  fontScale: number
}) {
  const year = item.release_date?.split('-')[0] || 'N/A'
  const posterUri = item.poster_path || null

  // Calculate progress percentage
  const progressPercent =
    showProgress && item.duration && item.duration > 0 && item.position
      ? (item.position / item.duration) * 100
      : 0

  // Calculate remaining time
  const remainingMinutes =
    showProgress && item.duration && item.position
      ? Math.ceil((item.duration - item.position) / 60)
      : 0

  return (
    <TouchableOpacity
      style={[styles.card, { width: posterWidth }]}
      onPress={onPress}
      activeOpacity={0.8}
    >
      <View
        style={[styles.posterContainer, { width: posterWidth, height: posterHeight }]}
      >
        {posterUri ? (
          <Image source={{ uri: posterUri }} style={styles.poster} resizeMode="cover" />
        ) : (
          <View style={styles.posterPlaceholder}>
            <Film size={32} color="#888" />
          </View>
        )}

        {/* Progress bar overlay */}
        {showProgress && progressPercent > 0 && (
          <View style={styles.progressOverlay}>
            <View style={styles.progressBar}>
              <View style={[styles.progressFill, { width: `${progressPercent}%` }]} />
            </View>
            {remainingMinutes > 0 && (
              <Text style={[styles.remainingText, { fontSize: scaledFont(9, fontScale) }]}>
                {remainingMinutes}m left
              </Text>
            )}
          </View>
        )}
      </View>
      <View style={styles.cardInfo}>
        <Text style={[styles.cardTitle, { fontSize: scaledFont(13, fontScale) }]} numberOfLines={2}>
          {item.title}
        </Text>
        <Text style={[styles.cardYear, { fontSize: scaledFont(11, fontScale) }]}>{year}</Text>
      </View>
    </TouchableOpacity>
  )
}

function EmptyState({ title, message }: { title: string; message: string }) {
  return (
    <View style={styles.emptyState}>
      <Tv size={48} color="#888" />
      <Text style={styles.emptyStateTitle}>{title}</Text>
      <Text style={styles.emptyStateMessage}>{message}</Text>
    </View>
  )
}

// Swipeable card wrapper for remove action (Recently Watched only)
function SwipeableCard({
  children,
  onRemove,
  screenWidth,
}: {
  children: React.ReactNode
  onRemove: () => void
  screenWidth: number
}) {
  const translateX = useRef(new Animated.Value(0)).current
  const swipeThreshold = -screenWidth * 0.25

  const panResponder = useRef(
    PanResponder.create({
      onStartShouldSetPanResponder: () => false,
      onMoveShouldSetPanResponder: (_, gestureState) => {
        return Math.abs(gestureState.dx) > 10 && Math.abs(gestureState.dy) < 20
      },
      onPanResponderMove: (_, gestureState) => {
        if (gestureState.dx < 0) {
          translateX.setValue(gestureState.dx)
        }
      },
      onPanResponderRelease: (_, gestureState) => {
        if (gestureState.dx < swipeThreshold) {
          // Swipe left to remove
          Animated.timing(translateX, {
            toValue: -screenWidth,
            duration: 200,
            useNativeDriver: true,
          }).start(() => {
            onRemove()
          })
        } else {
          // Reset position
          Animated.spring(translateX, {
            toValue: 0,
            useNativeDriver: true,
          }).start()
        }
      },
    }),
  ).current

  return (
    <View style={styles.swipeableContainer}>
      {/* Background remove action */}
      <View style={styles.removeAction}>
        <Trash2 size={24} color="#fff" />
      </View>
      {/* Card content */}
      <Animated.View
        style={[styles.swipeableCard, { transform: [{ translateX }] }]}
        {...panResponder.panHandlers}
      >
        {children}
      </Animated.View>
    </View>
  )
}

// Convert UserData to ProfileMediaItem
function userDataToMediaItem(userData: UserData, index: number): ProfileMediaItem {
  return {
    id: userData.media_id,
    title: userData.media_title || `Media ${userData.media_id}`,
    poster_path: userData.media_poster || null,
    backdrop_path: null,
    release_date: undefined,
    media_type: 'movie', // Default to movie, episodes would come from different source
    position: userData.position,
    duration: userData.media_duration,
    completed: userData.completed,
  }
}

export function ProfileScreen() {
  const navigation = useNavigation<NavigationProp>()
  const layout = useResponsiveLayout()
  const [activeTab, setActiveTab] = useState<'favorites' | 'recently'>('favorites')
  const [recentlyFilter, setRecentlyFilter] = useState<'all' | 'in_progress' | 'completed'>(
    'all',
  )

  // Responsive grid
  const numColumns = layout.gridColumns
  const posterW = cardWidth(numColumns, layout.width, layout.cardGap, layout.screenPadding)
  const posterH = posterW * 1.5
  // Max-width constraint for tablet content
  const contentMaxWidth = layout.device === 'tablet' || layout.device === 'tv' ? 960 : undefined

  // Favorites
  const {
    data: favoritesData,
    isLoading: favoritesLoading,
    refetch: refetchFavorites,
  } = useFavorites({ limit: 50 })
  const toggleFavorite = useToggleFavorite()

  // Recently Watched
  const {
    data: recentlyData,
    isLoading: recentlyLoading,
    refetch: refetchRecently,
  } = useRecentlyWatched({ limit: 50 })

  const isLoading = activeTab === 'favorites' ? favoritesLoading : recentlyLoading
  const data = activeTab === 'favorites' ? favoritesData : recentlyData

  // Convert UserData to ProfileMediaItem
  let mediaItems: ProfileMediaItem[] =
    data?.map((item, index) => userDataToMediaItem(item, index)) || []

  // Filter recently watched items
  if (activeTab === 'recently') {
    mediaItems = mediaItems.filter((item) => {
      if (recentlyFilter === 'all') return true
      if (recentlyFilter === 'in_progress') {
        return !item.completed && (item.position || 0) > 0
      }
      if (recentlyFilter === 'completed') {
        return item.completed
      }
      return true
    })
  }

  const handleRefresh = () => {
    if (activeTab === 'favorites') {
      refetchFavorites()
    } else {
      refetchRecently()
    }
  }

  const handleMediaPress = (media: ProfileMediaItem) => {
    if (media.media_type === 'episode') {
      navigation.navigate('Episode', { id: media.id })
    } else {
      navigation.navigate('Media', { id: media.id })
    }
  }

  const handleToggleFavorite = (mediaId: number) => {
    toggleFavorite.mutate(mediaId)
  }

  const dismissProgress = useDismissProgress()

  const handleRemoveFromHistory = (mediaId: number) => {
    dismissProgress.mutate(mediaId)
  }

  const renderItem = ({ item }: { item: ProfileMediaItem }) => {
    if (activeTab === 'recently') {
      return (
        <SwipeableCard onRemove={() => handleRemoveFromHistory(item.id)} screenWidth={layout.width}>
          <MediaCard
            item={item}
            onPress={() => handleMediaPress(item)}
            showProgress
            posterWidth={posterW}
            posterHeight={posterH}
            fontScale={layout.fontScale}
          />
        </SwipeableCard>
      )
    }
    return (
      <MediaCard
        item={item}
        onPress={() => handleMediaPress(item)}
        showProgress={false}
        posterWidth={posterW}
        posterHeight={posterH}
        fontScale={layout.fontScale}
      />
    )
  }

  return (
    <View style={styles.container}>
      <View style={[styles.header, { paddingHorizontal: layout.screenPadding }]}>
        <Text
          style={[styles.headerTitle, { fontSize: scaledFont(28, layout.fontScale) }]}
        >
          Profile
        </Text>
      </View>

      {/* Tabs */}
      <View style={[styles.tabs, { paddingHorizontal: layout.screenPadding }]}>
        <TouchableOpacity
          style={[styles.tab, activeTab === 'favorites' && styles.tabActive]}
          onPress={() => setActiveTab('favorites')}
        >
          <Text
            style={[
              styles.tabText,
              { fontSize: scaledFont(14, layout.fontScale) },
              activeTab === 'favorites' && styles.tabTextActive,
            ]}
          >
            Favorites ({favoritesData?.length || 0})
          </Text>
        </TouchableOpacity>
        <TouchableOpacity
          style={[styles.tab, activeTab === 'recently' && styles.tabActive]}
          onPress={() => setActiveTab('recently')}
        >
          <Text
            style={[
              styles.tabText,
              { fontSize: scaledFont(14, layout.fontScale) },
              activeTab === 'recently' && styles.tabTextActive,
            ]}
          >
            Recently Watched ({recentlyData?.length || 0})
          </Text>
        </TouchableOpacity>
      </View>

      {/* Filter Chips for Recently Watched */}
      {activeTab === 'recently' && (
        <View style={[styles.filterRow, { paddingHorizontal: layout.screenPadding }]}>
          <TouchableOpacity
            style={[styles.filterChip, recentlyFilter === 'all' && styles.filterChipActive]}
            onPress={() => setRecentlyFilter('all')}
          >
            <Text
              style={[
                styles.filterChipText,
                { fontSize: scaledFont(13, layout.fontScale) },
                recentlyFilter === 'all' && styles.filterChipTextActive,
              ]}
            >
              All
            </Text>
          </TouchableOpacity>
          <TouchableOpacity
            style={[
              styles.filterChip,
              recentlyFilter === 'in_progress' && styles.filterChipActive,
            ]}
            onPress={() => setRecentlyFilter('in_progress')}
          >
            <Text
              style={[
                styles.filterChipText,
                { fontSize: scaledFont(13, layout.fontScale) },
                recentlyFilter === 'in_progress' && styles.filterChipTextActive,
              ]}
            >
              In Progress
            </Text>
          </TouchableOpacity>
          <TouchableOpacity
            style={[
              styles.filterChip,
              recentlyFilter === 'completed' && styles.filterChipActive,
            ]}
            onPress={() => setRecentlyFilter('completed')}
          >
            <Text
              style={[
                styles.filterChipText,
                { fontSize: scaledFont(13, layout.fontScale) },
                recentlyFilter === 'completed' && styles.filterChipTextActive,
              ]}
            >
              Completed
            </Text>
          </TouchableOpacity>
        </View>
      )}

      {/* Content */}
      <View
        style={[
          styles.contentWrapper,
          contentMaxWidth ? { maxWidth: contentMaxWidth, alignSelf: 'center', width: '100%' } : undefined,
        ]}
      >
        {isLoading && mediaItems.length === 0 ? (
          <GridSkeleton
            columns={numColumns}
            rows={5}
            cardWidth={posterW}
            cardHeight={posterH}
          />
        ) : (
          <FlatList
            key={`grid-${numColumns}`}
            data={mediaItems}
            renderItem={renderItem}
            keyExtractor={(item) => String(item.id)}
            numColumns={numColumns}
            contentContainerStyle={[
              styles.listContent,
              { paddingHorizontal: layout.screenPadding },
            ]}
            columnWrapperStyle={[styles.columnWrapper, { gap: layout.cardGap }]}
            showsVerticalScrollIndicator={false}
            refreshControl={
              <RefreshControl
                refreshing={isLoading}
                onRefresh={handleRefresh}
                tintColor="#e50914"
                colors={['#e50914']}
              />
            }
            ListEmptyComponent={
              <EmptyState
                title={activeTab === 'favorites' ? 'No Favorites Yet' : 'No Recently Watched'}
                message={
                  activeTab === 'favorites'
                    ? 'Start adding movies and shows to your favorites'
                    : 'Your watch history will appear here'
                }
              />
            }
          />
        )}
      </View>
    </View>
  )
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
    backgroundColor: '#0a0a0a',
  },
  header: {
    paddingHorizontal: 20,
    paddingTop: 60,
    paddingBottom: 20,
    backgroundColor: '#141414',
  },
  headerTitle: {
    color: '#fff',
    fontSize: 28,
    fontWeight: 'bold',
  },
  tabs: {
    flexDirection: 'row',
    backgroundColor: '#141414',
    paddingHorizontal: 20,
    paddingBottom: 16,
  },
  tab: {
    flex: 1,
    paddingVertical: 12,
    alignItems: 'center',
    borderBottomWidth: 2,
    borderBottomColor: 'transparent',
  },
  tabActive: {
    borderBottomColor: '#e50914',
  },
  tabText: {
    color: '#888',
    fontSize: 14,
    fontWeight: '500',
  },
  tabTextActive: {
    color: '#fff',
  },
  // Filter chips
  filterRow: {
    flexDirection: 'row',
    paddingHorizontal: 16,
    paddingBottom: 12,
    gap: 8,
  },
  filterChip: {
    paddingHorizontal: 14,
    paddingVertical: 8,
    borderRadius: 20,
    backgroundColor: '#1f1f1f',
  },
  filterChipActive: {
    backgroundColor: '#e50914',
  },
  filterChipText: {
    fontSize: 13,
    color: '#888',
  },
  filterChipTextActive: {
    color: '#fff',
    fontWeight: '600',
  },
  contentWrapper: {
    flex: 1,
  },
  loadingContainer: {
    flex: 1,
    justifyContent: 'center',
    alignItems: 'center',
  },
  listContent: {
    paddingTop: 16,
    paddingBottom: 100,
  },
  columnWrapper: {
    justifyContent: 'flex-start',
    gap: 12,
    marginBottom: 16,
  },
  card: {
    // width set dynamically
  },
  posterContainer: {
    borderRadius: 8,
    overflow: 'hidden',
    backgroundColor: '#1a1a1a',
  },
  poster: {
    width: '100%',
    height: '100%',
  },
  posterPlaceholder: {
    width: '100%',
    height: '100%',
    justifyContent: 'center',
    alignItems: 'center',
    backgroundColor: '#2a2a2a',
  },
  // Progress bar styles
  progressOverlay: {
    position: 'absolute',
    bottom: 0,
    left: 0,
    right: 0,
    backgroundColor: 'rgba(0, 0, 0, 0.7)',
    paddingHorizontal: 6,
    paddingVertical: 4,
  },
  progressBar: {
    height: 3,
    backgroundColor: 'rgba(255, 255, 255, 0.3)',
    borderRadius: 1.5,
    overflow: 'hidden',
  },
  progressFill: {
    height: '100%',
    backgroundColor: '#e50914',
    borderRadius: 1.5,
  },
  remainingText: {
    color: '#fff',
    fontSize: 9,
    marginTop: 2,
    textAlign: 'center',
  },
  cardInfo: {
    marginTop: 8,
  },
  cardTitle: {
    color: '#fff',
    fontSize: 13,
    fontWeight: '500',
  },
  cardYear: {
    color: '#888',
    fontSize: 11,
    marginTop: 2,
  },
  emptyState: {
    flex: 1,
    justifyContent: 'center',
    alignItems: 'center',
    paddingVertical: 80,
    paddingHorizontal: 40,
  },
  emptyStateTitle: {
    marginTop: 16,
    color: '#fff',
    fontSize: 18,
    fontWeight: '600',
    marginBottom: 8,
    textAlign: 'center',
  },
  emptyStateMessage: {
    color: '#888',
    fontSize: 14,
    textAlign: 'center',
  },
  // Swipeable card styles
  swipeableContainer: {
    position: 'relative',
  },
  swipeableCard: {
    position: 'relative',
  },
  removeAction: {
    position: 'absolute',
    top: 0,
    bottom: 0,
    right: 0,
    width: 80,
    backgroundColor: '#e50914',
    borderRadius: 8,
    justifyContent: 'center',
    alignItems: 'center',
  },
})
