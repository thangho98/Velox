import { useState } from 'react'
import {
  View,
  ScrollView,
  RefreshControl,
  StyleSheet,
  Text,
  TouchableOpacity,
} from 'react-native'
import { useNavigation } from '@react-navigation/native'
import type { NativeStackNavigationProp } from '@react-navigation/native-stack'
import {
  useContinueWatching,
  useNextUp,
  useMediaList,
  useSeriesList,
} from '@velox/shared/hooks'
import type { ContinueWatchingItem, NextUpItem } from '@velox/shared/types/playback'
import type { MediaListItem } from '@velox/shared/types'
import type { RootStackParamList } from '../../App'
import { useAuthStore } from '../stores'
import { SectionHeader } from '../components/SectionHeader'
import { HorizontalMediaRow } from '../components/HorizontalMediaRow'
import { LibraryGrid } from '../components/LibraryGrid'
import { MediaCard } from '../components/MediaCard'
import { HeroSkeleton, SectionSkeleton } from '../components/SkeletonLoader'
import { useResponsiveLayout, scaledFont, scaledSpacing } from '../lib/responsive'

type NavigationProp = NativeStackNavigationProp<RootStackParamList>

export function HomeScreen() {
  const navigation = useNavigation<NavigationProp>()
  const layout = useResponsiveLayout()
  const [refreshing, setRefreshing] = useState(false)
  const { user } = useAuthStore()

  // Fetch data
  const { data: continueWatching, isLoading: loadingContinue } = useContinueWatching({ limit: 10 })
  const { data: nextUp, isLoading: loadingNextUp } = useNextUp({ limit: 10 })
  const { data: recentMovies, isLoading: loadingMovies } = useMediaList({ type: 'movie', limit: 12 })
  const { data: rawSeries, isLoading: loadingSeries } = useSeriesList({ limit: 12 })

  // Map series data to MediaListItem
  const recentSeries: MediaListItem[] = (rawSeries || []).map((s) => ({
    id: s.id,
    title: s.title,
    sort_title: s.sort_title,
    poster_path: s.poster_path,
    media_type: 'episode' as const,
    type: 'series' as const,
    genres: [],
    release_date: s.first_air_date,
    series_id: s.id,
  }))

  const onRefresh = async () => {
    setRefreshing(true)
    setTimeout(() => setRefreshing(false), 1000)
  }

  const handleContinueWatchingPress = (item: ContinueWatchingItem) => {
    if (item.media_type === 'movie') {
      navigation.navigate('Media', { id: item.media_id })
    } else {
      // Navigate directly to video player, like web
      navigation.navigate('Episode', { id: item.media_id, seriesId: item.series_id })
    }
  }

  const handleNextUpPress = (item: NextUpItem) => {
    // Navigate directly to video player, like web
    navigation.navigate('Episode', { id: item.media_id, seriesId: item.series_id })
  }

  const handleMediaPress = (id: number, type: string | undefined) => {
    if (type === 'movie' || type === undefined) {
      navigation.navigate('Media', { id })
    } else {
      navigation.navigate('SeriesDetail', { id })
    }
  }

  const handleSeriesPress = (id: number) => {
    navigation.navigate('SeriesDetail', { id })
  }

  const isLoading =
    loadingContinue || loadingNextUp || loadingMovies || loadingSeries

  return (
    <ScrollView
      style={styles.container}
      refreshControl={
        <RefreshControl
          refreshing={refreshing}
          onRefresh={onRefresh}
          tintColor="#e50914"
        />
      }
    >
      {/* Hero Section */}
      <View style={[styles.hero, { padding: layout.screenPadding, paddingTop: scaledSpacing(16, layout.fontScale), paddingBottom: scaledSpacing(24, layout.fontScale) }]}>
        <Text style={[styles.greeting, { fontSize: scaledFont(26, layout.fontScale) }]}>
          Welcome back{user?.display_name ? `, ${user.display_name}` : ''}
        </Text>
        <Text style={[styles.subtitle, { fontSize: scaledFont(14, layout.fontScale), marginBottom: scaledSpacing(20, layout.fontScale) }]}>Your personal media server.</Text>
        <View style={[styles.ctaRow, { gap: scaledSpacing(12, layout.fontScale) }]}>
          <TouchableOpacity
            style={[styles.ctaButton, { paddingVertical: scaledSpacing(12, layout.fontScale), paddingHorizontal: scaledSpacing(24, layout.fontScale) }]}
            onPress={() => navigation.navigate('HomeTabs', { screen: 'Movies' } as any)}
          >
            <Text style={[styles.ctaButtonText, { fontSize: scaledFont(15, layout.fontScale) }]}>Movies</Text>
          </TouchableOpacity>
          <TouchableOpacity
            style={[styles.ctaButton, styles.ctaButtonOutline, { paddingVertical: scaledSpacing(12, layout.fontScale), paddingHorizontal: scaledSpacing(24, layout.fontScale) }]}
            onPress={() => navigation.navigate('HomeTabs', { screen: 'Series' } as any)}
          >
            <Text style={[styles.ctaButtonText, styles.ctaButtonTextOutline, { fontSize: scaledFont(15, layout.fontScale) }]}>Series</Text>
          </TouchableOpacity>
        </View>
      </View>

      {/* Continue Watching */}
      {continueWatching && continueWatching.length > 0 && (
        <View style={styles.section}>
          <SectionHeader title="Continue Watching" />
          <ScrollView
            horizontal
            showsHorizontalScrollIndicator={false}
            contentContainerStyle={[styles.horizontalContent, { paddingHorizontal: layout.screenPadding }]}
          >
            {continueWatching.map((item) => (
              <MediaCard
                key={item.media_id}
                item={{
                  id: item.media_id,
                  title: item.title,
                  sort_title: item.title,
                  poster_path: item.poster_path,
                  media_type: item.media_type,
                  genres: [],
                }}
                onPress={() => handleContinueWatchingPress(item)}
                continueWatching={item}
              />
            ))}
          </ScrollView>
        </View>
      )}

      {/* Next Up */}
      {nextUp && nextUp.length > 0 && (
        <View style={styles.section}>
          <SectionHeader
            title="Next Up"
            onSeeAll={() => navigation.navigate('HomeTabs', { screen: 'Series' } as any)}
          />
          <ScrollView
            horizontal
            showsHorizontalScrollIndicator={false}
            contentContainerStyle={[styles.horizontalContent, { paddingHorizontal: layout.screenPadding }]}
          >
            {nextUp.map((item) => (
              <MediaCard
                key={item.media_id}
                item={{
                  id: item.media_id,
                  title: item.series_title,
                  sort_title: item.series_title,
                  poster_path: item.series_poster,
                  media_type: 'episode',
                  genres: [],
                }}
                onPress={() => handleNextUpPress(item)}
                continueWatching={{
                  media_id: item.media_id,
                  series_id: item.series_id,
                  position: 0,
                  completed: false,
                  title: item.episode_title,
                  poster_path: item.series_poster,
                  backdrop_path: item.backdrop_path,
                  media_type: 'episode',
                  duration: item.duration,
                  series_title: item.series_title,
                  season_number: item.season_number,
                  episode_number: item.episode_number,
                }}
              />
            ))}
          </ScrollView>
        </View>
      )}

      {/* Recently Added Movies */}
      {recentMovies && recentMovies.length > 0 && (
        <View style={styles.section}>
          <SectionHeader
            title="Movies"
            onSeeAll={() => navigation.navigate('HomeTabs', { screen: 'Movies' } as any)}
          />
          <HorizontalMediaRow
            items={recentMovies}
            onItemPress={(item) => handleMediaPress(item.id, item.type)}
            showBadge
          />
        </View>
      )}

      {/* Recently Added Series */}
      {recentSeries.length > 0 && (
        <View style={styles.section}>
          <SectionHeader
            title="Series"
            onSeeAll={() => navigation.navigate('HomeTabs', { screen: 'Series' } as any)}
          />
          <HorizontalMediaRow
            items={recentSeries}
            onItemPress={(item) => handleSeriesPress(item.id)}
            showBadge
          />
        </View>
      )}

      {/* Libraries */}
      {/*{libraries && libraries.length > 0 && (
        <View style={styles.section}>
          <SectionHeader title="Libraries" />
          <LibraryGrid libraries={libraries} />
        </View>
      )}*/}

      {/* Loading state - show skeletons */}
      {isLoading && (
        <View>
          <HeroSkeleton />
          <SectionSkeleton titleWidth={160} />
          <SectionSkeleton titleWidth={120} />
          <SectionSkeleton titleWidth={140} />
        </View>
      )}

      {/* Empty state */}
      {!isLoading &&
        !continueWatching?.length &&
        !nextUp?.length &&
        !recentMovies?.length &&
        !recentSeries.length && (
          <View style={styles.centered}>
            <Text style={[styles.emptyText, { fontSize: scaledFont(18, layout.fontScale) }]}>No media yet</Text>
            <Text style={[styles.emptySubtext, { fontSize: scaledFont(14, layout.fontScale) }]}>
              Add some movies or TV shows to get started
            </Text>
          </View>
        )}
    </ScrollView>
  )
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
    backgroundColor: '#141414',
  },
  hero: {
    padding: 20,
    paddingTop: 16,
    paddingBottom: 24,
  },
  greeting: {
    fontSize: 26,
    fontWeight: 'bold',
    color: '#fff',
    marginBottom: 4,
  },
  subtitle: {
    fontSize: 14,
    color: '#888',
    marginBottom: 20,
  },
  ctaRow: {
    flexDirection: 'row',
    gap: 12,
  },
  ctaButton: {
    backgroundColor: '#e50914',
    paddingVertical: 12,
    paddingHorizontal: 24,
    borderRadius: 8,
    alignItems: 'center',
    flex: 1,
  },
  ctaButtonOutline: {
    backgroundColor: 'transparent',
    borderWidth: 1,
    borderColor: 'rgba(255, 255, 255, 0.2)',
  },
  ctaButtonText: {
    color: '#fff',
    fontSize: 15,
    fontWeight: '600',
  },
  ctaButtonTextOutline: {
    color: '#fff',
  },
  section: {
    marginBottom: 24,
  },
  horizontalContent: {
    paddingHorizontal: 16,
  },
  centered: {
    padding: 40,
    alignItems: 'center',
  },
  emptyText: {
    color: '#fff',
    fontSize: 18,
    fontWeight: '600',
  },
  emptySubtext: {
    color: '#888',
    fontSize: 14,
    marginTop: 8,
    textAlign: 'center',
  },
})
