import { useState, useRef } from 'react'
import { useSearchParams } from 'react-router'
import {
  LuUser,
  LuSettings,
  LuShield,
  LuMonitor,
  LuCaptions,
  LuServer,
  LuLibrary,
  LuUsers,
  LuSave,
  LuCheck,
  LuEye,
  LuEyeOff,
  LuPlus,
  LuRefreshCw,
  LuTrash2,
  LuX,
  LuFolder,
  LuFilm,
  LuTv,
  LuList,
  LuActivity,
  LuClock,
  LuWebhook,
  LuPlay,
  LuPause,
  LuGlobe,
} from 'react-icons/lu'
import { useAuthStore } from '@/stores/auth'
import { useUIStore } from '@/stores/ui'
import {
  useProfile,
  useUpdateProfile,
  usePreferences,
  useUpdatePreferences,
  useSessions,
  useRevokeSession,
  useChangePassword,
} from '@/hooks/stores/useAuth'
import {
  useLibraries,
  useCreateLibrary,
  useDeleteLibrary,
  useScanLibrary,
} from '@/hooks/stores/useMedia'
import { useUsers, useCreateUser, useUpdateUser, useDeleteUser } from '@/hooks/stores/useUsers'
import {
  useOpenSubsSettings,
  useUpdateOpenSubsSettings,
  useTMDbSettings,
  useUpdateTMDbSettings,
  useOMDbSettings,
  useUpdateOMDbSettings,
  useTVDBSettings,
  useUpdateTVDBSettings,
  useFanartSettings,
  useUpdateFanartSettings,
  useBulkRefreshRatings,
  useSubdlSettings,
  useUpdateSubdlSettings,
  useDeepLSettings,
  useUpdateDeepLSettings,
  usePlaybackSettings,
  useUpdatePlaybackSettings,
  useAutoSubSettings,
  useUpdateAutoSubSettings,
  useCinemaSettings,
  useUpdateCinemaSettings,
  useUploadCinemaIntro,
} from '@/hooks/stores/useSettings'
import {
  useServerInfo,
  useLibraryStats,
  useActivity,
  useWebhooks,
  useCreateWebhook,
  useUpdateWebhook,
  useDeleteWebhook,
  useScheduledTasks,
  useRunTask,
} from '@/hooks/stores/useAdmin'
import { DirectoryPicker } from '@/components/DirectoryPicker'
import { LanguageSwitcher } from '@/components/LanguageSwitcher'
import { useTranslation } from '@/hooks/useTranslation'
import type { User, Webhook } from '@/types/api'

// ── Section Definitions ───────────────────────────────────────────────────────

interface Section {
  id: string
  labelKey: string
  icon: React.ReactNode
  group: string
  adminOnly?: boolean
}

const ALL_SECTIONS: Section[] = [
  // Web Settings (all users)
  {
    id: 'profile',
    labelKey: 'sections.profile.title',
    icon: <LuUser size={18} />,
    group: 'Web Settings',
  },
  {
    id: 'preferences',
    labelKey: 'sections.preferences.title',
    icon: <LuSettings size={18} />,
    group: 'Web Settings',
  },
  {
    id: 'security',
    labelKey: 'sections.security.title',
    icon: <LuShield size={18} />,
    group: 'Web Settings',
  },
  {
    id: 'sessions',
    labelKey: 'sections.sessions.title',
    icon: <LuMonitor size={18} />,
    group: 'Web Settings',
  },
  // Admin Preferences (admin only)
  {
    id: 'metadata',
    labelKey: 'sections.metadata.title',
    icon: <LuFilm size={18} />,
    group: 'Admin Preferences',
    adminOnly: true,
  },
  {
    id: 'subtitles',
    labelKey: 'sections.subtitles.title',
    icon: <LuCaptions size={18} />,
    group: 'Admin Preferences',
    adminOnly: true,
  },
  {
    id: 'playback',
    labelKey: 'sections.playback.title',
    icon: <LuPlay size={18} />,
    group: 'Admin Preferences',
    adminOnly: true,
  },
  {
    id: 'cinema',
    labelKey: 'sections.cinema.title',
    icon: <LuFilm size={18} />,
    group: 'Admin Preferences',
    adminOnly: true,
  },
  // Velox Server (admin only)
  {
    id: 'general',
    labelKey: 'sections.general.title',
    icon: <LuServer size={18} />,
    group: 'Velox Server',
    adminOnly: true,
  },
  {
    id: 'libraries',
    labelKey: 'sections.libraries.title',
    icon: <LuLibrary size={18} />,
    group: 'Velox Server',
    adminOnly: true,
  },
  {
    id: 'users',
    labelKey: 'sections.users.title',
    icon: <LuUsers size={18} />,
    group: 'Velox Server',
    adminOnly: true,
  },
  {
    id: 'activity',
    labelKey: 'sections.activity.title',
    icon: <LuActivity size={18} />,
    group: 'Velox Server',
    adminOnly: true,
  },
  {
    id: 'tasks',
    labelKey: 'sections.tasks.title',
    icon: <LuClock size={18} />,
    group: 'Velox Server',
    adminOnly: true,
  },
  {
    id: 'webhooks',
    labelKey: 'sections.webhooks.title',
    icon: <LuWebhook size={18} />,
    group: 'Velox Server',
    adminOnly: true,
  },
]

// ── Main Page ─────────────────────────────────────────────────────────────────

export function SettingsPage() {
  const { user } = useAuthStore()
  const [searchParams, setSearchParams] = useSearchParams()
  const activeSection = searchParams.get('section') || 'profile'
  const isAdmin = user?.is_admin ?? false
  const { t } = useTranslation('settings')

  const setSection = (id: string) => setSearchParams({ section: id })

  const sections = ALL_SECTIONS.filter((s) => !s.adminOnly || isAdmin)
  const groups = sections.reduce<Record<string, Section[]>>((acc, s) => {
    ;(acc[s.group] ??= []).push(s)
    return acc
  }, {})

  return (
    <div className="flex min-h-[calc(100vh-4rem)]">
      {/* Sidebar */}
      <aside className="w-56 shrink-0 border-r border-netflix-gray/50 bg-netflix-black/50">
        <div
          className="sticky top-16 overflow-y-auto py-4"
          style={{ maxHeight: 'calc(100vh - 4rem)' }}
        >
          {Object.entries(groups).map(([group, items], gi) => (
            <div key={group} className={gi > 0 ? 'mt-3' : ''}>
              <SidebarGroupHeader label={t(`groups.${group.toLowerCase().replace(/ /g, '')}`)} />
              {items.map((item) => (
                <button
                  key={item.id}
                  onClick={() => setSection(item.id)}
                  className={`flex w-full items-center gap-2.5 px-4 py-2 text-[13px] transition-colors ${
                    activeSection === item.id
                      ? 'bg-netflix-red/90 text-white font-medium'
                      : 'text-gray-400 hover:bg-netflix-gray/40 hover:text-white'
                  }`}
                >
                  <span className="shrink-0">{item.icon}</span>
                  <span className="flex-1 text-left">{t(item.labelKey)}</span>
                </button>
              ))}
            </div>
          ))}
        </div>
      </aside>

      {/* Content */}
      <main className="flex-1 overflow-y-auto p-6 lg:p-8">
        {activeSection === 'profile' && <ProfileSection />}
        {activeSection === 'preferences' && <PreferencesSection />}
        {activeSection === 'security' && <SecuritySection />}
        {activeSection === 'sessions' && <SessionsSection />}
        {activeSection === 'metadata' && <MetadataSection />}
        {activeSection === 'subtitles' && <SubtitlesSection />}
        {activeSection === 'general' && <GeneralSection />}
        {activeSection === 'libraries' && <LibrariesSection />}
        {activeSection === 'users' && <UsersSection />}
        {activeSection === 'activity' && <ActivitySection />}
        {activeSection === 'tasks' && <TasksSection />}
        {activeSection === 'webhooks' && <WebhooksSection />}
        {activeSection === 'playback' && <PlaybackSection />}
        {activeSection === 'cinema' && <CinemaSection />}
      </main>
    </div>
  )
}

// ── Sidebar Group Header ──────────────────────────────────────────────────────

function SidebarGroupHeader({ label }: { label: string }) {
  return (
    <p className="px-4 py-1.5 text-[10px] font-semibold uppercase tracking-wider text-gray-500">
      {label}
    </p>
  )
}

// ── Web Settings: Profile ─────────────────────────────────────────────────────

function ProfileSection() {
  const { data: profile } = useProfile()
  const { mutate: updateProfile, isPending } = useUpdateProfile()
  const [edited, setEdited] = useState<string | null>(null)
  const displayName = edited ?? profile?.display_name ?? ''
  const setDisplayName = (v: string) => setEdited(v)
  const [success, setSuccess] = useState('')

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault()
    setSuccess('')
    updateProfile(
      { display_name: displayName },
      { onSuccess: () => setSuccess('Profile updated successfully') },
    )
  }

  return (
    <div className="max-w-xl">
      <SectionHeader title="Profile" description="Manage your account information" />
      <form onSubmit={handleSubmit} className="mt-6 space-y-5">
        {success && <SuccessMsg>{success}</SuccessMsg>}
        <Field label="Username">
          <input type="text" value={profile?.username || ''} disabled className={inputDisabled} />
          <p className="mt-1 text-xs text-gray-500">Username cannot be changed</p>
        </Field>
        <Field label="Display Name">
          <input
            type="text"
            value={displayName}
            onChange={(e) => setDisplayName(e.target.value)}
            className={inputClass}
          />
        </Field>
        <Field label="Role">
          <span
            className={`inline-block rounded px-3 py-1 text-sm ${
              profile?.is_admin
                ? 'bg-purple-500/20 text-purple-400'
                : 'bg-blue-500/20 text-blue-400'
            }`}
          >
            {profile?.is_admin ? 'Administrator' : 'User'}
          </span>
        </Field>
        <SaveButton isPending={isPending} />
      </form>
    </div>
  )
}

// ── Web Settings: Preferences ─────────────────────────────────────────────────

