/**
 * AdminScreen - Basic admin panel for mobile
 *
 * Shows: Server info, Libraries list, Trigger scan, Activity log
 * Only visible to admin users.
 */

import { useState } from 'react'
import {
  View,
  Text,
  StyleSheet,
  ScrollView,
  TouchableOpacity,
  Alert,
  ActivityIndicator,
  TextInput,
} from 'react-native'
import { Film, Tv, Folder, Lock, RefreshCw } from 'lucide-react-native'
import {
  useServerInfo,
  useLibraryStats,
  useActivity,
} from '@velox/shared/hooks/admin'
import { useLibraries, useScanLibrary } from '@velox/shared/hooks/media'
import { useMutation } from '@tanstack/react-query'
import { api } from '@velox/shared/api'
import { useAuthStore } from '../stores'
import { Skeleton, ListSkeleton } from '../components/SkeletonLoader'
import { Toast } from '../components/Toast'
import {
  PretranscodeSection,
  MarkersSection,
  WebhooksSection,
  TasksSection,
  LibraryStatsSection,
} from '../components/admin'
import { useResponsiveLayout, scaledFont, scaledSpacing } from '../lib/responsive'

// Action types for activity log filter
const ACTION_TYPES = [
  { value: '', label: 'All Actions' },
  { value: 'login', label: 'Login' },
  { value: 'logout', label: 'Logout' },
  { value: 'scan', label: 'Scan' },
  { value: 'play', label: 'Play' },
  { value: 'transcode', label: 'Transcode' },
  { value: 'subtitle_download', label: 'Subtitle Download' },
  { value: 'user_create', label: 'User Create' },
  { value: 'user_update', label: 'User Update' },
  { value: 'user_delete', label: 'User Delete' },
  { value: 'settings_update', label: 'Settings Update' },
]

// =============================================================================
// Admin Screen
// =============================================================================

export function AdminScreen() {
  const { user } = useAuthStore()
  const layout = useResponsiveLayout()

  // Check admin permission
  if (!user?.is_admin) {
    return (
      <View style={styles.container}>
        <View style={[styles.header, { paddingHorizontal: layout.screenPadding }]}>
          <Text style={[styles.headerTitle, { fontSize: scaledFont(28, layout.fontScale) }]}>
            Admin
          </Text>
        </View>
        <View style={styles.emptyState}>
          <Lock size={48} color="#888" />
          <Text style={[styles.emptyTitle, { fontSize: scaledFont(20, layout.fontScale) }]}>
            Access Denied
          </Text>
          <Text style={[styles.emptyText, { fontSize: scaledFont(14, layout.fontScale) }]}>
            Admin privileges required
          </Text>
        </View>
      </View>
    )
  }

  // On tablet: use 2-column grid for top sections
  const useGrid = layout.device === 'tablet' || layout.device === 'tv'

  return (
    <View style={styles.container}>
      <View style={[styles.header, { paddingHorizontal: layout.screenPadding }]}>
        <Text style={[styles.headerTitle, { fontSize: scaledFont(28, layout.fontScale) }]}>
          Admin
        </Text>
      </View>
      <ScrollView style={styles.content} showsVerticalScrollIndicator={false}>
        {useGrid ? (
          // Tablet: 2-column grid for sections
          <>
            <View style={[styles.gridRow, { paddingHorizontal: layout.screenPadding }]}>
              <View style={styles.gridCol}>
                <ServerInfoSection fontScale={layout.fontScale} />
              </View>
              <View style={styles.gridCol}>
                <LibrariesSection fontScale={layout.fontScale} />
              </View>
            </View>
            <View style={[styles.gridRow, { paddingHorizontal: layout.screenPadding }]}>
              <View style={styles.gridCol}>
                <View style={styles.section}>
                  <LibraryStatsSection />
                </View>
              </View>
              <View style={styles.gridCol}>
                <View style={styles.section}>
                  <PretranscodeSection />
                </View>
              </View>
            </View>
            <View style={[styles.gridRow, { paddingHorizontal: layout.screenPadding }]}>
              <View style={styles.gridCol}>
                <View style={styles.section}>
                  <MarkersSection />
                </View>
              </View>
              <View style={styles.gridCol}>
                <View style={styles.section}>
                  <TasksSection />
                </View>
              </View>
            </View>
            <View style={[styles.gridRow, { paddingHorizontal: layout.screenPadding }]}>
              <View style={styles.gridCol}>
                <View style={styles.section}>
                  <WebhooksSection />
                </View>
              </View>
              <View style={styles.gridCol}>
                <View style={styles.section}>
                  <BulkMetadataRefreshSection fontScale={layout.fontScale} />
                </View>
              </View>
            </View>
            <View style={{ paddingHorizontal: layout.screenPadding }}>
              <ActivitySection fontScale={layout.fontScale} />
            </View>
          </>
        ) : (
          // Phone: single column (existing layout)
          <>
            <ServerInfoSection fontScale={layout.fontScale} />
            <LibrariesSection fontScale={layout.fontScale} />
            <View style={styles.section}>
              <LibraryStatsSection />
            </View>
            <View style={styles.section}>
              <PretranscodeSection />
            </View>
            <View style={styles.section}>
              <MarkersSection />
            </View>
            <View style={styles.section}>
              <TasksSection />
            </View>
            <View style={styles.section}>
              <WebhooksSection />
            </View>
            <View style={styles.section}>
              <BulkMetadataRefreshSection fontScale={layout.fontScale} />
            </View>
            <ActivitySection fontScale={layout.fontScale} />
          </>
        )}
      </ScrollView>
    </View>
  )
}

