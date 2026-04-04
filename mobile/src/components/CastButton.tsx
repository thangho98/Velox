/**
 * Chromecast button component
 *
 * Shows the native Cast device picker when tapped.
 * Changes color based on cast state:
 * - Gray: available but not connected
 * - White: no active media
 * - Blue (#4a9eff): actively casting media
 * - Hidden: no cast devices available
 */

import { TouchableOpacity, StyleSheet, type ViewStyle } from 'react-native'
import { Cast } from 'lucide-react-native'
import { useChromecast } from '../hooks/useChromecast'

interface CastButtonProps {
  style?: ViewStyle
  size?: number
}

export function CastButton({ style, size = 18 }: CastButtonProps) {
  const { available, connected, casting, showCastPicker } = useChromecast()

  const iconColor = casting ? '#4a9eff' : connected ? '#fff' : 'rgba(255,255,255,0.5)'

  return (
    <TouchableOpacity
      style={[styles.button, style]}
      onPress={showCastPicker}
      accessibilityLabel={connected ? 'Disconnect Chromecast' : 'Connect to Chromecast'}
      accessibilityRole="button"
    >
      <Cast size={size} color={iconColor} />
    </TouchableOpacity>
  )
}

const styles = StyleSheet.create({
  button: {
    width: 44,
    height: 44,
    borderRadius: 22,
    backgroundColor: 'rgba(255, 255, 255, 0.2)',
    justifyContent: 'center',
    alignItems: 'center',
  },
})
