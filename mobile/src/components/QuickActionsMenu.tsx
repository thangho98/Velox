/**
 * QuickActionsMenu - Bottom sheet with quick actions triggered by long press
 */

import { useState } from 'react'
import {
  View,
  Text,
  StyleSheet,
  Modal,
  TouchableOpacity,
  ScrollView,
} from 'react-native'
import { X } from 'lucide-react-native'
import { useResponsiveLayout, scaledFont } from '../lib/responsive'

// ── Types ─────────────────────────────────────────────────────────────────────

export interface QuickAction {
  id: string
  label: string
  icon: React.ReactNode
  destructive?: boolean
  disabled?: boolean
}

interface QuickActionsMenuProps {
  visible: boolean
  title: string
  subtitle?: string
  actions: QuickAction[]
  onAction: (actionId: string) => void
  onClose: () => void
}

// ── Component ─────────────────────────────────────────────────────────────────

export function QuickActionsMenu({
  visible,
  title,
  subtitle,
  actions,
  onAction,
  onClose,
}: QuickActionsMenuProps) {
  const layout = useResponsiveLayout()
  const isWideScreen = layout.device !== 'phone'

  return (
    <Modal
      visible={visible}
      transparent
      animationType={isWideScreen ? 'fade' : 'slide'}
      onRequestClose={onClose}
    >
      <TouchableOpacity style={[styles.backdrop, isWideScreen && { justifyContent: 'center', alignItems: 'center' }]} activeOpacity={1} onPress={onClose} />
      <View style={[
        styles.sheet,
        isWideScreen && {
          maxWidth: layout.largeControls ? 600 : 500,
          alignSelf: 'center',
          width: '90%',
          borderRadius: 16,
          position: 'absolute',
          top: layout.height * 0.2,
        },
      ]}>
        {/* Header */}
        <View style={styles.header}>
          <View style={styles.titleContainer}>
            <Text style={[styles.headerTitle, { fontSize: scaledFont(18, layout.fontScale) }]} numberOfLines={1}>{title}</Text>
            {subtitle && (
              <Text style={[styles.headerSubtitle, { fontSize: scaledFont(13, layout.fontScale) }]} numberOfLines={1}>{subtitle}</Text>
            )}
          </View>
          <TouchableOpacity style={styles.closeButton} onPress={onClose}>
            <X size={16} color="#888" />
          </TouchableOpacity>
        </View>

        {/* Actions */}
        <ScrollView style={styles.content} showsVerticalScrollIndicator={false}>
          {actions.map((action) => (
            <TouchableOpacity
              key={action.id}
              style={[
                styles.actionItem,
                action.destructive && styles.actionItemDestructive,
                action.disabled && styles.actionItemDisabled,
                { paddingVertical: layout.largeControls ? 18 : 14 },
              ]}
              onPress={() => {
                if (!action.disabled) {
                  onAction(action.id)
                  onClose()
                }
              }}
              disabled={action.disabled}
            >
              <View style={styles.actionIcon}>{action.icon}</View>
              <Text
                style={[
                  styles.actionLabel,
                  action.destructive && styles.actionLabelDestructive,
                  action.disabled && styles.actionLabelDisabled,
                  { fontSize: scaledFont(16, layout.fontScale) },
                ]}
              >
                {action.label}
              </Text>
            </TouchableOpacity>
          ))}
        </ScrollView>
      </View>
    </Modal>
  )
}

// ── Styles ─────────────────────────────────────────────────────────────────────

const styles = StyleSheet.create({
  backdrop: {
    flex: 1,
    backgroundColor: 'rgba(0, 0, 0, 0.6)',
  },
  sheet: {
    backgroundColor: '#141414',
    borderTopLeftRadius: 20,
    borderTopRightRadius: 20,
    maxHeight: '60%',
  },
  header: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'flex-start',
    padding: 16,
    borderBottomWidth: 1,
    borderBottomColor: '#2a2a2a',
  },
  titleContainer: {
    flex: 1,
    marginRight: 12,
  },
  headerTitle: {
    color: '#fff',
    fontSize: 18,
    fontWeight: '600',
  },
  headerSubtitle: {
    color: '#888',
    fontSize: 13,
    marginTop: 2,
  },
  closeButton: {
    width: 32,
    height: 32,
    borderRadius: 16,
    backgroundColor: '#2a2a2a',
    justifyContent: 'center',
    alignItems: 'center',
  },
  content: {
    padding: 12,
  },
  actionItem: {
    flexDirection: 'row',
    alignItems: 'center',
    paddingVertical: 14,
    paddingHorizontal: 16,
    backgroundColor: '#1f1f1f',
    borderRadius: 10,
    marginBottom: 8,
    gap: 14,
  },
  actionItemDestructive: {
    backgroundColor: 'rgba(229, 9, 20, 0.1)',
  },
  actionItemDisabled: {
    opacity: 0.5,
  },
  actionIcon: {
    width: 20,
    height: 20,
    justifyContent: 'center' as const,
    alignItems: 'center' as const,
  },
  actionLabel: {
    color: '#fff',
    fontSize: 16,
  },
  actionLabelDestructive: {
    color: '#e50914',
  },
  actionLabelDisabled: {
    color: '#666',
  },
})
