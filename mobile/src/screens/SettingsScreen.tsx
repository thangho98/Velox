import { useState, useEffect } from 'react'
import {
  View,
  Text,
  StyleSheet,
  TouchableOpacity,
  ScrollView,
  Switch,
  Alert,
  Modal,
  TextInput,
} from 'react-native'
import { useNavigation } from '@react-navigation/native'
import { useAuthStore } from '../stores'
import {
  useLogout,
  useProfile,
  useUpdateProfile,
  usePreferences,
  useUpdatePreferences,
  useChangePassword,
  useSessions,
  useRevokeSession,
  useServerInfo,
} from '@velox/shared/hooks'
import { CinemaSection, TranslateSection } from '../components/admin'
import { useResponsiveLayout, scaledFont, scaledSpacing } from '../lib/responsive'

// =============================================================================
// Tab Types
// =============================================================================

type TabType =
  | 'profile'
  | 'preferences'
  | 'security'
  | 'sessions'
  | 'server'
  | 'cinema'
  | 'subtitles'

// =============================================================================
// Settings Screen (Tab-based)
// =============================================================================

export function SettingsScreen() {
  const navigation = useNavigation()
  const { logout, user } = useAuthStore()
  const [activeTab, setActiveTab] = useState<TabType>('profile')
  const layout = useResponsiveLayout()

  const useSidebarLayout = layout.sideBySideDetail

  const tabs: { key: TabType; label: string }[] = [
    { key: 'profile', label: 'Profile' },
    { key: 'preferences', label: 'Preferences' },
    { key: 'security', label: 'Security' },
    { key: 'sessions', label: 'Sessions' },
    ...(user?.is_admin
      ? [
          { key: 'server' as TabType, label: 'Server' },
          { key: 'cinema' as TabType, label: 'Cinema' },
          { key: 'subtitles' as TabType, label: 'Subtitles' },
        ]
      : []),
  ]

  return (
    <View style={styles.container}>
      {/* Header */}
      <View style={[styles.header, { paddingHorizontal: layout.screenPadding }]}>
        <Text style={[styles.headerTitle, { fontSize: scaledFont(28, layout.fontScale) }]}>
          Settings
        </Text>
        <TouchableOpacity onPress={logout} style={styles.logoutHeaderBtn}>
          <Text style={[styles.logoutHeaderText, { fontSize: scaledFont(14, layout.fontScale) }]}>
            Logout
          </Text>
        </TouchableOpacity>
      </View>

      {useSidebarLayout ? (
        // Tablet landscape: sidebar + content
        <View style={styles.sidebarLayout}>
          {/* Left sidebar */}
          <View style={styles.sidebar}>
            {tabs.map((tab) => (
              <TouchableOpacity
                key={tab.key}
                style={[
                  styles.sidebarItem,
                  activeTab === tab.key && styles.sidebarItemActive,
                ]}
                onPress={() => setActiveTab(tab.key)}
              >
                <Text
                  style={[
                    styles.sidebarItemText,
                    { fontSize: scaledFont(15, layout.fontScale) },
                    activeTab === tab.key && styles.sidebarItemTextActive,
                  ]}
                >
                  {tab.label}
                </Text>
                {activeTab === tab.key && <View style={styles.sidebarIndicator} />}
              </TouchableOpacity>
            ))}
          </View>

          {/* Right content */}
          <ScrollView style={styles.sidebarContent} showsVerticalScrollIndicator={false}>
            {activeTab === 'profile' && <ProfileTab fontScale={layout.fontScale} />}
            {activeTab === 'preferences' && <PreferencesTab fontScale={layout.fontScale} />}
            {activeTab === 'security' && <SecurityTab fontScale={layout.fontScale} />}
            {activeTab === 'sessions' && <SessionsTab fontScale={layout.fontScale} />}
            {activeTab === 'server' && <ServerTab fontScale={layout.fontScale} />}
            {activeTab === 'cinema' && <CinemaTab />}
            {activeTab === 'subtitles' && <SubtitlesTab />}
          </ScrollView>
        </View>
      ) : (
        // Phone: horizontal tab bar + content
        <>
          {/* Tab Bar */}
          <ScrollView
            horizontal
            showsHorizontalScrollIndicator={false}
            style={styles.tabBar}
            contentContainerStyle={styles.tabBarContent}
          >
            {tabs.map((tab) => (
              <TouchableOpacity
                key={tab.key}
                style={[styles.tab, activeTab === tab.key && styles.tabActive]}
                onPress={() => setActiveTab(tab.key)}
              >
                <Text
                  style={[
                    styles.tabText,
                    { fontSize: scaledFont(14, layout.fontScale) },
                    activeTab === tab.key && styles.tabTextActive,
                  ]}
                >
                  {tab.label}
                </Text>
                {activeTab === tab.key && <View style={styles.tabIndicator} />}
              </TouchableOpacity>
            ))}
          </ScrollView>

          {/* Tab Content */}
          <ScrollView style={styles.content} showsVerticalScrollIndicator={false}>
            {activeTab === 'profile' && <ProfileTab fontScale={layout.fontScale} />}
            {activeTab === 'preferences' && <PreferencesTab fontScale={layout.fontScale} />}
            {activeTab === 'security' && <SecurityTab fontScale={layout.fontScale} />}
            {activeTab === 'sessions' && <SessionsTab fontScale={layout.fontScale} />}
            {activeTab === 'server' && <ServerTab fontScale={layout.fontScale} />}
            {activeTab === 'cinema' && <CinemaTab />}
            {activeTab === 'subtitles' && <SubtitlesTab />}
          </ScrollView>
        </>
      )}
    </View>
  )
}