function PreferencesSection() {
  const { data: preferences } = usePreferences()
  const { mutate: updatePreferences, isPending } = useUpdatePreferences()
  const { theme, setTheme } = useUIStore()

  type PrefsEdits = {
    subtitle_language?: string
    audio_language?: string
    max_streaming_quality?: string
    theme?: 'light' | 'dark' | 'system'
  }
  const [edits, setEdits] = useState<PrefsEdits>({})
  const prefs = {
    subtitle_language: edits.subtitle_language ?? preferences?.subtitle_language ?? '',
    audio_language: edits.audio_language ?? preferences?.audio_language ?? '',
    max_streaming_quality:
      edits.max_streaming_quality ?? preferences?.max_streaming_quality ?? 'original',
    theme: edits.theme ?? theme,
  }
  const setPrefs = (patch: PrefsEdits) => setEdits((prev) => ({ ...prev, ...patch }))

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault()
    updatePreferences({
      user_id: preferences?.user_id || 0,
      subtitle_language: prefs.subtitle_language,
      audio_language: prefs.audio_language,
      max_streaming_quality: prefs.max_streaming_quality,
      theme: prefs.theme,
      language: preferences?.language || 'en',
    })
    setTheme(prefs.theme)
  }

  return (
    <div className="max-w-xl">
      <SectionHeader title="Preferences" description="Customize your viewing experience" />
      <form onSubmit={handleSubmit} className="mt-6 space-y-5">
        <Field label="Subtitle Language">
          <select
            value={prefs.subtitle_language}
            onChange={(e) => setPrefs({ subtitle_language: e.target.value })}
            className={selectClass}
          >
            <option value="">Auto</option>
            <option value="vi">Vietnamese</option>
            <option value="en">English</option>
          </select>
        </Field>
        <Field label="Audio Language">
          <select
            value={prefs.audio_language}
            onChange={(e) => setPrefs({ audio_language: e.target.value })}
            className={selectClass}
          >
            <option value="">Auto</option>
            <option value="vi">Vietnamese</option>
            <option value="en">English</option>
          </select>
        </Field>
        <Field label="Max Streaming Quality">
          <select
            value={prefs.max_streaming_quality}
            onChange={(e) => setPrefs({ max_streaming_quality: e.target.value })}
            className={selectClass}
          >
            <option value="original">Original</option>
            <option value="4k">4K</option>
            <option value="1080p">1080p</option>
            <option value="720p">720p</option>
            <option value="480p">480p</option>
          </select>
        </Field>
        <Field label="Theme">
          <select
            value={prefs.theme}
            onChange={(e) => setPrefs({ theme: e.target.value as 'light' | 'dark' | 'system' })}
            className={selectClass}
          >
            <option value="system">System</option>
            <option value="dark">Dark</option>
            <option value="light">Light</option>
          </select>
        </Field>
        <Field label="Language">
          <LanguageSwitcher />
        </Field>
        <SaveButton isPending={isPending} />
      </form>
    </div>
  )
}

// ── Web Settings: Security ────────────────────────────────────────────────────

function SecuritySection() {
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
        onError: (err: Error) => setError(err.message),
      },
    )
  }

  return (
    <div className="max-w-xl">
      <SectionHeader title="Security" description="Change your password" />
      <form onSubmit={handleSubmit} className="mt-6 space-y-5">
        {error && <ErrorMsg>{error}</ErrorMsg>}
        {success && <SuccessMsg>{success}</SuccessMsg>}
        <Field label="Current Password">
          <input
            type="password"
            value={oldPassword}
            onChange={(e) => setOldPassword(e.target.value)}
            className={inputClass}
            required
          />
        </Field>
        <Field label="New Password">
          <input
            type="password"
            value={newPassword}
            onChange={(e) => setNewPassword(e.target.value)}
            className={inputClass}
            required
            minLength={8}
          />
        </Field>
        <Field label="Confirm Password">
          <input
            type="password"
            value={confirmPassword}
            onChange={(e) => setConfirmPassword(e.target.value)}
            className={inputClass}
            required
          />
        </Field>
        <SaveButton isPending={isPending} label="Change Password" />
      </form>
    </div>
  )
}

// ── Web Settings: Sessions ────────────────────────────────────────────────────

function SessionsSection() {
  const { data: sessions, isLoading } = useSessions()
  const { mutate: revokeSession } = useRevokeSession()

  return (
    <div className="max-w-2xl">
      <SectionHeader title="Sessions" description="Manage your active sessions" />
      {isLoading ? (
        <Spinner />
      ) : (
        <div className="mt-6 space-y-3">
          {sessions?.map((session) => (
            <div
              key={session.id}
              className="flex items-center justify-between rounded-lg bg-netflix-dark p-4"
            >
              <div className="flex items-center gap-4">
                <div className="flex h-10 w-10 items-center justify-center rounded bg-netflix-gray">
                  <LuMonitor size={20} className="text-gray-400" />
                </div>
                <div>
                  <p className="text-sm font-medium text-white">
                    {session.device_name || 'Unknown Device'}
                  </p>
                  <p className="text-xs text-gray-400">{session.ip_address}</p>
                  <p className="text-xs text-gray-500">
                    Last active: {new Date(session.last_active_at).toLocaleString()}
                  </p>
                </div>
              </div>
              <button
                onClick={() => revokeSession(session.id)}
                className="rounded bg-netflix-gray px-3 py-1.5 text-sm text-white transition-colors hover:bg-netflix-red"
              >
                Revoke
              </button>
            </div>
          ))}
          {sessions?.length === 0 && (
            <p className="py-8 text-center text-sm text-gray-400">No active sessions</p>
          )}
        </div>
      )}
    </div>
  )
}

// ── Admin Preferences: Metadata (TMDb) ───────────────────────────────────────

function MetadataSection() {
  const { data: tmdbSettings, isLoading: tmdbLoading } = useTMDbSettings()
  const { mutate: updateTmdb, isPending: tmdbSaving } = useUpdateTMDbSettings()
  const { data: omdbSettings, isLoading: omdbLoading } = useOMDbSettings()
  const { mutate: updateOmdb, isPending: omdbSaving } = useUpdateOMDbSettings()
  const { data: tvdbSettings, isLoading: tvdbLoading } = useTVDBSettings()
  const { mutate: updateTvdb, isPending: tvdbSaving } = useUpdateTVDBSettings()
  const { data: fanartSettings, isLoading: fanartLoading } = useFanartSettings()
  const { mutate: updateFanart, isPending: fanartSaving } = useUpdateFanartSettings()
  const {
    mutate: bulkRefresh,
    isPending: isRefreshing,
    data: refreshResult,
    error: refreshError,
  } = useBulkRefreshRatings()

  const [tmdbEdited, setTmdbEdited] = useState<string | null>(null)
  const [tmdbSaved, setTmdbSaved] = useState(false)
  const [omdbEdited, setOmdbEdited] = useState<string | null>(null)
  const [omdbSaved, setOmdbSaved] = useState(false)
  const [tvdbEdited, setTvdbEdited] = useState<string | null>(null)
  const [tvdbSaved, setTvdbSaved] = useState(false)
  const [fanartEdited, setFanartEdited] = useState<string | null>(null)
  const [fanartSaved, setFanartSaved] = useState(false)

  const tmdbKey = tmdbEdited ?? tmdbSettings?.api_key ?? ''
  const omdbKey = omdbEdited ?? omdbSettings?.api_key ?? ''

  const handleTmdbSave = (e: React.FormEvent) => {
    e.preventDefault()
    updateTmdb(
      { api_key: tmdbKey },
      {
        onSuccess: () => {
          setTmdbEdited(null)
          setTmdbSaved(true)
          setTimeout(() => setTmdbSaved(false), 2000)
        },
      },
    )
  }

  const handleOmdbSave = (e: React.FormEvent) => {
    e.preventDefault()
    updateOmdb(
      { api_key: omdbKey },
      {
        onSuccess: () => {
          setOmdbEdited(null)
          setOmdbSaved(true)
          setTimeout(() => setOmdbSaved(false), 2000)
        },
      },
    )
  }

  const tvdbKey = tvdbEdited ?? tvdbSettings?.api_key ?? ''

  const handleTvdbSave = (e: React.FormEvent) => {
    e.preventDefault()
    updateTvdb(
      { api_key: tvdbKey },
      {
        onSuccess: () => {
          setTvdbEdited(null)
          setTvdbSaved(true)
          setTimeout(() => setTvdbSaved(false), 2000)
        },
      },
    )
  }

  const fanartKey = fanartEdited ?? fanartSettings?.api_key ?? ''

  const handleFanartSave = (e: React.FormEvent) => {
    e.preventDefault()
    updateFanart(
      { api_key: fanartKey },
      {
        onSuccess: () => {
          setFanartEdited(null)
          setFanartSaved(true)
          setTimeout(() => setFanartSaved(false), 2000)
        },
      },
    )
  }

  if (tmdbLoading || omdbLoading || tvdbLoading || fanartLoading) return <Spinner />

  return (
    <div className="max-w-xl space-y-6">
      <SectionHeader
        title="Metadata"
        description="Configure metadata providers for movies and TV shows"
      />

      <div className="rounded-lg bg-netflix-dark p-5">
        <div className="mb-1 flex items-center gap-2">
          <h3 className="text-sm font-semibold text-white">TMDb (The Movie Database)</h3>
          <span className="rounded bg-green-500/20 px-2 py-0.5 text-[10px] font-medium text-green-400">
            {tmdbSettings?.api_key ? 'Custom Key' : 'Built-in Key'}
          </span>
        </div>
        <p className="mb-5 text-xs text-gray-400">
          TMDb provides posters, backdrops, plot summaries, cast, and genres for your media. A
          built-in API key is included — override below only if you want to use your own.
        </p>

        <form onSubmit={handleTmdbSave} className="space-y-5">
          <div className="space-y-3">
            <span className="text-xs font-medium text-gray-300">
              Custom API Key (v4 Read Access Token)
            </span>
            <p className="text-[11px] text-gray-500">
              Optional. Leave blank to use the built-in key. Get your own free key at themoviedb.org
              if needed.
            </p>
            <input
              type="text"
              value={tmdbKey}
              onChange={(e) => setTmdbEdited(e.target.value)}
              placeholder="Leave blank to use built-in key"
              className={inputClass}
            />
          </div>

          <div className="pt-1">
            <button
              type="submit"
              disabled={tmdbSaving}
              className="flex items-center gap-2 rounded bg-netflix-red px-4 py-2 text-sm font-medium text-white transition-colors hover:bg-netflix-red-hover disabled:opacity-50"
            >
              {tmdbSaved ? (
                <>
                  <LuCheck size={14} /> Saved
                </>
              ) : (
                <>
                  <LuSave size={14} /> {tmdbSaving ? 'Saving...' : 'Save'}
                </>
              )}
            </button>
          </div>
        </form>
      </div>

      <div className="rounded-lg bg-netflix-dark p-5">
        <div className="mb-1 flex items-center gap-2">
          <h3 className="text-sm font-semibold text-white">OMDb (Open Movie Database)</h3>
          <span className="rounded bg-green-500/20 px-2 py-0.5 text-[10px] font-medium text-green-400">
            {omdbSettings?.api_key ? 'Custom Key' : 'Built-in Key'}
          </span>
        </div>
        <p className="mb-5 text-xs text-gray-400">
          OMDb provides IMDb ratings, Rotten Tomatoes scores, and Metacritic scores. A built-in API
          key is included — override below only if you want to use your own.
        </p>

        <form onSubmit={handleOmdbSave} className="space-y-5">
          <div className="space-y-3">
            <span className="text-xs font-medium text-gray-300">Custom API Key</span>
            <p className="text-[11px] text-gray-500">
              Optional. Leave blank to use the built-in key. Get your own free key at omdbapi.com.
            </p>
            <input
              type="text"
              value={omdbKey}
              onChange={(e) => setOmdbEdited(e.target.value)}
              placeholder="Leave blank to use built-in key"
              className={inputClass}
            />
          </div>

          <div className="pt-1">
            <button
              type="submit"
              disabled={omdbSaving}
              className="flex items-center gap-2 rounded bg-netflix-red px-4 py-2 text-sm font-medium text-white transition-colors hover:bg-netflix-red-hover disabled:opacity-50"
            >
              {omdbSaved ? (
                <>
                  <LuCheck size={14} /> Saved
                </>
              ) : (
                <>
                  <LuSave size={14} /> {omdbSaving ? 'Saving...' : 'Save'}
                </>
              )}
            </button>
          </div>
        </form>
      </div>

      <div className="rounded-lg bg-netflix-dark p-5">
        <div className="mb-1 flex items-center gap-2">
          <h3 className="text-sm font-semibold text-white">TheTVDB</h3>
          <span className="rounded bg-green-500/20 px-2 py-0.5 text-[10px] font-medium text-green-400">
            {tvdbSettings?.api_key ? 'Custom Key' : 'Built-in Key'}
          </span>
        </div>
        <p className="mb-5 text-xs text-gray-400">
          TheTVDB provides additional TV series metadata including TVDB IDs and supplementary
          episode data. A built-in API key is included — override below only if you want to use your
          own.
        </p>

        <form onSubmit={handleTvdbSave} className="space-y-5">
          <div className="space-y-3">
            <span className="text-xs font-medium text-gray-300">Custom API Key</span>
            <p className="text-[11px] text-gray-500">
              Optional. Leave blank to use the built-in key. Get your own free key at thetvdb.com.
            </p>
            <input
              type="text"
              value={tvdbKey}
              onChange={(e) => setTvdbEdited(e.target.value)}
              placeholder="Leave blank to use built-in key"
              className={inputClass}
            />
          </div>

          <div className="pt-1">
            <button
              type="submit"
              disabled={tvdbSaving}
              className="flex items-center gap-2 rounded bg-netflix-red px-4 py-2 text-sm font-medium text-white transition-colors hover:bg-netflix-red-hover disabled:opacity-50"
            >
              {tvdbSaved ? (
                <>
                  <LuCheck size={14} /> Saved
                </>
              ) : (
                <>
                  <LuSave size={14} /> {tvdbSaving ? 'Saving...' : 'Save'}
                </>
              )}
            </button>
          </div>
        </form>
      </div>

      <div className="rounded-lg bg-netflix-dark p-5">
        <div className="mb-1 flex items-center gap-2">
          <h3 className="text-sm font-semibold text-white">Fanart.tv</h3>
          <span className="rounded bg-green-500/20 px-2 py-0.5 text-[10px] font-medium text-green-400">
            {fanartSettings?.api_key ? 'Custom Key' : 'Built-in Key'}
          </span>
        </div>
        <p className="mb-5 text-xs text-gray-400">
          Fanart.tv provides high-quality artwork including clearlogos, thumbnails, and HD art. A
          built-in API key is included — override below only if you want to use your own.
        </p>

        <form onSubmit={handleFanartSave} className="space-y-5">
          <div className="space-y-3">
            <span className="text-xs font-medium text-gray-300">Custom API Key</span>
            <p className="text-[11px] text-gray-500">
              Optional. Leave blank to use the built-in key. Get your own free key at fanart.tv.
            </p>
            <input
              type="text"
              value={fanartKey}
              onChange={(e) => setFanartEdited(e.target.value)}
              placeholder="Leave blank to use built-in key"
              className={inputClass}
            />
          </div>

          <div className="pt-1">
            <button
              type="submit"
              disabled={fanartSaving}
              className="flex items-center gap-2 rounded bg-netflix-red px-4 py-2 text-sm font-medium text-white transition-colors hover:bg-netflix-red-hover disabled:opacity-50"
            >
              {fanartSaved ? (
                <>
                  <LuCheck size={14} /> Saved
                </>
              ) : (
                <>
                  <LuSave size={14} /> {fanartSaving ? 'Saving...' : 'Save'}
                </>
              )}
            </button>
          </div>
        </form>
      </div>

      <div className="rounded-lg bg-netflix-dark p-5">
        <div className="mb-1 flex items-center gap-2">
          <h3 className="text-sm font-semibold text-white">Refresh All Metadata</h3>
        </div>
        <p className="mb-4 text-xs text-gray-400">
          Auto-match unmatched media against TMDb, then fetch IMDb, Rotten Tomatoes, and Metacritic
          ratings from OMDb. This does not require a library rescan.
        </p>
        <div className="flex items-center gap-3">
          <button
            onClick={() => bulkRefresh()}
            disabled={isRefreshing}
            className="flex items-center gap-2 rounded bg-netflix-red px-4 py-2 text-sm font-medium text-white transition-colors hover:bg-netflix-red-hover disabled:opacity-50"
          >
            {isRefreshing ? (
              <>
                <LuRefreshCw size={14} className="animate-spin" /> Refreshing...
              </>
            ) : (
              <>
                <LuRefreshCw size={14} /> Refresh All Metadata
              </>
            )}
          </button>
          {refreshResult && !isRefreshing && (
            <span className="text-xs text-green-400">Updated {refreshResult.updated} items</span>
          )}
          {refreshError && !isRefreshing && (
            <span className="text-xs text-red-400">Error: {refreshError.message}</span>
          )}
        </div>
      </div>

      <div className="rounded-lg bg-netflix-dark p-5">
        <h3 className="mb-1 text-sm font-semibold text-white">How it works</h3>
        <ul className="space-y-1 text-xs text-gray-400">
          <li>• During library scan, Velox matches filenames to TMDb entries</li>
          <li>• NFO files (if present) take priority over TMDb search</li>
          <li>• Matched metadata: poster, backdrop, overview, cast, genres</li>
          <li>• OMDb enriches with IMDb rating, Rotten Tomatoes %, and Metacritic score</li>
          <li>• TheTVDB provides additional TV series data (TVDB IDs, IMDb cross-reference)</li>
          <li>• Fanart.tv enriches with clearlogos, thumbnails, and HD artwork</li>
          <li>
            • TVmaze adds network info and cross-references external IDs (free, no key needed)
          </li>
          <li>• Re-scan a library to fetch metadata for existing items</li>
        </ul>
      </div>
    </div>
  )
}

