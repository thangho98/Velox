/**
 * TranslateSection - DeepL, Subdl API keys, and auto-subtitle language preferences
 */

import { useState } from 'react'
import { View, Text, StyleSheet, TextInput, TouchableOpacity } from 'react-native'
import {
  useDeepLSettings,
  useUpdateDeepLSettings,
  useSubdlSettings,
  useUpdateSubdlSettings,
  useAutoSubSettings,
  useUpdateAutoSubSettings,
} from '@velox/shared/hooks/settings'
import { CollapsibleSection } from './CollapsibleSection'
import { Toast } from '../Toast'

const COMMON_LANGUAGES = [
  { code: 'en', label: 'English' },
  { code: 'vi', label: 'Vietnamese' },
  { code: 'fr', label: 'French' },
  { code: 'de', label: 'German' },
  { code: 'es', label: 'Spanish' },
  { code: 'pt', label: 'Portuguese' },
  { code: 'ja', label: 'Japanese' },
  { code: 'ko', label: 'Korean' },
  { code: 'zh', label: 'Chinese' },
  { code: 'th', label: 'Thai' },
]

export function TranslateSection() {
  return (
    <CollapsibleSection title="Subtitles & Translation" subtitle="API keys and auto-download">
      <DeepLCard />
      <View style={styles.divider} />
      <SubdlCard />
      <View style={styles.divider} />
      <AutoSubCard />
    </CollapsibleSection>
  )
}

// ── DeepL API Key ────────────────────────────────────────────────────────────

function DeepLCard() {
  const { data: settings, isLoading } = useDeepLSettings()
  const { mutate: updateSettings, isPending } = useUpdateDeepLSettings()
  const [editedKey, setEditedKey] = useState<string | null>(null)
  const apiKey = editedKey ?? settings?.api_key ?? ''

  const handleSave = () => {
    updateSettings(
      { api_key: apiKey },
      {
        onSuccess: () => {
          setEditedKey(null)
          Toast.success('DeepL settings saved')
        },
        onError: () => Toast.error('Failed to save'),
      },
    )
  }

  if (isLoading) return <Text style={styles.loadingText}>Loading...</Text>

  return (
    <View>
      <View style={styles.cardHeader}>
        <Text style={styles.cardTitle}>DeepL Translation</Text>
        <View style={[styles.badge, settings?.api_key ? styles.badgeBlue : styles.badgeGray]}>
          <Text style={[styles.badgeText, settings?.api_key ? styles.badgeTextBlue : styles.badgeTextGray]}>
            {settings?.api_key ? 'DeepL' : 'Google Translate'}
          </Text>
        </View>
      </View>
      <Text style={styles.cardDesc}>
        Enter a DeepL API key for higher quality translation. Leave empty to use Google Translate.
      </Text>
      <TextInput
        style={styles.input}
        value={apiKey}
        onChangeText={setEditedKey}
        placeholder="DeepL API key (optional)"
        placeholderTextColor="#666"
        autoCapitalize="none"
        autoCorrect={false}
      />
      <TouchableOpacity
        style={[styles.saveBtn, isPending && { opacity: 0.5 }]}
        onPress={handleSave}
        disabled={isPending || editedKey === null}
      >
        <Text style={styles.saveBtnText}>{isPending ? 'Saving...' : 'Save'}</Text>
      </TouchableOpacity>
    </View>
  )
}

// ── Subdl API Key ────────────────────────────────────────────────────────────

function SubdlCard() {
  const { data: settings, isLoading } = useSubdlSettings()
  const { mutate: updateSettings, isPending } = useUpdateSubdlSettings()
  const [editedKey, setEditedKey] = useState<string | null>(null)
  const apiKey = editedKey ?? settings?.api_key ?? ''

  const handleSave = () => {
    updateSettings(
      { api_key: apiKey },
      {
        onSuccess: () => {
          setEditedKey(null)
          Toast.success('Subdl settings saved')
        },
        onError: () => Toast.error('Failed to save'),
      },
    )
  }

  if (isLoading) return <Text style={styles.loadingText}>Loading...</Text>

  const statusLabel = settings?.api_key
    ? 'Custom Key'
    : settings?.has_builtin
      ? 'Built-in Key'
      : 'Not Configured'

  return (
    <View>
      <View style={styles.cardHeader}>
        <Text style={styles.cardTitle}>Subdl Provider</Text>
        <View
          style={[
            styles.badge,
            settings?.api_key
              ? styles.badgeBlue
              : settings?.has_builtin
                ? styles.badgeGreen
                : styles.badgeGray,
          ]}
        >
          <Text
            style={[
              styles.badgeText,
              settings?.api_key
                ? styles.badgeTextBlue
                : settings?.has_builtin
                  ? styles.badgeTextGreen
                  : styles.badgeTextGray,
            ]}
          >
            {statusLabel}
          </Text>
        </View>
      </View>
      <Text style={styles.cardDesc}>
        Subdl subtitle provider API key.
        {settings?.has_builtin ? ' A built-in key is available.' : ' No built-in key.'}
      </Text>
      <TextInput
        style={styles.input}
        value={apiKey}
        onChangeText={setEditedKey}
        placeholder={settings?.has_builtin ? 'Optional (built-in available)' : 'Subdl API key'}
        placeholderTextColor="#666"
        autoCapitalize="none"
        autoCorrect={false}
      />
      <TouchableOpacity
        style={[styles.saveBtn, isPending && { opacity: 0.5 }]}
        onPress={handleSave}
        disabled={isPending || editedKey === null}
      >
        <Text style={styles.saveBtnText}>{isPending ? 'Saving...' : 'Save'}</Text>
      </TouchableOpacity>
    </View>
  )
}