// =============================================================================
// Profile Tab
// =============================================================================

function ProfileTab({ fontScale = 1.0 }: { fontScale?: number }) {
  const { data: profile, isLoading } = useProfile()
  const { mutate: updateProfile, isPending } = useUpdateProfile()
  const [displayName, setDisplayName] = useState('')
  const [success, setSuccess] = useState('')

  // Initialize display name when profile loads
  useEffect(() => {
    if (profile?.display_name) {
      setDisplayName(profile.display_name)
    }
  }, [profile?.display_name])

  const handleSave = () => {
    setSuccess('')
    updateProfile(
      { display_name: displayName },
      {
        onSuccess: () => {
          setSuccess('Profile updated successfully')
        },
      },
    )
  }

  if (isLoading) {
    return (
      <View style={styles.loadingContainer}>
        <Text style={styles.loadingText}>Loading...</Text>
      </View>
    )
  }

  return (
    <View style={styles.tabContent}>
      {success ? (
        <View style={styles.successBanner}>
          <Text style={[styles.successText, { fontSize: scaledFont(14, fontScale) }]}>
            {success}
          </Text>
        </View>
      ) : null}

      <View style={styles.fieldGroup}>
        <Text style={[styles.fieldLabel, { fontSize: scaledFont(12, fontScale) }]}>Username</Text>
        <TextInput
          style={[styles.input, styles.inputDisabled, { fontSize: scaledFont(16, fontScale) }]}
          value={profile?.username || ''}
          editable={false}
          placeholderTextColor="#666"
        />
        <Text style={[styles.fieldHint, { fontSize: scaledFont(12, fontScale) }]}>
          Username cannot be changed
        </Text>
      </View>

      <View style={styles.fieldGroup}>
        <Text style={[styles.fieldLabel, { fontSize: scaledFont(12, fontScale) }]}>
          Display Name
        </Text>
        <TextInput
          style={[styles.input, { fontSize: scaledFont(16, fontScale) }]}
          value={displayName}
          onChangeText={setDisplayName}
          placeholder="Enter display name"
          placeholderTextColor="#666"
        />
      </View>

      <View style={styles.fieldGroup}>
        <Text style={[styles.fieldLabel, { fontSize: scaledFont(12, fontScale) }]}>Role</Text>
        <View style={styles.roleBadgeContainer}>
          <View
            style={[
              styles.roleBadge,
              profile?.is_admin ? styles.roleBadgeAdmin : styles.roleBadgeUser,
            ]}
          >
            <Text
              style={[
                styles.roleBadgeText,
                { fontSize: scaledFont(14, fontScale) },
                profile?.is_admin ? styles.roleBadgeTextAdmin : styles.roleBadgeTextUser,
              ]}
            >
              {profile?.is_admin ? 'Administrator' : 'User'}
            </Text>
          </View>
        </View>
      </View>

      <TouchableOpacity
        style={[styles.saveButton, isPending && styles.saveButtonDisabled]}
        onPress={handleSave}
        disabled={isPending}
      >
        <Text style={[styles.saveButtonText, { fontSize: scaledFont(16, fontScale) }]}>
          {isPending ? 'Saving...' : 'Save Changes'}
        </Text>
      </TouchableOpacity>
    </View>
  )
}

