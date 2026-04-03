import { View, Text, ScrollView, TouchableOpacity, StyleSheet, Image } from 'react-native'
import { Film, Tv, Folder, Heart, Play } from 'lucide-react-native'
import { useLibraries } from '@velox/shared/hooks'
import type { Library } from '@velox/shared/types'
import { mediaImage } from '@velox/shared/lib'
import { Skeleton } from '../components/SkeletonLoader'
import { useResponsiveLayout, scaledFont, cardWidth } from '../lib/responsive'

interface LibraryCardProps {
  library: Library
  onPress: () => void
  cardW: number
  fontScale: number
}

function LibraryCard({ library, onPress, cardW, fontScale }: LibraryCardProps) {
  const iconSize = scaledFont(48, fontScale)

  const getLibraryIcon = (type: string) => {
    switch (type) {
      case 'movie':
        return <Film size={iconSize} color="#888" />
      case 'series':
        return <Tv size={iconSize} color="#888" />
      default:
        return <Folder size={iconSize} color="#888" />
    }
  }

  // Try to get poster from paths if available (paths[0] might be a folder path)
  const posterUrl = library.paths?.[0] ? mediaImage(library.paths[0], 'w500') : null

  return (
    <TouchableOpacity style={[styles.libraryCard, { width: cardW }]} onPress={onPress} activeOpacity={0.8}>
      {posterUrl ? (
        <Image source={{ uri: posterUrl }} style={styles.libraryPoster} resizeMode="cover" />
      ) : (
        <View style={styles.libraryPosterPlaceholder}>
          {getLibraryIcon(library.type)}
        </View>
      )}
      <View style={styles.libraryOverlay}>
        <Text style={[styles.libraryName, { fontSize: scaledFont(14, fontScale) }]}>{library.name}</Text>
        <Text style={[styles.libraryType, { fontSize: scaledFont(12, fontScale) }]}>{library.type}</Text>
      </View>
      <View style={styles.libraryBadge}>
        <View style={{ flexDirection: 'row', alignItems: 'center', gap: 4 }}>
          <Folder size={10} color="#fff" />
          <Text style={styles.libraryBadgeText}>Browse</Text>
        </View>
      </View>
    </TouchableOpacity>
  )
}

