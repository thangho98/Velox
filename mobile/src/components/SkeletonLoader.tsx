/**
 * SkeletonLoader - Loading placeholders that match component shapes
 */

import { View, StyleSheet, Animated } from 'react-native'
import { useEffect, useRef } from 'react'
import { useResponsiveLayout, scaledSpacing } from '../lib/responsive'

// ── Animated Skeleton Block ──────────────────────────────────────────────────

interface SkeletonProps {
  width?: number | string
  height?: number | string
  borderRadius?: number
  style?: any
}

export function Skeleton({ width = '100%', height = 16, borderRadius = 4, style }: SkeletonProps) {
  const opacity = useRef(new Animated.Value(0.3)).current

  useEffect(() => {
    const animation = Animated.loop(
      Animated.sequence([
        Animated.timing(opacity, {
          toValue: 0.7,
          duration: 800,
          useNativeDriver: true,
        }),
        Animated.timing(opacity, {
          toValue: 0.3,
          duration: 800,
          useNativeDriver: true,
        }),
      ]),
    )
    animation.start()
    return () => animation.stop()
  }, [opacity])

  return (
    <Animated.View
      style={[
        {
          width,
          height,
          borderRadius,
          backgroundColor: '#2a2a2a',
        },
        { opacity },
        style,
      ]}
    />
  )
}

// ── Card Skeleton ─────────────────────────────────────────────────────────────

interface CardSkeletonProps {
  width?: number
  height?: number
}

export function CardSkeleton({ width = 110, height = 165 }: CardSkeletonProps) {
  const layout = useResponsiveLayout()
  const scaledWidth = typeof width === 'number' ? scaledSpacing(width, layout.fontScale) : width
  const scaledHeight = typeof height === 'number' ? scaledSpacing(height, layout.fontScale) : height

  return (
    <View style={[styles.card, { width: scaledWidth, height: scaledHeight }]}>
      <Skeleton width="100%" height="100%" borderRadius={8} />
      <View style={styles.cardText}>
        <Skeleton width="80%" height={scaledSpacing(14, layout.fontScale)} borderRadius={4} />
        <Skeleton width="50%" height={scaledSpacing(10, layout.fontScale)} borderRadius={4} style={{ marginTop: 6 }} />
      </View>
    </View>
  )
}

// ── Row Skeleton (HorizontalMediaRow placeholder) ─────────────────────────────

interface RowSkeletonProps {
  count?: number
  cardWidth?: number
  cardHeight?: number
}

export function RowSkeleton({ count = 5, cardWidth = 110, cardHeight = 165 }: RowSkeletonProps) {
  const layout = useResponsiveLayout()
  const displayCount = layout.device === 'tv' ? 8 : layout.device === 'tablet' ? 6 : count

  return (
    <View style={[styles.row, { paddingHorizontal: layout.screenPadding }]}>
      {Array.from({ length: displayCount }).map((_, i) => (
        <CardSkeleton key={i} width={cardWidth} height={cardHeight} />
      ))}
    </View>
  )
}

// ── Section Skeleton (Home section placeholder) ───────────────────────────────

interface SectionSkeletonProps {
  titleWidth?: number
  showSeeAll?: boolean
}

export function SectionSkeleton({ titleWidth = 140, showSeeAll = true }: SectionSkeletonProps) {
  const layout = useResponsiveLayout()

  return (
    <View style={styles.section}>
      <View style={[styles.sectionHeader, { paddingHorizontal: layout.screenPadding }]}>
        <Skeleton width={scaledSpacing(titleWidth, layout.fontScale)} height={scaledSpacing(20, layout.fontScale)} borderRadius={4} />
        {showSeeAll && <Skeleton width={scaledSpacing(50, layout.fontScale)} height={scaledSpacing(14, layout.fontScale)} borderRadius={4} />}
      </View>
      <RowSkeleton />
    </View>
  )
}

// ── Detail Page Skeleton ──────────────────────────────────────────────────────