// =============================================================================
// Preferences Tab
// =============================================================================

function PreferencesTab({ fontScale = 1.0 }: { fontScale?: number }) {
  const { data: preferences, isLoading } = usePreferences()
  const { mutate: updatePreferences, isPending } = useUpdatePreferences()
  const [prefs, setPrefs] = useState({
    subtitle_language: '',
    audio_language: '',
    max_streaming_quality: 'original',
  })

  useEffect(() => {
    if (preferences) {
      setPrefs({
        subtitle_language: preferences.subtitle_language || '',
        audio_language: preferences.audio_language || '',
        max_streaming_quality: preferences.max_streaming_quality || 'original',
      })
    }
  }, [preferences])

  const handleSave = () => {
    updatePreferences({
      user_id: preferences?.user_id || 0,
      subtitle_language: prefs.subtitle_language,
      audio_language: prefs.audio_language,
      max_streaming_quality: prefs.max_streaming_quality,
      theme: 'dark', // mobile always dark
      language: preferences?.language || 'en',
    })
  }

  if (isLoading) {
    return (
      <View style={styles.loadingContainer}>
        <Text style={styles.loadingText}>Loading...</Text>
      </View>
    )
  }

  return (
    <View style={styles.tabContent}>
      <View style={styles.fieldGroup}>
        <Text style={[styles.fieldLabel, { fontSize: scaledFont(12, fontScale) }]}>
          Subtitle Language
        </Text>
        <TouchableOpacity
          style={styles.selectButton}
          onPress={() =>
            showPickerModal('subtitle', prefs.subtitle_language, (val) =>
              setPrefs({ ...prefs, subtitle_language: val }),
            )
          }
        >
          <Text style={[styles.selectButtonText, { fontSize: scaledFont(16, fontScale) }]}>
            {prefs.subtitle_language
              ? languageLabels[prefs.subtitle_language] || prefs.subtitle_language
              : 'Auto'}
          </Text>
          <Text style={styles.selectArrow}>▼</Text>
        </TouchableOpacity>
      </View>

      <View style={styles.fieldGroup}>
        <Text style={[styles.fieldLabel, { fontSize: scaledFont(12, fontScale) }]}>
          Audio Language
        </Text>
        <TouchableOpacity
          style={styles.selectButton}
          onPress={() =>
            showPickerModal('audio', prefs.audio_language, (val) =>
              setPrefs({ ...prefs, audio_language: val }),
            )
          }
        >
          <Text style={[styles.selectButtonText, { fontSize: scaledFont(16, fontScale) }]}>
            {prefs.audio_language
              ? languageLabels[prefs.audio_language] || prefs.audio_language
              : 'Auto'}
          </Text>
          <Text style={styles.selectArrow}>▼</Text>
        </TouchableOpacity>
      </View>

      <View style={styles.fieldGroup}>
        <Text style={[styles.fieldLabel, { fontSize: scaledFont(12, fontScale) }]}>
          Max Streaming Quality
        </Text>
        <TouchableOpacity
          style={styles.selectButton}
          onPress={() =>
            showPickerModal('quality', prefs.max_streaming_quality, (val) =>
              setPrefs({ ...prefs, max_streaming_quality: val }),
            )
          }
        >
          <Text style={[styles.selectButtonText, { fontSize: scaledFont(16, fontScale) }]}>
            {qualityLabels[prefs.max_streaming_quality] || prefs.max_streaming_quality}
          </Text>
          <Text style={styles.selectArrow}>▼</Text>
        </TouchableOpacity>
      </View>

      <TouchableOpacity
        style={[styles.saveButton, isPending && styles.saveButtonDisabled]}
        onPress={handleSave}
        disabled={isPending}
      >
        <Text style={[styles.saveButtonText, { fontSize: scaledFont(16, fontScale) }]}>
          {isPending ? 'Saving...' : 'Save Preferences'}
        </Text>
      </TouchableOpacity>
    </View>
  )
}