// =============================================================================
// Server Info Section
// =============================================================================

function ServerInfoSection({ fontScale = 1.0 }: { fontScale?: number }) {
  const { data: info, isLoading } = useServerInfo()

  if (isLoading) {
    return (
      <View style={styles.section}>
        <Text style={[styles.sectionTitle, { fontSize: scaledFont(12, fontScale) }]}>
          Server Info
        </Text>
        <View style={styles.card}>
          <Skeleton width="100%" height={120} borderRadius={8} />
        </View>
      </View>
    )
  }

  if (!info) return null

  const formatBytes = (bytes: number) => {
    if (bytes === 0) return '0 B'
    const k = 1024
    const sizes = ['B', 'KB', 'MB', 'GB', 'TB']
    const i = Math.floor(Math.log(bytes) / Math.log(k))
    return `${(bytes / Math.pow(k, i)).toFixed(1)} ${sizes[i]}`
  }

  return (
    <View style={styles.section}>
      <Text style={[styles.sectionTitle, { fontSize: scaledFont(12, fontScale) }]}>
        Server Info
      </Text>
      <View style={styles.card}>
        <View style={styles.infoRow}>
          <Text style={[styles.infoLabel, { fontSize: scaledFont(14, fontScale) }]}>Version</Text>
          <Text style={[styles.infoValue, { fontSize: scaledFont(14, fontScale) }]}>
            {info.version}
          </Text>
        </View>
        <View style={styles.infoRow}>
          <Text style={[styles.infoLabel, { fontSize: scaledFont(14, fontScale) }]}>Uptime</Text>
          <Text style={[styles.infoValue, { fontSize: scaledFont(14, fontScale) }]}>
            {info.uptime}
          </Text>
        </View>
        <View style={styles.infoRow}>
          <Text style={[styles.infoLabel, { fontSize: scaledFont(14, fontScale) }]}>
            Media Files
          </Text>
          <Text style={[styles.infoValue, { fontSize: scaledFont(14, fontScale) }]}>
            {info.media_count.toLocaleString()}
          </Text>
        </View>
        <View style={styles.infoRow}>
          <Text style={[styles.infoLabel, { fontSize: scaledFont(14, fontScale) }]}>Series</Text>
          <Text style={[styles.infoValue, { fontSize: scaledFont(14, fontScale) }]}>
            {info.series_count.toLocaleString()}
          </Text>
        </View>
        <View style={styles.infoRow}>
          <Text style={[styles.infoLabel, { fontSize: scaledFont(14, fontScale) }]}>Users</Text>
          <Text style={[styles.infoValue, { fontSize: scaledFont(14, fontScale) }]}>
            {info.user_count}
          </Text>
        </View>
        <View style={styles.infoRow}>
          <Text style={[styles.infoLabel, { fontSize: scaledFont(14, fontScale) }]}>Storage</Text>
          <Text style={[styles.infoValue, { fontSize: scaledFont(14, fontScale) }]}>
            {formatBytes(info.total_size_bytes)}
          </Text>
        </View>
        <View style={styles.infoRow}>
          <Text style={[styles.infoLabel, { fontSize: scaledFont(14, fontScale) }]}>
            Hardware Acceleration
          </Text>
          <Text style={[styles.infoValue, { fontSize: scaledFont(14, fontScale) }]}>
            {info.hw_accel}
          </Text>
        </View>
        <View style={styles.infoRow}>
          <Text style={[styles.infoLabel, { fontSize: scaledFont(14, fontScale) }]}>FFmpeg</Text>
          <Text style={[styles.infoValue, { fontSize: scaledFont(14, fontScale) }]}>
            {info.ffmpeg_version}
          </Text>
        </View>
      </View>
    </View>
  )
}