export function DetailPageSkeleton() {
  const layout = useResponsiveLayout()
  const backdropHeight = layout.device === 'tv' ? 300 : layout.device === 'tablet' ? 250 : 200
  const posterWidth = scaledSpacing(120, layout.fontScale)
  const posterHeight = scaledSpacing(180, layout.fontScale)
  const buttonHeight = layout.largeControls ? 56 : 44

  return (
    <View style={styles.detailPage}>
      {/* Backdrop */}
      <Skeleton width="100%" height={backdropHeight} borderRadius={0} />

      {/* Content */}
      <View style={[styles.detailContent, { padding: layout.screenPadding }]}>
        {/* Poster + Info row */}
        <View style={styles.detailRow}>
          <Skeleton width={posterWidth} height={posterHeight} borderRadius={8} />
          <View style={styles.detailInfo}>
            <Skeleton width="70%" height={scaledSpacing(24, layout.fontScale)} borderRadius={4} />
            <Skeleton width="50%" height={scaledSpacing(16, layout.fontScale)} borderRadius={4} style={{ marginTop: 8 }} />
            <Skeleton width="40%" height={scaledSpacing(14, layout.fontScale)} borderRadius={4} style={{ marginTop: 12 }} />
          </View>
        </View>

        {/* Action buttons */}
        <View style={styles.actionRow}>
          <Skeleton width={scaledSpacing(120, layout.fontScale)} height={buttonHeight} borderRadius={8} />
          <Skeleton width={buttonHeight} height={buttonHeight} borderRadius={buttonHeight / 2} style={{ marginLeft: 12 }} />
          <Skeleton width={buttonHeight} height={buttonHeight} borderRadius={buttonHeight / 2} style={{ marginLeft: 12 }} />
        </View>

        {/* Overview */}
        <View style={styles.overview}>
          <Skeleton width="100%" height={scaledSpacing(14, layout.fontScale)} borderRadius={4} />
          <Skeleton width="90%" height={scaledSpacing(14, layout.fontScale)} borderRadius={4} style={{ marginTop: 6 }} />
          <Skeleton width="70%" height={scaledSpacing(14, layout.fontScale)} borderRadius={4} style={{ marginTop: 6 }} />
        </View>
      </View>
    </View>
  )
}

// ── List Skeleton (for Settings, Sessions) ───────────────────────────────────

interface ListSkeletonProps {
  count?: number
}

export function ListSkeleton({ count = 5 }: ListSkeletonProps) {
  const layout = useResponsiveLayout()
  const avatarSize = layout.largeControls ? 56 : 44

  return (
    <View style={[styles.list, { padding: layout.screenPadding }]}>
      {Array.from({ length: count }).map((_, i) => (
        <View key={i} style={styles.listItem}>
          <Skeleton width={avatarSize} height={avatarSize} borderRadius={avatarSize / 2} />
          <View style={styles.listItemInfo}>
            <Skeleton width="60%" height={scaledSpacing(16, layout.fontScale)} borderRadius={4} />
            <Skeleton width="40%" height={scaledSpacing(12, layout.fontScale)} borderRadius={4} style={{ marginTop: 6 }} />
          </View>
          <Skeleton width={scaledSpacing(60, layout.fontScale)} height={scaledSpacing(32, layout.fontScale)} borderRadius={6} />
        </View>
      ))}
    </View>
  )
}

// ── Tab Content Skeleton ──────────────────────────────────────────────────────

export function TabContentSkeleton() {
  const layout = useResponsiveLayout()
  const inputHeight = layout.largeControls ? 56 : 48

  return (
    <View style={[styles.tabContent, { padding: layout.screenPadding }]}>
      <Skeleton width="100%" height={inputHeight} borderRadius={8} style={{ marginBottom: 24 }} />
      <Skeleton width="70%" height={scaledSpacing(20, layout.fontScale)} borderRadius={4} style={{ marginBottom: 16 }} />
      <Skeleton width="100%" height={inputHeight} borderRadius={8} style={{ marginBottom: 24 }} />
      <Skeleton width="60%" height={scaledSpacing(20, layout.fontScale)} borderRadius={4} style={{ marginBottom: 16 }} />
      <Skeleton width="100%" height={inputHeight} borderRadius={8} />
    </View>
  )
}

