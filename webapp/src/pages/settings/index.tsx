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
  LuActivity,
  LuClock,
  LuWebhook,
  LuPlay,
  LuHardDrive,
  LuBell,
  LuFilm,
  LuSkipForward,
} from 'react-icons/lu'
import { useAuthStore } from '@/stores/auth'
import { useTranslation } from '@/hooks/useTranslation'
import { Select } from '@/components/ui/Select'
import { ProfileSection } from './components/ProfileSection'
import { PreferencesSection } from './components/PreferencesSection'
import { SecuritySection } from './components/SecuritySection'
import { SessionsSection } from './components/SessionsSection'
import { NotificationsSection } from './components/NotificationsSection'
import { MetadataSection } from './components/MetadataSection'
import { SubtitlesSection } from './components/SubtitlesSection'
import { PlaybackSection } from './components/PlaybackSection'
import { GeneralSection } from './components/GeneralSection'
import { LibrariesSection } from './components/LibrariesSection'
import { UsersSection } from './components/UsersSection'
import { ActivitySection } from './components/ActivitySection'
import { TasksSection } from './components/TasksSection'
import { WebhooksSection } from './components/WebhooksSection'
import { CinemaSection } from './components/CinemaSection'
import { MarkersSection } from './components/MarkersSection'
import { PretranscodeSection } from './components/PretranscodeSection'

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
  {
    id: 'notifications',
    labelKey: 'sections.notifications.title',
    icon: <LuBell size={18} />,
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
  {
    id: 'pretranscode',
    labelKey: 'sections.pretranscode.title',
    icon: <LuHardDrive size={18} />,
    group: 'Admin Preferences',
    adminOnly: true,
  },
  {
    id: 'markers',
    labelKey: 'sections.markers.title',
    icon: <LuSkipForward size={18} />,
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

const SECTION_COMPONENTS: Record<string, React.FC> = {
  profile: ProfileSection,
  preferences: PreferencesSection,
  security: SecuritySection,
  sessions: SessionsSection,
  notifications: NotificationsSection,
  metadata: MetadataSection,
  subtitles: SubtitlesSection,
  playback: PlaybackSection,
  cinema: CinemaSection,
  pretranscode: PretranscodeSection,
  markers: MarkersSection,
  general: GeneralSection,
  libraries: LibrariesSection,
  users: UsersSection,
  activity: ActivitySection,
  tasks: TasksSection,
  webhooks: WebhooksSection,
}

// ── Main Page ─────────────────────────────────────────────────────────────────

export default function SettingsPage() {
  const { user } = useAuthStore()
  const [searchParams, setSearchParams] = useSearchParams()
  const activeSection = searchParams.get('section') || 'profile'
  const isAdmin = user?.is_admin ?? false
  const { t } = useTranslation('settings')

  const setSection = (id: string) => setSearchParams({ section: id })

  const sections = ALL_SECTIONS.filter((s) => !s.adminOnly || isAdmin)
  const isAllowed = sections.some((s) => s.id === activeSection)
  const actualSection = isAllowed ? activeSection : 'profile'

  const groups = sections.reduce<Record<string, Section[]>>((acc, s) => {
    ;(acc[s.group] ??= []).push(s)
    return acc
  }, {})

  const ActiveComponent = SECTION_COMPONENTS[actualSection]

  return (
    <div className="flex min-h-[calc(100vh-4rem)] flex-col md:flex-row">
      {/* Mobile: dropdown navigation */}
      <div className="border-b border-netflix-gray/50 bg-netflix-black/50 p-3 md:hidden">
        <Select
          value={actualSection}
          onChange={(e) => setSection(e.target.value)}
          className="w-full bg-[#1a1a1a] px-4 py-3 font-medium transition-all focus:border-netflix-red"
        >
          {Object.entries(groups).map(([group, items]) => (
            <optgroup key={group} label={t(`groups.${group.toLowerCase().replace(/ /g, '')}`)}>
              {items.map((item) => (
                <option key={item.id} value={item.id}>
                  {t(item.labelKey)}
                </option>
              ))}
            </optgroup>
          ))}
        </Select>
      </div>

      {/* Desktop: sidebar */}
      <aside className="hidden w-56 shrink-0 border-r border-netflix-gray/50 bg-netflix-black/50 md:block">
        <div
          className="sticky top-16 overflow-y-auto py-4"
          style={{ maxHeight: 'calc(100vh - 4rem)' }}
        >
          {Object.entries(groups).map(([group, items], gi) => (
            <div key={group} className={gi > 0 ? 'mt-3' : ''}>
              <p className="px-4 py-1.5 text-[10px] font-semibold uppercase tracking-wider text-gray-500">
                {t(`groups.${group.toLowerCase().replace(/ /g, '')}`)}
              </p>
              {items.map((item) => (
                <button
                  key={item.id}
                  onClick={() => setSection(item.id)}
                  className={`flex w-full items-center gap-2.5 px-4 py-2 text-[13px] transition-colors ${
                    actualSection === item.id
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
      <main className="flex-1 overflow-y-auto p-4 md:p-6 lg:p-8">
        {ActiveComponent && <ActiveComponent />}
      </main>
    </div>
  )
}
