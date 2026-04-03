/**
 * Animation utilities for React Native
 *
 * Screen transitions, card press feedback, and common animations.
 */

import { useRef, useCallback, useEffect } from 'react'
import { Animated, Pressable, ViewStyle, StyleProp, View } from 'react-native'

// ── Screen Transitions ────────────────────────────────────────────────────────

/**
 * Fade in animation hook
 */
export function useFadeIn(duration = 300) {
  const opacity = useRef(new Animated.Value(0)).current

  const fadeIn = useCallback(
    (callback?: () => void) => {
      Animated.timing(opacity, {
        toValue: 1,
        duration,
        useNativeDriver: true,
      }).start(callback)
    },
    [opacity, duration],
  )

  return { opacity, fadeIn }
}

/**
 * Slide up animation hook
 */
export function useSlideUp(duration = 300) {
  const translateY = useRef(new Animated.Value(50)).current
  const opacity = useRef(new Animated.Value(0)).current

  const slideUp = useCallback(
    (callback?: () => void) => {
      Animated.parallel([
        Animated.timing(translateY, {
          toValue: 0,
          duration,
          useNativeDriver: true,
        }),
        Animated.timing(opacity, {
          toValue: 1,
          duration,
          useNativeDriver: true,
        }),
      ]).start(callback)
    },
    [translateY, opacity, duration],
  )

  return { translateY, opacity, slideUp }
}

// ── Card Press Animation ──────────────────────────────────────────────────────

interface PressAnimationConfig {
  scale?: number
  opacity?: number
}

const DEFAULT_PRESS_CONFIG: PressAnimationConfig = {
  scale: 0.96,
  opacity: 0.8,
}

/**
 * Card press animation wrapper
 * Wraps any Pressable with scale + opacity feedback
 */
export function AnimatedCard({
  children,
  style,
  onPress,
  pressConfig = DEFAULT_PRESS_CONFIG,
  disabled,
}: {
  children: React.ReactNode
  style?: StyleProp<ViewStyle>
  onPress?: () => void
  pressConfig?: PressAnimationConfig
  disabled?: boolean
}) {
  const scale = useRef(new Animated.Value(1)).current
  const opacity = useRef(new Animated.Value(1)).current

  const handlePressIn = () => {
    if (disabled) return
    Animated.parallel([
      Animated.spring(scale, {
        toValue: pressConfig.scale ?? 0.96,
        useNativeDriver: true,
        friction: 8,
      }),
      Animated.timing(opacity, {
        toValue: pressConfig.opacity ?? 0.8,
        duration: 100,
        useNativeDriver: true,
      }),
    ]).start()
  }

  const handlePressOut = () => {
    if (disabled) return
    Animated.parallel([
      Animated.spring(scale, {
        toValue: 1,
        useNativeDriver: true,
        friction: 8,
      }),
      Animated.timing(opacity, {
        toValue: 1,
        duration: 100,
        useNativeDriver: true,
      }),
    ]).start()
  }

  return (
    <Pressable
      onPress={onPress}
      onPressIn={handlePressIn}
      onPressOut={handlePressOut}
      disabled={disabled}
    >
      <Animated.View
        style={[
          style,
          {
            transform: [{ scale }],
            opacity,
          },
        ]}
      >
        {children}
      </Animated.View>
    </Pressable>
  )
}

// ── Skeleton Pulse ────────────────────────────────────────────────────────────

/**
 * Skeleton loading pulse animation hook
 */
export function useSkeletonPulse() {
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

  return opacity
}

// ── Stagger Animation ─────────────────────────────────────────────────────────

interface StaggerItem {
  translateY: Animated.Value
  opacity: Animated.Value
}

/**
 * Create staggered animation values for list items
 */
export function createStaggerAnimation(
  count: number,
  itemHeight: number,
  baseDuration = 200,
): StaggerItem[] {
  return Array.from({ length: count }).map((_, index) => ({
    translateY: new Animated.Value(itemHeight),
    opacity: new Animated.Value(0),
  }))
}

/**
 * Run staggered animations for a list
 */
export function runStaggerAnimation(
  items: StaggerItem[],
  baseDuration = 200,
  onComplete?: () => void,
) {
  const animations = items.map((item, index) => {
    const delay = index * 50
    return Animated.parallel([
      Animated.timing(item.translateY, {
        toValue: 0,
        duration: baseDuration,
        delay,
        useNativeDriver: true,
      }),
      Animated.timing(item.opacity, {
        toValue: 1,
        duration: baseDuration,
        delay,
        useNativeDriver: true,
      }),
    ])
  })

  Animated.parallel(animations).start(onComplete)
}

// ── Spring Config ─────────────────────────────────────────────────────────────

export const SPRING_CONFIG = {
  default: {
    friction: 7,
    tension: 40,
  },
  gentle: {
    friction: 12,
    tension: 30,
  },
  bouncy: {
    friction: 4,
    tension: 60,
  },
} as const

// ── FadeIn Component ───────────────────────────────────────────────────────────

interface FadeInProps {
  children: React.ReactNode
  style?: StyleProp<ViewStyle>
  delay?: number
  duration?: number
}

/**
 * Fade in wrapper for any content
 */
export function FadeIn({ children, style, delay = 0, duration = 300 }: FadeInProps) {
  const opacity = useRef(new Animated.Value(0)).current
  const translateY = useRef(new Animated.Value(20)).current

  useEffect(() => {
    Animated.parallel([
      Animated.timing(opacity, {
        toValue: 1,
        duration,
        delay,
        useNativeDriver: true,
      }),
      Animated.timing(translateY, {
        toValue: 0,
        duration,
        delay,
        useNativeDriver: true,
      }),
    ]).start()
  }, [opacity, translateY, delay, duration])

  return (
    <Animated.View style={[{ opacity, transform: [{ translateY }] }, style]}>
      {children}
    </Animated.View>
  )
}