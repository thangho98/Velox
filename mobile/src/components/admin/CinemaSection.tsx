/**
 * CinemaSection - Cinema mode settings (enable toggle + max trailers)
 *
 * Uses createSettingsHooks factory from @velox/shared.
 */

import { useState } from 'react'
import { View, Text, Switch, StyleSheet, TouchableOpacity, Alert } from 'react-native'
import { createSettingsHooks } from '@velox/shared/hooks/settings'
import { CollapsibleSection } from './CollapsibleSection'
import { Toast } from '../Toast'

interface CinemaSettings {
  enabled: boolean
  max_trailers: number
  has_intro: boolean
}

const [useCinemaSettings, useUpdateCinemaSettings] = createSettingsHooks<CinemaSettings>(
  'cinema',
)

export function CinemaSection() {
  const { data: settings, isLoading } = useCinemaSettings()
  const { mutate: updateSettings, isPending } = useUpdateCinemaSettings()

  const handleToggle = (enabled: boolean) => {
    updateSettings(
      { enabled },
      {
        onSuccess: () => Toast.success(enabled ? 'Cinema mode enabled' : 'Cinema mode disabled'),
        onError: () => Toast.error('Failed to update cinema settings'),
      },
    )
  }

  const handleMaxTrailers = () => {
    const options = [0, 1, 2, 3, 4, 5]
    Alert.alert(
      'Max Trailers',
      'How many trailers to show before playback?',
      [
        ...options.map((n) => ({
          text: n === 0 ? 'Disabled (0)' : `${n} trailer${n > 1 ? 's' : ''}`,
          onPress: () => {
            updateSettings(
              { max_trailers: n },
              {
                onSuccess: () => Toast.success(`Max trailers set to ${n}`),
                onError: () => Toast.error('Failed to update'),
              },
            )
          },
        })),
        { text: 'Cancel', style: 'cancel' as const },
      ],
    )
  }

  if (isLoading) {
    return (
      <CollapsibleSection title="Cinema Mode" subtitle="Loading...">
        <Text style={styles.loadingText}>Loading settings...</Text>
      </CollapsibleSection>
    )
  }

  return (
    <CollapsibleSection
      title="Cinema Mode"
      subtitle={settings?.enabled ? 'Enabled' : 'Disabled'}
    >
      {/* Enable/Disable */}
      <View style={styles.row}>
        <View style={styles.rowInfo}>
          <Text style={styles.rowTitle}>Enable Cinema Mode</Text>
          <Text style={styles.rowDesc}>
            Show YouTube trailers on media detail pages
          </Text>
        </View>
        <Switch
          value={settings?.enabled ?? false}
          onValueChange={handleToggle}
          trackColor={{ false: '#3a3a3a', true: '#e50914' }}
          thumbColor="#fff"
          disabled={isPending}
        />
      </View>

      {/* Max trailers selector */}
      <View style={styles.divider} />
      <View style={styles.row}>
        <View style={styles.rowInfo}>
          <Text style={styles.rowTitle}>Max Trailers</Text>
          <Text style={styles.rowDesc}>
            Number of trailers to show (0 = disabled)
          </Text>
        </View>
        <TouchableOpacity style={styles.valueButton} onPress={handleMaxTrailers}>
          <Text style={styles.valueText}>{settings?.max_trailers ?? 2}</Text>
          <Text style={styles.valueArrow}>▼</Text>
        </TouchableOpacity>
      </View>
    </CollapsibleSection>
  )
}

const styles = StyleSheet.create({
  row: {
    flexDirection: 'row',
    alignItems: 'center',
    justifyContent: 'space-between',
    paddingVertical: 12,
  },
  rowInfo: {
    flex: 1,
    marginRight: 16,
  },
  rowTitle: {
    color: '#fff',
    fontSize: 14,
    fontWeight: '500',
  },
  rowDesc: {
    color: '#888',
    fontSize: 12,
    marginTop: 2,
  },
  divider: {
    height: 1,
    backgroundColor: '#2a2a2a',
  },
  valueButton: {
    flexDirection: 'row',
    alignItems: 'center',
    backgroundColor: '#2a2a2a',
    paddingHorizontal: 14,
    paddingVertical: 8,
    borderRadius: 8,
    gap: 8,
  },
  valueText: {
    color: '#fff',
    fontSize: 14,
    fontWeight: '600',
  },
  valueArrow: {
    color: '#888',
    fontSize: 10,
  },
  loadingText: {
    color: '#888',
    fontSize: 14,
  },
})
