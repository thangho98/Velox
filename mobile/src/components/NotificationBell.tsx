/**
 * Notification Bell - Mobile version
 */

import { useState, useRef } from 'react'
import {
  View,
  Text,
  StyleSheet,
  TouchableOpacity,
  Modal,
  FlatList,
  Pressable,
} from 'react-native'
import { Search, Film, Eye, Check, Trash2, X, Bell } from 'lucide-react-native'
import { useResponsiveLayout, scaledFont } from '../lib/responsive'
import { useNavigation } from '@react-navigation/native'
import type { NativeStackNavigationProp } from '@react-navigation/native-stack'
import {
  useNotifications,
  useUnreadCount,
  useMarkNotificationsAsRead,
  useMarkAllNotificationsAsRead,
  useDeleteNotifications,
  type Notification,
} from '@velox/shared/hooks/media/useNotifications'
import type { RootStackParamList } from '../../App'

type NavigationProp = NativeStackNavigationProp<RootStackParamList>

const NOTIFICATION_ICONS: Record<string, React.ReactNode> = {
  scan_complete: <Search size={20} color="#fff" />,
  media_added: <Film size={20} color="#fff" />,
  transcode_complete: <Check size={20} color="#22c55e" />,
  transcode_failed: <X size={20} color="#ef4444" />,
  subtitle_downloaded: <Text style={{ fontSize: 20 }}>📝</Text>,
  identify_complete: <Text style={{ fontSize: 20 }}>🆔</Text>,
  library_watcher: <Eye size={20} color="#fff" />,
}

export function NotificationBell() {
  const navigation = useNavigation<NavigationProp>()
  const layout = useResponsiveLayout()
  const isWideScreen = layout.device !== 'phone'
  const [visible, setVisible] = useState(false)
  const { data: unreadData } = useUnreadCount()
  const { data: notificationsData } = useNotifications(false, 20, 0)
  const { mutate: markAsRead } = useMarkNotificationsAsRead()
  const { mutate: markAllAsRead } = useMarkAllNotificationsAsRead()
  const { mutate: deleteNotification } = useDeleteNotifications()

  const unreadCount = unreadData?.count ?? 0
  const notifications = notificationsData?.notifications ?? []

  const handleNotificationPress = (n: Notification) => {
    if (!n.read) {
      markAsRead([n.id])
    }

    if (n.data?.media_id) {
      navigation.navigate('Media', { id: n.data.media_id })
    } else if (n.data?.series_id) {
      navigation.navigate('SeriesDetail', { id: n.data.series_id })
    } else if (n.data?.library_id) {
      // Navigate to browse
    }

    setVisible(false)
  }

  const handleMarkAllAsRead = () => {
    markAllAsRead()
  }

  const handleDelete = (id: number) => {
    deleteNotification([id])
  }

  const formatTime = (isoString: string): string => {
    const date = new Date(isoString)
    const now = new Date()
    const diff = now.getTime() - date.getTime()

    if (diff < 60000) return 'Just now'
    if (diff < 3600000) return `${Math.floor(diff / 60000)}m ago`
    if (diff < 86400000) return `${Math.floor(diff / 3600000)}h ago`
    if (diff < 604800000) return `${Math.floor(diff / 86400000)}d ago`
    return date.toLocaleDateString()
  }

  const renderNotification = ({ item }: { item: Notification }) => (
    <Pressable
      style={[styles.notificationItem, !item.read && styles.notificationUnread]}
      onPress={() => handleNotificationPress(item)}
    >
      <View style={[styles.notificationIcon, layout.largeControls && { width: 52, height: 52, borderRadius: 26 }]}>
        {NOTIFICATION_ICONS[item.type] || <Bell size={scaledFont(20, layout.fontScale)} color="#fff" />}
      </View>
      <View style={styles.notificationContent}>
        <Text style={[styles.notificationTitle, !item.read && styles.titleUnread, { fontSize: scaledFont(14, layout.fontScale) }]}>
          {item.title}
        </Text>
        <Text style={[styles.notificationMessage, { fontSize: scaledFont(13, layout.fontScale) }]} numberOfLines={2}>
          {item.message}
        </Text>
        <Text style={[styles.notificationTime, { fontSize: scaledFont(12, layout.fontScale) }]}>{formatTime(item.created_at)}</Text>
      </View>
      <View style={styles.notificationActions}>
        {!item.read && (
          <TouchableOpacity
            style={[styles.actionButton, layout.largeControls && { width: 40, height: 40, borderRadius: 20 }]}
            onPress={() => markAsRead([item.id])}
          >
            <Check size={scaledFont(14, layout.fontScale)} color="#fff" />
          </TouchableOpacity>
        )}
        <TouchableOpacity
          style={[styles.actionButton, layout.largeControls && { width: 40, height: 40, borderRadius: 20 }]}
          onPress={() => handleDelete(item.id)}
        >
          <Trash2 size={scaledFont(14, layout.fontScale)} color="#fff" />
        </TouchableOpacity>
      </View>
    </Pressable>
  )

  const bellSize = layout.largeControls ? 48 : layout.device === 'tablet' ? 42 : 36
  const bellIconSize = layout.largeControls ? 24 : layout.device === 'tablet' ? 22 : 18
  const badgeSize = layout.largeControls ? 24 : layout.device === 'tablet' ? 22 : 18

  return (
    <>
      <View style={styles.bellWrapper}>
        <TouchableOpacity
          style={styles.bellButton}
          onPress={() => setVisible(true)}
        >
          <Bell size={18} color="#fff" />
        </TouchableOpacity>
        {unreadCount && unreadCount > 0 ? (
          <View style={styles.badge} pointerEvents="none">
            <Text style={styles.badgeText}>{unreadCount > 9 ? '9+' : unreadCount}</Text>
          </View>
        ) : null}
      </View>

      <Modal
        visible={visible}
        animationType={isWideScreen ? 'fade' : 'slide'}
        transparent={true}
        onRequestClose={() => setVisible(false)}
      >
        <View style={[styles.modalOverlay, isWideScreen && { justifyContent: 'center', alignItems: 'center' }]}>
          <View style={[
            styles.modalContent,
            isWideScreen && {
              maxWidth: layout.largeControls ? 600 : 500,
              width: '90%',
              borderRadius: 16,
              maxHeight: '70%',
            },
          ]}>
            {/* Header */}
            <View style={styles.modalHeader}>
              <Text style={[styles.modalTitle, { fontSize: scaledFont(18, layout.fontScale) }]}>Notifications</Text>
              <View style={styles.headerActions}>
                {unreadCount && unreadCount > 0 && (
                  <TouchableOpacity onPress={handleMarkAllAsRead}>
                    <Text style={styles.markAllRead}>Mark all read</Text>
                  </TouchableOpacity>
                )}
                <TouchableOpacity
                  style={styles.closeButton}
                  onPress={() => setVisible(false)}
                >
                  <X size={14} color="#fff" />
                </TouchableOpacity>
              </View>
            </View>

            {/* Notification List */}
            {notifications.length === 0 ? (
              <View style={styles.emptyState}>
                <Bell size={48} color="#888" />
                <Text style={styles.emptyText}>No notifications yet</Text>
              </View>
            ) : (
              <FlatList
                data={notifications}
                renderItem={renderNotification}
                keyExtractor={(item) => String(item.id)}
                style={styles.list}
                showsVerticalScrollIndicator={false}
              />
            )}
          </View>
        </View>
      </Modal>
    </>
  )
}

