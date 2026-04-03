/**
 * Toast - Temporary notification messages
 *
 * Replaces Alert.dialog with non-blocking toast notifications.
 * Usage: Toast.show('Message', 'success' | 'error' | 'info')
 */

import { useEffect, useRef, useState } from 'react'
import { Animated, Text, StyleSheet, TouchableOpacity } from 'react-native'
import { X } from 'lucide-react-native'
import { useResponsiveLayout, scaledFont } from '../lib/responsive'

// ── Toast Context & Provider ──────────────────────────────────────────────────

import { createContext, useContext, useCallback } from 'react'

type ToastType = 'success' | 'error' | 'info'

interface ToastOptions {
  message: string
  type?: ToastType
  duration?: number
}

interface ToastItem extends ToastOptions {
  id: string
}

interface ToastContextValue {
  show: (options: ToastOptions) => void
}

const ToastContext = createContext<ToastContextValue>({ show: () => {} })

export function useToast() {
  return useContext(ToastContext)
}

// ── Toast Manager ─────────────────────────────────────────────────────────────

let toastHandler: ((options: ToastOptions) => void) | null = null

export function showToast(options: ToastOptions) {
  if (toastHandler) {
    toastHandler(options)
  }
}

// ── Toast Item Component ───────────────────────────────────────────────────────

interface ToastItemProps {
  toast: ToastItem
  onDismiss: (id: string) => void
}

function ToastItemComponent({ toast, onDismiss }: ToastItemProps) {
  const layout = useResponsiveLayout()
  const opacity = useRef(new Animated.Value(0)).current
  const translateY = useRef(new Animated.Value(-20)).current

  useEffect(() => {
    // Animate in
    Animated.parallel([
      Animated.timing(opacity, { toValue: 1, duration: 200, useNativeDriver: true }),
      Animated.timing(translateY, { toValue: 0, duration: 200, useNativeDriver: true }),
    ]).start()

    // Auto dismiss
    const timer = setTimeout(() => {
      Animated.parallel([
        Animated.timing(opacity, { toValue: 0, duration: 200, useNativeDriver: true }),
        Animated.timing(translateY, { toValue: -20, duration: 200, useNativeDriver: true }),
      ]).start(() => onDismiss(toast.id))
    }, toast.duration || 3000)

    return () => clearTimeout(timer)
  }, [toast, onDismiss, opacity, translateY])

  const colors = {
    success: { bg: '#1a4d1a', border: '#22c55e', text: '#22c55e' },
    error: { bg: '#4d1a1a', border: '#ef4444', text: '#ef4444' },
    info: { bg: '#1a2a4d', border: '#3b82f6', text: '#3b82f6' },
  }

  const color = colors[toast.type || 'info']

  return (
    <Animated.View
      style={[
        styles.toastItem,
        {
          backgroundColor: color.bg,
          borderColor: color.border,
          opacity,
          transform: [{ translateY }],
          maxWidth: layout.device !== 'phone' ? 500 : 400,
          paddingVertical: layout.largeControls ? 18 : 14,
        },
      ]}
    >
      <Text style={[styles.toastText, { color: color.text, fontSize: scaledFont(14, layout.fontScale) }]}>{toast.message}</Text>
      <TouchableOpacity onPress={() => onDismiss(toast.id)} style={styles.toastClose}>
        <X size={scaledFont(14, layout.fontScale)} color="#888" />
      </TouchableOpacity>
    </Animated.View>
  )
}

// ── Toast Container ───────────────────────────────────────────────────────────

export function ToastContainer() {
  const [toasts, setToasts] = useState<ToastItem[]>([])

  useEffect(() => {
    toastHandler = (options: ToastOptions) => {
      const id = Date.now().toString() + Math.random().toString(36).substr(2, 9)
      setToasts((prev) => [...prev, { ...options, id }])
    }
    return () => {
      toastHandler = null
    }
  }, [])

  const handleDismiss = useCallback((id: string) => {
    setToasts((prev) => prev.filter((t) => t.id !== id))
  }, [])

  if (toasts.length === 0) return null

  return (
    <Animated.View style={styles.container} pointerEvents="box-none">
      {toasts.map((toast) => (
        <ToastItemComponent key={toast.id} toast={toast} onDismiss={handleDismiss} />
      ))}
    </Animated.View>
  )
}

// ── Helper function for quick use ──────────────────────────────────────────────

export const Toast = {
  show: showToast,
  success: (message: string) => showToast({ message, type: 'success' }),
  error: (message: string) => showToast({ message, type: 'error' }),
  info: (message: string) => showToast({ message, type: 'info' }),
}

// ── Styles ────────────────────────────────────────────────────────────────────

const styles = StyleSheet.create({
  container: {
    position: 'absolute',
    top: 60,
    left: 20,
    right: 20,
    zIndex: 9999,
    alignItems: 'center',
  },
  toastItem: {
    flexDirection: 'row',
    alignItems: 'center',
    paddingHorizontal: 16,
    paddingVertical: 14,
    borderRadius: 12,
    borderWidth: 1,
    marginBottom: 8,
    maxWidth: 400,
  },
  toastText: {
    flex: 1,
    fontSize: 14,
    fontWeight: '500',
  },
  toastClose: {
    marginLeft: 12,
    padding: 4,
  },
})