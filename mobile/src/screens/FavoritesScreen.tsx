import { useState } from 'react'
import {
  View,
  Text,
  FlatList,
  TouchableOpacity,
  StyleSheet,
} from 'react-native'
import { Play, Info, Trash2, Heart } from 'lucide-react-native'
import { useNavigation } from '@react-navigation/native'
import type { NativeStackNavigationProp } from '@react-navigation/native-stack'
import { useFavorites, useToggleFavorite } from '@velox/shared/hooks'
import type { UserData } from '@velox/shared/types/playback'
import type { RootStackParamList } from '../../App'
import { MediaCard } from '../components/MediaCard'
import { GridSkeleton } from '../components/SkeletonLoader'
import { QuickActionsMenu } from '../components/QuickActionsMenu'
import { Toast } from '../components/Toast'
import { useResponsiveLayout, scaledFont } from '../lib/responsive'

type NavigationProp = NativeStackNavigationProp<RootStackParamList>

export function FavoritesScreen() {
  const navigation = useNavigation<NavigationProp>()
  const layout = useResponsiveLayout()
  const numColumns = layout.gridColumns
  const { data: favorites, isLoading } = useFavorites({ limit: 100 })
  const toggleFavorite = useToggleFavorite()

  // Quick actions menu state
  const [selectedItem, setSelectedItem] = useState<UserData | null>(null)
  const [showQuickActions, setShowQuickActions] = useState(false)

  const handleItemPress = (item: UserData) => {
    navigation.navigate('Media', { id: item.media_id })
  }

  const handleItemLongPress = (item: UserData) => {
    setSelectedItem(item)
    setShowQuickActions(true)
  }

  const handleQuickAction = (actionId: string) => {
    if (!selectedItem) return

    switch (actionId) {
      case 'unfavorite':
        toggleFavorite.mutate(selectedItem.media_id, {
          onSuccess: () => {
            Toast.info('Removed from favorites')
          },
        })
        break
      case 'info':
        navigation.navigate('Media', { id: selectedItem.media_id })
        break
      case 'play':
        navigation.navigate('Episode', { id: selectedItem.media_id })
        break
    }
  }

  const quickActions = [
    { id: 'play', label: 'Play', icon: <Play size={20} color="#fff" /> },
    { id: 'info', label: 'View Details', icon: <Info size={20} color="#fff" /> },
    { id: 'unfavorite', label: 'Remove from Favorites', icon: <Trash2 size={20} color="#e50914" />, destructive: true },
  ]

  if (isLoading) {
    return (
      <View style={styles.container}>
        <View style={[styles.header, { paddingHorizontal: layout.screenPadding }]}>
          <Text style={[styles.headerTitle, { fontSize: scaledFont(20, layout.fontScale) }]}>Favorites</Text>
        </View>
        <GridSkeleton columns={numColumns} rows={5} cardWidth={110} cardHeight={165} />
      </View>
    )
  }

  if (!favorites || favorites.length === 0) {
    return (
      <View style={styles.centered}>
        <Heart size={64} color="#e50914" fill="#e50914" />
        <Text style={styles.emptyTitle}>No favorites yet</Text>
        <Text style={styles.emptySubtext}>
          Tap the heart icon on any movie or series to add it here
        </Text>
      </View>
    )
  }

  return (
    <View style={styles.container}>
      <View style={[styles.header, { paddingHorizontal: layout.screenPadding }]}>
        <Text style={[styles.headerTitle, { fontSize: scaledFont(20, layout.fontScale) }]}>Favorites</Text>
        <Text style={[styles.headerCount, { fontSize: scaledFont(14, layout.fontScale) }]}>{favorites.length} items</Text>
      </View>
      <FlatList
        key={`grid-${numColumns}`}
        data={favorites}
        keyExtractor={(item) => String(item.media_id)}
        numColumns={numColumns}
        contentContainerStyle={[styles.grid, { paddingHorizontal: layout.screenPadding }]}
        columnWrapperStyle={numColumns > 1 ? { justifyContent: 'flex-start', gap: layout.cardGap } : undefined}
        renderItem={({ item }) => {
          // Calculate progress percentage
          const progressPercent =
            item.media_duration && item.media_duration > 0
              ? (item.position / item.media_duration) * 100
              : 0

          return (
            <View style={[styles.gridItem, { marginBottom: layout.cardGap }]}>
              <MediaCard
                item={{
                  id: item.media_id,
                  title: item.media_title || 'Unknown',
                  sort_title: item.media_title || 'Unknown',
                  poster_path: item.media_poster || '',
                  media_type: 'movie',
                  genres: [],
                  release_date: undefined,
                  rating: item.rating,
                }}
                onPress={() => handleItemPress(item)}
                onLongPress={() => handleItemLongPress(item)}
                size="medium"
                showBadge
                showRating={item.rating !== undefined}
                columns={numColumns}
                containerWidth={layout.width}
                gap={layout.cardGap}
                padding={layout.screenPadding}
                progress={
                  progressPercent > 0
                    ? {
                        position: item.position,
                        duration: item.media_duration || 1,
                        completed: item.completed,
                      }
                    : undefined
                }
              />
              {/* Favorite heart indicator */}
              <View style={styles.favoriteIndicator}>
                <Heart size={16} color="#e50914" fill="#e50914" />
              </View>
            </View>
          )
        }}
        ListEmptyComponent={
          <View style={styles.centered}>
            <Heart size={64} color="#e50914" fill="#e50914" />
            <Text style={styles.emptyTitle}>No favorites yet</Text>
            <Text style={styles.emptySubtext}>
              Tap the heart icon on any movie or series to add it here
            </Text>
          </View>
        }
      />

      {/* Quick Actions Menu */}
      <QuickActionsMenu
        visible={showQuickActions}
        title={selectedItem?.media_title || 'Options'}
        actions={quickActions}
        onAction={handleQuickAction}
        onClose={() => setShowQuickActions(false)}
      />
    </View>
  )
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
    backgroundColor: '#141414',
  },
  centered: {
    flex: 1,
    justifyContent: 'center',
    alignItems: 'center',
    backgroundColor: '#141414',
    padding: 40,
  },
  header: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
    paddingHorizontal: 16,
    paddingVertical: 12,
  },
  headerTitle: {
    fontSize: 20,
    fontWeight: 'bold',
    color: '#fff',
  },
  headerCount: {
    fontSize: 14,
    color: '#888',
  },
  grid: {
    paddingHorizontal: 12,
    paddingBottom: 20,
  },
  rowMulti: {
    justifyContent: 'flex-start',
    gap: 8,
  },
  gridItem: {
    position: 'relative',
    marginBottom: 8,
  },
  favoriteIndicator: {
    position: 'absolute',
    top: 8,
    right: 8,
    zIndex: 10,
  },
  moreButton: {
    position: 'absolute',
    top: 8,
    left: 8,
    width: 24,
    height: 24,
    borderRadius: 12,
    backgroundColor: 'rgba(0, 0, 0, 0.5)',
    justifyContent: 'center',
    alignItems: 'center',
    zIndex: 10,
  },
  moreButtonText: {
    color: '#fff',
    fontSize: 14,
    fontWeight: 'bold',
  },
  emptyTitle: {
    marginTop: 16,
    fontSize: 18,
    fontWeight: '600',
    color: '#fff',
    marginBottom: 8,
  },
  emptySubtext: {
    fontSize: 14,
    color: '#888',
    textAlign: 'center',
  },
})
