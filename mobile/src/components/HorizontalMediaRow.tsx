import { useRef, useState, useCallback } from 'react'
import {
  FlatList,
  StyleSheet,
  View,
  TouchableOpacity,
  type NativeSyntheticEvent,
  type NativeScrollEvent,
  type LayoutChangeEvent,
} from 'react-native'
import { ChevronLeft, ChevronRight } from 'lucide-react-native'
import type { MediaListItem } from '@velox/shared/types'
import { MediaCard } from './MediaCard'
import { useResponsiveLayout, scaledSpacing } from '../lib/responsive'

interface HorizontalMediaRowProps {
  items: MediaListItem[]
  onItemPress?: (item: MediaListItem) => void
  cardSize?: 'small' | 'medium' | 'large'
  showBadge?: boolean
}

export function HorizontalMediaRow({
  items,
  onItemPress,
  cardSize = 'medium',
  showBadge = false,
}: HorizontalMediaRowProps) {
  const layout = useResponsiveLayout()
  const arrowSize = layout.largeControls ? 24 : 18
  const indicatorWidth = layout.largeControls ? 44 : 32
  const indicatorInnerSize = layout.largeControls ? 36 : 28
  const flatListRef = useRef<FlatList>(null)
  const [canScrollLeft, setCanScrollLeft] = useState(false)
  const [canScrollRight, setCanScrollRight] = useState(false)
  const containerWidthRef = useRef(0)
  const contentWidthRef = useRef(0)
  const scrollXRef = useRef(0)

  const updateScrollIndicators = useCallback(() => {
    const scrollX = scrollXRef.current
    const containerWidth = containerWidthRef.current
    const contentWidth = contentWidthRef.current

    if (containerWidth <= 0 || contentWidth <= 0) return

    setCanScrollLeft(scrollX > 5)
    setCanScrollRight(scrollX < contentWidth - containerWidth - 5)
  }, [])

  const handleScroll = useCallback(
    (event: NativeSyntheticEvent<NativeScrollEvent>) => {
      scrollXRef.current = event.nativeEvent.contentOffset.x
      updateScrollIndicators()
    },
    [updateScrollIndicators],
  )

  const handleLayout = useCallback(
    (event: LayoutChangeEvent) => {
      containerWidthRef.current = event.nativeEvent.layout.width
      updateScrollIndicators()
    },
    [updateScrollIndicators],
  )

  const handleContentSizeChange = useCallback(
    (width: number) => {
      contentWidthRef.current = width
      updateScrollIndicators()
    },
    [updateScrollIndicators],
  )

  const scrollBy = useCallback((direction: 'left' | 'right') => {
    const amount = containerWidthRef.current * 0.75
    const newX =
      direction === 'right'
        ? scrollXRef.current + amount
        : scrollXRef.current - amount

    flatListRef.current?.scrollToOffset({
      offset: Math.max(0, newX),
      animated: true,
    })
  }, [])

  return (
    <View style={styles.wrapper} onLayout={handleLayout}>
      <FlatList
        ref={flatListRef}
        horizontal
        data={items}
        keyExtractor={(item) => String(item.id)}
        renderItem={({ item }) => (
          <MediaCard
            item={item}
            onPress={() => onItemPress?.(item)}
            size={cardSize}
            showBadge={showBadge}
            columns={layout.gridColumns}
            containerWidth={layout.width}
            gap={layout.cardGap}
            padding={layout.screenPadding}
          />
        )}
        showsHorizontalScrollIndicator={false}
        contentContainerStyle={[styles.content, { paddingHorizontal: layout.screenPadding }]}
        onScroll={handleScroll}
        scrollEventThrottle={16}
        onContentSizeChange={handleContentSizeChange}
      />

      {/* Left scroll indicator */}
      {canScrollLeft && (
        <TouchableOpacity
          style={[styles.indicator, styles.indicatorLeft, { width: indicatorWidth }]}
          onPress={() => scrollBy('left')}
          activeOpacity={0.8}
        >
          <View style={[styles.indicatorInner, { width: indicatorInnerSize, height: indicatorInnerSize, borderRadius: indicatorInnerSize / 2 }]}>
            <ChevronLeft size={arrowSize} color="#fff" />
          </View>
        </TouchableOpacity>
      )}

      {/* Right scroll indicator */}
      {canScrollRight && (
        <TouchableOpacity
          style={[styles.indicator, styles.indicatorRight, { width: indicatorWidth }]}
          onPress={() => scrollBy('right')}
          activeOpacity={0.8}
        >
          <View style={[styles.indicatorInner, { width: indicatorInnerSize, height: indicatorInnerSize, borderRadius: indicatorInnerSize / 2 }]}>
            <ChevronRight size={arrowSize} color="#fff" />
          </View>
        </TouchableOpacity>
      )}
    </View>
  )
}

const styles = StyleSheet.create({
  wrapper: {
    position: 'relative',
  },
  content: {
    // paddingHorizontal set dynamically via inline style
  },
  indicator: {
    position: 'absolute',
    top: 0,
    bottom: 0,
    justifyContent: 'center',
    alignItems: 'center',
  },
  indicatorLeft: {
    left: 0,
    backgroundColor: 'rgba(20, 20, 20, 0.7)',
    borderTopRightRadius: 4,
    borderBottomRightRadius: 4,
  },
  indicatorRight: {
    right: 0,
    backgroundColor: 'rgba(20, 20, 20, 0.7)',
    borderTopLeftRadius: 4,
    borderBottomLeftRadius: 4,
  },
  indicatorInner: {
    // width, height, borderRadius set dynamically via inline style
    backgroundColor: 'rgba(255, 255, 255, 0.15)',
    justifyContent: 'center',
    alignItems: 'center',
  },
})
