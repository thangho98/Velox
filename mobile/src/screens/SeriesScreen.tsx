import { useState, useMemo, useRef } from 'react'
import {
  View,
  FlatList,
  TextInput,
  TouchableOpacity,
  Text,
  StyleSheet,
  RefreshControl,
  ScrollView,
} from 'react-native'
import { Play, Info } from 'lucide-react-native'
import { useNavigation } from '@react-navigation/native'
import type { NativeStackNavigationProp } from '@react-navigation/native-stack'
import { useMediaList, useGenres } from '@velox/shared/hooks'
import type { MediaListItem } from '@velox/shared/types'
import type { RootStackParamList } from '../../App'
import { MediaCard } from '../components/MediaCard'
import { FilterBottomSheet } from '../components/FilterBottomSheet'
import { GridSkeleton } from '../components/SkeletonLoader'
import { QuickActionsMenu } from '../components/QuickActionsMenu'
import { useResponsiveLayout, scaledFont } from '../lib/responsive'

type NavigationProp = NativeStackNavigationProp<RootStackParamList>

const ALPHABET = 'ABCDEFGHIJKLMNOPQRSTUVWXYZ#'.split('')

type SortOption = 'newest' | 'oldest' | 'rating' | 'title'

export function SeriesScreen() {
  const navigation = useNavigation<NavigationProp>()
  const flatListRef = useRef<FlatList>(null)
  const layout = useResponsiveLayout()
  const numColumns = layout.gridColumns

  const [search, setSearch] = useState('')
  const [sort, setSort] = useState<SortOption>('newest')
  const [genre, setGenre] = useState<string | undefined>()
  const [year, setYear] = useState<string | undefined>()
  const [refreshing, setRefreshing] = useState(false)
  const [activeLetter, setActiveLetter] = useState<string | null>(null)

  // Filter bottom sheet states
  const [showGenreSheet, setShowGenreSheet] = useState(false)
  const [showYearSheet, setShowYearSheet] = useState(false)

  // Quick actions menu state
  const [selectedItem, setSelectedItem] = useState<MediaListItem | null>(null)
  const [showQuickActions, setShowQuickActions] = useState(false)

  const quickActions = [
    { id: 'play', label: 'Play First Episode', icon: <Play size={20} color="#fff" /> },
    { id: 'info', label: 'View Details', icon: <Info size={20} color="#fff" /> },
  ]

  const handleQuickAction = (actionId: string) => {
    if (!selectedItem) return

    // Navigate to Series detail
    const seriesId = selectedItem.series_id || selectedItem.id

    switch (actionId) {
      case 'play':
        navigation.navigate('Episode', { id: seriesId })
        break
      case 'info':
        navigation.navigate('Series', { id: seriesId })
        break
    }
  }

  const handleItemLongPress = (item: MediaListItem) => {
    setSelectedItem(item)
    setShowQuickActions(true)
  }

  // Series use 'episode' type for API
  const { data: series, isLoading, refetch } = useMediaList({
    type: 'episode',
    search: search || undefined,
    genre,
    year,
    sort,
    limit: 500,
  })

  const { data: genreList } = useGenres('series')
  const genres = useMemo(() => genreList?.map((g) => g.name) ?? [], [genreList])

  // Generate year options (current year down to 1970)
  const years = useMemo(() => {
    const currentYear = new Date().getFullYear()
    return Array.from({ length: currentYear - 1969 }, (_, i) => String(currentYear - i))
  }, [])

  const onRefresh = async () => {
    setRefreshing(true)
    await refetch()
    setRefreshing(false)
  }

  const handleMediaPress = (item: MediaListItem) => {
    // Episodes navigate to Series detail
    if (item.series_id) {
      navigation.navigate('Series', { id: item.series_id })
    } else if (item.media_type === 'episode' || item.type === 'series') {
      navigation.navigate('Series', { id: item.id })
    } else {
      navigation.navigate('Media', { id: item.id })
    }
  }

  const handleClearFilters = () => {
    setGenre(undefined)
    setYear(undefined)
  }

  const hasActiveFilters = genre !== undefined || year !== undefined

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

  const renderFilterButton = (
    label: string,
    activeValue: string | undefined,
    onPress: () => void
  ) => (
    <TouchableOpacity
      style={[styles.filterButton, activeValue && styles.filterButtonActive]}
      onPress={onPress}
    >
      <Text style={[styles.filterButtonText, activeValue && styles.filterButtonTextActive]}>
        {activeValue || label}
      </Text>
      <Text style={styles.filterChevron}>▼</Text>
    </TouchableOpacity>
  )

  const seriesCount = series?.length ?? 0

  // Group series by first letter for A-Z index
  const { groupedSeries, letterIndex } = useMemo(() => {
    if (!series) return { groupedSeries: [], letterIndex: {} }

    const index: Record<string, number> = {}
    const groups: { letter: string; data: MediaListItem[] }[] = []
    let currentLetter = ''
    let currentGroup: MediaListItem[] = []

    series.forEach((item, idx) => {
      const firstChar = (item.title || 'Unknown').charAt(0).toUpperCase()
      const letter = /[A-Z]/.test(firstChar) ? firstChar : '#'

      if (letter !== currentLetter) {
        if (currentGroup.length > 0) {
          groups.push({ letter: currentLetter, data: currentGroup })
        }
        currentLetter = letter
        currentGroup = [item]
        index[letter] = idx
      } else {
        currentGroup.push(item)
      }
    })

    if (currentGroup.length > 0) {
      groups.push({ letter: currentLetter, data: currentGroup })
    }

    return { groupedSeries: groups, letterIndex: index }
  }, [series])

  const handleLetterPress = (letter: string) => {
    setActiveLetter(letter)
    const idx = letterIndex[letter]
    if (idx !== undefined && flatListRef.current) {
      flatListRef.current.scrollToIndex({ index: idx, animated: true })
    }
    setTimeout(() => setActiveLetter(null), 500)
  }

  const renderLetterSection = ({ item: group }: { item: { letter: string; data: MediaListItem[] } }) => (
    <View>
      <View style={[styles.letterHeader, { paddingHorizontal: layout.screenPadding }]}>
        <Text style={[styles.letterHeaderText, { fontSize: scaledFont(18, layout.fontScale) }]}>{group.letter}</Text>
      </View>
      <View style={[styles.letterRow, { paddingHorizontal: layout.cardGap / 2 }]}>
        {group.data.map((item) => (
          <View key={item.id} style={[styles.gridItem, { maxWidth: `${100 / numColumns}%`, padding: layout.cardGap / 2 }]}>
            <MediaCard
              item={item}
              onPress={() => handleMediaPress(item)}
              onLongPress={() => handleItemLongPress(item)}
              size="medium"
              showBadge
              showRating
              columns={numColumns}
              containerWidth={layout.width}
              gap={layout.cardGap}
              padding={layout.screenPadding}
            />
          </View>
        ))}
      </View>
    </View>
  )

  return (
    <View style={styles.container}>
      {/* Search Bar */}
      <View style={[styles.searchContainer, { padding: layout.screenPadding, paddingBottom: layout.cardGap }]}>
        <TextInput
          style={[styles.searchInput, { fontSize: scaledFont(16, layout.fontScale) }]}
          placeholder="Search series..."
          placeholderTextColor="#666"
          value={search}
          onChangeText={setSearch}
        />
      </View>

      {/* Filter Row */}
      <View style={[styles.filterRow, { paddingHorizontal: layout.screenPadding, gap: layout.cardGap }]}>
        {renderFilterButton('Genre', genre, () => setShowGenreSheet(true))}
        {renderFilterButton('Year', year, () => setShowYearSheet(true))}
        {hasActiveFilters && (
          <TouchableOpacity style={styles.clearButton} onPress={handleClearFilters}>
            <Text style={styles.clearButtonText}>Clear</Text>
          </TouchableOpacity>
        )}
      </View>

      {/* Sort Options */}
      <View style={[styles.sortContainer, { paddingHorizontal: layout.screenPadding, gap: layout.cardGap }]}>
        {renderSortButton('newest', 'Newest')}
        {renderSortButton('oldest', 'Oldest')}
        {renderSortButton('rating', 'Rating')}
        {renderSortButton('title', 'A-Z')}
      </View>

      {/* Count Header */}
      <View style={[styles.countHeader, { paddingHorizontal: layout.screenPadding }]}>
        <Text style={[styles.countText, { fontSize: scaledFont(14, layout.fontScale) }]}>
          {seriesCount} {seriesCount === 1 ? 'series' : 'series'}
        </Text>
      </View>

      {/* Media Grid */}
      {isLoading && !series ? (
        <GridSkeleton columns={numColumns} rows={5} cardWidth={140} cardHeight={210} />
      ) : (
        <View style={styles.gridWithIndex}>
          <FlatList
            ref={flatListRef}
            data={groupedSeries}
            keyExtractor={(item) => item.letter}
            renderItem={renderLetterSection}
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
                  {hasActiveFilters ? 'No series match your filters.' : 'No series in your library yet.'}
                </Text>
              </View>
            }
          />

          {/* A-Z Sidebar */}
          {sort === 'title' && groupedSeries.length > 0 && (
            <ScrollView
              style={styles.alphabetSidebar}
              contentContainerStyle={styles.alphabetContent}
              showsVerticalScrollIndicator={false}
            >
              {ALPHABET.map((letter) => (
                <TouchableOpacity
                  key={letter}
                  style={[styles.alphabetLetter, activeLetter === letter && styles.alphabetLetterActive]}
                  onPress={() => handleLetterPress(letter)}
                >
                  <Text style={[styles.alphabetText, activeLetter === letter && styles.alphabetTextActive]}>
                    {letter}
                  </Text>
                </TouchableOpacity>
              ))}
            </ScrollView>
          )}
        </View>
      )}

      {/* Genre Filter Bottom Sheet */}
      <FilterBottomSheet
        visible={showGenreSheet}
        title="Select Genre"
        options={genres}
        selectedValue={genre}
        onSelect={setGenre}
        onClose={() => setShowGenreSheet(false)}
      />

      {/* Year Filter Bottom Sheet */}
      <FilterBottomSheet
        visible={showYearSheet}
        title="Select Year"
        options={years}
        selectedValue={year}
        onSelect={setYear}
        onClose={() => setShowYearSheet(false)}
      />

      {/* Quick Actions Menu */}
      <QuickActionsMenu
        visible={showQuickActions}
        title={selectedItem?.title || 'Options'}
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
  searchContainer: {
    padding: 16,
    paddingBottom: 8,
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
  filterRow: {
    flexDirection: 'row',
    paddingHorizontal: 16,
    paddingBottom: 8,
    gap: 8,
  },
  filterButton: {
    flexDirection: 'row',
    alignItems: 'center',
    paddingHorizontal: 14,
    paddingVertical: 8,
    borderRadius: 8,
    backgroundColor: '#1f1f1f',
    gap: 6,
  },
  filterButtonActive: {
    backgroundColor: '#e50914',
  },
  filterButtonText: {
    fontSize: 13,
    color: '#888',
  },
  filterButtonTextActive: {
    color: '#fff',
    fontWeight: '600',
  },
  filterChevron: {
    fontSize: 8,
    color: '#888',
  },
  clearButton: {
    paddingHorizontal: 14,
    paddingVertical: 8,
    borderRadius: 8,
    backgroundColor: 'rgba(229, 9, 20, 0.2)',
  },
  clearButtonText: {
    fontSize: 13,
    color: '#e50914',
    fontWeight: '600',
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
  countHeader: {
    paddingHorizontal: 16,
    paddingBottom: 8,
  },
  countText: {
    fontSize: 14,
    color: '#888',
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
  // A-Z Index styles
  gridWithIndex: {
    flexDirection: 'row',
    flex: 1,
  },
  letterHeader: {
    paddingHorizontal: 16,
    paddingVertical: 8,
    backgroundColor: '#0a0a0a',
  },
  letterHeaderText: {
    color: '#e50914',
    fontSize: 18,
    fontWeight: 'bold',
  },
  letterRow: {
    flexDirection: 'row',
    flexWrap: 'wrap',
    paddingHorizontal: 8,
  },
  alphabetSidebar: {
    width: 24,
    backgroundColor: 'rgba(0, 0, 0, 0.3)',
    paddingVertical: 8,
  },
  alphabetContent: {
    alignItems: 'center',
  },
  alphabetLetter: {
    width: 24,
    height: 18,
    justifyContent: 'center',
    alignItems: 'center',
  },
  alphabetLetterActive: {
    backgroundColor: '#e50914',
    borderRadius: 4,
  },
  alphabetText: {
    color: '#666',
    fontSize: 11,
    fontWeight: '600',
  },
  alphabetTextActive: {
    color: '#fff',
  },
})
