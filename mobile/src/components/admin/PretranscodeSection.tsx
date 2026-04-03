/**
 * PretranscodeSection - Pre-transcode dashboard with status, profiles, and controls
 */

import { useState } from 'react'
import {
  View,
  Text,
  Switch,
  StyleSheet,
  TouchableOpacity,
  ActivityIndicator,
  Alert,
} from 'react-native'
import {
  usePretranscodeSettings,
  useUpdatePretranscodeSettings,
  usePretranscodeStatus,
  usePretranscodeProfiles,
  useTogglePretranscodeProfile,
  useStartPretranscode,
  useStopPretranscode,
  useResumePretranscode,
  useCleanupPretranscode,
} from '@velox/shared/hooks/settings'
import { CollapsibleSection } from './CollapsibleSection'
import { Toast } from '../Toast'

function formatBytes(bytes: number): string {
  if (bytes === 0) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return `${(bytes / Math.pow(k, i)).toFixed(1)} ${sizes[i]}`
}

export function PretranscodeSection() {
  const { data: settings } = usePretranscodeSettings()
  const { data: status } = usePretranscodeStatus()
  const { data: profiles } = usePretranscodeProfiles()
  const updateSettings = useUpdatePretranscodeSettings()
  const toggleProfile = useTogglePretranscodeProfile()
  const startEncode = useStartPretranscode()
  const stopEncode = useStopPretranscode()
  const resumeEncode = useResumePretranscode()
  const cleanupFiles = useCleanupPretranscode()

  const isEncoding = status && (status.encoding > 0 || status.queued > 0) && !status.paused
  const progressPercent =
    status && status.total > 0 ? Math.round((status.done / status.total) * 100) : 0

  const statusLabel = status?.paused
    ? 'Paused'
    : isEncoding
      ? 'Encoding...'
      : settings?.enabled
        ? 'Idle'
        : 'Disabled'

  const handleCleanup = () => {
    Alert.alert(
      'Delete Pre-transcode Files',
      'This will permanently delete all pre-encoded files. Your original media files are NOT affected.',
      [
        { text: 'Cancel', style: 'cancel' },
        {
          text: 'Delete All',
          style: 'destructive',
          onPress: () => {
            cleanupFiles.mutate(undefined, {
              onSuccess: () => Toast.success('Pre-transcode files cleaned up'),
              onError: () => Toast.error('Cleanup failed'),
            })
          },
        },
      ],
    )
  }

  return (
    <CollapsibleSection title="Pre-transcode" subtitle={statusLabel}>
      {/* Enable Toggle */}
      <View style={styles.row}>
        <View style={styles.rowInfo}>
          <Text style={styles.rowTitle}>Offline Encoding</Text>
          <Text style={styles.rowDesc}>
            Pre-encode your library for instant playback
          </Text>
        </View>
        <Switch
          value={settings?.enabled ?? false}
          onValueChange={(v) => updateSettings.mutate({ enabled: v })}
          trackColor={{ false: '#3a3a3a', true: '#e50914' }}
          thumbColor="#fff"
        />
      </View>

      {settings?.enabled && (
        <>
          {/* Quality Profiles */}
          <View style={styles.divider} />
          <Text style={styles.sectionLabel}>Quality Profiles</Text>
          {profiles?.map((p) => (
            <View key={p.id} style={styles.profileRow}>
              <Switch
                value={p.enabled}
                onValueChange={() => toggleProfile.mutate({ id: p.id, enabled: !p.enabled })}
                trackColor={{ false: '#3a3a3a', true: '#e50914' }}
                thumbColor="#fff"
                style={styles.profileSwitch}
              />
              <View style={styles.profileInfo}>
                <Text style={styles.profileName}>{p.name}</Text>
                <Text style={styles.profileDetail}>
                  {p.height}p, {p.video_bitrate / 1000}Mbps + {p.audio_bitrate}kbps
                </Text>
              </View>
            </View>
          ))}

          {/* Progress Dashboard */}
          {status && status.total > 0 && (
            <>
              <View style={styles.divider} />
              <Text style={styles.sectionLabel}>Progress</Text>

              {/* Progress bar */}
              <View style={styles.progressBarBg}>
                <View
                  style={[styles.progressBarFill, { width: `${progressPercent}%` }]}
                />
              </View>
              <View style={styles.progressMeta}>
                <Text style={styles.metaText}>
                  {status.done}/{status.total} files
                </Text>
                <Text style={styles.metaText}>{progressPercent}%</Text>
              </View>

              {/* Current file */}
              {status.current_file ? (
                <Text style={styles.currentFile} numberOfLines={1}>
                  Encoding: {status.current_file.split('/').pop()}
                  {status.speed ? ` (${status.speed})` : ''}
                </Text>
              ) : null}

              {/* Stats row */}
              <View style={styles.statsRow}>
                <Text style={styles.statGreen}>Done: {status.done}</Text>
                {status.failed > 0 && (
                  <Text style={styles.statRed}>Failed: {status.failed}</Text>
                )}
                <Text style={styles.statGray}>Queued: {status.queued}</Text>
                <Text style={styles.statGray}>Disk: {formatBytes(status.disk_used)}</Text>
              </View>
            </>
          )}

          {/* Action Buttons */}
          <View style={styles.divider} />
          <View style={styles.actionsRow}>
            {status?.paused ? (
              <TouchableOpacity
                style={[styles.actionBtn, styles.actionBtnGreen]}
                onPress={() => resumeEncode.mutate()}
              >
                <Text style={styles.actionBtnText}>Resume</Text>
              </TouchableOpacity>
            ) : isEncoding ? (
              <TouchableOpacity
                style={[styles.actionBtn, styles.actionBtnYellow]}
                onPress={() => stopEncode.mutate()}
              >
                <Text style={styles.actionBtnTextDark}>Pause</Text>
              </TouchableOpacity>
            ) : (
              <TouchableOpacity
                style={[styles.actionBtn, styles.actionBtnRed]}
                onPress={() =>
                  startEncode.mutate(undefined, {
                    onSuccess: () => Toast.success('Encoding started'),
                    onError: () => Toast.error('Failed to start encoding'),
                  })
                }
                disabled={startEncode.isPending}
              >
                {startEncode.isPending ? (
                  <ActivityIndicator size="small" color="#fff" />
                ) : (
                  <Text style={styles.actionBtnText}>Start Encoding</Text>
                )}
              </TouchableOpacity>
            )}

            {isEncoding && (
              <TouchableOpacity
                style={[styles.actionBtn, styles.actionBtnGray]}
                onPress={() => stopEncode.mutate()}
              >
                <Text style={styles.actionBtnText}>Stop</Text>
              </TouchableOpacity>
            )}

            <TouchableOpacity
              style={[styles.actionBtn, styles.actionBtnGray]}
              onPress={handleCleanup}
            >
              <Text style={styles.actionBtnText}>Cleanup</Text>
            </TouchableOpacity>
          </View>
        </>
      )}
    </CollapsibleSection>
  )
}

