import { useState } from 'react'
import {
  useProfile,
  useUpdateProfile,
  usePreferences,
  useUpdatePreferences,
  useSessions,
  useRevokeSession,
  useChangePassword,
} from '@/hooks/stores/useAuth'
import { useAuthStore } from '@/stores/auth'
import { useUIStore } from '@/stores/ui'

export function ProfilePage() {
  const { logout } = useAuthStore()

  const [activeTab, setActiveTab] = useState<'profile' | 'preferences' | 'security' | 'sessions'>(
    'profile',
  )

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold text-white">Profile Settings</h1>
          <p className="text-gray-400">Manage your account and preferences</p>
        </div>
        <button
          onClick={logout}
          className="flex items-center gap-2 rounded bg-netflix-gray px-4 py-2 text-white transition-colors hover:bg-red-600"
        >
          <svg className="h-4 w-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path
              strokeLinecap="round"
              strokeLinejoin="round"
              strokeWidth={2}
              d="M17 16l4-4m0 0l-4-4m4 4H7m6 4v1a3 3 0 01-3 3H6a3 3 0 01-3-3V7a3 3 0 013-3h4a3 3 0 013 3v1"
            />
          </svg>
          Sign out
        </button>
      </div>

      {/* Tabs */}
      <div className="border-b border-netflix-gray">
        <div className="flex gap-6">
          {(['profile', 'preferences', 'security', 'sessions'] as const).map((tab) => (
            <button
              key={tab}
              onClick={() => setActiveTab(tab)}
              className={`relative py-3 text-sm font-medium capitalize transition-colors ${
                activeTab === tab ? 'text-white' : 'text-gray-400 hover:text-white'
              }`}
            >
              {tab}
              {activeTab === tab && (
                <span className="absolute bottom-0 left-0 right-0 h-0.5 bg-netflix-red" />
              )}
            </button>
          ))}
        </div>
      </div>

      {/* Tab Content */}
      <div className="max-w-2xl">
        {activeTab === 'profile' && <ProfileTab />}
        {activeTab === 'preferences' && <PreferencesTab />}
        {activeTab === 'security' && <SecurityTab />}
        {activeTab === 'sessions' && <SessionsTab />}
      </div>
    </div>
  )
}

function ProfileTab() {
  const { data: profile } = useProfile()
  const { mutate: updateProfile, isPending } = useUpdateProfile()
  const [displayName, setDisplayName] = useState(profile?.display_name || '')
  const [success, setSuccess] = useState('')

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault()
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

  return (
    <form onSubmit={handleSubmit} className="space-y-6">
      {success && (
        <div className="rounded-lg bg-green-500/20 p-3 text-sm text-green-400">{success}</div>
      )}

      <div>
        <label className="mb-2 block text-sm font-medium text-gray-400">Username</label>
        <input
          type="text"
          value={profile?.username || ''}
          disabled
          className="w-full rounded bg-netflix-gray px-4 py-3 text-gray-400"
        />
        <p className="mt-1 text-xs text-gray-500">Username cannot be changed</p>
      </div>

      <div>
        <label className="mb-2 block text-sm font-medium text-gray-400">Display Name</label>
        <input
          type="text"
          value={displayName}
          onChange={(e) => setDisplayName(e.target.value)}
          className="w-full rounded bg-netflix-gray px-4 py-3 text-white outline-none ring-1 ring-transparent transition-all focus:ring-netflix-red"
        />
      </div>

      <div>
        <label className="mb-2 block text-sm font-medium text-gray-400">Role</label>
        <div className="flex items-center gap-2">
          <span
            className={`rounded px-3 py-1 text-sm ${
              profile?.is_admin
                ? 'bg-purple-500/20 text-purple-400'
                : 'bg-blue-500/20 text-blue-400'
            }`}
          >
            {profile?.is_admin ? 'Administrator' : 'User'}
          </span>
        </div>
      </div>

      <button
        type="submit"
        disabled={isPending}
        className="rounded bg-netflix-red px-6 py-2 font-semibold text-white transition-colors hover:bg-netflix-red-hover disabled:opacity-50"
      >
        {isPending ? 'Saving...' : 'Save Changes'}
      </button>
    </form>
  )
}