// =============================================================================
// Security Tab
// =============================================================================

function SecurityTab({ fontScale = 1.0 }: { fontScale?: number }) {
  const { mutate: changePassword, isPending } = useChangePassword()
  const [oldPassword, setOldPassword] = useState('')
  const [newPassword, setNewPassword] = useState('')
  const [confirmPassword, setConfirmPassword] = useState('')
  const [error, setError] = useState('')
  const [success, setSuccess] = useState('')

  const handleSubmit = () => {
    setError('')
    setSuccess('')

    if (newPassword !== confirmPassword) {
      setError('Passwords do not match')
      return
    }

    if (newPassword.length < 8) {
      setError('Password must be at least 8 characters')
      return
    }

    changePassword(
      { old_password: oldPassword, new_password: newPassword },
      {
        onSuccess: () => {
          setSuccess('Password changed successfully')
          setOldPassword('')
          setNewPassword('')
          setConfirmPassword('')
        },
        onError: (err: any) => {
          setError(err?.message || 'Failed to change password')
        },
      },
    )
  }

  return (
    <View style={styles.tabContent}>
      {error ? (
        <View style={styles.errorBanner}>
          <Text style={[styles.errorText, { fontSize: scaledFont(14, fontScale) }]}>{error}</Text>
        </View>
      ) : null}
      {success ? (
        <View style={styles.successBanner}>
          <Text style={[styles.successText, { fontSize: scaledFont(14, fontScale) }]}>
            {success}
          </Text>
        </View>
      ) : null}

      <View style={styles.fieldGroup}>
        <Text style={[styles.fieldLabel, { fontSize: scaledFont(12, fontScale) }]}>
          Current Password
        </Text>
        <TextInput
          style={[styles.input, { fontSize: scaledFont(16, fontScale) }]}
          value={oldPassword}
          onChangeText={setOldPassword}
          placeholder="Enter current password"
          placeholderTextColor="#666"
          secureTextEntry
        />
      </View>

      <View style={styles.fieldGroup}>
        <Text style={[styles.fieldLabel, { fontSize: scaledFont(12, fontScale) }]}>
          New Password
        </Text>
        <TextInput
          style={[styles.input, { fontSize: scaledFont(16, fontScale) }]}
          value={newPassword}
          onChangeText={setNewPassword}
          placeholder="Enter new password"
          placeholderTextColor="#666"
          secureTextEntry
        />
      </View>

      <View style={styles.fieldGroup}>
        <Text style={[styles.fieldLabel, { fontSize: scaledFont(12, fontScale) }]}>
          Confirm Password
        </Text>
        <TextInput
          style={[styles.input, { fontSize: scaledFont(16, fontScale) }]}
          value={confirmPassword}
          onChangeText={setConfirmPassword}
          placeholder="Confirm new password"
          placeholderTextColor="#666"
          secureTextEntry
        />
      </View>

      <TouchableOpacity
        style={[styles.saveButton, isPending && styles.saveButtonDisabled]}
        onPress={handleSubmit}
        disabled={isPending}
      >
        <Text style={[styles.saveButtonText, { fontSize: scaledFont(16, fontScale) }]}>
          {isPending ? 'Changing...' : 'Change Password'}
        </Text>
      </TouchableOpacity>
    </View>
  )
}

// =============================================================================
// Sessions Tab
// =============================================================================