// =============================================================================
// Libraries Section
// =============================================================================

function LibrariesSection({ fontScale = 1.0 }: { fontScale?: number }) {
  const { data: libraries, isLoading } = useLibraries()
  const { mutate: scanLibrary } = useScanLibrary()
  const [scanningId, setScanningId] = useState<number | null>(null)

  const handleScan = (id: number, name: string, force = false) => {
    Alert.alert(
      force ? 'Force Rescan' : 'Scan Library',
      force
        ? `Force rescan "${name}"? This will re-scan all files regardless of modification time.`
        : `Start scanning "${name}"? This may take a while.`,
      [
        { text: 'Cancel', style: 'cancel' },
        {
          text: force ? 'Force Rescan' : 'Scan',
          onPress: () => {
            setScanningId(id)
            scanLibrary(
              { id, force },
              {
                onSettled: () => setScanningId(null),
                onSuccess: () =>
                  Toast.success(force ? 'Force rescan started' : 'Scan started'),
                onError: () => Toast.error('Scan failed'),
              },
            )
          },
        },
      ],
    )
  }

  if (isLoading) {
    return (
      <View style={styles.section}>
        <Text style={[styles.sectionTitle, { fontSize: scaledFont(12, fontScale) }]}>
          Libraries
        </Text>
        <View style={styles.card}>
          <ListSkeleton count={3} />
        </View>
      </View>
    )
  }

  return (
    <View style={styles.section}>
      <Text style={[styles.sectionTitle, { fontSize: scaledFont(12, fontScale) }]}>
        Libraries ({libraries?.length || 0})
      </Text>
      <View style={styles.card}>
        {libraries?.length === 0 ? (
          <View style={styles.emptyCard}>
            <Text style={[styles.emptyCardText, { fontSize: scaledFont(14, fontScale) }]}>
              No libraries configured
            </Text>
          </View>
        ) : (
          libraries?.map((lib) => (
            <View key={lib.id} style={styles.libraryItem}>
              <View style={styles.libraryIcon}>
                {lib.type === 'movies' ? (
                  <Film size={18} color="#fff" />
                ) : lib.type === 'tvshows' ? (
                  <Tv size={18} color="#fff" />
                ) : (
                  <Folder size={18} color="#fff" />
                )}
              </View>
              <View style={styles.libraryInfo}>
                <Text style={[styles.libraryName, { fontSize: scaledFont(15, fontScale) }]}>
                  {lib.name}
                </Text>
                <Text style={[styles.libraryMeta, { fontSize: scaledFont(13, fontScale) }]}>
                  {lib.type} · {lib.paths.length} path(s)
                </Text>
              </View>
              <View style={styles.libraryActions}>
                <TouchableOpacity
                  style={styles.scanButton}
                  onPress={() => handleScan(lib.id, lib.name, false)}
                  disabled={scanningId === lib.id}
                >
                  {scanningId === lib.id ? (
                    <ActivityIndicator size="small" color="#e50914" />
                  ) : (
                    <Text
                      style={[styles.scanButtonText, { fontSize: scaledFont(13, fontScale) }]}
                    >
                      Scan
                    </Text>
                  )}
                </TouchableOpacity>
                <TouchableOpacity
                  style={[styles.scanButton, styles.forceRescanButton]}
                  onPress={() => handleScan(lib.id, lib.name, true)}
                  disabled={scanningId === lib.id}
                >
                  {scanningId === lib.id ? (
                    <ActivityIndicator size="small" color="#f59e0b" />
                  ) : (
                    <Text
                      style={[
                        styles.forceRescanButtonText,
                        { fontSize: scaledFont(13, fontScale) },
                      ]}
                    >
                      Force
                    </Text>
                  )}
                </TouchableOpacity>
              </View>
            </View>
          ))
        )}
      </View>
    </View>
  )
}

// =============================================================================
// Activity Section (Recent Logs)
// =============================================================================

