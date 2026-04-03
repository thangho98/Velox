import { useState, useCallback } from 'react'
import {
  View,
  Text,
  TextInput,
  FlatList,
  TouchableOpacity,
  StyleSheet,
  ActivityIndicator,
} from 'react-native'
import { Search, Film, Tv, Clock } from 'lucide-react-native'
import { useNavigation } from '@react-navigation/native'
import type { NativeStackNavigationProp } from '@react-navigation/native-stack'
import { useSearch } from '@velox/shared/hooks'
import type { MediaListItem } from '@velox/shared/types'
import type { SeriesListItem } from '@velox/shared/types/series'
import type { RootStackParamList } from '../../App'
import { MediaCard } from '../components/MediaCard'
import { useResponsiveLayout, scaledFont } from '../lib/responsive'

type NavigationProp = NativeStackNavigationProp<RootStackParamList>
type TypeFilter = '' | 'movie' | 'series'

export function SearchScreen() {
  const navigation = useNavigation<NavigationProp>()
  const layout = useResponsiveLayout()
  const numColumns = layout.gridColumns

  const [query, setQuery] = useState('')
  const [typeFilter, setTypeFilter] = useState<TypeFilter>('')
  const [recentSearches, setRecentSearches] = useState<string[]>([])

  const { data: searchResult, isLoading } = useSearch(query, 50)

  const movies = searchResult?.movies ?? []
  const series = searchResult?.series ?? []

  const filteredMovies = typeFilter === 'series' ? [] : movies
  const filteredSeries = typeFilter === 'movie' ? [] : series
  const totalCount = filteredMovies.length + filteredSeries.length

  const handleMoviePress = (item: MediaListItem) => {
    navigation.navigate('Media', { id: item.id })
  }

  const handleSeriesPress = (item: SeriesListItem) => {
    navigation.navigate('Series', { id: item.id })
  }

  const handleSubmitSearch = () => {
    if (query.trim()) {
      setRecentSearches((prev) => {
        const filtered = prev.filter((s) => s !== query.trim())
        return [query.trim(), ...filtered].slice(0, 5)
      })
    }
  }

  const handleRecentSearchPress = (searchTerm: string) => {
    setQuery(searchTerm)
  }

  const clearRecentSearches = () => {
    setRecentSearches([])
  }

  const renderTypeChipIcon = (filter: TypeFilter): React.ReactNode => {
    const color = typeFilter === filter ? '#fff' : '#888'
    switch (filter) {
      case 'series':
        return <Tv size={14} color={color} />
      default:
        return <Film size={14} color={color} />
    }
  }

  const renderTypeChip = (filter: TypeFilter, label: string) => (
    <TouchableOpacity
      style={[styles.chip, typeFilter === filter && styles.chipActive]}
      onPress={() => setTypeFilter(filter)}
    >
      {renderTypeChipIcon(filter)}
      <Text style={[styles.chipText, typeFilter === filter && styles.chipTextActive]}>
        {label}
      </Text>
    </TouchableOpacity>
  )

  const renderMovieItem = ({ item }: { item: MediaListItem }) => (
    <View style={styles.gridItem}>
      <MediaCard
        item={item}
        onPress={() => handleMoviePress(item)}
        size="medium"
        showBadge
        showRating
      />
    </View>
  )

  const renderSeriesItem = ({ item }: { item: SeriesListItem }) => (
    <View style={styles.gridItem}>
      <MediaCard
        item={{
          id: item.id,
          title: item.title,
          sort_title: item.sort_title,
          poster_path: item.poster_path,
          media_type: 'episode',
          type: 'series',
          genres: item.genres,
          release_date: item.first_air_date,
        }}
        onPress={() => handleSeriesPress(item)}
        size="medium"
        showBadge
        showRating
      />
    </View>
  )

  const renderSectionIcon = (title: string): React.ReactNode => {
    if (title === 'Series') return <Tv size={18} color="#fff" />
    return <Film size={18} color="#fff" />
  }

  const renderSectionHeader = (title: string, count: number) => (
    <View style={[styles.sectionHeader, { paddingHorizontal: layout.screenPadding }]}>
      <View style={{ flexDirection: 'row', alignItems: 'center', gap: 8 }}>
        {renderSectionIcon(title)}
        <Text style={[styles.sectionTitle, { fontSize: scaledFont(18, layout.fontScale) }]}>
          {title} ({count})
        </Text>
      </View>
    </View>
  )

  // Empty state - no query
  if (!query) {
    return (
      <View style={styles.container}>
        <View style={styles.searchContainer}>
          <TextInput
            style={styles.searchInput}
            placeholder="Search movies and series..."
            placeholderTextColor="#666"
            value={query}
            onChangeText={setQuery}
            onSubmitEditing={handleSubmitSearch}
            autoFocus
            returnKeyType="search"
          />
        </View>

        {recentSearches.length > 0 && (
          <View style={styles.recentSection}>
            <View style={styles.recentHeader}>
              <Text style={styles.recentTitle}>Recent Searches</Text>
              <TouchableOpacity onPress={clearRecentSearches}>
                <Text style={styles.clearText}>Clear</Text>
              </TouchableOpacity>
            </View>
            {recentSearches.map((search, index) => (
              <TouchableOpacity
                key={index}
                style={styles.recentItem}
                onPress={() => handleRecentSearchPress(search)}
              >
                <Clock size={16} color="#888" />
                <Text style={styles.recentText}>{search}</Text>
              </TouchableOpacity>
            ))}
          </View>
        )}

        {recentSearches.length === 0 && (
          <View style={styles.emptyState}>
            <Search size={64} color="#888" />
            <Text style={styles.emptyText}>
              Use the search bar above to find movies and series
            </Text>
          </View>
        )}
      </View>
    )
  }

  // Loading state
  if (isLoading) {
    return (
      <View style={styles.container}>
        <View style={styles.searchContainer}>
          <TextInput
            style={styles.searchInput}
            placeholder="Search movies and series..."
            placeholderTextColor="#666"
            value={query}
            onChangeText={setQuery}
            autoFocus
          />
        </View>
        <View style={styles.loadingContainer}>
          <ActivityIndicator size="large" color="#e50914" />
        </View>
      </View>
    )
  }

  // No results
  if (totalCount === 0) {
    return (
      <View style={styles.container}>
        <View style={styles.searchContainer}>
          <TextInput
            style={styles.searchInput}
            placeholder="Search movies and series..."
            placeholderTextColor="#666"
            value={query}
            onChangeText={setQuery}
            autoFocus
          />
        </View>
        <View style={styles.noResultsContainer}>
          <Search size={64} color="#888" />
          <Text style={styles.noResultsText}>No results for "{query}"</Text>
        </View>
      </View>
    )
  }

  return (
    <View style={styles.container}>
      {/* Search Input */}
      <View style={[styles.searchContainer, { padding: layout.screenPadding, paddingBottom: layout.cardGap }]}>
        <TextInput
          style={[styles.searchInput, { fontSize: scaledFont(16, layout.fontScale) }]}
          placeholder="Search movies and series..."
          placeholderTextColor="#666"
          value={query}
          onChangeText={setQuery}
          autoFocus
          returnKeyType="search"
          onSubmitEditing={handleSubmitSearch}
        />
      </View>

      {/* Results Count */}
      <View style={[styles.resultsHeader, { paddingHorizontal: layout.screenPadding }]}>
        <Text style={[styles.resultsCount, { fontSize: scaledFont(14, layout.fontScale) }]}>
          Found {totalCount} results for "{query}"
        </Text>
      </View>

      {/* Type Filter Chips */}
      <View style={[styles.chipRow, { paddingHorizontal: layout.screenPadding, gap: layout.cardGap }]}>
        {renderTypeChip('', 'All')}
        {renderTypeChip('movie', 'Movies')}
        {renderTypeChip('series', 'Series')}
      </View>

      {/* Results */}
      <FlatList
        data={[]}
        renderItem={() => null}
        ListHeaderComponent={
          <>
            {/* Movies Section */}
            {filteredMovies.length > 0 && (
              <>
                {renderSectionHeader('Movies', filteredMovies.length)}
                <View style={[styles.grid, { paddingHorizontal: layout.screenPadding - layout.cardGap / 2 }]}>
                  {filteredMovies.map((item) => (
                    <View key={`movie-${item.id}`} style={[styles.gridItem, { width: `${100 / numColumns}%`, padding: layout.cardGap / 2 }]}>
                      <MediaCard
                        item={item}
                        onPress={() => handleMoviePress(item)}
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
              </>
            )}

            {/* Series Section */}
            {filteredSeries.length > 0 && (
              <>
                {renderSectionHeader('Series', filteredSeries.length)}
                <View style={[styles.grid, { paddingHorizontal: layout.screenPadding - layout.cardGap / 2 }]}>
                  {filteredSeries.map((item) => (
                    <View key={`series-${item.id}`} style={[styles.gridItem, { width: `${100 / numColumns}%`, padding: layout.cardGap / 2 }]}>
                      <MediaCard
                        item={{
                          id: item.id,
                          title: item.title,
                          sort_title: item.sort_title,
                          poster_path: item.poster_path,
                          media_type: 'episode',
                          type: 'series',
                          genres: item.genres,
                          release_date: item.first_air_date,
                        }}
                        onPress={() => handleSeriesPress(item)}
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
              </>
            )}
          </>
        }
        contentContainerStyle={styles.listContent}
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
  chipRow: {
    flexDirection: 'row',
    paddingHorizontal: 16,
    paddingBottom: 12,
    gap: 8,
  },
  chip: {
    flexDirection: 'row',
    alignItems: 'center',
    paddingHorizontal: 14,
    paddingVertical: 8,
    borderRadius: 20,
    backgroundColor: '#1f1f1f',
    gap: 6,
  },
  chipActive: {
    backgroundColor: '#e50914',
  },
  chipText: {
    fontSize: 13,
    color: '#888',
  },
  chipTextActive: {
    color: '#fff',
    fontWeight: '600',
  },
  resultsHeader: {
    paddingHorizontal: 16,
    paddingBottom: 8,
  },
  resultsCount: {
    fontSize: 14,
    color: '#888',
  },
  listContent: {
    paddingBottom: 20,
  },
  sectionHeader: {
    paddingHorizontal: 16,
    paddingVertical: 12,
  },
  sectionTitle: {
    fontSize: 18,
    fontWeight: 'bold',
    color: '#fff',
  },
  grid: {
    flexDirection: 'row',
    flexWrap: 'wrap',
    paddingHorizontal: 12,
  },
  gridItem: {
    // width and padding set dynamically via inline style
  },
  loadingContainer: {
    flex: 1,
    justifyContent: 'center',
    alignItems: 'center',
  },
  emptyState: {
    flex: 1,
    justifyContent: 'center',
    alignItems: 'center',
    padding: 40,
  },
  emptyText: {
    fontSize: 14,
    color: '#888',
    textAlign: 'center',
  },
  noResultsContainer: {
    flex: 1,
    justifyContent: 'center',
    alignItems: 'center',
    padding: 40,
  },
  noResultsText: {
    fontSize: 16,
    color: '#888',
    textAlign: 'center',
  },
  recentSection: {
    paddingHorizontal: 16,
  },
  recentHeader: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
    paddingVertical: 12,
  },
  recentTitle: {
    fontSize: 16,
    fontWeight: '600',
    color: '#fff',
  },
  clearText: {
    fontSize: 14,
    color: '#e50914',
  },
  recentItem: {
    flexDirection: 'row',
    alignItems: 'center',
    paddingVertical: 12,
    borderBottomWidth: 1,
    borderBottomColor: '#2a2a2a',
    gap: 12,
  },
  recentText: {
    fontSize: 14,
    color: '#ccc',
  },
})