// ── Admin Preferences: Subtitles ──────────────────────────────────────────────

function SubtitlesSection() {
  const { data: settings, isLoading } = useOpenSubsSettings()
  const { mutate: updateSettings, isPending: isSaving } = useUpdateOpenSubsSettings()

  const [editedApiKey, setEditedApiKey] = useState<string | null>(null)
  const [editedUsername, setEditedUsername] = useState<string | null>(null)
  const apiKey = editedApiKey ?? settings?.api_key ?? ''
  const setApiKey = (v: string) => setEditedApiKey(v)
  const username = editedUsername ?? settings?.username ?? ''
  const setUsername = (v: string) => setEditedUsername(v)
  const [password, setPassword] = useState('')
  const [showPassword, setShowPassword] = useState(false)
  const [saved, setSaved] = useState(false)

  const handleSave = (e: React.FormEvent) => {
    e.preventDefault()
    updateSettings(
      { api_key: apiKey, username, password },
      {
        onSuccess: () => {
          setSaved(true)
          setPassword('')
          setTimeout(() => setSaved(false), 2000)
        },
      },
    )
  }

  if (isLoading) return <Spinner />

  return (
    <div className="max-w-xl space-y-6">
      <SectionHeader title="Subtitles" description="Configure external subtitle providers" />

      {/* OpenSubtitles */}
      <div className="rounded-lg bg-netflix-dark p-5">
        <div className="mb-1 flex items-center gap-2">
          <h3 className="text-sm font-semibold text-white">OpenSubtitles.com</h3>
          {settings?.password_set && settings?.api_key && (
            <span className="rounded bg-green-500/20 px-2 py-0.5 text-[10px] font-medium text-green-400">
              Connected
            </span>
          )}
        </div>
        <p className="mb-5 text-xs text-gray-400">
          Connect your OpenSubtitles account to search and download subtitles.
        </p>

        <form onSubmit={handleSave} className="space-y-5">
          {/* Step 1: API Key */}
          <div className="space-y-3">
            <div className="flex items-center gap-2">
              <span className="flex h-5 w-5 shrink-0 items-center justify-center rounded-full bg-netflix-gray text-[10px] font-bold text-white">
                1
              </span>
              <span className="text-xs font-medium text-gray-300">API Key</span>
            </div>
            <p className="pl-7 text-[11px] text-gray-500">
              Get a free API key at opensubtitles.com
            </p>
            <div className="pl-7">
              <input
                type="text"
                value={apiKey}
                onChange={(e) => setApiKey(e.target.value)}
                placeholder="Your OpenSubtitles API key"
                className={inputClass}
              />
            </div>
          </div>

          <div className="border-t border-netflix-gray/30" />

          {/* Step 2: Account */}
          <div className="space-y-3">
            <div className="flex items-center gap-2">
              <span className="flex h-5 w-5 shrink-0 items-center justify-center rounded-full bg-netflix-gray text-[10px] font-bold text-white">
                2
              </span>
              <span className="text-xs font-medium text-gray-300">Account</span>
            </div>
            <p className="pl-7 text-[11px] text-gray-500">Your OpenSubtitles login credentials</p>
            <div className="space-y-3 pl-7">
              <Field label="Username" compact>
                <input
                  type="text"
                  value={username}
                  onChange={(e) => setUsername(e.target.value)}
                  placeholder="OpenSubtitles username"
                  className={inputClass}
                />
              </Field>
              <Field
                label={
                  <>
                    Password
                    {settings?.password_set && (
                      <span className="ml-2 text-xs text-green-400">(configured)</span>
                    )}
                  </>
                }
                compact
              >
                <div className="relative">
                  <input
                    type={showPassword ? 'text' : 'password'}
                    value={password}
                    onChange={(e) => setPassword(e.target.value)}
                    placeholder={
                      settings?.password_set
                        ? 'Leave blank to keep current'
                        : 'OpenSubtitles password'
                    }
                    className={`${inputClass} pr-10`}
                  />
                  <button
                    type="button"
                    onClick={() => setShowPassword(!showPassword)}
                    className="absolute right-3 top-1/2 -translate-y-1/2 text-gray-500 hover:text-gray-300"
                  >
                    {showPassword ? <LuEyeOff size={16} /> : <LuEye size={16} />}
                  </button>
                </div>
              </Field>
            </div>
          </div>

          <div className="pt-1">
            <button
              type="submit"
              disabled={isSaving}
              className="flex items-center gap-2 rounded bg-netflix-red px-4 py-2 text-sm font-medium text-white transition-colors hover:bg-netflix-red-hover disabled:opacity-50"
            >
              {saved ? (
                <>
                  <LuCheck size={14} /> Saved
                </>
              ) : (
                <>
                  <LuSave size={14} /> {isSaving ? 'Saving...' : 'Save'}
                </>
              )}
            </button>
          </div>
        </form>
      </div>

      {/* Subdl */}
      <SubdlCard />

      {/* DeepL Translation */}
      <DeepLCard />

      {/* Auto-Download */}
      <AutoSubCard />
    </div>
  )
}

