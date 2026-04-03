/**
 * SubtitleSearchModal - Search and download subtitles from online providers
 *
 * Full-screen modal with language selector chips and results list.
 * Uses useSubtitleSearch and useDownloadSubtitle from @velox/shared.
 */

import { useState } from 'react'
import {
  View,
  Text,
  Modal,
  TouchableOpacity,
  FlatList,
  ScrollView,
  ActivityIndicator,
  StyleSheet,
  SafeAreaView,
} from 'react-native'
import { X, Download, Check } from 'lucide-react-native'
import { useSubtitleSearch, useDownloadSubtitle } from '@velox/shared/hooks'
import { Toast } from './Toast'
import type { SubtitleSearchResult } from '@velox/shared/types'
import { useResponsiveLayout, scaledFont } from '../lib/responsive'

const LANGUAGES = [
  { code: 'en', name: 'English' },
  { code: 'vi', name: 'Vietnamese' },
  { code: 'fr', name: 'French' },
  { code: 'de', name: 'German' },
  { code: 'es', name: 'Spanish' },
  { code: 'pt', name: 'Portuguese' },
  { code: 'it', name: 'Italian' },
  { code: 'ja', name: 'Japanese' },
  { code: 'ko', name: 'Korean' },
  { code: 'zh', name: 'Chinese' },
  { code: 'nl', name: 'Dutch' },
  { code: 'pl', name: 'Polish' },
  { code: 'ru', name: 'Russian' },
  { code: 'ar', name: 'Arabic' },
  { code: 'tr', name: 'Turkish' },
  { code: 'sv', name: 'Swedish' },
  { code: 'th', name: 'Thai' },
  { code: 'id', name: 'Indonesian' },
]

// Map ISO 639-2 (3-letter) codes to ISO 639-1 (2-letter) for provider APIs
const ISO_639_2_TO_1: Record<string, string> = {
  eng: 'en',
  vie: 'vi',
  fra: 'fr',
  fre: 'fr',
  deu: 'de',
  ger: 'de',
  spa: 'es',
  por: 'pt',
  ita: 'it',
  jpn: 'ja',
  kor: 'ko',
  zho: 'zh',
  chi: 'zh',
  nld: 'nl',
  dut: 'nl',
  pol: 'pl',
  rus: 'ru',
  ara: 'ar',
  tur: 'tr',
  swe: 'sv',
  tha: 'th',
  ind: 'id',
}

function normalizeLangCode(code: string | null | undefined): string {
  if (!code) return 'en'
  const lower = code.toLowerCase()
  return ISO_639_2_TO_1[lower] ?? lower
}

function subtitleFormatPriority(format: string): number {
  switch (format.trim().toLowerCase()) {
    case 'srt':
    case 'subrip':
    case 'vtt':
    case 'webvtt':
    case 'ass':
    case 'ssa':
      return 0
    case 'pgs':
    case 'sup':
      return 1
    default:
      return 2
  }
}

function providerLabel(provider: string): string {
  switch (provider) {
    case 'opensubtitles':
      return 'OpenSub'
    case 'subdl':
      return 'Subdl'
    case 'podnapisi':
      return 'Podnapisi'
    case 'bsplayer':
      return 'BSPlayer'
    case 'subscene':
      return 'Subscene'
    default:
      return provider
  }
}

function providerColor(provider: string): { bg: string; text: string } {
  switch (provider) {
    case 'opensubtitles':
      return { bg: 'rgba(34, 197, 94, 0.15)', text: '#4ade80' }
    case 'subdl':
      return { bg: 'rgba(245, 158, 11, 0.15)', text: '#fbbf24' }
    case 'podnapisi':
      return { bg: 'rgba(59, 130, 246, 0.15)', text: '#60a5fa' }
    case 'bsplayer':
      return { bg: 'rgba(168, 85, 247, 0.15)', text: '#c084fc' }
    case 'subscene':
      return { bg: 'rgba(6, 182, 212, 0.15)', text: '#22d3ee' }
    default:
      return { bg: 'rgba(255, 255, 255, 0.1)', text: 'rgba(255, 255, 255, 0.6)' }
  }
}

function formatDownloads(count: number): string {
  if (count > 1000) return `${Math.floor(count / 1000)}k`
  return String(count)
}

interface SubtitleSearchModalProps {
  mediaId: number
  defaultLang?: string | null
  visible: boolean
  onClose: () => void
  onSubtitleDownloaded: () => void
}

