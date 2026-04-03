/**
 * PlayOptionsBottomSheet - Choose subtitle/audio/quality before playing
 */

import { useState } from 'react'
import {
  View,
  Text,
  StyleSheet,
  Modal,
  TouchableOpacity,
  ScrollView,
} from 'react-native'
import { Check, X, Play } from 'lucide-react-native'
import { useResponsiveLayout, scaledFont } from '../lib/responsive'

// ── Types ─────────────────────────────────────────────────────────────────────

interface SubtitleTrack {
  id: string
  label: string
  language: string
}

interface AudioTrack {
  id: string
  label: string
  language: string
}

interface QualityOption {
  label: string
  height: number
}

interface PlayOptionsBottomSheetProps {
  visible: boolean
  title: string
  subtitles?: SubtitleTrack[]
  audioTracks?: AudioTrack[]
  qualities?: QualityOption[]
  onPlay: (options: PlayOptions) => void
  onClose: () => void
}

export interface PlayOptions {
  subtitleId?: string
  audioId?: string
  quality?: string
}

// ── Component ─────────────────────────────────────────────────────────────────

export function PlayOptionsBottomSheet({
  visible,
  title,
  subtitles = [],
  audioTracks = [],
  qualities = [],
  onPlay,
  onClose,
}: PlayOptionsBottomSheetProps) {
  const layout = useResponsiveLayout()
  const isWideScreen = layout.device !== 'phone'

  const [selectedSubtitle, setSelectedSubtitle] = useState<string | undefined>(undefined)
  const [selectedAudio, setSelectedAudio] = useState<string | undefined>(undefined)
  const [selectedQuality, setSelectedQuality] = useState<string>('Auto')

  const handlePlay = () => {
    onPlay({
      subtitleId: selectedSubtitle,
      audioId: selectedAudio,
      quality: selectedQuality,
    })
    onClose()
  }

  const checkSize = layout.largeControls ? 22 : 16
  const optionPadding = layout.largeControls ? 16 : 12
  const buttonHeight = layout.largeControls ? 56 : 44

  const renderOption = (
    label: string,
    isSelected: boolean,
    onPress: () => void
  ) => (
    <TouchableOpacity
      style={[styles.option, isSelected && styles.optionSelected, { paddingVertical: optionPadding }]}
      onPress={onPress}
    >
      <Text style={[styles.optionText, isSelected && styles.optionTextSelected, { fontSize: scaledFont(15, layout.fontScale) }]}>
        {label}
      </Text>
      {isSelected && <Check size={checkSize} color="#e50914" />}
    </TouchableOpacity>
  )

  const renderNoneOption = (isSelected: boolean, onPress: () => void) => (
    <TouchableOpacity
      style={[styles.option, isSelected && styles.optionSelected, { paddingVertical: optionPadding }]}
      onPress={onPress}
    >
      <Text style={[styles.optionText, isSelected && styles.optionTextSelected, { fontSize: scaledFont(15, layout.fontScale) }]}>
        None
      </Text>
      {isSelected && <Check size={checkSize} color="#e50914" />}
    </TouchableOpacity>
  )

  return (
    <Modal
      visible={visible}
      transparent
      animationType={isWideScreen ? 'fade' : 'slide'}
      onRequestClose={onClose}
    >
      <TouchableOpacity style={[styles.backdrop, isWideScreen && { justifyContent: 'center', alignItems: 'center' }]} activeOpacity={1} onPress={onClose} />
      <View style={[
        styles.sheet,
        isWideScreen && {
          maxWidth: layout.largeControls ? 600 : 500,
          alignSelf: 'center',
          width: '90%',
          borderRadius: 16,
          position: 'absolute',
          top: layout.height * 0.15,
        },
      ]}>
        {/* Header */}
        <View style={styles.header}>
          <Text style={[styles.headerTitle, { fontSize: scaledFont(18, layout.fontScale) }]}>Playback Options</Text>
          <TouchableOpacity style={styles.closeButton} onPress={onClose}>
            <X size={16} color="#888" />
          </TouchableOpacity>
        </View>

        <ScrollView style={styles.content} showsVerticalScrollIndicator={false}>
          {/* Title */}
          <Text style={[styles.mediaTitle, { fontSize: scaledFont(16, layout.fontScale) }]} numberOfLines={2}>{title}</Text>

          {/* Subtitles Section */}
          {subtitles.length > 0 && (
            <View style={styles.section}>
              <Text style={[styles.sectionTitle, { fontSize: scaledFont(12, layout.fontScale) }]}>Subtitles</Text>
              {renderNoneOption(selectedSubtitle === undefined, () => setSelectedSubtitle(undefined))}
              {subtitles.map((track) =>
                renderOption(
                  `${track.label} ${track.language ? `(${track.language})` : ''}`,
                  selectedSubtitle === track.id,
                  () => setSelectedSubtitle(track.id)
                )
              )}
            </View>
          )}

          {/* Audio Section */}
          {audioTracks.length > 0 && (
            <View style={styles.section}>
              <Text style={[styles.sectionTitle, { fontSize: scaledFont(12, layout.fontScale) }]}>Audio</Text>
              {audioTracks.map((track) =>
                renderOption(
                  `${track.label} ${track.language ? `(${track.language})` : ''}`,
                  selectedAudio === track.id,
                  () => setSelectedAudio(track.id)
                )
              )}
            </View>
          )}

          {/* Quality Section */}
          {qualities.length > 0 && (
            <View style={styles.section}>
              <Text style={[styles.sectionTitle, { fontSize: scaledFont(12, layout.fontScale) }]}>Quality</Text>
              {qualities.map((q) =>
                renderOption(
                  q.label,
                  selectedQuality === q.label,
                  () => setSelectedQuality(q.label)
                )
              )}
            </View>
          )}

          {/* No options available */}
          {subtitles.length === 0 && audioTracks.length === 0 && qualities.length === 0 && (
            <View style={styles.noOptions}>
              <Text style={styles.noOptionsText}>
                No additional options available.{'\n'}Playing with default settings.
              </Text>
            </View>
          )}
        </ScrollView>

        {/* Play Button */}
        <TouchableOpacity style={[styles.playButton, { paddingVertical: layout.largeControls ? 20 : 16 }]} onPress={handlePlay}>
          <View style={{ flexDirection: 'row', alignItems: 'center', gap: 8 }}>
            <Play size={scaledFont(18, layout.fontScale)} color="#fff" />
            <Text style={[styles.playButtonText, { fontSize: scaledFont(17, layout.fontScale) }]}>Play</Text>
          </View>
        </TouchableOpacity>
      </View>
    </Modal>
  )
}

