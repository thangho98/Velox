/**
 * SubtitleTranslate - Modal for translating existing subtitles via DeepL
 *
 * Select a source subtitle track and target language, then translate.
 * Uses useTranslateSubtitle from @velox/shared.
 */

import { useState } from 'react'
import {
  View,
  Text,
  Modal,
  TouchableOpacity,
  ScrollView,
  ActivityIndicator,
  StyleSheet,
  SafeAreaView,
} from 'react-native'
import { X, Check } from 'lucide-react-native'
import { useTranslateSubtitle } from '@velox/shared/hooks'
import { Toast } from './Toast'
import type { PlaybackSubtitleTrack } from '@velox/shared/types'
import { useResponsiveLayout, scaledFont } from '../lib/responsive'

const TRANSLATE_LANGS = [
  { code: 'vi', label: 'Vietnamese' },
  { code: 'en', label: 'English' },
  { code: 'fr', label: 'French' },
  { code: 'de', label: 'German' },
  { code: 'es', label: 'Spanish' },
  { code: 'ja', label: 'Japanese' },
  { code: 'ko', label: 'Korean' },
  { code: 'zh', label: 'Chinese' },
  { code: 'pt', label: 'Portuguese' },
  { code: 'ru', label: 'Russian' },
  { code: 'th', label: 'Thai' },
]

interface SubtitleTranslateProps {
  subtitles: PlaybackSubtitleTrack[]
  visible: boolean
  onClose: () => void
  onTranslated: () => void
}

export function SubtitleTranslate({
  subtitles,
  visible,
  onClose,
  onTranslated,
}: SubtitleTranslateProps) {
  const layout = useResponsiveLayout()
  const isWideScreen = layout.device !== 'phone'

  const textSubs = subtitles.filter((s) => !s.is_image)
  const [sourceId, setSourceId] = useState<number | null>(null)
  const [targetLang, setTargetLang] = useState<string | null>(null)
  const translateMutation = useTranslateSubtitle()

  const selectedSourceId = sourceId ?? textSubs[0]?.id ?? null

  const handleTranslate = () => {
    if (!selectedSourceId || !targetLang) return
    translateMutation.mutate(
      { subtitleId: selectedSourceId, targetLanguage: targetLang },
      {
        onSuccess: () => {
          Toast.success('Subtitle translated successfully')
          onTranslated()
          onClose()
          setTargetLang(null)
        },
        onError: (error) => {
          const message = error instanceof Error ? error.message : 'Translation failed'
          Toast.error(message)
        },
      },
    )
  }

  return (
    <Modal visible={visible} animationType={isWideScreen ? 'fade' : 'slide'} transparent onRequestClose={onClose}>
      <View style={[styles.overlay, isWideScreen && { justifyContent: 'center', alignItems: 'center' }]}>
        <View style={[
          styles.sheet,
          isWideScreen && {
            maxWidth: layout.largeControls ? 600 : 500,
            width: '90%',
            borderRadius: 16,
            borderTopLeftRadius: 16,
            borderTopRightRadius: 16,
          },
        ]}>
          <SafeAreaView>
            {/* Header */}
            <View style={styles.header}>
              <Text style={[styles.headerTitle, { fontSize: scaledFont(18, layout.fontScale) }]}>Translate Subtitle</Text>
              <TouchableOpacity style={styles.closeButton} onPress={onClose}>
                <X size={scaledFont(22, layout.fontScale)} color="#fff" />
              </TouchableOpacity>
            </View>

            <ScrollView contentContainerStyle={styles.content}>
              {textSubs.length === 0 ? (
                <View style={styles.emptyState}>
                  <Text style={styles.emptyText}>No text subtitles available to translate</Text>
                </View>
              ) : (
                <>
                  {/* Source subtitle selector */}
                  <Text style={[styles.sectionLabel, { fontSize: scaledFont(12, layout.fontScale) }]}>Source Subtitle</Text>
                  <View style={styles.optionsList}>
                    {textSubs.map((sub) => {
                      const isSelected = sub.id === selectedSourceId
                      return (
                        <TouchableOpacity
                          key={sub.id}
                          style={[styles.optionItem, isSelected && styles.optionItemSelected]}
                          onPress={() => setSourceId(sub.id)}
                        >
                          <View style={styles.optionContent}>
                            <Text
                              style={[
                                styles.optionText,
                                isSelected && styles.optionTextSelected,
                              ]}
                            >
                              {sub.label || sub.language}
                            </Text>
                            <Text style={styles.optionSubtext}>
                              {sub.format.toUpperCase()}
                            </Text>
                          </View>
                          {isSelected ? <Check size={16} color="#e50914" /> : null}
                        </TouchableOpacity>
                      )
                    })}
                  </View>

                  {/* Target language selector */}
                  <Text style={[styles.sectionLabel, { marginTop: 20, fontSize: scaledFont(12, layout.fontScale) }]}>Translate To</Text>
                  <View style={styles.optionsList}>
                    {TRANSLATE_LANGS.map((l) => {
                      const isSelected = l.code === targetLang
                      return (
                        <TouchableOpacity
                          key={l.code}
                          style={[styles.optionItem, isSelected && styles.optionItemSelected]}
                          onPress={() => setTargetLang(l.code)}
                        >
                          <Text
                            style={[
                              styles.optionText,
                              isSelected && styles.optionTextSelected,
                            ]}
                          >
                            {l.label}
                          </Text>
                          {isSelected ? <Check size={16} color="#e50914" /> : null}
                        </TouchableOpacity>
                      )
                    })}
                  </View>

                  {/* Error message */}
                  {translateMutation.isError ? (
                    <Text style={styles.errorText}>
                      {(translateMutation.error as Error)?.message || 'Translation failed'}
                    </Text>
                  ) : null}

                  {/* Action buttons */}
                  <View style={styles.actions}>
                    <TouchableOpacity
                      style={[
                        styles.translateButton,
                        (!targetLang || translateMutation.isPending) &&
                          styles.translateButtonDisabled,
                        { paddingVertical: layout.largeControls ? 18 : 14 },
                      ]}
                      onPress={handleTranslate}
                      disabled={!targetLang || translateMutation.isPending}
                    >
                      {translateMutation.isPending ? (
                        <ActivityIndicator size="small" color="#fff" />
                      ) : null}
                      <Text style={[styles.translateButtonText, { fontSize: scaledFont(16, layout.fontScale) }]}>
                        {translateMutation.isPending ? 'Translating...' : 'Translate'}
                      </Text>
                    </TouchableOpacity>

                    <TouchableOpacity style={[styles.cancelButton, { paddingVertical: layout.largeControls ? 16 : 12 }]} onPress={onClose}>
                      <Text style={[styles.cancelButtonText, { fontSize: scaledFont(15, layout.fontScale) }]}>Cancel</Text>
                    </TouchableOpacity>
                  </View>
                </>
              )}
            </ScrollView>
          </SafeAreaView>
        </View>
      </View>
    </Modal>
  )
}