const styles = StyleSheet.create({
  bellWrapper: {
    width: 36,
    height: 36,
  },
  bellButton: {
    width: 36,
    height: 36,
    borderRadius: 18,
    backgroundColor: 'rgba(255, 255, 255, 0.1)',
    justifyContent: 'center',
    alignItems: 'center',
  },
  badge: {
    position: 'absolute',
    top: -6,
    right: -8,
    minWidth: 18,
    height: 18,
    borderRadius: 9,
    backgroundColor: '#e50914',
    justifyContent: 'center',
    alignItems: 'center',
    paddingHorizontal: 4,
  },
  badgeText: {
    color: '#fff',
    fontSize: 10,
    fontWeight: 'bold',
  },
  modalOverlay: {
    flex: 1,
    backgroundColor: 'rgba(0, 0, 0, 0.7)',
    justifyContent: 'flex-end',
  },
  modalContent: {
    backgroundColor: '#1a1a1a',
    borderTopLeftRadius: 20,
    borderTopRightRadius: 20,
    maxHeight: '80%',
    minHeight: '50%',
  },
  modalHeader: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
    padding: 16,
    borderBottomWidth: 1,
    borderBottomColor: 'rgba(255, 255, 255, 0.1)',
  },
  modalTitle: {
    color: '#fff',
    fontSize: 18,
    fontWeight: 'bold',
  },
  headerActions: {
    flexDirection: 'row',
    alignItems: 'center',
    gap: 16,
  },
  markAllRead: {
    color: '#e50914',
    fontSize: 14,
  },
  closeButton: {
    width: 28,
    height: 28,
    borderRadius: 14,
    backgroundColor: 'rgba(255, 255, 255, 0.1)',
    justifyContent: 'center',
    alignItems: 'center',
  },
  list: {
    flex: 1,
  },
  notificationItem: {
    flexDirection: 'row',
    padding: 16,
    borderBottomWidth: 1,
    borderBottomColor: 'rgba(255, 255, 255, 0.05)',
    alignItems: 'flex-start',
  },
  notificationUnread: {
    backgroundColor: 'rgba(229, 9, 20, 0.05)',
  },
  notificationIcon: {
    width: 40,
    height: 40,
    borderRadius: 20,
    backgroundColor: 'rgba(255, 255, 255, 0.1)',
    justifyContent: 'center',
    alignItems: 'center',
    marginRight: 12,
  },
  notificationContent: {
    flex: 1,
  },
  notificationTitle: {
    color: '#ccc',
    fontSize: 14,
    fontWeight: '500',
  },
  titleUnread: {
    color: '#fff',
    fontWeight: '600',
  },
  notificationMessage: {
    color: '#888',
    fontSize: 13,
    marginTop: 4,
    lineHeight: 18,
  },
  notificationTime: {
    color: '#666',
    fontSize: 12,
    marginTop: 6,
  },
  notificationActions: {
    flexDirection: 'row',
    gap: 8,
  },
  actionButton: {
    width: 32,
    height: 32,
    borderRadius: 16,
    backgroundColor: 'rgba(255, 255, 255, 0.1)',
    justifyContent: 'center',
    alignItems: 'center',
  },
  emptyState: {
    flex: 1,
    justifyContent: 'center',
    alignItems: 'center',
    paddingVertical: 60,
  },
  emptyText: {
    color: '#888',
    fontSize: 16,
  },
})
