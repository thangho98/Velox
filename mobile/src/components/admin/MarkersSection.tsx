/**
 * MarkersSection - Intro/credits marker detection stats and backfill
 */

import { useState } from 'react'
import {
  View,
  Text,
  StyleSheet,
  TouchableOpacity,
  ActivityIndicator,
  Alert,
} from 'react-native'
import { useMarkerStats, useBackfillMarkers } from '@velox/shared/hooks/admin'
import { useLibraries } from '@velox/shared/hooks/media'
import { CollapsibleSection } from './CollapsibleSection'
import { Toast } from '../Toast'

export function MarkersSection() {
  const { data: stats, isLoading } = useMarkerStats()
  const { data: libraries } = useLibraries()
  const backfill = useBackfillMarkers()
  const [selectedLibrary, setSelectedLibrary] = useState(0)

  // Auto-select first library
  if (selectedLibrary === 0 && libraries && libraries.length > 0) {
    setSelectedLibrary(libraries[0].id)
  }

  const introCoverage =
    stats && stats.total_files > 0
      ? Math.round((stats.files_with_intro / stats.total_files) * 100)
      : 0
  const creditsCoverage =
    stats && stats.total_files > 0
      ? Math.round((stats.files_with_credits / stats.total_files) * 100)
      : 0

  const handleSelectLibrary = () => {
    if (!libraries || libraries.length === 0) return
    Alert.alert(
      'Select Library',
      undefined,
      [
        ...libraries.map((lib) => ({
          text: lib.name,
          onPress: () => setSelectedLibrary(lib.id),
        })),
        { text: 'Cancel', style: 'cancel' as const },
      ],
    )
  }

  const handleBackfill = () => {
    if (selectedLibrary === 0) {
      Toast.error('Select a library first')
      return
    }
    backfill.mutate(
      { library_id: selectedLibrary },
      {
        onSuccess: (result) => {
          Toast.success(`Processed: ${result.processed}, Skipped: ${result.skipped}`)
        },
        onError: () => Toast.error('Marker detection failed'),
      },
    )
  }

  const subtitle = stats
    ? `${stats.total_markers} markers`
    : 'Loading...'

  return (
    <CollapsibleSection title="Marker Detection" subtitle={subtitle}>
      {isLoading ? (
        <Text style={styles.loadingText}>Loading stats...</Text>
      ) : stats && stats.total_markers > 0 ? (
        <>
          {/* Stats Grid */}
          <View style={styles.statsGrid}>
            <View style={styles.statCard}>
              <Text style={styles.statValue}>{stats.total_markers}</Text>
              <Text style={styles.statLabel}>Total</Text>
            </View>
            <View style={styles.statCard}>
              <Text style={[styles.statValue, { color: '#3b82f6' }]}>
                {stats.intro_markers}
              </Text>
              <Text style={styles.statLabel}>Intros</Text>
            </View>
            <View style={styles.statCard}>
              <Text style={[styles.statValue, { color: '#a855f7' }]}>
                {stats.credits_markers}
              </Text>
              <Text style={styles.statLabel}>Credits</Text>
            </View>
            <View style={styles.statCard}>
              <Text style={[styles.statValue, { color: '#888' }]}>
                {stats.total_files}
              </Text>
              <Text style={styles.statLabel}>Files</Text>
            </View>
          </View>

          {/* Coverage Bars */}
          <View style={styles.coverageSection}>
            <Text style={styles.coverageLabel}>
              Intro Coverage ({stats.files_with_intro}/{stats.total_files})
            </Text>
            <View style={styles.coverageBarBg}>
              <View
                style={[
                  styles.coverageBarFill,
                  { width: `${introCoverage}%`, backgroundColor: '#3b82f6' },
                ]}
              />
            </View>
            <Text style={styles.coveragePercent}>{introCoverage}%</Text>
          </View>

          <View style={styles.coverageSection}>
            <Text style={styles.coverageLabel}>
              Credits Coverage ({stats.files_with_credits}/{stats.total_files})
            </Text>
            <View style={styles.coverageBarBg}>
              <View
                style={[
                  styles.coverageBarFill,
                  { width: `${creditsCoverage}%`, backgroundColor: '#a855f7' },
                ]}
              />
            </View>
            <Text style={styles.coveragePercent}>{creditsCoverage}%</Text>
          </View>

          {/* Source Breakdown */}
          <View style={styles.sourceRow}>
            {stats.chapter_source > 0 && (
              <Text style={[styles.sourceTag, { color: '#22c55e' }]}>
                Chapter: {stats.chapter_source}
              </Text>
            )}
            {stats.fingerprint_source > 0 && (
              <Text style={[styles.sourceTag, { color: '#3b82f6' }]}>
                Fingerprint: {stats.fingerprint_source}
              </Text>
            )}
            {stats.manual_source > 0 && (
              <Text style={[styles.sourceTag, { color: '#eab308' }]}>
                Manual: {stats.manual_source}
              </Text>
            )}
          </View>
        </>
      ) : (
        <Text style={styles.emptyText}>No markers detected yet</Text>
      )}

      {/* Backfill Controls */}
      <View style={styles.divider} />
      <Text style={styles.sectionLabel}>Run Detection</Text>

      <View style={styles.backfillRow}>
        <TouchableOpacity style={styles.librarySelector} onPress={handleSelectLibrary}>
          <Text style={styles.librarySelectorText}>
            {libraries?.find((l) => l.id === selectedLibrary)?.name || 'Select Library'}
          </Text>
          <Text style={styles.librarySelectorArrow}>▼</Text>
        </TouchableOpacity>

        <TouchableOpacity
          style={[styles.backfillBtn, backfill.isPending && styles.backfillBtnDisabled]}
          onPress={handleBackfill}
          disabled={backfill.isPending || selectedLibrary === 0}
        >
          {backfill.isPending ? (
            <ActivityIndicator size="small" color="#fff" />
          ) : (
            <Text style={styles.backfillBtnText}>Detect</Text>
          )}
        </TouchableOpacity>
      </View>
    </CollapsibleSection>
  )
}