function PreferencesTab() {
  const { data: preferences } = usePreferences()
  const { mutate: updatePreferences, isPending } = useUpdatePreferences()
  const { theme, setTheme } = useUIStore()

  const [prefs, setPrefs] = useState({
    subtitle_language: preferences?.subtitle_language || '',
    audio_language: preferences?.audio_language || '',
    max_streaming_quality: preferences?.max_streaming_quality || 'original',
    theme: theme,
  })

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault()
    updatePreferences({
      user_id: preferences?.user_id || 0,
      subtitle_language: prefs.subtitle_language,
      audio_language: prefs.audio_language,
      max_streaming_quality: prefs.max_streaming_quality,
      theme: prefs.theme,
    })
    setTheme(prefs.theme)
  }

  return (
    <form onSubmit={handleSubmit} className="space-y-6">
      <div>
        <label className="mb-2 block text-sm font-medium text-gray-400">Subtitle Language</label>
        <select
          value={prefs.subtitle_language}
          onChange={(e) => setPrefs({ ...prefs, subtitle_language: e.target.value })}
          className="w-full rounded bg-netflix-gray px-4 py-3 text-white outline-none ring-1 ring-transparent transition-all focus:ring-netflix-red"
        >
          <option value="">Auto</option>
          <option value="vi">Vietnamese</option>
          <option value="en">English</option>
        </select>
      </div>

      <div>
        <label className="mb-2 block text-sm font-medium text-gray-400">Audio Language</label>
        <select
          value={prefs.audio_language}
          onChange={(e) => setPrefs({ ...prefs, audio_language: e.target.value })}
          className="w-full rounded bg-netflix-gray px-4 py-3 text-white outline-none ring-1 ring-transparent transition-all focus:ring-netflix-red"
        >
          <option value="">Auto</option>
          <option value="vi">Vietnamese</option>
          <option value="en">English</option>
        </select>
      </div>

      <div>
        <label className="mb-2 block text-sm font-medium text-gray-400">
          Max Streaming Quality
        </label>
        <select
          value={prefs.max_streaming_quality}
          onChange={(e) => setPrefs({ ...prefs, max_streaming_quality: e.target.value })}
          className="w-full rounded bg-netflix-gray px-4 py-3 text-white outline-none ring-1 ring-transparent transition-all focus:ring-netflix-red"
        >
          <option value="original">Original</option>
          <option value="4k">4K</option>
          <option value="1080p">1080p</option>
          <option value="720p">720p</option>
          <option value="480p">480p</option>
        </select>
      </div>

      <div>
        <label className="mb-2 block text-sm font-medium text-gray-400">Theme</label>
        <select
          value={theme}
          onChange={(e) =>
            setPrefs({ ...prefs, theme: e.target.value as 'light' | 'dark' | 'system' })
          }
          className="w-full rounded bg-netflix-gray px-4 py-3 text-white outline-none ring-1 ring-transparent transition-all focus:ring-netflix-red"
        >
          <option value="system">System</option>
          <option value="light">Light</option>
          <option value="dark">Dark</option>
        </select>
      </div>

      <button
        type="submit"
        disabled={isPending}
        className="rounded bg-netflix-red px-6 py-2 font-semibold text-white transition-colors hover:bg-netflix-red-hover disabled:opacity-50"
      >
        {isPending ? 'Saving...' : 'Save Preferences'}
      </button>
    </form>
  )
}

function SecurityTab() {
  const { mutate: changePassword, isPending } = useChangePassword()
  const [oldPassword, setOldPassword] = useState('')
  const [newPassword, setNewPassword] = useState('')
  const [confirmPassword, setConfirmPassword] = useState('')
  const [error, setError] = useState('')
  const [success, setSuccess] = useState('')

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault()
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
        onError: (err: Error) => {
          setError(err.message)
        },
      },
    )
  }

  return (
    <form onSubmit={handleSubmit} className="space-y-6">
      {error && (
        <div className="rounded-lg bg-netflix-red/20 p-3 text-sm text-netflix-red">{error}</div>
      )}
      {success && (
        <div className="rounded-lg bg-green-500/20 p-3 text-sm text-green-400">{success}</div>
      )}

      <div>
        <label className="mb-2 block text-sm font-medium text-gray-400">Current Password</label>
        <input
          type="password"
          value={oldPassword}
          onChange={(e) => setOldPassword(e.target.value)}
          className="w-full rounded bg-netflix-gray px-4 py-3 text-white outline-none ring-1 ring-transparent transition-all focus:ring-netflix-red"
          required
        />
      </div>

      <div>
        <label className="mb-2 block text-sm font-medium text-gray-400">New Password</label>
        <input
          type="password"
          value={newPassword}
          onChange={(e) => setNewPassword(e.target.value)}
          className="w-full rounded bg-netflix-gray px-4 py-3 text-white outline-none ring-1 ring-transparent transition-all focus:ring-netflix-red"
          required
          minLength={8}
        />
      </div>

      <div>
        <label className="mb-2 block text-sm font-medium text-gray-400">Confirm Password</label>
        <input
          type="password"
          value={confirmPassword}
          onChange={(e) => setConfirmPassword(e.target.value)}
          className="w-full rounded bg-netflix-gray px-4 py-3 text-white outline-none ring-1 ring-transparent transition-all focus:ring-netflix-red"
          required
        />
      </div>

      <button
        type="submit"
        disabled={isPending}
        className="rounded bg-netflix-red px-6 py-2 font-semibold text-white transition-colors hover:bg-netflix-red-hover disabled:opacity-50"
      >
        {isPending ? 'Changing...' : 'Change Password'}
      </button>
    </form>
  )
}

function SessionsTab() {
  const { data: sessions, isLoading } = useSessions()
  const { mutate: revokeSession } = useRevokeSession()

  if (isLoading) {
    return (
      <div className="flex h-32 items-center justify-center">
        <div className="h-8 w-8 animate-spin rounded-full border-2 border-netflix-red border-t-transparent" />
      </div>
    )
  }

  return (
    <div className="space-y-4">
      {sessions?.map((session) => (
        <div
          key={session.id}
          className="flex items-center justify-between rounded-lg bg-netflix-dark p-4"
        >
          <div className="flex items-center gap-4">
            <div className="flex h-10 w-10 items-center justify-center rounded bg-netflix-gray">
              <svg
                className="h-5 w-5 text-gray-400"
                fill="none"
                stroke="currentColor"
                viewBox="0 0 24 24"
              >
                <path
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  strokeWidth={2}
                  d="M9.75 17L9 20l-1 1h8l-1-1-.75-3M3 13h18M5 17h14a2 2 0 002-2V5a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z"
                />
              </svg>
            </div>
            <div>
              <p className="font-medium text-white">{session.device_name || 'Unknown Device'}</p>
              <p className="text-sm text-gray-400">{session.ip_address}</p>
              <p className="text-xs text-gray-500">
                Last active: {new Date(session.last_active_at).toLocaleString()}
              </p>
            </div>
          </div>
          <button
            onClick={() => revokeSession(session.id)}
            className="rounded bg-netflix-gray px-3 py-1 text-sm text-white transition-colors hover:bg-netflix-red"
          >
            Revoke
          </button>
        </div>
      ))}
    </div>
  )
}
