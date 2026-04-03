import { ScrollView, TouchableOpacity, Text, View, StyleSheet } from 'react-native'
import { House, ChevronRight } from 'lucide-react-native'

interface BreadcrumbProps {
  libraryName: string
  path: string // e.g. "Action/Marvel" — split by "/" for segments
  onNavigate: (path: string) => void // called with the path to navigate to
  onHome: () => void // navigate back to library list
}

export function Breadcrumb({ libraryName, path, onNavigate, onHome }: BreadcrumbProps) {
  // Build segments from the path
  const pathSegments = path ? path.split('/').filter(Boolean) : []

  // Build breadcrumb items: Home > Library > Folder segments...
  const segments: { label: string; path: string; isCurrent: boolean }[] = []

  // Library is always the first segment after home
  segments.push({
    label: libraryName,
    path: '',
    isCurrent: pathSegments.length === 0,
  })

  // Add folder path segments
  pathSegments.forEach((segment, index) => {
    const segmentPath = pathSegments.slice(0, index + 1).join('/')
    segments.push({
      label: segment,
      path: segmentPath,
      isCurrent: index === pathSegments.length - 1,
    })
  })

  return (
    <View style={styles.container}>
      <ScrollView
        horizontal
        showsHorizontalScrollIndicator={false}
        contentContainerStyle={styles.content}
      >
        {/* Home icon */}
        <TouchableOpacity
          style={styles.homeButton}
          onPress={onHome}
          activeOpacity={0.7}
        >
          <House size={14} color="#888" />
        </TouchableOpacity>

        {/* Segments */}
        {segments.map((segment, index) => (
          <View key={segment.path + index} style={styles.segmentRow}>
            <ChevronRight size={12} color="#555" />
            {segment.isCurrent ? (
              <Text style={styles.currentText} numberOfLines={1}>
                {segment.label}
              </Text>
            ) : (
              <TouchableOpacity
                onPress={() => onNavigate(segment.path)}
                activeOpacity={0.7}
              >
                <Text style={styles.linkText} numberOfLines={1}>
                  {segment.label}
                </Text>
              </TouchableOpacity>
            )}
          </View>
        ))}
      </ScrollView>
    </View>
  )
}

const styles = StyleSheet.create({
  container: {
    height: 36,
    backgroundColor: '#141414',
    borderBottomWidth: StyleSheet.hairlineWidth,
    borderBottomColor: '#1f1f1f',
  },
  content: {
    alignItems: 'center',
    paddingHorizontal: 16,
    gap: 4,
  },
  homeButton: {
    width: 28,
    height: 28,
    borderRadius: 14,
    backgroundColor: '#1f1f1f',
    justifyContent: 'center',
    alignItems: 'center',
  },
  segmentRow: {
    flexDirection: 'row',
    alignItems: 'center',
    gap: 4,
  },
  linkText: {
    fontSize: 13,
    color: '#888',
    maxWidth: 120,
  },
  currentText: {
    fontSize: 13,
    color: '#fff',
    fontWeight: '600',
    maxWidth: 160,
  },
})