function ActivitySection({ fontScale = 1.0 }: { fontScale?: number }) {
  const [filters, setFilters] = useState({
    dateFrom: '',
    dateTo: '',
    action: '',
  })
  const [showFilters, setShowFilters] = useState(false)

  const filterParams: Record<string, string> = { limit: '50' }
  if (filters.dateFrom) filterParams.from = filters.dateFrom
  if (filters.dateTo) filterParams.to = filters.dateTo
  if (filters.action) filterParams.action = filters.action

  const { data: activity, isLoading } = useActivity(filterParams)

  const formatTime = (dateStr: string) => {
    const date = new Date(dateStr)
    const now = new Date()
    const diffMs = now.getTime() - date.getTime()
    const diffMins = Math.floor(diffMs / 60000)

    if (diffMins < 1) return 'Just now'
    if (diffMins < 60) return `${diffMins}m ago`
    if (diffMins < 1440) return `${Math.floor(diffMins / 60)}h ago`
    return date.toLocaleDateString()
  }

  const handleActionFilter = () => {
    Alert.alert(
      'Filter by Action',
      undefined,
      [
        ...ACTION_TYPES.map((type) => ({
          text: type.label,
          onPress: () => setFilters({ ...filters, action: type.value }),
        })),
        { text: 'Cancel', style: 'cancel' as const },
      ],
    )
  }

  const clearFilters = () => {
    setFilters({ dateFrom: '', dateTo: '', action: '' })
  }

  const hasActiveFilters = filters.dateFrom || filters.dateTo || filters.action

  return (
    <View style={styles.section}>
      <View style={styles.sectionHeader}>
        <Text style={[styles.sectionTitle, { fontSize: scaledFont(12, fontScale) }]}>
          Activity Log
        </Text>
        <TouchableOpacity onPress={() => setShowFilters(!showFilters)}>
          <Text
            style={[
              styles.filterToggle,
              { fontSize: scaledFont(12, fontScale) },
              showFilters && styles.filterToggleActive,
            ]}
          >
            {showFilters ? 'Hide Filters' : 'Show Filters'}
          </Text>
        </TouchableOpacity>
      </View>

      {showFilters && (
        <View style={styles.filterCard}>
          <View style={styles.filterRow}>
            <View style={styles.filterField}>
              <Text style={[styles.filterLabel, { fontSize: scaledFont(11, fontScale) }]}>
                Date From
              </Text>
              <TextInput
                style={[styles.filterInput, { fontSize: scaledFont(14, fontScale) }]}
                value={filters.dateFrom}
                onChangeText={(text) => setFilters({ ...filters, dateFrom: text })}
                placeholder="YYYY-MM-DD"
                placeholderTextColor="#666"
              />
            </View>
            <View style={styles.filterField}>
              <Text style={[styles.filterLabel, { fontSize: scaledFont(11, fontScale) }]}>
                Date To
              </Text>
              <TextInput
                style={[styles.filterInput, { fontSize: scaledFont(14, fontScale) }]}
                value={filters.dateTo}
                onChangeText={(text) => setFilters({ ...filters, dateTo: text })}
                placeholder="YYYY-MM-DD"
                placeholderTextColor="#666"
              />
            </View>
          </View>
          <View style={styles.filterRow}>
            <View style={[styles.filterField, styles.filterFieldFull]}>
              <Text style={[styles.filterLabel, { fontSize: scaledFont(11, fontScale) }]}>
                Action Type
              </Text>
              <TouchableOpacity style={styles.filterSelect} onPress={handleActionFilter}>
                <Text style={[styles.filterSelectText, { fontSize: scaledFont(14, fontScale) }]}>
                  {ACTION_TYPES.find((a) => a.value === filters.action)?.label || 'All Actions'}
                </Text>
                <Text style={styles.filterArrow}>▼</Text>
              </TouchableOpacity>
            </View>
          </View>
          {hasActiveFilters && (
            <TouchableOpacity style={styles.clearFiltersBtn} onPress={clearFilters}>
              <Text
                style={[styles.clearFiltersText, { fontSize: scaledFont(13, fontScale) }]}
              >
                Clear Filters
              </Text>
            </TouchableOpacity>
          )}
        </View>
      )}

      <View style={styles.card}>
        {isLoading ? (
          <ListSkeleton count={5} />
        ) : activity?.length === 0 ? (
          <View style={styles.emptyCard}>
            <Text style={[styles.emptyCardText, { fontSize: scaledFont(14, fontScale) }]}>
              No activity logs found
            </Text>
          </View>
        ) : (
          activity?.map((log) => (
            <View key={log.id} style={styles.activityItem}>
              <View style={styles.activityIcon}>
                <Text>📋</Text>
              </View>
              <View style={styles.activityInfo}>
                <Text style={[styles.activityAction, { fontSize: scaledFont(14, fontScale) }]}>
                  {log.username || 'System'}{' '}
                  <Text style={styles.activityDetail}>{log.action}</Text>
                </Text>
                {log.media_title && (
                  <Text
                    style={[styles.activityMedia, { fontSize: scaledFont(13, fontScale) }]}
                  >
                    {log.media_title}
                  </Text>
                )}
                <Text style={[styles.activityTime, { fontSize: scaledFont(12, fontScale) }]}>
                  {formatTime(log.created_at)} · {log.ip_address}
                </Text>
              </View>
            </View>
          ))
        )}
      </View>
    </View>
  )
}