// ── Auto-Subtitle Languages ──────────────────────────────────────────────────

function AutoSubCard() {
  const { data: settings, isLoading } = useAutoSubSettings()
  const { mutate: updateSettings, isPending } = useUpdateAutoSubSettings()
  const [edited, setEdited] = useState<string[] | null>(null)

  const selected =
    edited ?? (settings?.languages ? settings.languages.split(',').filter(Boolean) : [])

  const toggleLang = (code: string) => {
    const current = [...selected]
    const idx = current.indexOf(code)
    if (idx >= 0) {
      current.splice(idx, 1)
    } else {
      current.push(code)
    }
    setEdited(current)
  }

  const handleSave = () => {
    updateSettings(
      { languages: selected.join(',') },
      {
        onSuccess: () => {
          setEdited(null)
          Toast.success('Auto-subtitle languages saved')
        },
        onError: () => Toast.error('Failed to save'),
      },
    )
  }

  if (isLoading) return <Text style={styles.loadingText}>Loading...</Text>

  return (
    <View>
      <View style={styles.cardHeader}>
        <Text style={styles.cardTitle}>Auto-Download Subtitles</Text>
        {selected.length > 0 && (
          <View style={styles.badgeBlue}>
            <Text style={styles.badgeTextBlue}>
              {selected.length} lang{selected.length !== 1 ? 's' : ''}
            </Text>
          </View>
        )}
      </View>
      <Text style={styles.cardDesc}>
        Automatically download subtitles in these languages during library scan.
      </Text>

      <View style={styles.langGrid}>
        {COMMON_LANGUAGES.map((lang) => (
          <TouchableOpacity
            key={lang.code}
            style={[
              styles.langChip,
              selected.includes(lang.code) && styles.langChipSelected,
            ]}
            onPress={() => toggleLang(lang.code)}
          >
            <Text
              style={[
                styles.langChipText,
                selected.includes(lang.code) && styles.langChipTextSelected,
              ]}
            >
              {lang.label}
            </Text>
          </TouchableOpacity>
        ))}
      </View>

      <TouchableOpacity
        style={[styles.saveBtn, (isPending || edited === null) && { opacity: 0.5 }]}
        onPress={handleSave}
        disabled={isPending || edited === null}
      >
        <Text style={styles.saveBtnText}>{isPending ? 'Saving...' : 'Save Languages'}</Text>
      </TouchableOpacity>
    </View>
  )
}

const styles = StyleSheet.create({
  loadingText: {
    color: '#888',
    fontSize: 14,
    paddingVertical: 8,
  },
  divider: {
    height: 1,
    backgroundColor: '#2a2a2a',
    marginVertical: 16,
  },
  cardHeader: {
    flexDirection: 'row',
    alignItems: 'center',
    gap: 8,
    marginBottom: 4,
  },
  cardTitle: {
    color: '#fff',
    fontSize: 14,
    fontWeight: '600',
  },
  cardDesc: {
    color: '#888',
    fontSize: 12,
    marginBottom: 10,
    lineHeight: 17,
  },
  badge: {
    paddingHorizontal: 8,
    paddingVertical: 2,
    borderRadius: 4,
  },
  badgeBlue: {
    backgroundColor: 'rgba(59, 130, 246, 0.2)',
    paddingHorizontal: 8,
    paddingVertical: 2,
    borderRadius: 4,
  },
  badgeGreen: {
    backgroundColor: 'rgba(34, 197, 94, 0.2)',
    paddingHorizontal: 8,
    paddingVertical: 2,
    borderRadius: 4,
  },
  badgeGray: {
    backgroundColor: 'rgba(136, 136, 136, 0.2)',
  },
  badgeText: {
    fontSize: 10,
    fontWeight: '600',
  },
  badgeTextBlue: {
    color: '#3b82f6',
    fontSize: 10,
    fontWeight: '600',
  },
  badgeTextGreen: {
    color: '#22c55e',
    fontSize: 10,
    fontWeight: '600',
  },
  badgeTextGray: {
    color: '#888',
    fontSize: 10,
    fontWeight: '600',
  },
  input: {
    backgroundColor: '#2a2a2a',
    borderRadius: 8,
    paddingHorizontal: 14,
    paddingVertical: 12,
    color: '#fff',
    fontSize: 14,
    marginBottom: 10,
  },
  saveBtn: {
    backgroundColor: '#e50914',
    borderRadius: 8,
    paddingVertical: 10,
    alignItems: 'center',
  },
  saveBtnText: {
    color: '#fff',
    fontSize: 13,
    fontWeight: '600',
  },
  langGrid: {
    flexDirection: 'row',
    flexWrap: 'wrap',
    gap: 6,
    marginBottom: 12,
  },
  langChip: {
    backgroundColor: '#2a2a2a',
    borderRadius: 16,
    paddingHorizontal: 12,
    paddingVertical: 6,
  },
  langChipSelected: {
    backgroundColor: '#e50914',
  },
  langChipText: {
    color: '#ccc',
    fontSize: 12,
    fontWeight: '500',
  },
  langChipTextSelected: {
    color: '#fff',
  },
})
