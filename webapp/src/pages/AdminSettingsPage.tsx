import { useState, useEffect } from 'react'
import { useSearchParams } from 'react-router'
import {
  LuCaptions,
  LuSave,
  LuCheck,
  LuEye,
  LuEyeOff,
  LuServer,
  LuChevronRight,
} from 'react-icons/lu'
import { useOpenSubsSettings, useUpdateOpenSubsSettings } from '@/hooks/stores/useSettings'

// ── Sidebar sections ──────────────────────────────────────────────────────────

interface SettingsSection {
  id: string
  label: string
  icon: React.ReactNode
  group: string
}

const SECTIONS: SettingsSection[] = [
  { id: 'general', label: 'General', icon: <LuServer size={18} />, group: 'Server' },
  { id: 'subtitles', label: 'Subtitles', icon: <LuCaptions size={18} />, group: 'Services' },
]

// ── Main Page ─────────────────────────────────────────────────────────────────

export function AdminSettingsPage() {
  const [searchParams, setSearchParams] = useSearchParams()
  const activeSection = searchParams.get('section') || 'general'

  const setSection = (id: string) => setSearchParams({ section: id })

  const groups = SECTIONS.reduce<Record<string, SettingsSection[]>>((acc, s) => {
    ;(acc[s.group] ??= []).push(s)
    return acc
  }, {})

  return (
    <div className="flex min-h-[calc(100vh-4rem)] gap-0">
      {/* Sidebar */}
      <aside className="w-56 shrink-0 border-r border-netflix-gray/50 bg-netflix-black/50 py-4">
        <div className="px-4 pb-3">
          <h1 className="text-lg font-bold text-white">Settings</h1>
        </div>
        {Object.entries(groups).map(([group, items]) => (
          <div key={group} className="mb-2">
            <p className="px-4 py-1.5 text-[10px] font-semibold uppercase tracking-wider text-gray-500">
              {group}
            </p>
            {items.map((item) => (
              <button
                key={item.id}
                onClick={() => setSection(item.id)}
                className={`flex w-full items-center gap-2.5 px-4 py-2 text-sm transition-colors ${
                  activeSection === item.id
                    ? 'bg-netflix-gray/60 text-white font-medium'
                    : 'text-gray-400 hover:bg-netflix-gray/30 hover:text-white'
                }`}
              >
                <span className="shrink-0">{item.icon}</span>
                <span className="flex-1 text-left">{item.label}</span>
                {activeSection === item.id && (
                  <LuChevronRight size={14} className="shrink-0 text-gray-500" />
                )}
              </button>
            ))}
          </div>
        ))}
      </aside>

      {/* Content */}
      <main className="flex-1 overflow-y-auto p-6">
        {activeSection === 'general' && <GeneralSection />}
        {activeSection === 'subtitles' && <SubtitlesSection />}
      </main>
    </div>
  )
}

// ── General Section ───────────────────────────────────────────────────────────

function GeneralSection() {
  return (
    <div className="max-w-2xl space-y-6">
      <SectionHeader title="General" description="Server information and status" />
      <div className="rounded-lg bg-netflix-dark p-6">
        <div className="space-y-3">
          <InfoRow label="Server" value="Velox Media Server" />
          <InfoRow label="Version" value="0.1.0-dev" />
          <InfoRow label="Database" value="SQLite (WAL mode)" />
          <InfoRow label="Transcoder" value="FFmpeg 8.0" />
        </div>
      </div>
    </div>
  )
}

function InfoRow({ label, value }: { label: string; value: string }) {
  return (
    <div className="flex items-center justify-between border-b border-netflix-gray/30 py-2.5 last:border-b-0">
      <span className="text-sm text-gray-400">{label}</span>
      <span className="text-sm font-medium text-white">{value}</span>
    </div>
  )
}

// ── Subtitles Section ─────────────────────────────────────────────────────────

