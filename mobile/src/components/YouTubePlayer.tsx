/**
 * YouTube Player for Cinema Mode - Mobile version
 * Uses YouTube app intent or web browser for playback
 */

import { useState } from 'react'
import { View, StyleSheet, TouchableOpacity, Text, Modal, Dimensions, Linking } from 'react-native'
import { Film, Play } from 'lucide-react-native'

interface YouTubePlayerProps {
  videoId: string
  visible: boolean
  onClose?: () => void
  onMutedChange?: (muted: boolean) => void
}

const { width: SCREEN_WIDTH } = Dimensions.get('window')

export function YouTubePlayer({ videoId, visible, onClose }: YouTubePlayerProps) {
  const [error, setError] = useState(false)

  if (!visible) return null

  const handleOpenYouTube = async () => {
    try {
      // Try to open YouTube app
      const youtubeUrl = `vnd.youtube://${videoId}`
      const canOpen = await Linking.canOpenURL(youtubeUrl)

      if (canOpen) {
        await Linking.openURL(youtubeUrl)
      } else {
        // Fallback to web
        await Linking.openURL(`https://www.youtube.com/watch?v=${videoId}`)
      }
      onClose?.()
    } catch {
      setError(true)
    }
  }

  const handleOpenWeb = async () => {
    try {
      await Linking.openURL(`https://www.youtube.com/watch?v=${videoId}`)
      onClose?.()
    } catch {
      setError(true)
    }
  }

  return (
    <Modal
      visible={visible}
      animationType="fade"
      transparent
      onRequestClose={onClose}
    >
      <View style={styles.container}>
        {/* Backdrop */}
        <TouchableOpacity style={styles.backdrop} onPress={onClose} activeOpacity={1} />

        {/* Player Card */}
        <View style={styles.playerCard}>
          <View style={{ flexDirection: 'row', alignItems: 'center', gap: 8, justifyContent: 'center' }}>
            <Film size={22} color="#fff" />
            <Text style={styles.title}>Watch Trailer</Text>
          </View>
          <Text style={styles.subtitle}>
            Choose how to watch the trailer
          </Text>

          {/* YouTube App button */}
          <TouchableOpacity style={styles.youtubeButton} onPress={handleOpenYouTube}>
            <Play size={20} color="#fff" />
            <Text style={styles.youtubeButtonText}>Open in YouTube App</Text>
          </TouchableOpacity>

          {/* Web browser button */}
          <TouchableOpacity style={styles.webButton} onPress={handleOpenWeb}>
            <Text style={styles.webButtonIcon}>🌐</Text>
            <Text style={styles.webButtonText}>Open in Browser</Text>
          </TouchableOpacity>

          {error && (
            <Text style={styles.errorText}>Could not open YouTube. Please try again.</Text>
          )}

          {/* Close button */}
          <TouchableOpacity style={styles.closeButton} onPress={onClose}>
            <Text style={styles.closeButtonText}>Cancel</Text>
          </TouchableOpacity>
        </View>
      </View>
    </Modal>
  )
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
    justifyContent: 'center',
    alignItems: 'center',
    backgroundColor: 'rgba(0, 0, 0, 0.9)',
  },
  backdrop: {
    ...StyleSheet.absoluteFillObject,
  },
  playerCard: {
    width: SCREEN_WIDTH - 60,
    backgroundColor: '#1a1a1a',
    borderRadius: 16,
    padding: 24,
    alignItems: 'center',
  },
  title: {
    color: '#fff',
    fontSize: 22,
    fontWeight: 'bold',
    marginBottom: 8,
  },
  subtitle: {
    color: '#888',
    fontSize: 14,
    marginBottom: 24,
    textAlign: 'center',
  },
  youtubeButton: {
    flexDirection: 'row',
    alignItems: 'center',
    backgroundColor: '#e50914',
    paddingVertical: 14,
    paddingHorizontal: 24,
    borderRadius: 8,
    width: '100%',
    marginBottom: 12,
    gap: 12,
  },
  youtubeButtonText: {
    color: '#fff',
    fontSize: 16,
    fontWeight: '600',
  },
  webButton: {
    flexDirection: 'row',
    alignItems: 'center',
    backgroundColor: '#2a2a2a',
    paddingVertical: 14,
    paddingHorizontal: 24,
    borderRadius: 8,
    width: '100%',
    marginBottom: 12,
    gap: 12,
    borderWidth: 1,
    borderColor: '#444',
  },
  webButtonIcon: {
    fontSize: 20,
  },
  webButtonText: {
    color: '#fff',
    fontSize: 16,
    fontWeight: '600',
  },
  errorText: {
    color: '#e50914',
    fontSize: 13,
    marginTop: 8,
    textAlign: 'center',
  },
  closeButton: {
    marginTop: 16,
    paddingVertical: 12,
  },
  closeButtonText: {
    color: '#888',
    fontSize: 15,
  },
})