// ── Styles ─────────────────────────────────────────────────────────────────────

const styles = StyleSheet.create({
  backdrop: {
    flex: 1,
    backgroundColor: 'rgba(0, 0, 0, 0.6)',
  },
  sheet: {
    backgroundColor: '#141414',
    borderTopLeftRadius: 20,
    borderTopRightRadius: 20,
    maxHeight: '80%',
  },
  header: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
    padding: 16,
    borderBottomWidth: 1,
    borderBottomColor: '#2a2a2a',
  },
  headerTitle: {
    color: '#fff',
    fontSize: 18,
    fontWeight: '600',
  },
  closeButton: {
    width: 32,
    height: 32,
    borderRadius: 16,
    backgroundColor: '#2a2a2a',
    justifyContent: 'center',
    alignItems: 'center',
  },
  content: {
    padding: 16,
    maxHeight: 400,
  },
  mediaTitle: {
    color: '#fff',
    fontSize: 16,
    fontWeight: '500',
    marginBottom: 16,
    paddingBottom: 16,
    borderBottomWidth: 1,
    borderBottomColor: '#2a2a2a',
  },
  section: {
    marginBottom: 20,
  },
  sectionTitle: {
    color: '#888',
    fontSize: 12,
    fontWeight: '600',
    textTransform: 'uppercase',
    marginBottom: 10,
  },
  option: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
    paddingVertical: 12,
    paddingHorizontal: 14,
    backgroundColor: '#1f1f1f',
    borderRadius: 8,
    marginBottom: 8,
  },
  optionSelected: {
    backgroundColor: 'rgba(229, 9, 20, 0.2)',
    borderWidth: 1,
    borderColor: '#e50914',
  },
  optionText: {
    color: '#ccc',
    fontSize: 15,
  },
  optionTextSelected: {
    color: '#fff',
    fontWeight: '500',
  },
  noOptions: {
    padding: 20,
    alignItems: 'center',
  },
  noOptionsText: {
    color: '#888',
    fontSize: 14,
    textAlign: 'center',
    lineHeight: 22,
  },
  playButton: {
    backgroundColor: '#e50914',
    margin: 16,
    marginBottom: 34,
    paddingVertical: 16,
    borderRadius: 8,
    alignItems: 'center',
  },
  playButtonText: {
    color: '#fff',
    fontSize: 17,
    fontWeight: '600',
  },
})