function SubtitlesSection() {
  const { data: settings, isLoading } = useOpenSubsSettings()
  const { mutate: updateSettings, isPending: isSaving } = useUpdateOpenSubsSettings()

  const [apiKey, setApiKey] = useState('')
  const [username, setUsername] = useState('')
  const [password, setPassword] = useState('')
  const [showPassword, setShowPassword] = useState(false)
  const [saved, setSaved] = useState(false)

  useEffect(() => {
    if (settings) {
      setApiKey(settings.api_key)
      setUsername(settings.username)
    }
  }, [settings])

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
    <div className="max-w-2xl space-y-6">
      <SectionHeader title="Subtitles" description="Configure external subtitle providers" />

      {/* OpenSubtitles */}
      <div className="rounded-lg bg-netflix-dark p-6">
        <div className="mb-1 flex items-center gap-2">
          <h3 className="text-base font-semibold text-white">OpenSubtitles.com</h3>
          {settings?.password_set && settings?.api_key && (
            <span className="rounded bg-green-500/20 px-2 py-0.5 text-[10px] font-medium text-green-400">
              Connected
            </span>
          )}
        </div>
        <p className="mb-5 text-sm text-gray-400">
          Connect your OpenSubtitles account to search and download subtitles. Get a free API key at{' '}
          <span className="text-blue-400">opensubtitles.com</span>.
        </p>

        <form onSubmit={handleSave} className="space-y-4">
          <div>
            <label className="mb-1.5 block text-sm font-medium text-gray-400">API Key</label>
            <input
              type="text"
              value={apiKey}
              onChange={(e) => setApiKey(e.target.value)}
              placeholder="Your OpenSubtitles API key"
              className="w-full rounded bg-netflix-gray px-4 py-2.5 text-sm text-white outline-none ring-1 ring-transparent transition-all placeholder:text-gray-500 focus:ring-netflix-red"
            />
          </div>

          <div>
            <label className="mb-1.5 block text-sm font-medium text-gray-400">Username</label>
            <input
              type="text"
              value={username}
              onChange={(e) => setUsername(e.target.value)}
              placeholder="OpenSubtitles username"
              className="w-full rounded bg-netflix-gray px-4 py-2.5 text-sm text-white outline-none ring-1 ring-transparent transition-all placeholder:text-gray-500 focus:ring-netflix-red"
            />
          </div>

          <div>
            <label className="mb-1.5 block text-sm font-medium text-gray-400">
              Password
              {settings?.password_set && (
                <span className="ml-2 text-xs text-green-400">(configured)</span>
              )}
            </label>
            <div className="relative">
              <input
                type={showPassword ? 'text' : 'password'}
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                placeholder={
                  settings?.password_set ? 'Leave blank to keep current' : 'OpenSubtitles password'
                }
                className="w-full rounded bg-netflix-gray px-4 py-2.5 pr-10 text-sm text-white outline-none ring-1 ring-transparent transition-all placeholder:text-gray-500 focus:ring-netflix-red"
              />
              <button
                type="button"
                onClick={() => setShowPassword(!showPassword)}
                className="absolute right-3 top-1/2 -translate-y-1/2 text-gray-500 hover:text-gray-300"
              >
                {showPassword ? <LuEyeOff size={16} /> : <LuEye size={16} />}
              </button>
            </div>
          </div>

          <div className="pt-2">
            <button
              type="submit"
              disabled={isSaving}
              className="flex items-center gap-2 rounded bg-netflix-red px-5 py-2.5 text-sm font-medium text-white transition-colors hover:bg-netflix-red-hover disabled:opacity-50"
            >
              {saved ? (
                <>
                  <LuCheck size={16} />
                  Saved
                </>
              ) : (
                <>
                  <LuSave size={16} />
                  {isSaving ? 'Saving...' : 'Save'}
                </>
              )}
            </button>
          </div>
        </form>
      </div>

      {/* Podnapisi */}
      <div className="rounded-lg bg-netflix-dark p-6">
        <div className="mb-1 flex items-center gap-2">
          <h3 className="text-base font-semibold text-white">Podnapisi</h3>
          <span className="rounded bg-green-500/20 px-2 py-0.5 text-[10px] font-medium text-green-400">
            Active
          </span>
        </div>
        <p className="text-sm text-gray-400">
          No configuration needed. Podnapisi is always available as a subtitle source.
        </p>
      </div>
    </div>
  )
}

// ── Shared Components ─────────────────────────────────────────────────────────

function SectionHeader({ title, description }: { title: string; description: string }) {
  return (
    <div>
      <h2 className="text-2xl font-bold text-white">{title}</h2>
      <p className="text-sm text-gray-400">{description}</p>
    </div>
  )
}

function Spinner() {
  return (
    <div className="flex h-48 items-center justify-center">
      <div className="h-8 w-8 animate-spin rounded-full border-2 border-netflix-red border-t-transparent" />
    </div>
  )
}