// ── Grid Skeleton (for Movies/Series screens) ─────────────────────────────────

interface GridSkeletonProps {
  columns?: number
  rows?: number
  cardWidth?: number
  cardHeight?: number
}

export function GridSkeleton({
  columns,
  rows = 4,
  cardWidth = 140,
  cardHeight = 210,
}: GridSkeletonProps) {
  const layout = useResponsiveLayout()
  const cols = columns ?? layout.gridColumns
  const items = Array.from({ length: cols * rows })

  return (
    <View style={[styles.gridSkeleton, { padding: layout.cardGap }]}>
      {items.map((_, i) => (
        <View key={i} style={[styles.gridItem, { width: `${100 / cols}%`, padding: layout.cardGap / 2 }]}>
          <Skeleton width={cardWidth} height={cardHeight} borderRadius={8} />
          <View style={styles.gridText}>
            <Skeleton width="80%" height={scaledSpacing(14, layout.fontScale)} borderRadius={4} style={{ marginTop: 8 }} />
            <Skeleton width="50%" height={scaledSpacing(10, layout.fontScale)} borderRadius={4} style={{ marginTop: 6 }} />
          </View>
        </View>
      ))}
    </View>
  )
}

// ── Hero Skeleton (for HomeScreen) ────────────────────────────────────────────

export function HeroSkeleton() {
  const layout = useResponsiveLayout()
  const buttonHeight = layout.largeControls ? 56 : 44

  return (
    <View style={[styles.heroSkeleton, { padding: layout.screenPadding }]}>
      <Skeleton width="60%" height={scaledSpacing(28, layout.fontScale)} borderRadius={4} />
      <Skeleton width="40%" height={scaledSpacing(16, layout.fontScale)} borderRadius={4} style={{ marginTop: 8 }} />
      <View style={styles.heroButtons}>
        <Skeleton width={scaledSpacing(100, layout.fontScale)} height={buttonHeight} borderRadius={8} />
        <Skeleton width={scaledSpacing(100, layout.fontScale)} height={buttonHeight} borderRadius={8} style={{ marginLeft: 12 }} />
      </View>
    </View>
  )
}

// ── Styles ────────────────────────────────────────────────────────────────────

const styles = StyleSheet.create({
  card: {
    backgroundColor: '#1f1f1f',
    borderRadius: 8,
    overflow: 'hidden',
    marginRight: 8,
  },
  cardText: {
    position: 'absolute',
    bottom: 10,
    left: 8,
    right: 8,
  },
  row: {
    flexDirection: 'row',
    paddingHorizontal: 20,
  },
  section: {
    marginTop: 24,
  },
  sectionHeader: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
    paddingHorizontal: 20,
    marginBottom: 12,
  },
  detailPage: {
    flex: 1,
  },
  detailContent: {
    padding: 20,
  },
  detailRow: {
    flexDirection: 'row',
    marginTop: -60,
  },
  detailInfo: {
    flex: 1,
    marginLeft: 16,
    paddingTop: 60,
  },
  actionRow: {
    flexDirection: 'row',
    marginTop: 20,
    marginBottom: 24,
  },
  overview: {
    marginTop: 16,
  },
  list: {
    padding: 20,
  },
  listItem: {
    flexDirection: 'row',
    alignItems: 'center',
    backgroundColor: '#1f1f1f',
    borderRadius: 12,
    padding: 16,
    marginBottom: 12,
  },
  listItemInfo: {
    flex: 1,
    marginLeft: 14,
  },
  tabContent: {
    padding: 20,
  },
  // Grid Skeleton styles
  gridSkeleton: {
    flexDirection: 'row',
    flexWrap: 'wrap',
    padding: 8,
  },
  gridItem: {
    padding: 8,
    alignItems: 'center',
  },
  gridText: {
    width: '100%',
    paddingHorizontal: 8,
  },
  // Hero Skeleton styles
  heroSkeleton: {
    padding: 20,
    paddingTop: 16,
    paddingBottom: 24,
  },
  heroButtons: {
    flexDirection: 'row',
    marginTop: 20,
  },
})