const styles = StyleSheet.create({
  row: {
    flexDirection: 'row',
    alignItems: 'center',
    justifyContent: 'space-between',
    paddingVertical: 8,
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
    marginVertical: 12,
  },
  sectionLabel: {
    color: '#888',
    fontSize: 11,
    fontWeight: '600',
    textTransform: 'uppercase',
    marginBottom: 8,
  },
  profileRow: {
    flexDirection: 'row',
    alignItems: 'center',
    paddingVertical: 8,
    borderBottomWidth: 1,
    borderBottomColor: '#2a2a2a',
  },
  profileSwitch: {
    marginRight: 12,
  },
  profileInfo: {
    flex: 1,
  },
  profileName: {
    color: '#fff',
    fontSize: 14,
    fontWeight: '500',
  },
  profileDetail: {
    color: '#888',
    fontSize: 12,
    marginTop: 1,
  },
  progressBarBg: {
    height: 8,
    backgroundColor: '#2a2a2a',
    borderRadius: 4,
    overflow: 'hidden',
    marginTop: 4,
  },
  progressBarFill: {
    height: '100%',
    backgroundColor: '#e50914',
    borderRadius: 4,
  },
  progressMeta: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    marginTop: 6,
  },
  metaText: {
    color: '#888',
    fontSize: 12,
  },
  currentFile: {
    color: '#ccc',
    fontSize: 12,
    marginTop: 8,
  },
  statsRow: {
    flexDirection: 'row',
    flexWrap: 'wrap',
    gap: 12,
    marginTop: 8,
  },
  statGreen: {
    color: '#22c55e',
    fontSize: 12,
  },
  statRed: {
    color: '#ef4444',
    fontSize: 12,
  },
  statGray: {
    color: '#888',
    fontSize: 12,
  },
  actionsRow: {
    flexDirection: 'row',
    flexWrap: 'wrap',
    gap: 8,
  },
  actionBtn: {
    paddingHorizontal: 16,
    paddingVertical: 10,
    borderRadius: 8,
    minWidth: 80,
    alignItems: 'center',
  },
  actionBtnRed: {
    backgroundColor: '#e50914',
  },
  actionBtnGreen: {
    backgroundColor: '#16a34a',
  },
  actionBtnYellow: {
    backgroundColor: '#ca8a04',
  },
  actionBtnGray: {
    backgroundColor: '#333',
  },
  actionBtnText: {
    color: '#fff',
    fontSize: 13,
    fontWeight: '600',
  },
  actionBtnTextDark: {
    color: '#fff',
    fontSize: 13,
    fontWeight: '600',
  },
})
