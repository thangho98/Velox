/**
 * EpisodeEditDialog - Bottom sheet modal for editing episode metadata
 *
 * Admin-only feature. Allows editing title, overview, air date, episode number,
 * and metadata lock toggle.
 */

import { useState, useEffect } from 'react'
import {
  View,
  Text,
  StyleSheet,
  Modal,
  TouchableOpacity,
  TextInput,
  ScrollView,
  KeyboardAvoidingView,
  Platform,
  Switch,
} from 'react-native'
import { X, Save } from 'lucide-react-native'
import type { Episode, EpisodeMetadataEditRequest } from '@velox/shared/types'
import { useResponsiveLayout, scaledFont } from '../lib/responsive'

interface EpisodeEditDialogProps {
  episode: Episode
  onSave: (episodeId: number, req: EpisodeMetadataEditRequest) => void
  isSaving: boolean
  onClose: () => void
}

export function EpisodeEditDialog({ episode, onSave, isSaving, onClose }: EpisodeEditDialogProps) {
  const layout = useResponsiveLayout()
  const isWideScreen = layout.device !== 'phone'

  const [title, setTitle] = useState(episode.title)
  const [overview, setOverview] = useState(episode.overview ?? '')
  const [airDate, setAirDate] = useState(episode.air_date ?? '')
  const [episodeNumber, setEpisodeNumber] = useState(String(episode.episode_number))
  const [metadataLocked, setMetadataLocked] = useState(false)

  useEffect(() => {
    setTitle(episode.title)
    setOverview(episode.overview ?? '')
    setAirDate(episode.air_date ?? '')
    setEpisodeNumber(String(episode.episode_number))
    setMetadataLocked(false)
  }, [episode])

  const handleSave = () => {
    const epNum = parseInt(episodeNumber, 10)
    const req: EpisodeMetadataEditRequest = {
      title,
      overview,
      air_date: airDate || undefined,
      episode_number: !isNaN(epNum) && epNum !== episode.episode_number ? epNum : undefined,
      metadata_locked: metadataLocked || undefined,
    }
    onSave(episode.id, req)
  }

  const canSave = title.trim().length > 0

  return (
    <Modal visible animationType="slide" transparent onRequestClose={onClose}>
      <KeyboardAvoidingView
        style={styles.container}
        behavior={Platform.OS === 'ios' ? 'padding' : undefined}
      >
        <TouchableOpacity style={styles.backdrop} onPress={onClose} activeOpacity={1} />

        <View style={[
          styles.sheet,
          isWideScreen && {
            maxWidth: layout.largeControls ? 600 : 500,
            alignSelf: 'center',
            width: '90%',
            borderRadius: 16,
          },
        ]}>
          {/* Header */}
          <View style={styles.header}>
            <Text style={[styles.headerTitle, { fontSize: scaledFont(18, layout.fontScale) }]}>Edit Episode {episode.episode_number}</Text>
            <TouchableOpacity style={styles.closeButton} onPress={onClose}>
              <X size={16} color="#888" />
            </TouchableOpacity>
          </View>

          <ScrollView style={styles.content} showsVerticalScrollIndicator={false}>
            {/* Title */}
            <View style={styles.fieldGroup}>
              <Text style={[styles.fieldLabel, { fontSize: scaledFont(12, layout.fontScale) }]}>Title</Text>
              <TextInput
                style={[styles.input, { fontSize: scaledFont(16, layout.fontScale), paddingVertical: layout.largeControls ? 16 : 12 }]}
                value={title}
                onChangeText={setTitle}
                placeholder="Enter title"
                placeholderTextColor="#666"
              />
            </View>

            {/* Episode Number */}
            <View style={styles.fieldGroup}>
              <Text style={[styles.fieldLabel, { fontSize: scaledFont(12, layout.fontScale) }]}>Episode Number</Text>
              <TextInput
                style={[styles.input, { fontSize: scaledFont(16, layout.fontScale), paddingVertical: layout.largeControls ? 16 : 12 }]}
                value={episodeNumber}
                onChangeText={(v) => {
                  // Allow only digits
                  const cleaned = v.replace(/[^0-9]/g, '')
                  setEpisodeNumber(cleaned)
                }}
                placeholder="1"
                placeholderTextColor="#666"
                keyboardType="numeric"
              />
            </View>

            {/* Air Date */}
            <View style={styles.fieldGroup}>
              <Text style={[styles.fieldLabel, { fontSize: scaledFont(12, layout.fontScale) }]}>Air Date</Text>
              <TextInput
                style={[styles.input, { fontSize: scaledFont(16, layout.fontScale), paddingVertical: layout.largeControls ? 16 : 12 }]}
                value={airDate}
                onChangeText={setAirDate}
                placeholder="YYYY-MM-DD"
                placeholderTextColor="#666"
              />
              <Text style={[styles.fieldHint, { fontSize: scaledFont(11, layout.fontScale) }]}>Format: YYYY-MM-DD (e.g. 2025-01-15)</Text>
            </View>

            {/* Overview */}
            <View style={styles.fieldGroup}>
              <Text style={[styles.fieldLabel, { fontSize: scaledFont(12, layout.fontScale) }]}>Overview</Text>
              <TextInput
                style={[styles.input, styles.textArea, { fontSize: scaledFont(16, layout.fontScale) }]}
                value={overview}
                onChangeText={setOverview}
                placeholder="Enter overview"
                placeholderTextColor="#666"
                multiline
                numberOfLines={4}
                textAlignVertical="top"
              />
            </View>

            {/* Metadata Lock */}
            <View style={styles.fieldGroup}>
              <View style={styles.switchRow}>
                <View>
                  <Text style={[styles.fieldLabel, { fontSize: scaledFont(12, layout.fontScale) }]}>Lock Metadata</Text>
                  <Text style={[styles.fieldHint, { fontSize: scaledFont(11, layout.fontScale) }]}>Prevent automatic metadata updates</Text>
                </View>
                <Switch
                  value={metadataLocked}
                  onValueChange={setMetadataLocked}
                  trackColor={{ false: '#333', true: '#e50914' }}
                  thumbColor={metadataLocked ? '#fff' : '#888'}
                />
              </View>
            </View>
          </ScrollView>

          {/* Actions */}
          <View style={styles.actions}>
            <TouchableOpacity style={[styles.cancelButton, { paddingVertical: layout.largeControls ? 18 : 14 }]} onPress={onClose}>
              <Text style={[styles.cancelButtonText, { fontSize: scaledFont(15, layout.fontScale) }]}>Cancel</Text>
            </TouchableOpacity>

            <TouchableOpacity
              style={[styles.saveButton, (!canSave || isSaving) && styles.saveButtonDisabled, { paddingVertical: layout.largeControls ? 18 : 14 }]}
              onPress={handleSave}
              disabled={!canSave || isSaving}
            >
              <View style={styles.saveButtonContent}>
                {!isSaving && <Save size={scaledFont(16, layout.fontScale)} color="#fff" />}
                <Text style={[styles.saveButtonText, { fontSize: scaledFont(15, layout.fontScale) }]}>
                  {isSaving ? 'Saving...' : 'Save'}
                </Text>
              </View>
            </TouchableOpacity>
          </View>
        </View>
      </KeyboardAvoidingView>
    </Modal>
  )
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
    justifyContent: 'flex-end',
  },
  backdrop: {
    ...StyleSheet.absoluteFillObject,
    backgroundColor: 'rgba(0, 0, 0, 0.6)',
  },
  sheet: {
    backgroundColor: '#141414',
    borderTopLeftRadius: 20,
    borderTopRightRadius: 20,
    maxHeight: '85%',
    minHeight: 400,
  },
  header: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
    paddingHorizontal: 20,
    paddingVertical: 16,
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
    padding: 20,
  },
  fieldGroup: {
    marginBottom: 20,
  },
  fieldLabel: {
    color: '#888',
    fontSize: 12,
    fontWeight: '600',
    textTransform: 'uppercase',
    marginBottom: 8,
  },
  fieldHint: {
    color: '#555',
    fontSize: 11,
    marginTop: 4,
  },
  input: {
    backgroundColor: '#1f1f1f',
    borderRadius: 8,
    paddingHorizontal: 16,
    paddingVertical: 12,
    color: '#fff',
    fontSize: 16,
    borderWidth: 1,
    borderColor: 'transparent',
  },
  textArea: {
    minHeight: 100,
    paddingTop: 12,
  },
  switchRow: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
  },
  actions: {
    flexDirection: 'row',
    padding: 20,
    paddingBottom: 34,
    borderTopWidth: 1,
    borderTopColor: '#2a2a2a',
    gap: 12,
  },
  cancelButton: {
    flex: 1,
    paddingVertical: 14,
    borderRadius: 8,
    alignItems: 'center',
    backgroundColor: '#2a2a2a',
  },
  cancelButtonText: {
    color: '#fff',
    fontSize: 15,
    fontWeight: '500',
  },
  saveButton: {
    flex: 2,
    paddingVertical: 14,
    borderRadius: 8,
    alignItems: 'center',
    backgroundColor: '#e50914',
  },
  saveButtonDisabled: {
    opacity: 0.5,
  },
  saveButtonContent: {
    flexDirection: 'row',
    alignItems: 'center',
    gap: 8,
  },
  saveButtonText: {
    color: '#fff',
    fontSize: 15,
    fontWeight: '600',
  },
})
