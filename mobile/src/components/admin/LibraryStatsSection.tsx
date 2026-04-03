/**
 * LibraryStatsSection - Per-library statistics (items, files, size, last scan)
 */

import { View, Text, StyleSheet } from 'react-native'
import { Film, Tv, Folder } from 'lucide-react-native'
import { useLibraryStats } from '@velox/shared/hooks/admin'
import { CollapsibleSection } from './CollapsibleSection'

function formatBytes(bytes: number): string {
  if (bytes === 0) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return `${(bytes / Math.pow(k, i)).toFixed(1)} ${sizes[i]}`
}

function timeAgo(dateStr: string): string {
  const date = new Date(dateStr)
  const now = new Date()
  const diffMs = now.getTime() - date.getTime()
  const diffMins = Math.floor(diffMs / 60000)

  if (diffMins < 1) return 'Just now'
  if (diffMins < 60) return `${diffMins}m ago`
  if (diffMins < 1440) return `${Math.floor(diffMins / 60)}h ago`
  if (diffMins < 10080) return `${Math.floor(diffMins / 1440)}d ago`
  return date.toLocaleDateString()
}

export function LibraryStatsSection() {
  const { data: stats, isLoading } = useLibraryStats()

  const totalSize = stats?.reduce((sum, s) => sum + s.total_size_bytes, 0) ?? 0
  const totalFiles = stats?.reduce((sum, s) => sum + s.file_count, 0) ?? 0

  return (
    <CollapsibleSection
      title="Library Statistics"
      subtitle={stats ? `${stats.length} libraries, ${formatBytes(totalSize)}` : 'Loading...'}
    >
      {isLoading ? (
        <Text style={styles.loadingText}>Loading stats...</Text>
      ) : !stats || stats.length === 0 ? (
        <Text style={styles.emptyText}>No libraries found</Text>
      ) : (
        <>
          {/* Summary row */}
          <View style={styles.summaryRow}>
            <View style={styles.summaryCard}>
              <Text style={styles.summaryValue}>{totalFiles.toLocaleString()}</Text>
              <Text style={styles.summaryLabel}>Total Files</Text>
            </View>
            <View style={styles.summaryCard}>
              <Text style={styles.summaryValue}>{formatBytes(totalSize)}</Text>
              <Text style={styles.summaryLabel}>Total Size</Text>
            </View>
          </View>

          {/* Per-library cards */}
          {stats.map((lib) => {
            const LibIcon =
              lib.type === 'movies' ? Film : lib.type === 'tvshows' ? Tv : Folder

            return (
              <View key={lib.id} style={styles.libCard}>
                <View style={styles.libHeader}>
                  <View style={styles.libIcon}>
                    <LibIcon size={16} color="#fff" />
                  </View>
                  <View style={styles.libTitleRow}>
                    <Text style={styles.libName}>{lib.name}</Text>
                    <Text style={styles.libType}>{lib.type}</Text>
                  </View>
                </View>

                <View style={styles.libStats}>
                  <View style={styles.libStatItem}>
                    <Text style={styles.libStatValue}>{lib.item_count.toLocaleString()}</Text>
                    <Text style={styles.libStatLabel}>Items</Text>
                  </View>
                  <View style={styles.libStatItem}>
                    <Text style={styles.libStatValue}>{lib.file_count.toLocaleString()}</Text>
                    <Text style={styles.libStatLabel}>Files</Text>
                  </View>
                  <View style={styles.libStatItem}>
                    <Text style={styles.libStatValue}>{formatBytes(lib.total_size_bytes)}</Text>
                    <Text style={styles.libStatLabel}>Size</Text>
                  </View>
                </View>

                <Text style={styles.lastScan}>
                  Last scanned: {lib.last_scanned ? timeAgo(lib.last_scanned) : 'Never'}
                </Text>
              </View>
            )
          })}
        </>
      )}
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
  summaryRow: {
    flexDirection: 'row',
    gap: 8,
    marginBottom: 12,
  },
  summaryCard: {
    flex: 1,
    backgroundColor: '#2a2a2a',
    borderRadius: 8,
    padding: 12,
    alignItems: 'center',
  },
  summaryValue: {
    color: '#fff',
    fontSize: 16,
    fontWeight: '700',
  },
  summaryLabel: {
    color: '#888',
    fontSize: 11,
    marginTop: 2,
  },
  libCard: {
    backgroundColor: '#2a2a2a',
    borderRadius: 8,
    padding: 12,
    marginBottom: 8,
  },
  libHeader: {
    flexDirection: 'row',
    alignItems: 'center',
    marginBottom: 10,
  },
  libIcon: {
    width: 32,
    height: 32,
    borderRadius: 16,
    backgroundColor: '#3a3a3a',
    alignItems: 'center',
    justifyContent: 'center',
    marginRight: 10,
  },
  libTitleRow: {
    flex: 1,
  },
  libName: {
    color: '#fff',
    fontSize: 14,
    fontWeight: '600',
  },
  libType: {
    color: '#888',
    fontSize: 12,
    marginTop: 1,
    textTransform: 'capitalize',
  },
  libStats: {
    flexDirection: 'row',
    gap: 8,
    marginBottom: 8,
  },
  libStatItem: {
    flex: 1,
    backgroundColor: '#1a1a1a',
    borderRadius: 6,
    padding: 8,
    alignItems: 'center',
  },
  libStatValue: {
    color: '#fff',
    fontSize: 13,
    fontWeight: '600',
  },
  libStatLabel: {
    color: '#888',
    fontSize: 10,
    marginTop: 2,
  },
  lastScan: {
    color: '#555',
    fontSize: 11,
  },
})