// =============================================================================
// Bulk Metadata Refresh Section
// =============================================================================

function BulkMetadataRefreshSection({ fontScale = 1.0 }: { fontScale?: number }) {
  const refreshAll = useMutation({
    mutationFn: () =>
      api.post<{ refreshed: number }>('/admin/metadata/refresh-all', {}),
  })

  const handleRefresh = () => {
    Alert.alert(
      'Refresh All Metadata',
      'This will re-fetch metadata for all media items from TMDb. This may take a while. Continue?',
      [
        { text: 'Cancel', style: 'cancel' },
        {
          text: 'Refresh All',
          style: 'destructive',
          onPress: () => {
            refreshAll.mutate(undefined, {
              onSuccess: () => Toast.success('Metadata refresh started'),
              onError: () => Toast.error('Failed to start metadata refresh'),
            })
          },
        },
      ],
    )
  }

  return (
    <View style={styles.card}>
      <View style={{ padding: 16 }}>
        <Text
          style={{
            color: '#fff',
            fontSize: scaledFont(16, fontScale),
            fontWeight: '600',
            marginBottom: 4,
          }}
        >
          Bulk Metadata Refresh
        </Text>
        <Text
          style={{
            color: '#888',
            fontSize: scaledFont(13, fontScale),
            marginBottom: 16,
            lineHeight: scaledFont(18, fontScale),
          }}
        >
          Re-fetch metadata for all media items from TMDb. Original files are not affected.
        </Text>
        <TouchableOpacity
          style={[
            styles.scanButton,
            {
              backgroundColor: 'rgba(229, 9, 20, 0.2)',
              paddingVertical: 12,
              paddingHorizontal: 20,
              borderRadius: 8,
              alignSelf: 'flex-start',
              minWidth: 120,
            },
            refreshAll.isPending && { opacity: 0.5 },
          ]}
          onPress={handleRefresh}
          disabled={refreshAll.isPending}
        >
          {refreshAll.isPending ? (
            <ActivityIndicator size="small" color="#e50914" />
          ) : (
            <Text
              style={{
                color: '#e50914',
                fontSize: scaledFont(14, fontScale),
                fontWeight: '600',
                textAlign: 'center',
              }}
            >
              Refresh All Metadata
            </Text>
          )}
        </TouchableOpacity>
      </View>
    </View>
  )
}

// =============================================================================
// Styles
// =============================================================================