export function SubtitleSearchModal({
  mediaId,
  defaultLang,
  visible,
  onClose,
  onSubtitleDownloaded,
}: SubtitleSearchModalProps) {
  const layout = useResponsiveLayout()
  const isWideScreen = layout.device !== 'phone'

  const [lang, setLang] = useState(normalizeLangCode(defaultLang))
  const [downloaded, setDownloaded] = useState<Set<string>>(new Set())

  const { data: results, isLoading } = useSubtitleSearch(mediaId, lang, visible)
  const downloadMutation = useDownloadSubtitle(mediaId)

  const downloadingKey = downloadMutation.isPending
    ? `${downloadMutation.variables?.provider}:${downloadMutation.variables?.external_id}`
    : null

  const sortedResults = [...(results ?? [])].sort((a, b) => {
    const fmt = subtitleFormatPriority(a.format) - subtitleFormatPriority(b.format)
    if (fmt !== 0) return fmt
    return (b.downloads || 0) - (a.downloads || 0)
  })

  const handleDownload = (result: SubtitleSearchResult) => {
    downloadMutation.mutate(
      {
        provider: result.provider,
        external_id: result.external_id,
        language: result.language || lang,
      },
      {
        onSuccess: () => {
          const key = `${result.provider}:${result.external_id}`
          setDownloaded((prev) => new Set(prev).add(key))
          Toast.success('Subtitle downloaded successfully')
          onSubtitleDownloaded()
        },
        onError: (error) => {
          const message = error instanceof Error ? error.message : 'Download failed'
          Toast.error(`${providerLabel(result.provider)}: ${message}`)
        },
      },
    )
  }

  const renderResultItem = ({ item }: { item: SubtitleSearchResult }) => {
    const key = `${item.provider}:${item.external_id}`
    const isDownloaded = downloaded.has(key)
    const isDownloading = downloadingKey === key
    const pColor = providerColor(item.provider)

    return (
      <View style={[styles.resultItem, isWideScreen && { paddingHorizontal: 24 }]}>
        <View style={styles.resultInfo}>
          <Text style={[styles.resultTitle, { fontSize: scaledFont(14, layout.fontScale) }]} numberOfLines={2}>
            {item.title || 'Untitled'}
          </Text>
          <View style={styles.resultBadges}>
            <View style={[styles.badge, { backgroundColor: pColor.bg }]}>
              <Text style={[styles.badgeText, { color: pColor.text }]}>
                {providerLabel(item.provider)}
              </Text>
            </View>
            {item.format ? (
              <Text style={styles.formatText}>{item.format.toUpperCase()}</Text>
            ) : null}
            {item.downloads > 0 ? (
              <Text style={styles.downloadCount}>{formatDownloads(item.downloads)} DL</Text>
            ) : null}
            {item.hearing_impaired ? (
              <View style={[styles.badge, { backgroundColor: 'rgba(234, 179, 8, 0.15)' }]}>
                <Text style={[styles.badgeText, { color: '#facc15' }]}>CC</Text>
              </View>
            ) : null}
            {item.forced ? (
              <View style={[styles.badge, { backgroundColor: 'rgba(168, 85, 247, 0.15)' }]}>
                <Text style={[styles.badgeText, { color: '#c084fc' }]}>Forced</Text>
              </View>
            ) : null}
            {item.ai_translated ? (
              <View style={[styles.badge, { backgroundColor: 'rgba(249, 115, 22, 0.15)' }]}>
                <Text style={[styles.badgeText, { color: '#fb923c' }]}>AI</Text>
              </View>
            ) : null}
          </View>
        </View>

        <TouchableOpacity
          style={[styles.downloadButton, layout.largeControls && { width: 52, height: 52 }]}
          onPress={() => handleDownload(item)}
          disabled={isDownloaded || isDownloading}
        >
          {isDownloaded ? (
            <Check size={scaledFont(18, layout.fontScale)} color="#4ade80" />
          ) : isDownloading ? (
            <ActivityIndicator size="small" color="#888" />
          ) : (
            <Download size={scaledFont(18, layout.fontScale)} color="rgba(255, 255, 255, 0.5)" />
          )}
        </TouchableOpacity>
      </View>
    )
  }

  return (
    <Modal visible={visible} animationType="slide" onRequestClose={onClose}>
      <SafeAreaView style={[styles.container, isWideScreen && { paddingHorizontal: layout.screenPadding }]}>
        {/* Header */}
        <View style={[styles.header, isWideScreen && { paddingHorizontal: 24 }]}>
          <Text style={[styles.headerTitle, { fontSize: scaledFont(18, layout.fontScale) }]}>Search Subtitles</Text>
          <TouchableOpacity style={styles.closeButton} onPress={onClose}>
            <X size={scaledFont(22, layout.fontScale)} color="#fff" />
          </TouchableOpacity>
        </View>

        {/* Language chips */}
        <ScrollView
          horizontal
          showsHorizontalScrollIndicator={false}
          contentContainerStyle={styles.languageChips}
          style={styles.languageScroll}
        >
          {LANGUAGES.map((l) => {
            const isSelected = l.code === lang
            return (
              <TouchableOpacity
                key={l.code}
                style={[styles.chip, isSelected && styles.chipSelected, layout.largeControls && { paddingHorizontal: 20, paddingVertical: 10 }]}
                onPress={() => setLang(l.code)}
              >
                <Text style={[styles.chipText, isSelected && styles.chipTextSelected, { fontSize: scaledFont(13, layout.fontScale) }]}>
                  {l.name}
                </Text>
              </TouchableOpacity>
            )
          })}
        </ScrollView>

        {/* Results */}
        {isLoading ? (
          <View style={styles.centerState}>
            <ActivityIndicator size="large" color="#e50914" />
            <Text style={styles.stateText}>Searching...</Text>
          </View>
        ) : sortedResults.length === 0 ? (
          <View style={styles.centerState}>
            <Text style={styles.stateText}>No subtitles found</Text>
            <Text style={styles.stateSubtext}>Try a different language</Text>
          </View>
        ) : (
          <FlatList
            data={sortedResults}
            keyExtractor={(item) => `${item.provider}:${item.external_id}`}
            renderItem={renderResultItem}
            contentContainerStyle={styles.resultsList}
          />
        )}

        {/* Footer */}
        <View style={styles.footer}>
          <Text style={[styles.footerText, { fontSize: scaledFont(11, layout.fontScale) }]}>
            {results?.length ?? 0} results from Subdl, Podnapisi & BSPlayer
          </Text>
        </View>
      </SafeAreaView>
    </Modal>
  )
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
    backgroundColor: '#141414',
  },
  header: {
    flexDirection: 'row',
    alignItems: 'center',
    justifyContent: 'space-between',
    paddingHorizontal: 16,
    paddingVertical: 14,
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
  languageScroll: {
    maxHeight: 52,
    borderBottomWidth: 1,
    borderBottomColor: 'rgba(255, 255, 255, 0.1)',
  },
  languageChips: {
    flexDirection: 'row',
    paddingHorizontal: 16,
    paddingVertical: 10,
    gap: 8,
  },
  chip: {
    paddingHorizontal: 14,
    paddingVertical: 6,
    borderRadius: 16,
    backgroundColor: 'rgba(255, 255, 255, 0.08)',
    borderWidth: 1,
    borderColor: 'rgba(255, 255, 255, 0.1)',
  },
  chipSelected: {
    backgroundColor: '#e50914',
    borderColor: '#e50914',
  },
  chipText: {
    fontSize: 13,
    color: 'rgba(255, 255, 255, 0.6)',
    fontWeight: '500',
  },
  chipTextSelected: {
    color: '#fff',
    fontWeight: '600',
  },
  centerState: {
    flex: 1,
    justifyContent: 'center',
    alignItems: 'center',
    gap: 8,
  },
  stateText: {
    fontSize: 15,
    color: 'rgba(255, 255, 255, 0.4)',
    marginTop: 8,
  },
  stateSubtext: {
    fontSize: 13,
    color: 'rgba(255, 255, 255, 0.25)',
  },
  resultsList: {
    paddingBottom: 16,
  },
  resultItem: {
    flexDirection: 'row',
    alignItems: 'center',
    paddingHorizontal: 16,
    paddingVertical: 12,
    borderBottomWidth: 1,
    borderBottomColor: 'rgba(255, 255, 255, 0.05)',
    gap: 12,
  },
  resultInfo: {
    flex: 1,
    gap: 6,
  },
  resultTitle: {
    fontSize: 14,
    color: 'rgba(255, 255, 255, 0.85)',
    lineHeight: 20,
  },
  resultBadges: {
    flexDirection: 'row',
    flexWrap: 'wrap',
    gap: 6,
    alignItems: 'center',
  },
  badge: {
    paddingHorizontal: 6,
    paddingVertical: 2,
    borderRadius: 4,
  },
  badgeText: {
    fontSize: 10,
    fontWeight: '600',
  },
  formatText: {
    fontSize: 10,
    color: 'rgba(255, 255, 255, 0.3)',
    fontWeight: '600',
  },
  downloadCount: {
    fontSize: 10,
    color: 'rgba(255, 255, 255, 0.3)',
  },
  downloadButton: {
    width: 40,
    height: 40,
    borderRadius: 8,
    justifyContent: 'center',
    alignItems: 'center',
  },
  footer: {
    paddingHorizontal: 16,
    paddingVertical: 10,
    borderTopWidth: 1,
    borderTopColor: 'rgba(255, 255, 255, 0.1)',
  },
  footerText: {
    fontSize: 11,
    color: 'rgba(255, 255, 255, 0.25)',
    textAlign: 'center',
  },
})