const styles = StyleSheet.create({
  overlay: {
    flex: 1,
    backgroundColor: 'rgba(0, 0, 0, 0.6)',
    justifyContent: 'flex-end',
  },
  sheet: {
    backgroundColor: '#1a1a1a',
    borderTopLeftRadius: 20,
    borderTopRightRadius: 20,
    maxHeight: '85%',
  },
  header: {
    flexDirection: 'row',
    alignItems: 'center',
    justifyContent: 'space-between',
    paddingHorizontal: 20,
    paddingVertical: 16,
    borderBottomWidth: 1,
    borderBottomColor: 'rgba(255, 255, 255, 0.1)',
  },
  headerTitle: {
    fontSize: 18,
    fontWeight: '700',
    color: '#fff',
  },
  closeButton: {
    padding: 4,
  },
  content: {
    padding: 20,
    paddingBottom: 40,
  },
  emptyState: {
    paddingVertical: 40,
    alignItems: 'center',
  },
  emptyText: {
    fontSize: 14,
    color: 'rgba(255, 255, 255, 0.4)',
  },
  sectionLabel: {
    fontSize: 12,
    fontWeight: '600',
    color: 'rgba(255, 255, 255, 0.4)',
    textTransform: 'uppercase',
    letterSpacing: 1,
    marginBottom: 10,
  },
  optionsList: {
    borderRadius: 12,
    overflow: 'hidden',
    backgroundColor: 'rgba(255, 255, 255, 0.05)',
  },
  optionItem: {
    flexDirection: 'row',
    alignItems: 'center',
    justifyContent: 'space-between',
    paddingHorizontal: 16,
    paddingVertical: 14,
    borderBottomWidth: 1,
    borderBottomColor: 'rgba(255, 255, 255, 0.05)',
  },
  optionItemSelected: {
    backgroundColor: 'rgba(229, 9, 20, 0.1)',
  },
  optionContent: {
    flexDirection: 'row',
    alignItems: 'center',
    gap: 8,
  },
  optionText: {
    fontSize: 15,
    color: 'rgba(255, 255, 255, 0.7)',
  },
  optionTextSelected: {
    color: '#fff',
    fontWeight: '600',
  },
  optionSubtext: {
    fontSize: 11,
    color: 'rgba(255, 255, 255, 0.3)',
    fontWeight: '500',
  },
  errorText: {
    fontSize: 13,
    color: '#ef4444',
    marginTop: 12,
    textAlign: 'center',
  },
  actions: {
    marginTop: 24,
    gap: 10,
  },
  translateButton: {
    flexDirection: 'row',
    alignItems: 'center',
    justifyContent: 'center',
    backgroundColor: '#2563eb',
    paddingVertical: 14,
    borderRadius: 10,
    gap: 8,
  },
  translateButtonDisabled: {
    opacity: 0.5,
  },
  translateButtonText: {
    fontSize: 16,
    fontWeight: '600',
    color: '#fff',
  },
  cancelButton: {
    alignItems: 'center',
    paddingVertical: 12,
    borderRadius: 10,
    backgroundColor: 'rgba(255, 255, 255, 0.08)',
  },
  cancelButtonText: {
    fontSize: 15,
    color: 'rgba(255, 255, 255, 0.5)',
  },
})