const styles = StyleSheet.create({
  container: {
    flex: 1,
    backgroundColor: '#0a0a0a',
  },
  header: {
    paddingHorizontal: 20,
    paddingTop: 60,
    paddingBottom: 16,
    backgroundColor: '#141414',
  },
  headerTitle: {
    color: '#fff',
    fontSize: 28,
    fontWeight: 'bold',
  },
  content: {
    flex: 1,
    paddingTop: 16,
  },
  // ── Grid layout (tablet) ──────────────────────────────────────────────────
  gridRow: {
    flexDirection: 'row',
    gap: 16,
  },
  gridCol: {
    flex: 1,
  },
  section: {
    paddingHorizontal: 20,
    marginBottom: 24,
  },
  sectionTitle: {
    color: '#888',
    fontSize: 12,
    fontWeight: '600',
    textTransform: 'uppercase',
    marginBottom: 8,
    marginLeft: 4,
  },
  sectionHeader: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
    marginBottom: 8,
    marginLeft: 4,
  },
  filterToggle: {
    color: '#888',
    fontSize: 12,
    fontWeight: '500',
  },
  filterToggleActive: {
    color: '#e50914',
  },
  filterCard: {
    backgroundColor: '#1f1f1f',
    borderRadius: 12,
    padding: 12,
    marginBottom: 12,
  },
  filterRow: {
    flexDirection: 'row',
    gap: 12,
  },
  filterField: {
    flex: 1,
  },
  filterFieldFull: {
    flex: 1,
  },
  filterLabel: {
    color: '#888',
    fontSize: 11,
    fontWeight: '600',
    textTransform: 'uppercase',
    marginBottom: 6,
  },
  filterInput: {
    backgroundColor: '#2a2a2a',
    borderRadius: 8,
    paddingHorizontal: 12,
    paddingVertical: 10,
    color: '#fff',
    fontSize: 14,
  },
  filterSelect: {
    backgroundColor: '#2a2a2a',
    borderRadius: 8,
    paddingHorizontal: 12,
    paddingVertical: 10,
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
  },
  filterSelectText: {
    color: '#fff',
    fontSize: 14,
  },
  filterArrow: {
    color: '#888',
    fontSize: 10,
  },
  clearFiltersBtn: {
    marginTop: 12,
    paddingVertical: 8,
    alignItems: 'center',
    borderTopWidth: 1,
    borderTopColor: '#2a2a2a',
  },
  clearFiltersText: {
    color: '#e50914',
    fontSize: 13,
    fontWeight: '500',
  },
  card: {
    backgroundColor: '#1f1f1f',
    borderRadius: 12,
    padding: 4,
  },
  infoRow: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
    paddingVertical: 12,
    paddingHorizontal: 16,
    borderBottomWidth: 1,
    borderBottomColor: '#2a2a2a',
  },
  infoLabel: {
    color: '#888',
    fontSize: 14,
  },
  infoValue: {
    color: '#fff',
    fontSize: 14,
    fontWeight: '500',
  },
  libraryItem: {
    flexDirection: 'row',
    alignItems: 'center',
    paddingVertical: 12,
    paddingHorizontal: 16,
    borderBottomWidth: 1,
    borderBottomColor: '#2a2a2a',
  },
  libraryIcon: {
    width: 40,
    height: 40,
    borderRadius: 20,
    backgroundColor: '#2a2a2a',
    justifyContent: 'center',
    alignItems: 'center',
    marginRight: 12,
  },
  libraryInfo: {
    flex: 1,
  },
  libraryName: {
    color: '#fff',
    fontSize: 15,
    fontWeight: '500',
  },
  libraryMeta: {
    color: '#888',
    fontSize: 13,
    marginTop: 2,
  },
  scanButton: {
    minWidth: 56,
    height: 36,
    borderRadius: 18,
    backgroundColor: '#2a2a2a',
    justifyContent: 'center',
    alignItems: 'center',
    paddingHorizontal: 12,
  },
  scanButtonText: {
    color: '#fff',
    fontSize: 13,
    fontWeight: '500',
  },
  libraryActions: {
    flexDirection: 'row',
    gap: 8,
  },
  forceRescanButton: {
    backgroundColor: 'rgba(245, 158, 11, 0.2)',
  },
  forceRescanButtonText: {
    color: '#f59e0b',
    fontSize: 13,
    fontWeight: '600',
  },
  activityItem: {
    flexDirection: 'row',
    paddingVertical: 12,
    paddingHorizontal: 16,
    borderBottomWidth: 1,
    borderBottomColor: '#2a2a2a',
  },
  activityIcon: {
    width: 32,
    height: 32,
    borderRadius: 16,
    backgroundColor: '#2a2a2a',
    justifyContent: 'center',
    alignItems: 'center',
    marginRight: 12,
  },
  activityInfo: {
    flex: 1,
  },
  activityAction: {
    color: '#fff',
    fontSize: 14,
    fontWeight: '500',
  },
  activityDetail: {
    color: '#888',
    fontWeight: '400',
  },
  activityMedia: {
    color: '#e50914',
    fontSize: 13,
    marginTop: 2,
  },
  activityTime: {
    color: '#555',
    fontSize: 12,
    marginTop: 4,
  },
  emptyCard: {
    padding: 24,
    alignItems: 'center',
  },
  emptyCardText: {
    color: '#888',
    fontSize: 14,
  },
  emptyState: {
    flex: 1,
    justifyContent: 'center',
    alignItems: 'center',
    padding: 40,
  },
  emptyTitle: {
    color: '#fff',
    fontSize: 20,
    fontWeight: '600',
    marginBottom: 8,
  },
  emptyText: {
    color: '#888',
    fontSize: 14,
  },
})