const styles = StyleSheet.create({
  loadingText: {
    color: '#888',
    fontSize: 14,
  },
  emptyText: {
    color: '#666',
    fontSize: 14,
    textAlign: 'center',
    paddingVertical: 16,
  },
  statsGrid: {
    flexDirection: 'row',
    gap: 8,
    marginBottom: 16,
  },
  statCard: {
    flex: 1,
    backgroundColor: '#2a2a2a',
    borderRadius: 8,
    padding: 10,
    alignItems: 'center',
  },
  statValue: {
    color: '#fff',
    fontSize: 20,
    fontWeight: '700',
  },
  statLabel: {
    color: '#888',
    fontSize: 10,
    fontWeight: '500',
    marginTop: 2,
  },
  coverageSection: {
    marginBottom: 10,
  },
  coverageLabel: {
    color: '#888',
    fontSize: 12,
    marginBottom: 4,
  },
  coverageBarBg: {
    height: 6,
    backgroundColor: '#2a2a2a',
    borderRadius: 3,
    overflow: 'hidden',
  },
  coverageBarFill: {
    height: '100%',
    borderRadius: 3,
  },
  coveragePercent: {
    color: '#fff',
    fontSize: 12,
    fontWeight: '600',
    textAlign: 'right',
    marginTop: 2,
  },
  sourceRow: {
    flexDirection: 'row',
    flexWrap: 'wrap',
    gap: 12,
    marginTop: 8,
  },
  sourceTag: {
    fontSize: 12,
    fontWeight: '500',
  },
  divider: {
    height: 1,
    backgroundColor: '#2a2a2a',
    marginVertical: 12,
  },
  sectionLabel: {
    color: '#888',
    fontSize: 11,
    fontWeight: '600',
    textTransform: 'uppercase',
    marginBottom: 10,
  },
  backfillRow: {
    flexDirection: 'row',
    gap: 10,
  },
  librarySelector: {
    flex: 1,
    flexDirection: 'row',
    alignItems: 'center',
    justifyContent: 'space-between',
    backgroundColor: '#2a2a2a',
    borderRadius: 8,
    paddingHorizontal: 14,
    paddingVertical: 10,
  },
  librarySelectorText: {
    color: '#fff',
    fontSize: 14,
  },
  librarySelectorArrow: {
    color: '#888',
    fontSize: 10,
  },
  backfillBtn: {
    backgroundColor: '#e50914',
    borderRadius: 8,
    paddingHorizontal: 20,
    paddingVertical: 10,
    alignItems: 'center',
    justifyContent: 'center',
    minWidth: 80,
  },
  backfillBtnDisabled: {
    opacity: 0.5,
  },
  backfillBtnText: {
    color: '#fff',
    fontSize: 14,
    fontWeight: '600',
  },
})