function SessionsTab({ fontScale = 1.0 }: { fontScale?: number }) {
  const { data: sessions, isLoading } = useSessions()
  const { mutate: revokeSession } = useRevokeSession()

  const handleRevoke = (sessionId: number) => {
    Alert.alert('Revoke Session', 'Are you sure you want to logout this session?', [
      { text: 'Cancel', style: 'cancel' },
      {
        text: 'Revoke',
        style: 'destructive',
        onPress: () => revokeSession(sessionId),
      },
    ])
  }

  if (isLoading) {
    return (
      <View style={styles.loadingContainer}>
        <Text style={styles.loadingText}>Loading...</Text>
      </View>
    )
  }

  return (
    <View style={styles.tabContent}>
      {sessions?.length === 0 ? (
        <View style={styles.emptyState}>
          <Text style={styles.emptyStateIcon}>💻</Text>
          <Text style={[styles.emptyStateText, { fontSize: scaledFont(16, fontScale) }]}>
            No active sessions
          </Text>
        </View>
      ) : (
        sessions?.map((session) => (
          <View key={session.id} style={styles.sessionItem}>
            <View style={styles.sessionIcon}>
              <Text>💻</Text>
            </View>
            <View style={styles.sessionInfo}>
              <Text
                style={[styles.sessionDevice, { fontSize: scaledFont(16, fontScale) }]}
              >
                {session.device_name || 'Unknown Device'}
              </Text>
              <Text style={[styles.sessionMeta, { fontSize: scaledFont(13, fontScale) }]}>
                {session.ip_address}
              </Text>
              <Text style={[styles.sessionMeta, { fontSize: scaledFont(13, fontScale) }]}>
                Last active: {new Date(session.last_active_at).toLocaleString()}
              </Text>
            </View>
            <TouchableOpacity
              style={styles.revokeButton}
              onPress={() => handleRevoke(session.id)}
            >
              <Text style={[styles.revokeButtonText, { fontSize: scaledFont(14, fontScale) }]}>
                Revoke
              </Text>
            </TouchableOpacity>
          </View>
        ))
      )}
    </View>
  )
}

// =============================================================================
// Server Tab (Admin only)
// =============================================================================

function ServerTab({ fontScale = 1.0 }: { fontScale?: number }) {
  const { data: info, isLoading } = useServerInfo()

  const formatBytes = (bytes: number) => {
    if (bytes === 0) return '0 B'
    const k = 1024
    const sizes = ['B', 'KB', 'MB', 'GB', 'TB']
    const i = Math.floor(Math.log(bytes) / Math.log(k))
    return `${(bytes / Math.pow(k, i)).toFixed(1)} ${sizes[i]}`
  }

  if (isLoading) {
    return (
      <View style={styles.tabContent}>
        <Text style={styles.loadingText}>Loading...</Text>
      </View>
    )
  }

  if (!info) {
    return (
      <View style={styles.tabContent}>
        <Text style={styles.emptyStateText}>Unable to load server info</Text>
      </View>
    )
  }

  const serverInfoRows = [
    { label: 'Version', value: info.version },
    { label: 'Uptime', value: info.uptime },
    { label: 'Go Version', value: info.go_version },
    { label: 'OS / Arch', value: `${info.os} / ${info.arch}` },
    { label: 'FFmpeg', value: info.ffmpeg_version },
    { label: 'Hardware Acceleration', value: info.hw_accel },
    { label: 'Database', value: info.database },
    { label: 'Media Files', value: info.media_count.toLocaleString() },
    { label: 'Series', value: info.series_count.toLocaleString() },
    { label: 'Users', value: info.user_count.toString() },
    { label: 'Total Storage', value: formatBytes(info.total_size_bytes) },
  ]

  return (
    <View style={styles.tabContent}>
      <View style={styles.serverInfoCard}>
        {serverInfoRows.map((row, index) => (
          <View
            key={row.label}
            style={[
              styles.serverInfoRow,
              index < serverInfoRows.length - 1 && styles.serverInfoRowBorder,
            ]}
          >
            <Text style={[styles.serverInfoLabel, { fontSize: scaledFont(14, fontScale) }]}>
              {row.label}
            </Text>
            <Text style={[styles.serverInfoValue, { fontSize: scaledFont(14, fontScale) }]}>
              {row.value}
            </Text>
          </View>
        ))}
      </View>
    </View>
  )
}

// =============================================================================
// Cinema Tab (Admin only)
// =============================================================================

function CinemaTab() {
  return (
    <View style={styles.tabContent}>
      <CinemaSection />
    </View>
  )
}

// =============================================================================
// Subtitles Tab (Admin only)
// =============================================================================

