/**
 * CinemaOverlay - Pre-roll cinema experience before main video playback
 *
 * Plays intro video (via expo-video) and/or YouTube trailers sequentially.
 * User can skip individual items or skip to the main movie at any time.
 */

import { useState, useEffect, useRef } from 'react'
import {
  View,
  Text,
  StyleSheet,
  TouchableOpacity,
  ActivityIndicator,
  Dimensions,
  Linking,
} from 'react-native'
import { Film, SkipForward, ChevronRight } from 'lucide-react-native'
import { useVideoPlayer, VideoView } from 'expo-video'
import type { CinemaItem } from '@velox/shared/hooks/media/useCinema'

export type { CinemaItem }

interface CinemaOverlayProps {
  items: CinemaItem[]
  onComplete: () => void
  serverUrl: string
  accessToken: string | null
}

const { width: SCREEN_WIDTH } = Dimensions.get('window')

/** Extract YouTube video key from embed URL */
function extractYouTubeKey(url: string): string | null {
  const match = url.match(/embed\/([^?]+)/)
  return match ? match[1] : null
}

export function CinemaOverlay({ items, onComplete, serverUrl, accessToken }: CinemaOverlayProps) {
  const [currentIndex, setCurrentIndex] = useState(0)
  const hasCompletedRef = useRef(false)

  // Guard: no items -> complete immediately
  useEffect(() => {
    if (items.length === 0 && !hasCompletedRef.current) {
      hasCompletedRef.current = true
      onComplete()
    }
  }, [items.length, onComplete])

  if (items.length === 0) return null

  const currentItem = items[currentIndex]
  if (!currentItem) {
    // All items played
    if (!hasCompletedRef.current) {
      hasCompletedRef.current = true
      onComplete()
    }
    return null
  }

  const isIntro = currentItem.type === 'intro'
  const isTrailer = currentItem.type === 'trailer'
  const trailerIndex = items.slice(0, currentIndex + 1).filter((i) => i.type === 'trailer').length
  const totalTrailers = items.filter((i) => i.type === 'trailer').length

  const handleSkip = () => {
    const nextIndex = currentIndex + 1
    if (nextIndex >= items.length) {
      if (!hasCompletedRef.current) {
        hasCompletedRef.current = true
        onComplete()
      }
    } else {
      setCurrentIndex(nextIndex)
    }
  }

  const handleSkipAll = () => {
    if (!hasCompletedRef.current) {
      hasCompletedRef.current = true
      onComplete()
    }
  }

  const handleItemEnded = () => {
    handleSkip()
  }

  // Label for the top bar
  const itemLabel = isIntro
    ? 'Cinema Intro'
    : `Trailer ${trailerIndex} of ${totalTrailers}`

  return (
    <View style={styles.container}>
      {/* Media content */}
      {isIntro && (
        <IntroPlayer
          serverUrl={serverUrl}
          accessToken={accessToken}
          onEnded={handleItemEnded}
        />
      )}
      {isTrailer && (
        <TrailerCard
          item={currentItem}
          onOpenTrailer={() => {
            const key = extractYouTubeKey(currentItem.url)
            if (key) {
              openYouTube(key)
            }
          }}
        />
      )}

      {/* Top bar: gradient + item info */}
      <View style={styles.topBar}>
        <View style={styles.topBarContent}>
          <Film size={16} color="#FFC107" />
          <Text style={styles.topBarLabel}>{itemLabel}</Text>
          <Text style={styles.topBarCounter}>
            {currentIndex + 1} / {items.length}
          </Text>
        </View>
      </View>

      {/* Bottom bar: skip controls */}
      <View style={styles.bottomBar}>
        <Text style={styles.itemTitle} numberOfLines={1}>
          {currentItem.title}
        </Text>
        <View style={styles.buttonRow}>
          {currentItem.skippable && (
            <TouchableOpacity style={styles.skipButton} onPress={handleSkip}>
              <SkipForward size={16} color="#fff" />
              <Text style={styles.skipButtonText}>Skip</Text>
            </TouchableOpacity>
          )}
          <TouchableOpacity style={styles.skipAllButton} onPress={handleSkipAll}>
            <Text style={styles.skipAllButtonText}>Skip to Movie</Text>
            <ChevronRight size={16} color="rgba(255,255,255,0.7)" />
          </TouchableOpacity>
        </View>
      </View>
    </View>
  )
}

// --- Sub-components ---

/** Plays the cinema intro video using expo-video */
function IntroPlayer({
  serverUrl,
  accessToken,
  onEnded,
}: {
  serverUrl: string
  accessToken: string | null
  onEnded: () => void
}) {
  const introUrl = accessToken
    ? `${serverUrl}/api/cinema/intro?token=${encodeURIComponent(accessToken)}`
    : `${serverUrl}/api/cinema/intro`

  const hasEndedRef = useRef(false)

  const player = useVideoPlayer(introUrl, (p) => {
    p.volume = 1
    p.muted = false
    p.play()
  })

  // Poll for playback end
  useEffect(() => {
    if (!player) return
    const interval = setInterval(() => {
      if (player.status === 'idle' && hasEndedRef.current) return
      // Check if video ended (duration reached or status changed)
      if (
        player.duration > 0 &&
        player.currentTime > 0 &&
        player.currentTime >= player.duration - 0.5
      ) {
        if (!hasEndedRef.current) {
          hasEndedRef.current = true
          onEnded()
        }
      }
    }, 500)
    return () => clearInterval(interval)
  }, [player, onEnded])

  if (!player) {
    return (
      <View style={styles.loadingContainer}>
        <ActivityIndicator size="large" color="#e50914" />
      </View>
    )
  }

  return (
    // @ts-ignore -- expo-video VideoView JSX type issue with strict mode
    <VideoView
      style={StyleSheet.absoluteFill}
      player={player}
      nativeControls={false}
      contentFit="contain"
    />
  )
}