function SubdlCard() {
  const { data: settings, isLoading } = useSubdlSettings()
  const { mutate: updateSettings, isPending: isSaving } = useUpdateSubdlSettings()
  const [editedApiKey, setEditedApiKey] = useState<string | null>(null)
  const apiKey = editedApiKey ?? settings?.api_key ?? ''
  const [saved, setSaved] = useState(false)

  const handleSave = (e: React.FormEvent) => {
    e.preventDefault()
    updateSettings(
      { api_key: apiKey },
      {
        onSuccess: () => {
          setSaved(true)
          setTimeout(() => setSaved(false), 2000)
        },
      },
    )
  }

  if (isLoading) return <Spinner />

  return (
    <div className="rounded-lg bg-netflix-dark p-5">
      <div className="mb-1 flex items-center gap-2">
        <h3 className="text-sm font-semibold text-white">Subdl</h3>
        <span className="rounded bg-green-500/20 px-2 py-0.5 text-[10px] font-medium text-green-400">
          {settings?.api_key ? 'Custom Key' : 'Built-in Key'}
        </span>
      </div>
      <p className="mb-5 text-xs text-gray-400">
        Free subtitle provider with large database. Uses a built-in API key by default (override
        below).
      </p>

      <form onSubmit={handleSave} className="space-y-4">
        <div>
          <input
            type="text"
            value={apiKey}
            onChange={(e) => setEditedApiKey(e.target.value)}
            placeholder="Your Subdl API key"
            className={inputClass}
          />
        </div>
        <button
          type="submit"
          disabled={isSaving}
          className="flex items-center gap-2 rounded bg-netflix-red px-4 py-2 text-sm font-medium text-white transition-colors hover:bg-netflix-red-hover disabled:opacity-50"
        >
          {saved ? (
            <>
              <LuCheck size={14} /> Saved
            </>
          ) : (
            <>
              <LuSave size={14} /> {isSaving ? 'Saving...' : 'Save'}
            </>
          )}
        </button>
      </form>
    </div>
  )
}

function DeepLCard() {
  const { data: settings, isLoading } = useDeepLSettings()
  const { mutate: updateSettings, isPending: isSaving } = useUpdateDeepLSettings()
  const [editedApiKey, setEditedApiKey] = useState<string | null>(null)
  const apiKey = editedApiKey ?? settings?.api_key ?? ''
  const [saved, setSaved] = useState(false)

  const handleSave = (e: React.FormEvent) => {
    e.preventDefault()
    updateSettings(
      { api_key: apiKey },
      {
        onSuccess: () => {
          setSaved(true)
          setTimeout(() => setSaved(false), 2000)
        },
      },
    )
  }

  if (isLoading) return <Spinner />

  return (
    <div className="rounded-lg bg-netflix-dark p-5">
      <div className="mb-1 flex items-center gap-2">
        <h3 className="text-sm font-semibold text-white">Subtitle Translation</h3>
        <span className="rounded bg-blue-500/20 px-2 py-0.5 text-[10px] font-medium text-blue-400">
          {settings?.api_key ? 'DeepL' : 'Google Translate'}
        </span>
      </div>
      <p className="mb-5 text-xs text-gray-400">
        Auto-translate subtitles on-demand. Uses Google Translate (free) by default. Add a DeepL API
        key for higher quality translations (500K chars/month free).
      </p>

      <form onSubmit={handleSave} className="space-y-4">
        <div>
          <span className="mb-1.5 block text-xs font-medium text-gray-300">
            DeepL API Key (optional)
          </span>
          <input
            type="text"
            value={apiKey}
            onChange={(e) => setEditedApiKey(e.target.value)}
            placeholder="Leave empty to use Google Translate"
            className={inputClass}
          />
          <p className="mt-1.5 text-[11px] text-gray-500">
            Get a free key at deepl.com/pro-api. Free tier: 500K chars/month.
          </p>
        </div>
        <button
          type="submit"
          disabled={isSaving}
          className="flex items-center gap-2 rounded bg-netflix-red px-4 py-2 text-sm font-medium text-white transition-colors hover:bg-netflix-red-hover disabled:opacity-50"
        >
          {saved ? (
            <>
              <LuCheck size={14} /> Saved
            </>
          ) : (
            <>
              <LuSave size={14} /> {isSaving ? 'Saving...' : 'Save'}
            </>
          )}
        </button>
      </form>
    </div>
  )
}

const COMMON_LANGUAGES = [
  { code: 'en', label: 'English' },
  { code: 'vi', label: 'Vietnamese' },
  { code: 'fr', label: 'French' },
  { code: 'de', label: 'German' },
  { code: 'es', label: 'Spanish' },
  { code: 'pt', label: 'Portuguese' },
  { code: 'ja', label: 'Japanese' },
  { code: 'ko', label: 'Korean' },
  { code: 'zh', label: 'Chinese' },
  { code: 'th', label: 'Thai' },
]

function AutoSubCard() {
  const { data: settings, isLoading } = useAutoSubSettings()
  const { mutate: updateSettings, isPending: isSaving } = useUpdateAutoSubSettings()
  const [edited, setEdited] = useState<string[] | null>(null)
  const [saved, setSaved] = useState(false)

  const selected =
    edited ?? (settings?.languages ? settings.languages.split(',').filter(Boolean) : [])

  const toggleLang = (code: string) => {
    const current = [...selected]
    const idx = current.indexOf(code)
    if (idx >= 0) {
      current.splice(idx, 1)
    } else {
      current.push(code)
    }
    setEdited(current)
  }

  const handleSave = () => {
    updateSettings(
      { languages: selected.join(',') },
      {
        onSuccess: () => {
          setEdited(null)
          setSaved(true)
          setTimeout(() => setSaved(false), 2000)
        },
      },
    )
  }

  if (isLoading) return <Spinner />

  return (
    <div className="rounded-lg bg-netflix-dark p-5">
      <div className="mb-1 flex items-center gap-2">
        <h3 className="text-sm font-semibold text-white">Auto-Download Subtitles</h3>
        {selected.length > 0 && (
          <span className="rounded bg-blue-500/20 px-2 py-0.5 text-[10px] font-medium text-blue-400">
            {selected.length} language{selected.length > 1 ? 's' : ''}
          </span>
        )}
      </div>
      <p className="mb-4 text-xs text-gray-400">
        Automatically download subtitles when new media is scanned. Select which languages to fetch.
      </p>

      <div className="mb-4 flex flex-wrap gap-2">
        {COMMON_LANGUAGES.map((lang) => (
          <button
            key={lang.code}
            type="button"
            onClick={() => toggleLang(lang.code)}
            className={`rounded-full px-3 py-1.5 text-xs font-medium transition-colors ${
              selected.includes(lang.code)
                ? 'bg-netflix-red text-white'
                : 'bg-netflix-gray text-gray-300 hover:bg-netflix-gray/80'
            }`}
          >
            {lang.label}
          </button>
        ))}
      </div>

      {selected.length === 0 && (
        <p className="mb-4 text-[11px] text-gray-500">
          No languages selected. Auto-download is disabled.
        </p>
      )}

      <button
        type="button"
        onClick={handleSave}
        disabled={isSaving || edited === null}
        className="flex items-center gap-2 rounded bg-netflix-red px-4 py-2 text-sm font-medium text-white transition-colors hover:bg-netflix-red-hover disabled:opacity-50"
      >
        {saved ? (
          <>
            <LuCheck size={14} /> Saved
          </>
        ) : (
          <>
            <LuSave size={14} /> {isSaving ? 'Saving...' : 'Save'}
          </>
        )}
      </button>
    </div>
  )
}

// ── Velox Server: General ─────────────────────────────────────────────────────

function formatBytes(bytes: number): string {
  if (bytes === 0) return '0 B'
  const units = ['B', 'KB', 'MB', 'GB', 'TB']
  const i = Math.floor(Math.log(bytes) / Math.log(1024))
  return `${(bytes / Math.pow(1024, i)).toFixed(i > 1 ? 2 : 0)} ${units[i]}`
}

function timeAgo(dateStr: string): string {
  const now = Date.now()
  const then = new Date(dateStr).getTime()
  const diff = now - then
  const mins = Math.floor(diff / 60000)
  if (mins < 1) return 'just now'
  if (mins < 60) return `${mins} min ago`
  const hours = Math.floor(mins / 60)
  if (hours < 24) return `${hours} hour${hours > 1 ? 's' : ''} ago`
  const days = Math.floor(hours / 24)
  return `${days} day${days > 1 ? 's' : ''} ago`
}

function PlaybackSection() {
  const { data: settings, isLoading } = usePlaybackSettings()
  const { mutate: updateSettings, isPending: isSaving } = useUpdatePlaybackSettings()
  const [saved, setSaved] = useState(false)

  const handleChange = (mode: 'auto' | 'direct_play') => {
    updateSettings(
      { playback_mode: mode },
      {
        onSuccess: () => {
          setSaved(true)
          setTimeout(() => setSaved(false), 2000)
        },
      },
    )
  }

  if (isLoading) return <Spinner />

  const current = settings?.playback_mode || 'auto'

  return (
    <div className="max-w-2xl space-y-6">
      <SectionHeader title="Playback" description="Server-wide playback policy" />

      <div className="rounded-lg bg-netflix-dark p-6">
        <h3 className="mb-1 text-base font-semibold text-white">Playback Mode</h3>
        <p className="mb-4 text-sm text-gray-400">
          Control how media is delivered to clients. Direct Play sends the original file without
          transcoding — best for compatible devices and saves server resources.
        </p>

        <div className="space-y-3">
          {[
            {
              value: 'auto' as const,
              label: 'Auto (Recommended)',
              description:
                'Automatically decide based on client capabilities. Transcodes when the client cannot play the original format.',
            },
            {
              value: 'direct_play' as const,
              label: 'Force Direct Play',
              description:
                'Always send the original file. No transcoding. Best when all your devices support your media formats.',
            },
          ].map((option) => (
            <button
              key={option.value}
              onClick={() => handleChange(option.value)}
              disabled={isSaving}
              className={`flex w-full items-start gap-3 rounded-lg border-2 p-4 text-left transition-colors ${
                current === option.value
                  ? 'border-netflix-red bg-netflix-red/10'
                  : 'border-netflix-gray/50 bg-netflix-black/30 hover:border-white/20'
              }`}
            >
              <div
                className={`mt-0.5 flex h-5 w-5 shrink-0 items-center justify-center rounded-full border-2 ${
                  current === option.value ? 'border-netflix-red bg-netflix-red' : 'border-gray-500'
                }`}
              >
                {current === option.value && <div className="h-2 w-2 rounded-full bg-white" />}
              </div>
              <div>
                <p className="text-sm font-medium text-white">{option.label}</p>
                <p className="mt-0.5 text-xs text-gray-400">{option.description}</p>
              </div>
            </button>
          ))}
        </div>

        {saved && (
          <div className="mt-3 flex items-center gap-1.5 text-sm text-green-400">
            <LuCheck size={16} /> Saved
          </div>
        )}
      </div>
    </div>
  )
}