function SubtitlesTab() {
  return (
    <View style={styles.tabContent}>
      <TranslateSection />
    </View>
  )
}

// =============================================================================
// Picker Modal Helper
// =============================================================================

const languageLabels: Record<string, string> = {
  '': 'Auto',
  vi: 'Vietnamese',
  en: 'English',
}

const qualityLabels: Record<string, string> = {
  original: 'Original',
  '4k': '4K',
  '1080p': '1080p',
  '720p': '720p',
  '480p': '480p',
}

function showPickerModal(
  type: 'subtitle' | 'audio' | 'quality',
  currentValue: string,
  onSelect: (value: string) => void,
) {
  let options: string[] = []
  let labels: Record<string, string> = {}

  if (type === 'subtitle' || type === 'audio') {
    options = ['', 'vi', 'en']
    labels = languageLabels
  } else {
    options = ['original', '4k', '1080p', '720p', '480p']
    labels = qualityLabels
  }

  Alert.alert(
    type === 'subtitle'
      ? 'Subtitle Language'
      : type === 'audio'
        ? 'Audio Language'
        : 'Max Streaming Quality',
    undefined,
    [
      ...options.map((option) => ({
        text: labels[option] || option,
        onPress: () => onSelect(option),
      })),
      { text: 'Cancel', style: 'cancel' as const },
    ],
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
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
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
  logoutHeaderBtn: {
    paddingHorizontal: 12,
    paddingVertical: 6,
    backgroundColor: '#2a2a2a',
    borderRadius: 6,
  },
  logoutHeaderText: {
    color: '#e50914',
    fontSize: 14,
    fontWeight: '500',
  },
  // ── Sidebar layout (tablet landscape) ─────────────────────────────────────
  sidebarLayout: {
    flex: 1,
    flexDirection: 'row',
  },
  sidebar: {
    width: 200,
    backgroundColor: '#141414',
    borderRightWidth: 1,
    borderRightColor: '#2a2a2a',
    paddingTop: 8,
  },
  sidebarItem: {
    position: 'relative',
    paddingVertical: 14,
    paddingHorizontal: 20,
  },
  sidebarItemActive: {
    backgroundColor: 'rgba(229, 9, 20, 0.1)',
  },
  sidebarItemText: {
    color: '#888',
    fontSize: 15,
    fontWeight: '500',
  },
  sidebarItemTextActive: {
    color: '#fff',
  },
  sidebarIndicator: {
    position: 'absolute',
    left: 0,
    top: 8,
    bottom: 8,
    width: 3,
    backgroundColor: '#e50914',
    borderRadius: 2,
  },
  sidebarContent: {
    flex: 1,
  },
  // ── Tab bar (phone) ───────────────────────────────────────────────────────
  tabBar: {
    backgroundColor: '#141414',
    borderBottomWidth: 1,
    borderBottomColor: '#2a2a2a',
    flexGrow: 0,
  },
  tabBarContent: {
    flexDirection: 'row',
    paddingHorizontal: 4,
  },
  tab: {
    paddingVertical: 14,
    paddingHorizontal: 16,
    alignItems: 'center',
    position: 'relative',
  },
  tabActive: {},
  tabText: {
    color: '#888',
    fontSize: 14,
    fontWeight: '500',
    textTransform: 'capitalize',
  },
  tabTextActive: {
    color: '#fff',
  },
  tabIndicator: {
    position: 'absolute',
    bottom: 0,
    left: 20,
    right: 20,
    height: 2,
    backgroundColor: '#e50914',
    borderRadius: 1,
  },
  // Content
  content: {
    flex: 1,
  },
  tabContent: {
    padding: 20,
  },
  loadingContainer: {
    padding: 40,
    alignItems: 'center',
  },
  loadingText: {
    color: '#888',
    fontSize: 16,
  },
  // Field groups
  fieldGroup: {
    marginBottom: 24,
  },
  fieldLabel: {
    color: '#888',
    fontSize: 12,
    fontWeight: '600',
    textTransform: 'uppercase',
    marginBottom: 8,
    marginLeft: 4,
  },
  fieldHint: {
    color: '#555',
    fontSize: 12,
    marginTop: 6,
    marginLeft: 4,
  },
  input: {
    backgroundColor: '#1f1f1f',
    borderRadius: 8,
    paddingHorizontal: 16,
    paddingVertical: 14,
    color: '#fff',
    fontSize: 16,
    borderWidth: 1,
    borderColor: 'transparent',
  },
  inputDisabled: {
    color: '#888',
  },
  // Select button (like dropdown)
  selectButton: {
    backgroundColor: '#1f1f1f',
    borderRadius: 8,
    paddingHorizontal: 16,
    paddingVertical: 14,
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
  },
  selectButtonText: {
    color: '#fff',
    fontSize: 16,
  },
  selectArrow: {
    color: '#888',
    fontSize: 12,
  },
  // Role badge
  roleBadgeContainer: {
    flexDirection: 'row',
  },
  roleBadge: {
    paddingHorizontal: 12,
    paddingVertical: 6,
    borderRadius: 16,
  },
  roleBadgeAdmin: {
    backgroundColor: 'rgba(168, 85, 247, 0.2)',
  },
  roleBadgeUser: {
    backgroundColor: 'rgba(59, 130, 246, 0.2)',
  },
  roleBadgeText: {
    fontSize: 14,
    fontWeight: '500',
  },
  roleBadgeTextAdmin: {
    color: '#a855f7',
  },
  roleBadgeTextUser: {
    color: '#3b82f6',
  },
  // Buttons
  saveButton: {
    backgroundColor: '#e50914',
    borderRadius: 8,
    paddingVertical: 16,
    alignItems: 'center',
    marginTop: 16,
  },
  saveButtonDisabled: {
    opacity: 0.5,
  },
  saveButtonText: {
    color: '#fff',
    fontSize: 16,
    fontWeight: '600',
  },
  // Banners
  successBanner: {
    backgroundColor: 'rgba(34, 197, 94, 0.2)',
    padding: 12,
    borderRadius: 8,
    marginBottom: 16,
  },
  successText: {
    color: '#22c55e',
    fontSize: 14,
  },
  errorBanner: {
    backgroundColor: 'rgba(239, 68, 68, 0.2)',
    padding: 12,
    borderRadius: 8,
    marginBottom: 16,
  },
  errorText: {
    color: '#ef4444',
    fontSize: 14,
  },
  // Session list
  sessionItem: {
    backgroundColor: '#1f1f1f',
    borderRadius: 12,
    padding: 16,
    flexDirection: 'row',
    alignItems: 'center',
    marginBottom: 12,
  },
  sessionIcon: {
    width: 44,
    height: 44,
    borderRadius: 22,
    backgroundColor: '#2a2a2a',
    justifyContent: 'center',
    alignItems: 'center',
    marginRight: 14,
  },
  sessionInfo: {
    flex: 1,
  },
  sessionDevice: {
    color: '#fff',
    fontSize: 16,
    fontWeight: '500',
    marginBottom: 2,
  },
  sessionMeta: {
    color: '#888',
    fontSize: 13,
    marginTop: 2,
  },
  revokeButton: {
    backgroundColor: '#2a2a2a',
    paddingHorizontal: 14,
    paddingVertical: 8,
    borderRadius: 6,
  },
  revokeButtonText: {
    color: '#fff',
    fontSize: 14,
  },
  // Empty state
  emptyState: {
    padding: 40,
    alignItems: 'center',
  },
  emptyStateIcon: {
    fontSize: 48,
    marginBottom: 12,
  },
  emptyStateText: {
    color: '#888',
    fontSize: 16,
  },
  // Server info card
  serverInfoCard: {
    backgroundColor: '#1f1f1f',
    borderRadius: 12,
    overflow: 'hidden',
  },
  serverInfoRow: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
    paddingVertical: 14,
    paddingHorizontal: 16,
  },
  serverInfoRowBorder: {
    borderBottomWidth: 1,
    borderBottomColor: '#2a2a2a',
  },
  serverInfoLabel: {
    color: '#888',
    fontSize: 14,
  },
  serverInfoValue: {
    color: '#fff',
    fontSize: 14,
    fontWeight: '500',
    textAlign: 'right',
    flex: 1,
    marginLeft: 16,
  },
})