export function BrowseScreen() {
  const layout = useResponsiveLayout()
  const libCols = layout.libraryColumns
  const libCardW = cardWidth(libCols, layout.width, layout.cardGap, layout.screenPadding)
  const quickCols = layout.libraryColumns
  const quickCardW = cardWidth(quickCols, layout.width, layout.cardGap, layout.screenPadding)

  const { data: libraries, isLoading } = useLibraries()

  const handleLibraryPress = (library: Library) => {
    // Navigation is handled by the stack navigator
    // The actual folder browsing is done in LibraryBrowseScreen
  }

  return (
    <ScrollView style={styles.container}>
      <View style={[styles.section, { padding: layout.screenPadding, paddingBottom: layout.cardGap }]}>
        <Text style={[styles.sectionTitle, { fontSize: scaledFont(22, layout.fontScale) }]}>Your Libraries</Text>
        <Text style={[styles.sectionSubtitle, { fontSize: scaledFont(14, layout.fontScale) }]}>
          Select a library to browse its contents
        </Text>

        {isLoading ? (
          <View style={[styles.libraryGrid, { gap: layout.cardGap }]}>
            {[1, 2].map((i) => (
              <View key={i} style={[styles.librarySkeleton, { width: libCardW }]}>
                <Skeleton width="100%" height="100%" borderRadius={12} />
              </View>
            ))}
          </View>
        ) : libraries && libraries.length > 0 ? (
          <View style={[styles.libraryGrid, { gap: layout.cardGap }]}>
            {libraries.map((library) => (
              <LibraryCard
                key={library.id}
                library={library}
                onPress={() => handleLibraryPress(library)}
                cardW={libCardW}
                fontScale={layout.fontScale}
              />
            ))}
          </View>
        ) : (
          <View style={styles.emptyContainer}>
            <Folder size={64} color="#888" />
            <Text style={styles.emptyTitle}>No libraries yet</Text>
            <Text style={styles.emptySubtext}>
              Contact your administrator to add libraries
            </Text>
          </View>
        )}
      </View>

      {/* Quick Access Section */}
      <View style={[styles.section, { padding: layout.screenPadding, paddingBottom: layout.cardGap }]}>
        <Text style={[styles.sectionTitle, { fontSize: scaledFont(22, layout.fontScale) }]}>Quick Access</Text>
        <View style={[styles.quickAccessGrid, { gap: layout.cardGap, marginTop: layout.cardGap }]}>
          <TouchableOpacity style={[styles.quickAccessCard, { width: quickCardW }]}>
            <Film size={scaledFont(32, layout.fontScale)} color="#fff" />
            <Text style={[styles.quickAccessLabel, { fontSize: scaledFont(14, layout.fontScale) }]}>Movies</Text>
          </TouchableOpacity>
          <TouchableOpacity style={[styles.quickAccessCard, { width: quickCardW }]}>
            <Tv size={scaledFont(32, layout.fontScale)} color="#fff" />
            <Text style={[styles.quickAccessLabel, { fontSize: scaledFont(14, layout.fontScale) }]}>Series</Text>
          </TouchableOpacity>
          <TouchableOpacity style={[styles.quickAccessCard, { width: quickCardW }]}>
            <Heart size={scaledFont(32, layout.fontScale)} color="#e50914" fill="#e50914" />
            <Text style={[styles.quickAccessLabel, { fontSize: scaledFont(14, layout.fontScale) }]}>Favorites</Text>
          </TouchableOpacity>
          <TouchableOpacity style={[styles.quickAccessCard, { width: quickCardW }]}>
            <Play size={scaledFont(32, layout.fontScale)} color="#fff" />
            <Text style={[styles.quickAccessLabel, { fontSize: scaledFont(14, layout.fontScale) }]}>Continue</Text>
          </TouchableOpacity>
        </View>
      </View>
    </ScrollView>
  )
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
    backgroundColor: '#141414',
  },
  section: {
    padding: 16,
    paddingBottom: 8,
  },
  sectionTitle: {
    fontSize: 22,
    fontWeight: 'bold',
    color: '#fff',
    marginBottom: 4,
  },
  sectionSubtitle: {
    fontSize: 14,
    color: '#888',
    marginBottom: 16,
  },
  libraryGrid: {
    flexDirection: 'row',
    flexWrap: 'wrap',
    gap: 12,
  },
  libraryCard: {
    // width set dynamically via inline style
    aspectRatio: 2 / 3,
    borderRadius: 12,
    overflow: 'hidden',
    backgroundColor: '#1f1f1f',
    position: 'relative',
  },
  libraryPoster: {
    width: '100%',
    height: '100%',
    backgroundColor: '#2a2a2a',
  },
  libraryPosterPlaceholder: {
    width: '100%',
    height: '100%',
    backgroundColor: '#1f1f1f',
    justifyContent: 'center',
    alignItems: 'center',
  },
  libraryOverlay: {
    position: 'absolute',
    bottom: 0,
    left: 0,
    right: 0,
    padding: 12,
    backgroundColor: 'rgba(0, 0, 0, 0.7)',
  },
  libraryName: {
    fontSize: 14,
    fontWeight: 'bold',
    color: '#fff',
  },
  libraryType: {
    fontSize: 12,
    color: '#aaa',
    textTransform: 'capitalize',
    marginTop: 2,
  },
  libraryBadge: {
    position: 'absolute',
    top: 8,
    right: 8,
    backgroundColor: 'rgba(0, 0, 0, 0.6)',
    paddingHorizontal: 8,
    paddingVertical: 4,
    borderRadius: 12,
  },
  libraryBadgeText: {
    fontSize: 10,
    color: '#fff',
  },
  librarySkeleton: {
    // width set dynamically via inline style
    aspectRatio: 2 / 3,
    borderRadius: 12,
  },
  emptyContainer: {
    padding: 40,
    alignItems: 'center',
  },
  emptyTitle: {
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
  quickAccessGrid: {
    flexDirection: 'row',
    flexWrap: 'wrap',
    gap: 12,
    marginTop: 8,
  },
  quickAccessCard: {
    // width set dynamically via inline style
    backgroundColor: '#1f1f1f',
    borderRadius: 12,
    padding: 20,
    alignItems: 'center',
  },
  quickAccessLabel: {
    marginTop: 8,
    fontSize: 14,
    fontWeight: '600',
    color: '#fff',
  },
})
