import { View, Text, TouchableOpacity, StyleSheet } from 'react-native'
import { Film, Tv, Folder } from 'lucide-react-native'
import { useNavigation } from '@react-navigation/native'
import type { NativeStackNavigationProp } from '@react-navigation/native-stack'
import type { Library } from '@velox/shared/types'
import type { RootStackParamList } from '../../App'
import { useResponsiveLayout, scaledFont, cardWidth } from '../lib/responsive'

interface LibraryGridProps {
  libraries: Library[]
}

type NavigationProp = NativeStackNavigationProp<RootStackParamList>

export function LibraryGrid({ libraries }: LibraryGridProps) {
  const navigation = useNavigation<NavigationProp>()
  const layout = useResponsiveLayout()
  const cols = layout.libraryColumns
  const cw = cardWidth(cols, layout.width, layout.cardGap, layout.screenPadding)

  const iconSize = scaledFont(36, layout.fontScale)

  const getLibraryIcon = (type: string) => {
    switch (type) {
      case 'movie':
        return <Film size={iconSize} color="#e50914" />
      case 'series':
        return <Tv size={iconSize} color="#e50914" />
      default:
        return <Folder size={iconSize} color="#e50914" />
    }
  }

  const handlePress = (library: Library) => {
    navigation.navigate('LibraryBrowse', { id: library.id, name: library.name })
  }

  return (
    <View style={[styles.container, { paddingHorizontal: layout.screenPadding, gap: layout.cardGap }]}>
      {libraries.map((library) => (
        <TouchableOpacity
          key={library.id}
          style={[styles.card, { width: cw }]}
          onPress={() => handlePress(library)}
          activeOpacity={0.8}
        >
          <View style={styles.icon}>{getLibraryIcon(library.type)}</View>
          <Text style={[styles.name, { fontSize: scaledFont(16, layout.fontScale) }]}>{library.name}</Text>
          <Text style={[styles.type, { fontSize: scaledFont(12, layout.fontScale) }]}>{library.type}</Text>
        </TouchableOpacity>
      ))}
    </View>
  )
}

const styles = StyleSheet.create({
  container: {
    flexDirection: 'row',
    flexWrap: 'wrap',
  },
  card: {
    backgroundColor: '#1f1f1f',
    borderRadius: 12,
    padding: 20,
    alignItems: 'center',
  },
  icon: {
    marginBottom: 12,
  },
  name: {
    fontSize: 16,
    fontWeight: '600',
    color: '#fff',
    textAlign: 'center',
  },
  type: {
    fontSize: 12,
    color: '#888',
    marginTop: 4,
    textTransform: 'capitalize',
  },
})