function GeneralSection() {
  const { data: serverInfo, isLoading: serverLoading } = useServerInfo()
  const { data: libraryStats, isLoading: statsLoading } = useLibraryStats()

  if (serverLoading || statsLoading) return <Spinner />

  return (
    <div className="max-w-3xl space-y-6">
      <SectionHeader title="Dashboard" description="Server information and status" />

      {/* Stats Cards */}
      <div className="grid grid-cols-2 gap-4 lg:grid-cols-4">
        <div className="rounded-lg bg-netflix-dark p-4">
          <p className="text-xs text-gray-400">Total Media</p>
          <p className="mt-1 text-2xl font-bold text-white">{serverInfo?.media_count ?? 0}</p>
        </div>
        <div className="rounded-lg bg-netflix-dark p-4">
          <p className="text-xs text-gray-400">Series</p>
          <p className="mt-1 text-2xl font-bold text-white">{serverInfo?.series_count ?? 0}</p>
        </div>
        <div className="rounded-lg bg-netflix-dark p-4">
          <p className="text-xs text-gray-400">Users</p>
          <p className="mt-1 text-2xl font-bold text-white">{serverInfo?.user_count ?? 0}</p>
        </div>
        <div className="rounded-lg bg-netflix-dark p-4">
          <p className="text-xs text-gray-400">Total Size</p>
          <p className="mt-1 text-2xl font-bold text-white">
            {formatBytes(serverInfo?.total_size_bytes ?? 0)}
          </p>
        </div>
      </div>

      {/* Server Info */}
      <div className="rounded-lg bg-netflix-dark p-5">
        <h3 className="mb-3 text-sm font-semibold text-white">Server Information</h3>
        <div className="space-y-0">
          <InfoRow label="Version" value={serverInfo?.version ?? 'Unknown'} />
          <InfoRow label="Uptime" value={serverInfo?.uptime ?? 'Unknown'} />
          <InfoRow label="Go Version" value={serverInfo?.go_version ?? 'Unknown'} />
          <InfoRow
            label="OS / Arch"
            value={`${serverInfo?.os ?? '?'} / ${serverInfo?.arch ?? '?'}`}
          />
          <InfoRow label="FFmpeg" value={serverInfo?.ffmpeg_version ?? 'Unknown'} />
          <InfoRow label="HW Acceleration" value={serverInfo?.hw_accel || 'None'} />
          <InfoRow label="Database" value={serverInfo?.database ?? 'SQLite'} />
        </div>
      </div>

      {/* Library Stats */}
      {libraryStats && libraryStats.length > 0 && (
        <div className="overflow-hidden rounded-lg bg-netflix-dark">
          <div className="px-5 pt-5 pb-3">
            <h3 className="text-sm font-semibold text-white">Library Statistics</h3>
          </div>
          <table className="w-full">
            <thead className="border-b border-netflix-gray bg-netflix-black/50">
              <tr>
                <th className="px-5 py-2.5 text-left text-xs font-medium text-gray-400">Library</th>
                <th className="px-5 py-2.5 text-left text-xs font-medium text-gray-400">Type</th>
                <th className="px-5 py-2.5 text-right text-xs font-medium text-gray-400">Items</th>
                <th className="px-5 py-2.5 text-right text-xs font-medium text-gray-400">Files</th>
                <th className="px-5 py-2.5 text-right text-xs font-medium text-gray-400">Size</th>
                <th className="px-5 py-2.5 text-right text-xs font-medium text-gray-400">
                  Last Scanned
                </th>
              </tr>
            </thead>
            <tbody>
              {libraryStats.map((lib) => (
                <tr
                  key={lib.id}
                  className="border-b border-netflix-gray/50 last:border-b-0 hover:bg-netflix-gray/30"
                >
                  <td className="px-5 py-3 text-sm font-medium text-white">{lib.name}</td>
                  <td className="px-5 py-3">
                    <span
                      className={`rounded px-2 py-0.5 text-xs font-medium ${
                        lib.type === 'movies'
                          ? 'bg-blue-500/20 text-blue-400'
                          : lib.type === 'tvshows'
                            ? 'bg-purple-500/20 text-purple-400'
                            : 'bg-green-500/20 text-green-400'
                      }`}
                    >
                      {lib.type}
                    </span>
                  </td>
                  <td className="px-5 py-3 text-right text-sm text-gray-300">{lib.item_count}</td>
                  <td className="px-5 py-3 text-right text-sm text-gray-300">{lib.file_count}</td>
                  <td className="px-5 py-3 text-right text-sm text-gray-300">
                    {formatBytes(lib.total_size_bytes)}
                  </td>
                  <td className="px-5 py-3 text-right text-sm text-gray-400">
                    {lib.last_scanned ? timeAgo(lib.last_scanned) : 'Never'}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
    </div>
  )
}

function InfoRow({ label, value }: { label: string; value: string }) {
  return (
    <div className="flex items-center justify-between border-b border-netflix-gray/30 py-3 last:border-b-0">
      <span className="text-sm text-gray-400">{label}</span>
      <span className="text-sm font-medium text-white">{value}</span>
    </div>
  )
}

// ── Velox Server: Libraries ───────────────────────────────────────────────────

const LIBRARY_TYPES = [
  { value: 'movies', label: 'Movies', description: 'Feature films', icon: <LuFilm size={20} /> },
  {
    value: 'tvshows',
    label: 'TV Shows',
    description: 'Series & episodes',
    icon: <LuTv size={20} />,
  },
  { value: 'mixed', label: 'Mixed', description: 'Movies & TV', icon: <LuList size={20} /> },
]

const TYPE_ICON_BG: Record<string, string> = {
  movies: 'bg-blue-500/20 text-blue-400',
  tvshows: 'bg-purple-500/20 text-purple-400',
  mixed: 'bg-green-500/20 text-green-400',
}

const TYPE_COLORS: Record<string, string> = {
  movies: 'bg-blue-500/20 text-blue-400 border-blue-500',
  tvshows: 'bg-purple-500/20 text-purple-400 border-purple-500',
  mixed: 'bg-green-500/20 text-green-400 border-green-500',
}

interface LibraryFormData {
  name: string
  paths: string[]
  type: string
}

const DEFAULT_LIB_FORM: LibraryFormData = { name: '', paths: [''], type: 'movies' }

function LibrariesSection() {
  const { data: libraries, isLoading } = useLibraries()
  const { mutate: createLibrary, isPending: isCreating } = useCreateLibrary()
  const { mutate: deleteLibrary } = useDeleteLibrary()
  const { mutate: scanLibrary } = useScanLibrary()

  const [showAddModal, setShowAddModal] = useState(false)
  const [dirPickerIndex, setDirPickerIndex] = useState<number | null>(null)
  const [formData, setFormData] = useState<LibraryFormData>(DEFAULT_LIB_FORM)
  const [formError, setFormError] = useState('')
  const [scanningId, setScanningId] = useState<number | null>(null)

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault()
    setFormError('')
    if (!formData.name.trim()) {
      setFormError('Library name is required')
      return
    }
    const validPaths = formData.paths.map((p) => p.trim()).filter(Boolean)
    if (validPaths.length === 0) {
      setFormError('At least one folder path is required')
      return
    }
    createLibrary(
      { name: formData.name.trim(), type: formData.type, paths: validPaths },
      {
        onSuccess: () => {
          setShowAddModal(false)
          setFormData(DEFAULT_LIB_FORM)
        },
        onError: (err: Error) => setFormError(err.message || 'Failed to create library'),
      },
    )
  }

  const handleDelete = (id: number, name: string) => {
    if (confirm(`Delete "${name}"? This cannot be undone.`)) deleteLibrary(id)
  }

  const handleScan = (id: number, force = false) => {
    setScanningId(id)
    scanLibrary({ id, force }, { onSettled: () => setScanningId(null) })
  }

  const setPath = (idx: number, value: string) => {
    const next = [...formData.paths]
    next[idx] = value
    setFormData({ ...formData, paths: next })
  }

  const addPath = () => setFormData({ ...formData, paths: [...formData.paths, ''] })
  const removePath = (idx: number) => {
    if (formData.paths.length <= 1) return
    setFormData({ ...formData, paths: formData.paths.filter((_, i) => i !== idx) })
  }

  const typeOption = (v: string) => LIBRARY_TYPES.find((t) => t.value === v)

  return (
    <div className="max-w-3xl">
      <div className="flex items-center justify-between">
        <SectionHeader
          title="Libraries"
          description={`${libraries?.length || 0} ${(libraries?.length || 0) === 1 ? 'library' : 'libraries'} configured`}
        />
        <button
          onClick={() => setShowAddModal(true)}
          className="flex items-center gap-2 rounded bg-netflix-red px-4 py-2 text-sm font-semibold text-white transition-colors hover:bg-netflix-red-hover"
        >
          <LuPlus size={16} />
          Add Library
        </button>
      </div>

      {isLoading ? (
        <Spinner />
      ) : libraries?.length === 0 ? (
        <div className="mt-6 flex h-40 flex-col items-center justify-center rounded-lg bg-netflix-dark">
          <LuLibrary size={36} className="text-gray-600" />
          <p className="mt-2 text-sm text-gray-400">No libraries configured</p>
          <button
            onClick={() => setShowAddModal(true)}
            className="mt-3 rounded bg-netflix-red px-4 py-2 text-sm font-medium text-white hover:bg-netflix-red-hover"
          >
            Add Library
          </button>
        </div>
      ) : (
        <div className="mt-6 space-y-3">
          {libraries?.map((lib) => {
            const opt = typeOption(lib.type)
            return (
              <div
                key={lib.id}
                className="flex items-center justify-between rounded-lg bg-netflix-dark p-4"
              >
                <div className="flex min-w-0 items-center gap-3">
                  <div
                    className={`flex h-10 w-10 shrink-0 items-center justify-center rounded-lg ${TYPE_ICON_BG[lib.type] ?? 'bg-gray-500/20 text-gray-400'}`}
                  >
                    {opt?.icon}
                  </div>
                  <div className="min-w-0">
                    <h3 className="text-sm font-semibold text-white">{lib.name}</h3>
                    <div className="mt-0.5">
                      {lib.paths?.map((p) => (
                        <p key={p} className="truncate font-mono text-xs text-gray-400">
                          {p}
                        </p>
                      ))}
                    </div>
                    <div className="mt-1 flex items-center gap-2">
                      <span
                        className={`rounded border px-1.5 py-0.5 text-[10px] ${TYPE_COLORS[lib.type] ?? 'bg-gray-500/20 text-gray-400 border-gray-500'}`}
                      >
                        {opt?.label ?? lib.type}
                      </span>
                    </div>
                  </div>
                </div>
                <div className="flex shrink-0 items-center gap-2">
                  <button
                    onClick={() => handleScan(lib.id)}
                    disabled={scanningId === lib.id}
                    className="flex items-center gap-1.5 rounded bg-netflix-gray px-3 py-1.5 text-xs text-white transition-colors hover:bg-blue-600 disabled:opacity-50"
                  >
                    {scanningId === lib.id ? (
                      <>
                        <div className="h-3 w-3 animate-spin rounded-full border-2 border-white border-t-transparent" />
                        Scanning
                      </>
                    ) : (
                      <>
                        <LuRefreshCw size={13} /> Scan
                      </>
                    )}
                  </button>
                  <button
                    onClick={() => handleScan(lib.id, true)}
                    disabled={scanningId === lib.id}
                    className="flex items-center gap-1.5 rounded bg-netflix-gray px-3 py-1.5 text-xs text-white transition-colors hover:bg-amber-600 disabled:opacity-50"
                    title="Re-parse all filenames and update titles"
                  >
                    <LuRefreshCw size={13} /> Force Rescan
                  </button>
                  <button
                    onClick={() => handleDelete(lib.id, lib.name)}
                    className="rounded bg-netflix-gray p-1.5 text-white transition-colors hover:bg-red-600"
                  >
                    <LuTrash2 size={13} />
                  </button>
                </div>
              </div>
            )
          })}
        </div>
      )}

      {dirPickerIndex !== null && (
        <DirectoryPicker
          onSelect={(path) => {
            setPath(dirPickerIndex, path)
            setDirPickerIndex(null)
          }}
          onClose={() => setDirPickerIndex(null)}
        />
      )}

      {showAddModal && (
        <Modal title="Add New Library" onClose={() => setShowAddModal(false)}>
          <form onSubmit={handleSubmit} className="space-y-5">
            {formError && <ErrorMsg>{formError}</ErrorMsg>}
            <Field label="Library Name">
              <input
                type="text"
                value={formData.name}
                onChange={(e) => setFormData({ ...formData, name: e.target.value })}
                placeholder="e.g., My Movies"
                className={inputClass}
                required
              />
            </Field>
            <Field label="Content Type">
              <div className="grid grid-cols-3 gap-2">
                {LIBRARY_TYPES.map((t) => {
                  const sel = formData.type === t.value
                  const c: Record<string, string> = {
                    movies: sel
                      ? 'border-blue-500 bg-blue-500/15 text-blue-300'
                      : 'border-white/10 bg-netflix-gray text-gray-400 hover:border-white/20',
                    tvshows: sel
                      ? 'border-purple-500 bg-purple-500/15 text-purple-300'
                      : 'border-white/10 bg-netflix-gray text-gray-400 hover:border-white/20',
                    mixed: sel
                      ? 'border-green-500 bg-green-500/15 text-green-300'
                      : 'border-white/10 bg-netflix-gray text-gray-400 hover:border-white/20',
                  }
                  return (
                    <button
                      key={t.value}
                      type="button"
                      onClick={() => setFormData({ ...formData, type: t.value })}
                      className={`flex flex-col items-center gap-1.5 rounded-lg border-2 px-3 py-3 text-center transition-colors ${c[t.value]}`}
                    >
                      {t.icon}
                      <span className="text-sm font-medium">{t.label}</span>
                      <span className="text-xs opacity-70">{t.description}</span>
                    </button>
                  )
                })}
              </div>
            </Field>
            <div>
              <div className="mb-2 flex items-center justify-between">
                <label className="text-sm font-medium text-gray-400">Folders</label>
                <button
                  type="button"
                  onClick={addPath}
                  className="flex items-center gap-1 text-xs text-gray-400 hover:text-white"
                >
                  <LuPlus size={14} /> Add folder
                </button>
              </div>
              <div className="space-y-2">
                {formData.paths.map((p, idx) => (
                  <div key={idx} className="flex gap-2">
                    <input
                      type="text"
                      value={p}
                      onChange={(e) => setPath(idx, e.target.value)}
                      placeholder="/media/movies"
                      className="min-w-0 flex-1 rounded bg-netflix-gray px-4 py-2.5 font-mono text-sm text-white outline-none ring-1 ring-transparent focus:ring-netflix-red"
                    />
                    <button
                      type="button"
                      onClick={() => setDirPickerIndex(idx)}
                      className="shrink-0 rounded bg-netflix-gray px-3 py-2.5 text-gray-300 hover:bg-gray-600 hover:text-white"
                    >
                      <LuFolder size={16} />
                    </button>
                    {formData.paths.length > 1 && (
                      <button
                        type="button"
                        onClick={() => removePath(idx)}
                        className="shrink-0 rounded bg-netflix-gray px-3 py-2.5 text-gray-500 hover:bg-red-600/20 hover:text-red-400"
                      >
                        <LuX size={16} />
                      </button>
                    )}
                  </div>
                ))}
              </div>
            </div>
            <div className="flex gap-3 pt-2">
              <button
                type="button"
                onClick={() => {
                  setShowAddModal(false)
                  setFormData(DEFAULT_LIB_FORM)
                }}
                className="flex-1 rounded bg-netflix-gray px-4 py-2.5 font-medium text-white hover:bg-gray-600"
              >
                Cancel
              </button>
              <button
                type="submit"
                disabled={isCreating}
                className="flex-1 rounded bg-netflix-red px-4 py-2.5 font-medium text-white hover:bg-netflix-red-hover disabled:opacity-50"
              >
                {isCreating ? 'Creating...' : 'Create Library'}
              </button>
            </div>
          </form>
        </Modal>
      )}
    </div>
  )
}

// ── Velox Server: Users ───────────────────────────────────────────────────────

function UsersSection() {
  const { user: currentUser } = useAuthStore()
  const { data: users, isLoading } = useUsers()
  const [isCreateOpen, setIsCreateOpen] = useState(false)
  const [editingUser, setEditingUser] = useState<User | null>(null)

  return (
    <div className="max-w-3xl">
      <div className="flex items-center justify-between">
        <SectionHeader
          title="Users"
          description={`${users?.length || 0} ${(users?.length || 0) === 1 ? 'user' : 'users'}`}
        />
        <button
          onClick={() => setIsCreateOpen(true)}
          className="flex items-center gap-2 rounded bg-netflix-red px-4 py-2 text-sm font-semibold text-white transition-colors hover:bg-netflix-red-hover"
        >
          <LuPlus size={16} />
          Add User
        </button>
      </div>

      {isLoading ? (
        <Spinner />
      ) : (
        <div className="mt-6 overflow-hidden rounded-xl bg-netflix-dark">
          <table className="w-full">
            <thead className="border-b border-netflix-gray bg-netflix-black/50">
              <tr>
                <th className="px-4 py-3 text-left text-xs font-medium text-gray-400">User</th>
                <th className="px-4 py-3 text-left text-xs font-medium text-gray-400">Role</th>
                <th className="px-4 py-3 text-left text-xs font-medium text-gray-400">Created</th>
                <th className="px-4 py-3 text-left text-xs font-medium text-gray-400">Actions</th>
              </tr>
            </thead>
            <tbody>
              {users?.map((u) => (
                <tr
                  key={u.id}
                  className="border-b border-netflix-gray/50 last:border-b-0 hover:bg-netflix-gray/30"
                >
                  <td className="px-4 py-3">
                    <div className="flex items-center gap-3">
                      <div className="flex h-8 w-8 items-center justify-center rounded-full bg-netflix-gray text-sm font-medium text-white">
                        {(u.display_name || u.username).charAt(0).toUpperCase()}
                      </div>
                      <div>
                        <p className="text-sm font-medium text-white">
                          {u.display_name || u.username}
                        </p>
                        <p className="text-xs text-gray-500">{u.username}</p>
                      </div>
                    </div>
                  </td>
                  <td className="px-4 py-3">
                    <span
                      className={`rounded px-2 py-0.5 text-xs font-medium ${
                        u.is_admin
                          ? 'bg-purple-500/20 text-purple-400'
                          : 'bg-blue-500/20 text-blue-400'
                      }`}
                    >
                      {u.is_admin ? 'Admin' : 'User'}
                    </span>
                  </td>
                  <td className="px-4 py-3 text-xs text-gray-400">
                    {new Date(u.created_at).toLocaleDateString()}
                  </td>
                  <td className="px-4 py-3">
                    <div className="flex gap-2">
                      <button
                        onClick={() => setEditingUser(u)}
                        className="rounded bg-netflix-gray px-3 py-1 text-xs text-white hover:bg-gray-600"
                      >
                        Edit
                      </button>
                      {currentUser && u.id !== currentUser.id && <DeleteUserBtn userId={u.id} />}
                    </div>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
          {users?.length === 0 && (
            <p className="py-8 text-center text-sm text-gray-400">No users found</p>
          )}
        </div>
      )}

      {isCreateOpen && <CreateUserModal onClose={() => setIsCreateOpen(false)} />}
      {editingUser && <EditUserModal user={editingUser} onClose={() => setEditingUser(null)} />}
    </div>
  )
}

function CreateUserModal({ onClose }: { onClose: () => void }) {
  const { mutate: createUser, isPending } = useCreateUser()
  const [form, setForm] = useState({
    username: '',
    password: '',
    display_name: '',
    is_admin: false,
  })
  const [error, setError] = useState('')

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault()
    setError('')
    if (!form.username || !form.password) {
      setError('Username and password are required')
      return
    }
    if (form.password.length < 8) {
      setError('Password must be at least 8 characters')
      return
    }
    createUser(form, { onSuccess: onClose, onError: (err: Error) => setError(err.message) })
  }

  return (
    <Modal title="Add User" onClose={onClose}>
      {error && <ErrorMsg>{error}</ErrorMsg>}
      <form onSubmit={handleSubmit} className="space-y-4">
        <Field label="Username *">
          <input
            type="text"
            value={form.username}
            onChange={(e) => setForm({ ...form, username: e.target.value })}
            className={inputClass}
            required
          />
        </Field>
        <Field label="Display Name">
          <input
            type="text"
            value={form.display_name}
            onChange={(e) => setForm({ ...form, display_name: e.target.value })}
            className={inputClass}
          />
        </Field>
        <Field label="Password *">
          <input
            type="password"
            value={form.password}
            onChange={(e) => setForm({ ...form, password: e.target.value })}
            className={inputClass}
            required
            minLength={8}
          />
        </Field>
        <div className="flex items-center gap-3">
          <input
            type="checkbox"
            id="new_is_admin"
            checked={form.is_admin}
            onChange={(e) => setForm({ ...form, is_admin: e.target.checked })}
            className="h-4 w-4 rounded"
          />
          <label htmlFor="new_is_admin" className="text-sm text-gray-300">
            Administrator
          </label>
        </div>
        <div className="flex justify-end gap-3 pt-2">
          <button
            type="button"
            onClick={onClose}
            className="rounded px-4 py-2 text-gray-300 hover:bg-netflix-gray hover:text-white"
          >
            Cancel
          </button>
          <button
            type="submit"
            disabled={isPending}
            className="rounded bg-netflix-red px-4 py-2 font-semibold text-white hover:bg-netflix-red-hover disabled:opacity-50"
          >
            {isPending ? 'Creating...' : 'Create'}
          </button>
        </div>
      </form>
    </Modal>
  )
}

function EditUserModal({ user, onClose }: { user: User; onClose: () => void }) {
  const { mutate: updateUser, isPending } = useUpdateUser()
  const [form, setForm] = useState({
    display_name: user.display_name || '',
    is_admin: user.is_admin,
    password: '',
  })
  const [error, setError] = useState('')

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault()
    setError('')
    const data: { display_name?: string; is_admin?: boolean; password?: string } = {
      display_name: form.display_name,
      is_admin: form.is_admin,
    }
    if (form.password) {
      if (form.password.length < 8) {
        setError('Password must be at least 8 characters')
        return
      }
      data.password = form.password
    }
    updateUser(
      { id: user.id, data },
      { onSuccess: onClose, onError: (err: Error) => setError(err.message) },
    )
  }

  return (
    <Modal title="Edit User" onClose={onClose}>
      {error && <ErrorMsg>{error}</ErrorMsg>}
      <form onSubmit={handleSubmit} className="space-y-4">
        <Field label="Username">
          <input type="text" value={user.username} disabled className={inputDisabled} />
        </Field>
        <Field label="Display Name">
          <input
            type="text"
            value={form.display_name}
            onChange={(e) => setForm({ ...form, display_name: e.target.value })}
            className={inputClass}
          />
        </Field>
        <Field label="New Password (optional)">
          <input
            type="password"
            value={form.password}
            onChange={(e) => setForm({ ...form, password: e.target.value })}
            placeholder="Leave blank to keep current"
            className={inputClass}
          />
        </Field>
        <div className="flex items-center gap-3">
          <input
            type="checkbox"
            id="edit_is_admin"
            checked={form.is_admin}
            onChange={(e) => setForm({ ...form, is_admin: e.target.checked })}
            className="h-4 w-4 rounded"
          />
          <label htmlFor="edit_is_admin" className="text-sm text-gray-300">
            Administrator
          </label>
        </div>
        <div className="flex justify-end gap-3 pt-2">
          <button
            type="button"
            onClick={onClose}
            className="rounded px-4 py-2 text-gray-300 hover:bg-netflix-gray hover:text-white"
          >
            Cancel
          </button>
          <button
            type="submit"
            disabled={isPending}
            className="rounded bg-netflix-red px-4 py-2 font-semibold text-white hover:bg-netflix-red-hover disabled:opacity-50"
          >
            {isPending ? 'Saving...' : 'Save'}
          </button>
        </div>
      </form>
    </Modal>
  )
}

function DeleteUserBtn({ userId }: { userId: number }) {
  const { mutate: deleteUser, isPending } = useDeleteUser()
  const [confirm, setConfirm] = useState(false)

  if (confirm) {
    return (
      <div className="flex gap-1">
        <button
          onClick={() => deleteUser(userId)}
          disabled={isPending}
          className="rounded bg-netflix-red px-2.5 py-1 text-xs text-white hover:bg-netflix-red-hover"
        >
          Confirm
        </button>
        <button
          onClick={() => setConfirm(false)}
          className="rounded bg-netflix-gray px-2.5 py-1 text-xs text-gray-300 hover:bg-gray-600"
        >
          Cancel
        </button>
      </div>
    )
  }

  return (
    <button
      onClick={() => setConfirm(true)}
      className="rounded bg-red-500/20 px-3 py-1 text-xs text-red-400 hover:bg-red-500/30"
    >
      Delete
    </button>
  )
}

// ── Velox Server: Activity ────────────────────────────────────────────────────

const ACTION_BADGES: Record<string, string> = {
  login: 'bg-blue-500/20 text-blue-400',
  play_start: 'bg-green-500/20 text-green-400',
  play_stop: 'bg-yellow-500/20 text-yellow-400',
  library_scan: 'bg-purple-500/20 text-purple-400',
  media_added: 'bg-teal-500/20 text-teal-400',
}

function ActivitySection() {
  const [action, setAction] = useState('')
  const [dateFrom, setDateFrom] = useState('')
  const [dateTo, setDateTo] = useState('')
  const [limit, setLimit] = useState('25')

  const filters: Record<string, string> = {}
  if (action) filters.action = action
  if (dateFrom) filters.from = dateFrom
  if (dateTo) filters.to = dateTo
  if (limit) filters.limit = limit

  const { data: logs, isLoading } = useActivity(filters)

  return (
    <div className="max-w-4xl">
      <SectionHeader title="Activity" description="Recent server activity and user actions" />

      {/* Filters */}
      <div className="mt-6 flex flex-wrap items-end gap-3">
        <div>
          <label className="mb-1 block text-xs text-gray-400">Action</label>
          <select
            value={action}
            onChange={(e) => setAction(e.target.value)}
            className={selectClass + ' !w-auto min-w-[140px]'}
          >
            <option value="">All Actions</option>
            <option value="login">Login</option>
            <option value="play_start">Play Start</option>
            <option value="play_stop">Play Stop</option>
            <option value="library_scan">Library Scan</option>
            <option value="media_added">Media Added</option>
          </select>
        </div>
        <div>
          <label className="mb-1 block text-xs text-gray-400">From</label>
          <input
            type="date"
            value={dateFrom}
            onChange={(e) => setDateFrom(e.target.value)}
            className={inputClass + ' !w-auto'}
          />
        </div>
        <div>
          <label className="mb-1 block text-xs text-gray-400">To</label>
          <input
            type="date"
            value={dateTo}
            onChange={(e) => setDateTo(e.target.value)}
            className={inputClass + ' !w-auto'}
          />
        </div>
        <div>
          <label className="mb-1 block text-xs text-gray-400">Limit</label>
          <select
            value={limit}
            onChange={(e) => setLimit(e.target.value)}
            className={selectClass + ' !w-auto min-w-[80px]'}
          >
            <option value="25">25</option>
            <option value="50">50</option>
            <option value="100">100</option>
          </select>
        </div>
      </div>

      {/* Activity Table */}
      {isLoading ? (
        <Spinner />
      ) : (
        <div className="mt-4 overflow-hidden rounded-xl bg-netflix-dark">
          <table className="w-full">
            <thead className="border-b border-netflix-gray bg-netflix-black/50">
              <tr>
                <th className="px-4 py-3 text-left text-xs font-medium text-gray-400">Time</th>
                <th className="px-4 py-3 text-left text-xs font-medium text-gray-400">User</th>
                <th className="px-4 py-3 text-left text-xs font-medium text-gray-400">Action</th>
                <th className="px-4 py-3 text-left text-xs font-medium text-gray-400">Media</th>
                <th className="px-4 py-3 text-left text-xs font-medium text-gray-400">IP</th>
              </tr>
            </thead>
            <tbody>
              {logs?.map((log) => (
                <tr
                  key={log.id}
                  className="border-b border-netflix-gray/50 last:border-b-0 hover:bg-netflix-gray/30"
                >
                  <td className="whitespace-nowrap px-4 py-3 text-xs text-gray-400">
                    {new Date(log.created_at).toLocaleString()}
                  </td>
                  <td className="px-4 py-3 text-sm text-white">{log.username ?? 'System'}</td>
                  <td className="px-4 py-3">
                    <span
                      className={`rounded px-2 py-0.5 text-xs font-medium ${ACTION_BADGES[log.action] ?? 'bg-gray-500/20 text-gray-400'}`}
                    >
                      {log.action}
                    </span>
                  </td>
                  <td className="max-w-[200px] truncate px-4 py-3 text-sm text-gray-300">
                    {log.media_title ?? '-'}
                  </td>
                  <td className="whitespace-nowrap px-4 py-3 font-mono text-xs text-gray-500">
                    {log.ip_address}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
          {(!logs || logs.length === 0) && (
            <p className="py-8 text-center text-sm text-gray-400">No activity found</p>
          )}
        </div>
      )}
    </div>
  )
}

// ── Velox Server: Tasks ──────────────────────────────────────────────────────

function TasksSection() {
  const { data: tasks, isLoading } = useScheduledTasks()
  const { mutate: runTask, isPending: isRunning } = useRunTask()
  const [runningTask, setRunningTask] = useState<string | null>(null)

  const handleRun = (name: string) => {
    setRunningTask(name)
    runTask(name, { onSettled: () => setRunningTask(null) })
  }

  return (
    <div className="max-w-3xl">
      <SectionHeader title="Scheduled Tasks" description="Background tasks and maintenance jobs" />

      {isLoading ? (
        <Spinner />
      ) : (
        <div className="mt-6 overflow-hidden rounded-xl bg-netflix-dark">
          <table className="w-full">
            <thead className="border-b border-netflix-gray bg-netflix-black/50">
              <tr>
                <th className="px-4 py-3 text-left text-xs font-medium text-gray-400">Task</th>
                <th className="px-4 py-3 text-left text-xs font-medium text-gray-400">Interval</th>
                <th className="px-4 py-3 text-left text-xs font-medium text-gray-400">Last Run</th>
                <th className="px-4 py-3 text-left text-xs font-medium text-gray-400">Next Run</th>
                <th className="px-4 py-3 text-left text-xs font-medium text-gray-400">Status</th>
                <th className="px-4 py-3 text-left text-xs font-medium text-gray-400">Actions</th>
              </tr>
            </thead>
            <tbody>
              {tasks?.map((task) => (
                <tr
                  key={task.name}
                  className="border-b border-netflix-gray/50 last:border-b-0 hover:bg-netflix-gray/30"
                >
                  <td className="px-4 py-3 text-sm font-medium text-white">{task.name}</td>
                  <td className="px-4 py-3 text-sm text-gray-300">{task.interval}</td>
                  <td className="px-4 py-3 text-xs text-gray-400">
                    {task.last_run ? timeAgo(task.last_run) : 'Never'}
                  </td>
                  <td className="px-4 py-3 text-xs text-gray-400">
                    {new Date(task.next_run).toLocaleString()}
                  </td>
                  <td className="px-4 py-3">
                    {task.running ? (
                      <span className="flex items-center gap-1.5 text-xs text-yellow-400">
                        <div className="h-3 w-3 animate-spin rounded-full border-2 border-yellow-400 border-t-transparent" />
                        Running
                      </span>
                    ) : (
                      <span className="text-xs text-gray-500">Idle</span>
                    )}
                  </td>
                  <td className="px-4 py-3">
                    <button
                      onClick={() => handleRun(task.name)}
                      disabled={task.running || (isRunning && runningTask === task.name)}
                      className="flex items-center gap-1.5 rounded bg-netflix-gray px-3 py-1.5 text-xs text-white transition-colors hover:bg-blue-600 disabled:opacity-50"
                    >
                      <LuPlay size={12} />
                      Run Now
                    </button>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
          {(!tasks || tasks.length === 0) && (
            <p className="py-8 text-center text-sm text-gray-400">No scheduled tasks</p>
          )}
        </div>
      )}
    </div>
  )
}

// ── Velox Server: Webhooks ────────────────────────────────────────────────────

const WEBHOOK_EVENTS = ['media_added', 'playback_start', 'scan_complete', 'error']

function WebhooksSection() {
  const { data: webhooks, isLoading } = useWebhooks()
  const { mutate: createWebhook, isPending: isCreating } = useCreateWebhook()
  const { mutate: updateWebhook } = useUpdateWebhook()
  const { mutate: deleteWebhook } = useDeleteWebhook()
  const [showAddModal, setShowAddModal] = useState(false)
  const [formUrl, setFormUrl] = useState('')
  const [formEvents, setFormEvents] = useState<string[]>([])
  const [formError, setFormError] = useState('')

  const toggleEvent = (event: string) => {
    setFormEvents((prev) =>
      prev.includes(event) ? prev.filter((e) => e !== event) : [...prev, event],
    )
  }

  const handleCreate = (e: React.FormEvent) => {
    e.preventDefault()
    setFormError('')
    if (!formUrl.trim()) {
      setFormError('URL is required')
      return
    }
    if (formEvents.length === 0) {
      setFormError('Select at least one event')
      return
    }
    createWebhook(
      { url: formUrl.trim(), events: formEvents.join(','), active: true },
      {
        onSuccess: () => {
          setShowAddModal(false)
          setFormUrl('')
          setFormEvents([])
        },
        onError: (err: Error) => setFormError(err.message),
      },
    )
  }

  const handleToggleActive = (webhook: Webhook) => {
    updateWebhook({ id: webhook.id, data: { active: !webhook.active } })
  }

  const handleDelete = (id: number) => {
    if (confirm('Delete this webhook?')) deleteWebhook(id)
  }

  return (
    <div className="max-w-3xl">
      <div className="flex items-center justify-between">
        <SectionHeader
          title="Webhooks"
          description={`${webhooks?.length || 0} webhook${(webhooks?.length || 0) !== 1 ? 's' : ''} configured`}
        />
        <button
          onClick={() => setShowAddModal(true)}
          className="flex items-center gap-2 rounded bg-netflix-red px-4 py-2 text-sm font-semibold text-white transition-colors hover:bg-netflix-red-hover"
        >
          <LuPlus size={16} />
          Add Webhook
        </button>
      </div>

      {isLoading ? (
        <Spinner />
      ) : webhooks?.length === 0 ? (
        <div className="mt-6 flex h-40 flex-col items-center justify-center rounded-lg bg-netflix-dark">
          <LuGlobe size={36} className="text-gray-600" />
          <p className="mt-2 text-sm text-gray-400">No webhooks configured</p>
        </div>
      ) : (
        <div className="mt-6 space-y-3">
          {webhooks?.map((wh) => (
            <div
              key={wh.id}
              className="flex items-center justify-between rounded-lg bg-netflix-dark p-4"
            >
              <div className="min-w-0 flex-1">
                <div className="flex items-center gap-2">
                  <LuGlobe size={14} className="shrink-0 text-gray-400" />
                  <p className="truncate font-mono text-sm text-white">{wh.url}</p>
                </div>
                <div className="mt-1.5 flex flex-wrap gap-1.5">
                  {wh.events.split(',').map((ev) => (
                    <span
                      key={ev}
                      className="rounded bg-netflix-gray px-2 py-0.5 text-[10px] font-medium text-gray-300"
                    >
                      {ev.trim()}
                    </span>
                  ))}
                </div>
              </div>
              <div className="ml-4 flex shrink-0 items-center gap-2">
                <button
                  onClick={() => handleToggleActive(wh)}
                  className={`flex items-center gap-1.5 rounded px-3 py-1.5 text-xs transition-colors ${
                    wh.active
                      ? 'bg-green-500/20 text-green-400 hover:bg-green-500/30'
                      : 'bg-netflix-gray text-gray-400 hover:bg-gray-600'
                  }`}
                >
                  {wh.active ? (
                    <>
                      <LuPause size={12} /> Active
                    </>
                  ) : (
                    <>
                      <LuPlay size={12} /> Inactive
                    </>
                  )}
                </button>
                <button
                  onClick={() => handleDelete(wh.id)}
                  className="rounded bg-netflix-gray p-1.5 text-white transition-colors hover:bg-red-600"
                >
                  <LuTrash2 size={13} />
                </button>
              </div>
            </div>
          ))}
        </div>
      )}

      {showAddModal && (
        <Modal title="Add Webhook" onClose={() => setShowAddModal(false)}>
          <form onSubmit={handleCreate} className="space-y-5">
            {formError && <ErrorMsg>{formError}</ErrorMsg>}
            <Field label="URL">
              <input
                type="url"
                value={formUrl}
                onChange={(e) => setFormUrl(e.target.value)}
                placeholder="https://example.com/webhook"
                className={inputClass}
                required
              />
            </Field>
            <Field label="Events">
              <div className="space-y-2">
                {WEBHOOK_EVENTS.map((event) => (
                  <label key={event} className="flex items-center gap-3">
                    <input
                      type="checkbox"
                      checked={formEvents.includes(event)}
                      onChange={() => toggleEvent(event)}
                      className="h-4 w-4 rounded"
                    />
                    <span className="text-sm text-gray-300">{event}</span>
                  </label>
                ))}
              </div>
            </Field>
            <div className="flex gap-3 pt-2">
              <button
                type="button"
                onClick={() => {
                  setShowAddModal(false)
                  setFormUrl('')
                  setFormEvents([])
                  setFormError('')
                }}
                className="flex-1 rounded bg-netflix-gray px-4 py-2.5 font-medium text-white hover:bg-gray-600"
              >
                Cancel
              </button>
              <button
                type="submit"
                disabled={isCreating}
                className="flex-1 rounded bg-netflix-red px-4 py-2.5 font-medium text-white hover:bg-netflix-red-hover disabled:opacity-50"
              >
                {isCreating ? 'Creating...' : 'Create Webhook'}
              </button>
            </div>
          </form>
        </Modal>
      )}
    </div>
  )
}

// ── Shared UI Primitives ──────────────────────────────────────────────────────

const inputClass =
  'w-full rounded bg-netflix-gray px-4 py-2.5 text-sm text-white outline-none ring-1 ring-transparent transition-all placeholder:text-gray-500 focus:ring-netflix-red'
const inputDisabled = 'w-full rounded bg-netflix-black px-4 py-2.5 text-sm text-gray-500'
const selectClass =
  'w-full rounded bg-netflix-gray px-4 py-2.5 text-sm text-white outline-none ring-1 ring-transparent transition-all focus:ring-netflix-red'

function SectionHeader({ title, description }: { title: string; description: string }) {
  return (
    <div>
      <h2 className="text-2xl font-bold text-white">{title}</h2>
      <p className="text-sm text-gray-400">{description}</p>
    </div>
  )
}

function Field({
  label,
  compact,
  children,
}: {
  label: React.ReactNode
  compact?: boolean
  children: React.ReactNode
}) {
  return (
    <div>
      <label
        className={`mb-1.5 block text-sm font-medium text-gray-400 ${compact ? 'text-xs' : ''}`}
      >
        {label}
      </label>
      {children}
    </div>
  )
}

function SaveButton({ isPending, label = 'Save Changes' }: { isPending: boolean; label?: string }) {
  return (
    <button
      type="submit"
      disabled={isPending}
      className="rounded bg-netflix-red px-6 py-2 text-sm font-semibold text-white transition-colors hover:bg-netflix-red-hover disabled:opacity-50"
    >
      {isPending ? 'Saving...' : label}
    </button>
  )
}

function Spinner() {
  return (
    <div className="flex h-40 items-center justify-center">
      <div className="h-8 w-8 animate-spin rounded-full border-2 border-netflix-red border-t-transparent" />
    </div>
  )
}

function ErrorMsg({ children }: { children: React.ReactNode }) {
  return <div className="rounded-lg bg-netflix-red/20 p-3 text-sm text-netflix-red">{children}</div>
}

function SuccessMsg({ children }: { children: React.ReactNode }) {
  return <div className="rounded-lg bg-green-500/20 p-3 text-sm text-green-400">{children}</div>
}

function Modal({
  title,
  onClose,
  children,
}: {
  title: string
  onClose: () => void
  children: React.ReactNode
}) {
  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/70 backdrop-blur-sm p-4">
      <div className="w-full max-w-lg rounded-xl bg-netflix-dark p-6 shadow-2xl">
        <div className="mb-5 flex items-center justify-between">
          <h2 className="text-lg font-bold text-white">{title}</h2>
          <button onClick={onClose} className="text-gray-400 hover:text-white">
            <LuX size={22} />
          </button>
        </div>
        {children}
      </div>
    </div>
  )
}

// ── Cinema Mode Settings ──────────────────────────────────────────────────────

function CinemaSection() {
  const { data: settings, isLoading } = useCinemaSettings()
  const { mutate: updateSettings } = useUpdateCinemaSettings()
  const { mutate: uploadIntro, isPending: isUploading } = useUploadCinemaIntro()
  const fileRef = useRef<HTMLInputElement>(null)
  const [saved, setSaved] = useState(false)

  if (isLoading) return <Spinner />

  return (
    <div className="max-w-2xl space-y-6">
      <SectionHeader
        title="Cinema Mode"
        description="Play trailers and a custom intro video before the main feature"
      />

      {/* Enable toggle */}
      <div className="rounded-lg bg-netflix-dark p-6">
        <div className="flex items-center justify-between">
          <div>
            <h3 className="text-base font-semibold text-white">Enable Cinema Mode</h3>
            <p className="mt-1 text-sm text-gray-400">
              Show trailers and intro before playback. Users can skip at any time.
            </p>
          </div>
          <button
            onClick={() => {
              updateSettings(
                { enabled: !settings?.enabled },
                {
                  onSuccess: () => {
                    setSaved(true)
                    setTimeout(() => setSaved(false), 2000)
                  },
                },
              )
            }}
            className={`relative h-7 w-12 rounded-full transition-colors ${
              settings?.enabled ? 'bg-blue-600' : 'bg-gray-600'
            }`}
          >
            <span
              className={`absolute left-1 top-1 h-5 w-5 rounded-full bg-white transition-transform ${
                settings?.enabled ? 'translate-x-5' : ''
              }`}
            />
          </button>
        </div>
        {saved && <p className="mt-2 text-sm text-green-400">Saved!</p>}
      </div>

      {/* Max trailers */}
      <div className="rounded-lg bg-netflix-dark p-6">
        <h3 className="mb-1 text-base font-semibold text-white">Number of Trailers</h3>
        <p className="mb-3 text-sm text-gray-400">
          How many YouTube trailers to show (fetched from TMDb)
        </p>
        <select
          value={settings?.max_trailers ?? '2'}
          onChange={(e) => updateSettings({ max_trailers: e.target.value })}
          className="rounded-lg bg-[#2a2a2a] px-4 py-2 text-white outline-none"
        >
          <option value="0">None</option>
          <option value="1">1 trailer</option>
          <option value="2">2 trailers</option>
          <option value="3">3 trailers</option>
        </select>
      </div>

      {/* Custom intro video */}
      <div className="rounded-lg bg-netflix-dark p-6">
        <h3 className="mb-1 text-base font-semibold text-white">Custom Intro Video</h3>
        <p className="mb-3 text-sm text-gray-400">
          Upload a short video (e.g. THX-style logo animation) that plays before trailers. Max
          100MB.
        </p>
        <div className="flex items-center gap-4">
          <input
            ref={fileRef}
            type="file"
            accept="video/mp4,video/webm,video/quicktime"
            className="hidden"
            onChange={(e) => {
              const file = e.target.files?.[0]
              if (file) uploadIntro(file)
            }}
          />
          <button
            onClick={() => fileRef.current?.click()}
            disabled={isUploading}
            className="rounded-lg bg-blue-600 px-4 py-2 text-sm font-medium text-white hover:bg-blue-500 disabled:opacity-50"
          >
            {isUploading ? 'Uploading...' : settings?.has_intro ? 'Replace Intro' : 'Upload Intro'}
          </button>
          {settings?.has_intro && (
            <span className="text-sm text-green-400">Intro video uploaded</span>
          )}
        </div>
      </div>
    </div>
  )
}