/** Shows trailer info card with button to open YouTube */
function TrailerCard({
  item,
  onOpenTrailer,
}: {
  item: CinemaItem
  onOpenTrailer: () => void
}) {
  return (
    <View style={styles.trailerContainer}>
      <View style={styles.trailerCard}>
        <Film size={40} color="#FFC107" />
        <Text style={styles.trailerTitle} numberOfLines={2}>
          {item.title}
        </Text>
        <Text style={styles.trailerSubtitle}>YouTube Trailer</Text>
        <TouchableOpacity style={styles.openYouTubeButton} onPress={onOpenTrailer}>
          <Text style={styles.openYouTubeText}>Watch in YouTube</Text>
        </TouchableOpacity>
      </View>
    </View>
  )
}

/** Open YouTube via deep link or browser */
async function openYouTube(videoId: string) {
  try {
    const youtubeUrl = `vnd.youtube://${videoId}`
    const canOpen = await Linking.canOpenURL(youtubeUrl)
    if (canOpen) {
      await Linking.openURL(youtubeUrl)
    } else {
      await Linking.openURL(`https://www.youtube.com/watch?v=${videoId}`)
    }
  } catch {
    // Silently fail — user can skip
  }
}

// --- Styles ---

const styles = StyleSheet.create({
  container: {
    flex: 1,
    backgroundColor: '#000',
  },
  loadingContainer: {
    flex: 1,
    justifyContent: 'center',
    alignItems: 'center',
    backgroundColor: '#000',
  },

  // Top bar
  topBar: {
    position: 'absolute',
    top: 0,
    left: 0,
    right: 0,
    paddingTop: 52,
    paddingHorizontal: 20,
    paddingBottom: 20,
    // Gradient effect via background
    backgroundColor: 'rgba(0, 0, 0, 0.6)',
  },
  topBarContent: {
    flexDirection: 'row',
    alignItems: 'center',
    gap: 10,
  },
  topBarLabel: {
    color: 'rgba(255, 255, 255, 0.85)',
    fontSize: 14,
    fontWeight: '600',
  },
  topBarCounter: {
    color: 'rgba(255, 255, 255, 0.4)',
    fontSize: 12,
  },

  // Bottom bar
  bottomBar: {
    position: 'absolute',
    bottom: 0,
    left: 0,
    right: 0,
    paddingBottom: 40,
    paddingHorizontal: 20,
    paddingTop: 20,
    backgroundColor: 'rgba(0, 0, 0, 0.6)',
  },
  itemTitle: {
    color: 'rgba(255, 255, 255, 0.6)',
    fontSize: 13,
    marginBottom: 12,
  },
  buttonRow: {
    flexDirection: 'row',
    alignItems: 'center',
    justifyContent: 'flex-end',
    gap: 12,
  },
  skipButton: {
    flexDirection: 'row',
    alignItems: 'center',
    gap: 6,
    backgroundColor: 'rgba(255, 255, 255, 0.2)',
    paddingHorizontal: 16,
    paddingVertical: 10,
    borderRadius: 8,
  },
  skipButtonText: {
    color: '#fff',
    fontSize: 14,
    fontWeight: '600',
  },
  skipAllButton: {
    flexDirection: 'row',
    alignItems: 'center',
    gap: 4,
    borderWidth: 1,
    borderColor: 'rgba(255, 255, 255, 0.3)',
    paddingHorizontal: 16,
    paddingVertical: 10,
    borderRadius: 8,
  },
  skipAllButtonText: {
    color: 'rgba(255, 255, 255, 0.7)',
    fontSize: 14,
  },

  // Trailer card
  trailerContainer: {
    flex: 1,
    justifyContent: 'center',
    alignItems: 'center',
    backgroundColor: '#000',
  },
  trailerCard: {
    width: SCREEN_WIDTH - 80,
    backgroundColor: '#1a1a1a',
    borderRadius: 16,
    padding: 32,
    alignItems: 'center',
  },
  trailerTitle: {
    color: '#fff',
    fontSize: 18,
    fontWeight: 'bold',
    marginTop: 16,
    textAlign: 'center',
  },
  trailerSubtitle: {
    color: '#888',
    fontSize: 13,
    marginTop: 6,
    marginBottom: 20,
  },
  openYouTubeButton: {
    backgroundColor: '#e50914',
    paddingHorizontal: 24,
    paddingVertical: 12,
    borderRadius: 8,
  },
  openYouTubeText: {
    color: '#fff',
    fontSize: 15,
    fontWeight: '600',
  },
})
