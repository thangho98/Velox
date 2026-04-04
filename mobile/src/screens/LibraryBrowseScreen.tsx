import { useState } from 'react'
import {
  View,
  FlatList,
  TextInput,
  TouchableOpacity,
  Text,
  StyleSheet,
  RefreshControl,
  ActivityIndicator,
} from 'react-native'
import { useNavigation, useRoute, RouteProp } from '@react-navigation/native'
import type { NativeStackNavigationProp } from '@react-navigation/native-stack'
import { useMediaList } from '@velox/shared/hooks'
import type { MediaListItem } from '@velox/shared/types'
import type { RootStackParamList } from '../../App'
import { MediaCard } from '../components/MediaCard'
import { Breadcrumb } from '../components/Breadcrumb'
import { useResponsiveLayout, scaledFont } from '../lib/responsive'

type NavigationProp = NativeStackNavigationProp<RootStackParamList>
type LibraryRouteProp = RouteProp<RootStackParamList, 'LibraryBrowse'>

type SortOption = 'newest' | 'oldest' | 'rating' | 'title'

interface LibraryBrowseScreenProps {
  /** Filter by media type when used as standalone screen (Movies/Series tabs) */
  type?: 'movie' | 'series'
}

export function LibraryBrowseScreen({ type }: LibraryBrowseScreenProps) {
  const navigation = useNavigation<NavigationProp>()
  const layout = useResponsiveLayout()
  const numColumns = layout.gridColumns
  const route = useRoute<LibraryRouteProp>()
  // When used from tab screens, route.params may be undefined
  const libraryId = route.params?.id
  const libraryName = route.params?.name ?? 'Library'

  const [search, setSearch] = useState('')
  const [sort, setSort] = useState<SortOption>('newest')
  const [refreshing, setRefreshing] = useState(false)
  const [currentPath, setCurrentPath] = useState('')

  // Convert 'series' to 'episode' for API (series content is stored as episodes)
  const apiType = type === 'series' ? 'episode' : type

  const { data: media, isLoading, refetch } = useMediaList({
    library_id: libraryId,
    type: apiType,
    search: search || undefined,
    sort,
    limit: 50,
  })

  const onRefresh = async () => {
    setRefreshing(true)
    await refetch()
    setRefreshing(false)
  }

  const handleMediaPress = (item: MediaListItem) => {
    if (item.media_type === 'movie' || item.type === 'movie') {
      navigation.navigate('Media', { id: item.id })
    } else {
      navigation.navigate('SeriesDetail', { id: item.id })
    }
  }

  const handleBreadcrumbNavigate = (path: string) => {
    setCurrentPath(path)
  }

  const handleBreadcrumbHome = () => {
    navigation.goBack()
  }

  const renderSortButton = (option: SortOption, label: string) => (
    <TouchableOpacity
      style={[styles.sortButton, sort === option && styles.sortButtonActive]}
      onPress={() => setSort(option)}
    >
      <Text style={[styles.sortButtonText, sort === option && styles.sortButtonTextActive]}>
        {label}
      </Text>
    </TouchableOpacity>
  )

  return (
    <View style={styles.container}>
      {/* Breadcrumb Navigation */}
      {libraryId && (
        <Breadcrumb
          libraryName={libraryName}
          path={currentPath}
          onNavigate={handleBreadcrumbNavigate}
          onHome={handleBreadcrumbHome}
        />
      )}

      {/* Search Bar */}
      <View style={[styles.searchContainer, { padding: layout.screenPadding }]}>
        <TextInput
          style={[styles.searchInput, { fontSize: scaledFont(16, layout.fontScale) }]}
          placeholder="Search..."
          placeholderTextColor="#666"
          value={search}
          onChangeText={setSearch}
        />
      </View>

      {/* Sort Options */}
      <View style={[styles.sortContainer, { paddingHorizontal: layout.screenPadding, gap: layout.cardGap }]}>
        {renderSortButton('newest', 'Newest')}
        {renderSortButton('oldest', 'Oldest')}
        {renderSortButton('rating', 'Rating')}
        {renderSortButton('title', 'A-Z')}
      </View>

      {/* Media Grid */}
      <FlatList
        key={`grid-${numColumns}`}
        data={media}
        keyExtractor={(item) => String(item.id)}
        numColumns={numColumns}
        renderItem={({ item }) => (
          <View style={[styles.gridItem, { maxWidth: `${100 / numColumns}%`, padding: layout.cardGap / 2 }]}>
            <MediaCard
              item={item}
              onPress={() => handleMediaPress(item)}
              size="medium"
              columns={numColumns}
              containerWidth={layout.width}
              gap={layout.cardGap}
              padding={layout.screenPadding}
            />
          </View>
        )}
        columnWrapperStyle={styles.row}
        contentContainerStyle={[styles.gridContent, { padding: layout.cardGap / 2 }]}
        refreshControl={
          <RefreshControl
            refreshing={refreshing}
            onRefresh={onRefresh}
            tintColor="#e50914"
          />
        }
        ListEmptyComponent={
          <View style={styles.empty}>
            <Text style={styles.emptyText}>
              {isLoading ? 'Loading...' : 'No media found'}
            </Text>
          </View>
        }
      />
    </View>
  )
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
    backgroundColor: '#141414',
  },
  searchContainer: {
    padding: 16,
  },
  searchInput: {
    backgroundColor: '#1f1f1f',
    borderRadius: 8,
    padding: 12,
    fontSize: 16,
    color: '#fff',
    borderWidth: 1,
    borderColor: '#333',
  },
  sortContainer: {
    flexDirection: 'row',
    paddingHorizontal: 16,
    paddingBottom: 12,
    gap: 8,
  },
  sortButton: {
    paddingHorizontal: 12,
    paddingVertical: 6,
    borderRadius: 16,
    backgroundColor: '#1f1f1f',
  },
  sortButtonActive: {
    backgroundColor: '#e50914',
  },
  sortButtonText: {
    fontSize: 12,
    color: '#888',
  },
  sortButtonTextActive: {
    color: '#fff',
    fontWeight: '600',
  },
  gridContent: {
    padding: 8,
  },
  row: {
    justifyContent: 'flex-start',
  },
  gridItem: {
    flex: 1,
    // maxWidth and padding set dynamically via inline style
  },
  empty: {
    flex: 1,
    padding: 40,
    alignItems: 'center',
  },
  emptyText: {
    color: '#888',
    fontSize: 14,
  },
})
