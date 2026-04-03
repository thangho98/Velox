/**
 * MetadataEditor - Bottom sheet for editing media/series metadata
 *
 * Admin-only feature. Allows editing title, overview, rating, genres.
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
} from 'react-native'
import { X, Save } from 'lucide-react-native'
import type { Media, Series, MetadataEditRequest, SeriesMetadataEditRequest } from '@velox/shared/types'
import { Toast } from './Toast'
import { useResponsiveLayout, scaledFont } from '../lib/responsive'

// ── Types ─────────────────────────────────────────────────────────────────────

interface MediaEditorProps {
  type: 'media'
  item: Media
  genres?: string[]
  onSave: (req: MetadataEditRequest) => void
  onRefresh: () => void
  isSaving: boolean
  isRefreshing: boolean
  onClose: () => void
}

interface SeriesEditorProps {
  type: 'series'
  item: Series
  genres?: string[]
  onSave: (req: SeriesMetadataEditRequest) => void
  onRefresh: () => void
  isSaving: boolean
  isRefreshing: boolean
  onClose: () => void
}

type MetadataEditorProps = MediaEditorProps | SeriesEditorProps

// ── Metadata Editor Component ─────────────────────────────────────────────────

export function MetadataEditor(props: MetadataEditorProps) {
  const layout = useResponsiveLayout()
  const isWideScreen = layout.device !== 'phone'

  const isMedia = props.type === 'media'
  const item = props.item
  const genresList = props.genres || []

  const [title, setTitle] = useState(item.title)
  const [overview, setOverview] = useState(item.overview)
  const [rating, setRating] = useState(
    isMedia && 'rating' in item ? (item.rating || 0) : 0,
  )
  const [genres, setGenres] = useState(genresList.join(', '))

  useEffect(() => {
    setTitle(item.title)
    setOverview(item.overview)
    setRating(isMedia && 'rating' in item ? (item.rating || 0) : 0)
    setGenres(genresList.join(', '))
  }, [item, genresList])

  const handleSave = () => {
    const genreArray = genres
      .split(',')
      .map((g: string) => g.trim())
      .filter(Boolean)

    if (isMedia) {
      const req: MetadataEditRequest = {
        title,
        overview,
        rating,
        genres: genreArray,
      }
      props.onSave(req)
    } else {
      const req: SeriesMetadataEditRequest = {
        title,
        overview,
        genres: genreArray,
      }
      props.onSave(req)
    }
  }

  const handleRefresh = () => {
    props.onRefresh()
    Toast.info('Refreshing metadata from TMDB...')
  }

  return (
    <Modal
      visible
      animationType="slide"
      transparent
      onRequestClose={props.onClose}
    >
      <KeyboardAvoidingView
        style={styles.container}
        behavior={Platform.OS === 'ios' ? 'padding' : undefined}
      >
        <TouchableOpacity style={styles.backdrop} onPress={props.onClose} activeOpacity={1} />

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
            <Text style={[styles.headerTitle, { fontSize: scaledFont(18, layout.fontScale) }]}>Edit Metadata</Text>
            <TouchableOpacity style={styles.closeButton} onPress={props.onClose}>
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

            {/* Rating */}
            <View style={styles.fieldGroup}>
              <Text style={[styles.fieldLabel, { fontSize: scaledFont(12, layout.fontScale) }]}>Rating (0-10)</Text>
              <TextInput
                style={[styles.input, { fontSize: scaledFont(16, layout.fontScale), paddingVertical: layout.largeControls ? 16 : 12 }]}
                value={rating.toString()}
                onChangeText={(v: string) => {
                  const num = parseFloat(v)
                  if (!isNaN(num) && num >= 0 && num <= 10) {
                    setRating(num)
                  } else if (v === '') {
                    setRating(0)
                  }
                }}
                placeholder="0"
                placeholderTextColor="#666"
                keyboardType="numeric"
              />
            </View>

            {/* Genres */}
            <View style={styles.fieldGroup}>
              <Text style={[styles.fieldLabel, { fontSize: scaledFont(12, layout.fontScale) }]}>Genres (comma separated)</Text>
              <TextInput
                style={[styles.input, { fontSize: scaledFont(16, layout.fontScale), paddingVertical: layout.largeControls ? 16 : 12 }]}
                value={genres}
                onChangeText={setGenres}
                placeholder="Action, Drama, Comedy"
                placeholderTextColor="#666"
              />
            </View>
          </ScrollView>

          {/* Actions */}
          <View style={styles.actions}>
            <TouchableOpacity
              style={[styles.actionButton, styles.refreshButton, { paddingVertical: layout.largeControls ? 18 : 14 }]}
              onPress={handleRefresh}
              disabled={props.isRefreshing}
            >
              <Text style={[styles.refreshButtonText, { fontSize: scaledFont(15, layout.fontScale) }]}>
                {props.isRefreshing ? 'Refreshing...' : '🔄 Refresh from TMDB'}
              </Text>
            </TouchableOpacity>

            <TouchableOpacity
              style={[styles.actionButton, styles.saveButton, { paddingVertical: layout.largeControls ? 18 : 14 }]}
              onPress={handleSave}
              disabled={props.isSaving}
            >
              <View style={{ flexDirection: 'row', alignItems: 'center', gap: 8 }}>
                {!props.isSaving && <Save size={scaledFont(16, layout.fontScale)} color="#fff" />}
                <Text style={[styles.saveButtonText, { fontSize: scaledFont(15, layout.fontScale) }]}>
                  {props.isSaving ? 'Saving...' : 'Save Changes'}
                </Text>
              </View>
            </TouchableOpacity>
          </View>
        </View>
      </KeyboardAvoidingView>
    </Modal>
  )
}

// ── Styles ─────────────────────────────────────────────────────────────────────

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
  actions: {
    padding: 20,
    paddingBottom: 34,
    borderTopWidth: 1,
    borderTopColor: '#2a2a2a',
  },
  actionButton: {
    paddingVertical: 14,
    borderRadius: 8,
    alignItems: 'center',
    marginBottom: 10,
  },
  refreshButton: {
    backgroundColor: '#2a2a2a',
  },
  refreshButtonText: {
    color: '#fff',
    fontSize: 15,
    fontWeight: '500',
  },
  saveButton: {
    backgroundColor: '#e50914',
  },
  saveButtonText: {
    color: '#fff',
    fontSize: 15,
    fontWeight: '600',
  },
})