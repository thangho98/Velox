import { useState } from 'react'
import { useNavigate, Navigate } from 'react-router'
import { useTranslation } from '@/hooks/useTranslation'
import { useCompleteWizard, useWizardStatus } from '@/hooks/stores/useAuth'
import {
  useTMDbSettings,
  useUpdateTMDbSettings,
  useOMDbSettings,
  useUpdateOMDbSettings,
  useTVDBSettings,
  useUpdateTVDBSettings,
  useFanartSettings,
  useUpdateFanartSettings,
  useOpenSubsSettings,
  useUpdateOpenSubsSettings,
  useSubdlSettings,
  useUpdateSubdlSettings,
  useDeepLSettings,
  useUpdateDeepLSettings,
  useAutoSubSettings,
  useUpdateAutoSubSettings,
  usePlaybackSettings,
  useUpdatePlaybackSettings,
  usePretranscodeSettings,
  useUpdatePretranscodeSettings,
  usePretranscodeProfiles,
  useTogglePretranscodeProfile,
  useCinemaSettings,
  useUpdateCinemaSettings,
} from '@/hooks/stores/useSettings'
import { useCreateLibrary } from '@/hooks/stores/useMedia'
import { Logo } from '@/components/Logo'
import { DirectoryPicker } from '@/components/DirectoryPicker'
import { Toggle } from '@/components/ui/Toggle'
import { Select } from '@/components/ui/Select'
import type { PretranscodeProfile } from '@/hooks/stores/useSettings'

const STEPS = ['metadata', 'subtitles', 'playback', 'pretranscode', 'cinema', 'summary'] as const
type Step = (typeof STEPS)[number]

// --- Progress Bar ---
function ProgressBar({ current, total }: { current: number; total: number }) {
  return (
    <div className="flex items-center gap-2">
      {Array.from({ length: total }, (_, i) => (
        <div
          key={i}
          className={`h-1 flex-1 rounded-full transition-colors ${
            i <= current ? 'bg-netflix-red' : 'bg-gray-700'
          }`}
        />
      ))}
    </div>
  )
}

// --- Step 1: Metadata ---
function MetadataStep() {
  const { t } = useTranslation('wizard')

  const { data: tmdb } = useTMDbSettings()
  const { data: omdb } = useOMDbSettings()
  const { data: tvdb } = useTVDBSettings()
  const { data: fanart } = useFanartSettings()

  const { mutate: updateTmdb } = useUpdateTMDbSettings()
  const { mutate: updateOmdb } = useUpdateOMDbSettings()
  const { mutate: updateTvdb } = useUpdateTVDBSettings()
  const { mutate: updateFanart } = useUpdateFanartSettings()

  const [keys, setKeys] = useState({
    tmdb: '',
    omdb: '',
    tvdb: '',
    fanart: '',
  })
  const [saved, setSaved] = useState<Record<string, boolean>>({})

  const providers = [
    {
      id: 'tmdb' as const,
      data: tmdb,
      update: updateTmdb,
      transform: (key: string) => ({ api_key: key }),
    },
    {
      id: 'omdb' as const,
      data: omdb,
      update: updateOmdb,
      transform: (key: string) => ({ api_key: key }),
    },
    {
      id: 'tvdb' as const,
      data: tvdb,
      update: updateTvdb,
      transform: (key: string) => ({ api_key: key }),
    },
    {
      id: 'fanart' as const,
      data: fanart,
      update: updateFanart,
      transform: (key: string) => ({ api_key: key }),
    },
  ]

  const handleSave = (id: 'tmdb' | 'omdb' | 'tvdb' | 'fanart') => {
    const provider = providers.find((p) => p.id === id)
    if (!provider || !keys[id]) return
    provider.update(provider.transform(keys[id]), {
      onSuccess: () => {
        setSaved((prev) => ({ ...prev, [id]: true }))
        setTimeout(() => setSaved((prev) => ({ ...prev, [id]: false })), 2000)
      },
    })
  }

  return (
    <div className="space-y-6">
      <div>
        <h2 className="text-2xl font-bold text-white">{t('metadata.title')}</h2>
        <p className="mt-2 text-sm text-gray-400">{t('metadata.description')}</p>
      </div>

      <div className="space-y-4">
        {providers.map(({ id, data }) => (
          <div key={id} className="rounded-lg bg-[#1a1a1a] p-4">
            <div className="mb-1 flex items-center justify-between">
              <div>
                <span className="font-medium text-white">{t(`metadata.${id}.label`)}</span>
                {data?.has_builtin && (
                  <span className="ml-2 rounded bg-green-900/50 px-2 py-0.5 text-xs text-green-400">
                    {t('metadata.builtinAvailable')}
                  </span>
                )}
                {(data?.api_key || keys[id]) && !data?.has_builtin && (
                  <span className="ml-2 rounded bg-blue-900/50 px-2 py-0.5 text-xs text-blue-400">
                    {t('metadata.customKey')}
                  </span>
                )}
              </div>
            </div>
            <p className="mb-3 text-xs text-gray-500">{t(`metadata.${id}.description`)}</p>
            <div className="flex gap-2">
              <input
                type="text"
                value={keys[id]}
                onChange={(e) => setKeys((prev) => ({ ...prev, [id]: e.target.value }))}
                placeholder={data?.api_key || t('metadata.apiKeyPlaceholder')}
                className="flex-1 rounded bg-[#0f0f0f] px-3 py-2 text-sm text-white placeholder-gray-500 outline-none ring-1 ring-gray-700 transition-colors focus:ring-netflix-red"
              />
              <button
                onClick={() => handleSave(id)}
                disabled={!keys[id]}
                className="rounded bg-gray-700 px-4 py-2 text-sm text-white transition-colors hover:bg-gray-600 disabled:opacity-30"
              >
                {saved[id] ? t('buttons.saved') : t('buttons.saving').replace('...', '')}
              </button>
            </div>
          </div>
        ))}
      </div>
    </div>
  )
}

// --- Step 2: Subtitles ---
function SubtitlesStep() {
  const { t } = useTranslation('wizard')

  const { data: openSubs } = useOpenSubsSettings()
  const { data: subdl } = useSubdlSettings()
  const { data: deepl } = useDeepLSettings()
  const { data: autoSub } = useAutoSubSettings()

  const { mutate: updateOpenSubs } = useUpdateOpenSubsSettings()
  const { mutate: updateSubdl } = useUpdateSubdlSettings()
  const { mutate: updateDeepL } = useUpdateDeepLSettings()
  const { mutate: updateAutoSub } = useUpdateAutoSubSettings()

  const [openSubsForm, setOpenSubsForm] = useState({ api_key: '', username: '', password: '' })
  const [subdlKey, setSubdlKey] = useState('')
  const [deeplKey, setDeeplKey] = useState('')
  const [autoLangs, setAutoLangs] = useState(autoSub?.languages || '')
  const [saved, setSaved] = useState<Record<string, boolean>>({})

  const markSaved = (key: string) => {
    setSaved((prev) => ({ ...prev, [key]: true }))
    setTimeout(() => setSaved((prev) => ({ ...prev, [key]: false })), 2000)
  }

  return (
    <div className="space-y-6">
      <div>
        <h2 className="text-2xl font-bold text-white">{t('subtitles.title')}</h2>
        <p className="mt-2 text-sm text-gray-400">{t('subtitles.description')}</p>
      </div>

      {/* OpenSubtitles */}
      <div className="rounded-lg bg-[#1a1a1a] p-4">
        <span className="font-medium text-white">{t('subtitles.opensubtitles.label')}</span>
        <p className="mb-3 text-xs text-gray-500">{t('subtitles.opensubtitles.description')}</p>
        <div className="space-y-2">
          <input
            type="text"
            value={openSubsForm.api_key}
            onChange={(e) => setOpenSubsForm((prev) => ({ ...prev, api_key: e.target.value }))}
            placeholder={openSubs?.api_key || t('metadata.apiKeyPlaceholder')}
            className="w-full rounded bg-[#0f0f0f] px-3 py-2 text-sm text-white placeholder-gray-500 outline-none ring-1 ring-gray-700 focus:ring-netflix-red"
          />
          <div className="flex gap-2">
            <input
              type="text"
              value={openSubsForm.username}
              onChange={(e) => setOpenSubsForm((prev) => ({ ...prev, username: e.target.value }))}
              placeholder={openSubs?.username || t('subtitles.username')}
              className="flex-1 rounded bg-[#0f0f0f] px-3 py-2 text-sm text-white placeholder-gray-500 outline-none ring-1 ring-gray-700 focus:ring-netflix-red"
            />
            <input
              type="password"
              value={openSubsForm.password}
              onChange={(e) => setOpenSubsForm((prev) => ({ ...prev, password: e.target.value }))}
              placeholder={openSubs?.password_set ? '••••••••' : t('subtitles.password')}
              className="flex-1 rounded bg-[#0f0f0f] px-3 py-2 text-sm text-white placeholder-gray-500 outline-none ring-1 ring-gray-700 focus:ring-netflix-red"
            />
          </div>
          <button
            onClick={() => updateOpenSubs(openSubsForm, { onSuccess: () => markSaved('openSubs') })}
            disabled={!openSubsForm.api_key && !openSubsForm.username}
            className="rounded bg-gray-700 px-4 py-2 text-sm text-white transition-colors hover:bg-gray-600 disabled:opacity-30"
          >
            {saved.openSubs ? t('buttons.saved') : 'Save'}
          </button>
        </div>
      </div>

      {/* Subdl */}
      <div className="rounded-lg bg-[#1a1a1a] p-4">
        <span className="font-medium text-white">{t('subtitles.subdl.label')}</span>
        {subdl?.has_builtin && (
          <span className="ml-2 rounded bg-green-900/50 px-2 py-0.5 text-xs text-green-400">
            {t('metadata.builtinAvailable')}
          </span>
        )}
        <p className="mb-3 text-xs text-gray-500">{t('subtitles.subdl.description')}</p>
        <div className="flex gap-2">
          <input
            type="text"
            value={subdlKey}
            onChange={(e) => setSubdlKey(e.target.value)}
            placeholder={subdl?.api_key || t('metadata.apiKeyPlaceholder')}
            className="flex-1 rounded bg-[#0f0f0f] px-3 py-2 text-sm text-white placeholder-gray-500 outline-none ring-1 ring-gray-700 focus:ring-netflix-red"
          />
          <button
            onClick={() =>
              updateSubdl({ api_key: subdlKey }, { onSuccess: () => markSaved('subdl') })
            }
            disabled={!subdlKey}
            className="rounded bg-gray-700 px-4 py-2 text-sm text-white transition-colors hover:bg-gray-600 disabled:opacity-30"
          >
            {saved.subdl ? t('buttons.saved') : 'Save'}
          </button>
        </div>
      </div>

      {/* DeepL */}
      <div className="rounded-lg bg-[#1a1a1a] p-4">
        <span className="font-medium text-white">{t('subtitles.deepl.label')}</span>
        <p className="mb-3 text-xs text-gray-500">{t('subtitles.deepl.description')}</p>
        <div className="flex gap-2">
          <input
            type="text"
            value={deeplKey}
            onChange={(e) => setDeeplKey(e.target.value)}
            placeholder={deepl?.api_key || t('metadata.apiKeyPlaceholder')}
            className="flex-1 rounded bg-[#0f0f0f] px-3 py-2 text-sm text-white placeholder-gray-500 outline-none ring-1 ring-gray-700 focus:ring-netflix-red"
          />
          <button
            onClick={() =>
              updateDeepL({ api_key: deeplKey }, { onSuccess: () => markSaved('deepl') })
            }
            disabled={!deeplKey}
            className="rounded bg-gray-700 px-4 py-2 text-sm text-white transition-colors hover:bg-gray-600 disabled:opacity-30"
          >
            {saved.deepl ? t('buttons.saved') : 'Save'}
          </button>
        </div>
      </div>

      {/* Auto-download languages */}
      <div className="rounded-lg bg-[#1a1a1a] p-4">
        <span className="font-medium text-white">{t('subtitles.autoDownload.label')}</span>
        <p className="mb-3 text-xs text-gray-500">{t('subtitles.autoDownload.description')}</p>
        <div className="flex gap-2">
          <input
            type="text"
            value={autoLangs}
            onChange={(e) => setAutoLangs(e.target.value)}
            placeholder={t('subtitles.autoDownload.placeholder')}
            className="flex-1 rounded bg-[#0f0f0f] px-3 py-2 text-sm text-white placeholder-gray-500 outline-none ring-1 ring-gray-700 focus:ring-netflix-red"
          />
          <button
            onClick={() =>
              updateAutoSub({ languages: autoLangs }, { onSuccess: () => markSaved('autoSub') })
            }
            className="rounded bg-gray-700 px-4 py-2 text-sm text-white transition-colors hover:bg-gray-600"
          >
            {saved.autoSub ? t('buttons.saved') : 'Save'}
          </button>
        </div>
      </div>
    </div>
  )
}

// --- Step 3: Playback ---
function PlaybackStep() {
  const { t } = useTranslation('wizard')
  const { data: playback } = usePlaybackSettings()
  const { mutate: updatePlayback } = useUpdatePlaybackSettings()
  const [mode, setMode] = useState<'auto' | 'direct_play'>(playback?.playback_mode || 'auto')

  const handleChange = (newMode: 'auto' | 'direct_play') => {
    setMode(newMode)
    updatePlayback({ playback_mode: newMode })
  }

  return (
    <div className="space-y-6">
      <div>
        <h2 className="text-2xl font-bold text-white">{t('playback.title')}</h2>
        <p className="mt-2 text-sm text-gray-400">{t('playback.description')}</p>
      </div>

      <div className="space-y-3">
        <button
          onClick={() => handleChange('auto')}
          className={`w-full rounded-lg p-4 text-left transition-colors ${
            mode === 'auto'
              ? 'bg-netflix-red/20 ring-1 ring-netflix-red'
              : 'bg-[#1a1a1a] hover:bg-[#222]'
          }`}
        >
          <div className="flex items-center gap-3">
            <div
              className={`h-4 w-4 rounded-full border-2 ${
                mode === 'auto' ? 'border-netflix-red bg-netflix-red' : 'border-gray-500'
              }`}
            >
              {mode === 'auto' && (
                <div className="flex h-full items-center justify-center">
                  <div className="h-1.5 w-1.5 rounded-full bg-white" />
                </div>
              )}
            </div>
            <div>
              <span className="font-medium text-white">{t('playback.mode.auto.label')}</span>
              <p className="mt-1 text-xs text-gray-400">{t('playback.mode.auto.description')}</p>
            </div>
          </div>
        </button>

        <button
          onClick={() => handleChange('direct_play')}
          className={`w-full rounded-lg p-4 text-left transition-colors ${
            mode === 'direct_play'
              ? 'bg-netflix-red/20 ring-1 ring-netflix-red'
              : 'bg-[#1a1a1a] hover:bg-[#222]'
          }`}
        >
          <div className="flex items-center gap-3">
            <div
              className={`h-4 w-4 rounded-full border-2 ${
                mode === 'direct_play' ? 'border-netflix-red bg-netflix-red' : 'border-gray-500'
              }`}
            >
              {mode === 'direct_play' && (
                <div className="flex h-full items-center justify-center">
                  <div className="h-1.5 w-1.5 rounded-full bg-white" />
                </div>
              )}
            </div>
            <div>
              <span className="font-medium text-white">{t('playback.mode.directPlay.label')}</span>
              <p className="mt-1 text-xs text-gray-400">
                {t('playback.mode.directPlay.description')}
              </p>
            </div>
          </div>
        </button>
      </div>
    </div>
  )
}

// --- Step 4: Pre-transcode ---
function PretranscodeStep() {
  const { t } = useTranslation('wizard')
  const { data: settings } = usePretranscodeSettings()
  const { data: profiles } = usePretranscodeProfiles()
  const { mutate: updateSettings } = useUpdatePretranscodeSettings()
  const { mutate: toggleProfile } = useTogglePretranscodeProfile()

  const [enabled, setEnabled] = useState(settings?.enabled ?? false)
  const [schedule, setSchedule] = useState(settings?.schedule || 'always')
  const [concurrency, setConcurrency] = useState(settings?.concurrency || '2')

  const handleToggle = (val: boolean) => {
    setEnabled(val)
    updateSettings({ enabled: val })
  }

  const handleSchedule = (val: string) => {
    setSchedule(val)
    updateSettings({ schedule: val })
  }

  const handleConcurrency = (val: string) => {
    setConcurrency(val)
    updateSettings({ concurrency: val })
  }

  return (
    <div className="space-y-6">
      <div>
        <h2 className="text-2xl font-bold text-white">{t('pretranscode.title')}</h2>
        <p className="mt-2 text-sm text-gray-400">{t('pretranscode.description')}</p>
      </div>

      <div className="rounded-lg bg-[#1a1a1a] p-4">
        <div className="flex items-center justify-between">
          <span className="font-medium text-white">{t('pretranscode.enable')}</span>
          <Toggle enabled={enabled} onChange={handleToggle} />
        </div>
      </div>

      {enabled && (
        <>
          {/* Profiles */}
          <div className="rounded-lg bg-[#1a1a1a] p-4">
            <span className="mb-1 block font-medium text-white">
              {t('pretranscode.profiles.label')}
            </span>
            <p className="mb-3 text-xs text-gray-500">{t('pretranscode.profiles.description')}</p>
            <div className="space-y-2">
              {profiles?.map((p: PretranscodeProfile) => (
                <label
                  key={p.id}
                  className="flex cursor-pointer items-center justify-between rounded bg-[#0f0f0f] px-3 py-2"
                >
                  <div>
                    <span className="text-sm text-white">{p.name}</span>
                    <span className="ml-2 text-xs text-gray-500">
                      {p.height}p @ {(p.video_bitrate / 1000000).toFixed(1)} Mbps
                    </span>
                  </div>
                  <input
                    type="checkbox"
                    checked={p.enabled}
                    onChange={() => toggleProfile({ id: p.id, enabled: !p.enabled })}
                    className="h-4 w-4 rounded border-gray-600 bg-gray-800 text-netflix-red accent-netflix-red"
                  />
                </label>
              ))}
            </div>
          </div>

          {/* Schedule + Concurrency */}
          <div className="grid grid-cols-2 gap-4">
            <div className="rounded-lg bg-[#1a1a1a] p-4">
              <span className="mb-2 block text-sm font-medium text-white">
                {t('pretranscode.schedule.label')}
              </span>
              <Select value={schedule} onChange={(e) => handleSchedule(e.target.value)}>
                <option value="always">{t('pretranscode.schedule.always')}</option>
                <option value="night">{t('pretranscode.schedule.night')}</option>
                <option value="idle">{t('pretranscode.schedule.idle')}</option>
              </Select>
            </div>
            <div className="rounded-lg bg-[#1a1a1a] p-4">
              <span className="mb-2 block text-sm font-medium text-white">
                {t('pretranscode.concurrency.label')}
              </span>
              <Select value={concurrency} onChange={(e) => handleConcurrency(e.target.value)}>
                <option value="1">1</option>
                <option value="2">2</option>
                <option value="3">3</option>
                <option value="4">4</option>
              </Select>
            </div>
          </div>
        </>
      )}
    </div>
  )
}

// --- Step 5: Cinema Mode ---
function CinemaStep() {
  const { t } = useTranslation('wizard')
  const { data: cinema } = useCinemaSettings()
  const { mutate: updateCinema } = useUpdateCinemaSettings()

  const [enabled, setEnabled] = useState(cinema?.enabled ?? false)
  const [maxTrailers, setMaxTrailers] = useState(cinema?.max_trailers || '2')

  const handleToggle = (val: boolean) => {
    setEnabled(val)
    updateCinema({ enabled: val })
  }

  const handleTrailers = (val: string) => {
    setMaxTrailers(val)
    updateCinema({ max_trailers: val })
  }

  return (
    <div className="space-y-6">
      <div>
        <h2 className="text-2xl font-bold text-white">{t('cinema.title')}</h2>
        <p className="mt-2 text-sm text-gray-400">{t('cinema.description')}</p>
      </div>

      <div className="rounded-lg bg-[#1a1a1a] p-4">
        <div className="flex items-center justify-between">
          <span className="font-medium text-white">{t('cinema.enable')}</span>
          <Toggle enabled={enabled} onChange={handleToggle} />
        </div>
      </div>

      {enabled && (
        <div className="rounded-lg bg-[#1a1a1a] p-4">
          <span className="mb-2 block text-sm font-medium text-white">
            {t('cinema.maxTrailers.label')}
          </span>
          <p className="mb-2 text-xs text-gray-500">{t('cinema.maxTrailers.description')}</p>
          <Select value={maxTrailers} onChange={(e) => handleTrailers(e.target.value)}>
            <option value="1">1</option>
            <option value="2">2</option>
            <option value="3">3</option>
            <option value="5">5</option>
          </Select>
        </div>
      )}
    </div>
  )
}

// --- Step 6: Summary ---
function SummaryStep() {
  const { t } = useTranslation('wizard')

  const { data: tmdb } = useTMDbSettings()
  const { data: omdb } = useOMDbSettings()
  const { data: tvdb } = useTVDBSettings()
  const { data: fanart } = useFanartSettings()
  const { data: playback } = usePlaybackSettings()
  const { data: pretranscode } = usePretranscodeSettings()
  const { data: cinema } = useCinemaSettings()
  const { data: autoSub } = useAutoSubSettings()

  const { mutate: createLibrary, isPending: isCreating } = useCreateLibrary()

  const [libraryForm, setLibraryForm] = useState({
    name: '',
    path: '',
    type: 'movie' as 'movie' | 'tvshow',
  })
  const [libraryAdded, setLibraryAdded] = useState(false)
  const [showPicker, setShowPicker] = useState(false)

  const handleAddLibrary = () => {
    if (!libraryForm.name || !libraryForm.path) return
    createLibrary(
      { name: libraryForm.name, path: libraryForm.path, type: libraryForm.type },
      {
        onSuccess: () => setLibraryAdded(true),
      },
    )
  }

  const summaryItems = [
    {
      label: t('steps.metadata'),
      status:
        tmdb?.has_builtin || tmdb?.api_key || omdb?.api_key || tvdb?.api_key || fanart?.api_key
          ? t('summary.configured')
          : t('summary.skipped'),
      active: !!(
        tmdb?.has_builtin ||
        tmdb?.api_key ||
        omdb?.api_key ||
        tvdb?.api_key ||
        fanart?.api_key
      ),
    },
    {
      label: t('steps.subtitles'),
      status: autoSub?.languages ? t('summary.configured') : t('summary.skipped'),
      active: !!autoSub?.languages,
    },
    {
      label: t('steps.playback'),
      status: playback?.playback_mode === 'auto' ? 'Auto' : 'Direct Play',
      active: true,
    },
    {
      label: t('steps.pretranscode'),
      status: pretranscode?.enabled ? t('summary.enabled') : t('summary.disabled'),
      active: !!pretranscode?.enabled,
    },
    {
      label: t('steps.cinema'),
      status: cinema?.enabled ? t('summary.enabled') : t('summary.disabled'),
      active: !!cinema?.enabled,
    },
  ]

  return (
    <div className="space-y-6">
      <div>
        <h2 className="text-2xl font-bold text-white">{t('summary.title')}</h2>
        <p className="mt-2 text-sm text-gray-400">{t('summary.description')}</p>
      </div>

      {/* Config summary */}
      <div className="rounded-lg bg-[#1a1a1a] p-4">
        <div className="space-y-3">
          {summaryItems.map((item) => (
            <div key={item.label} className="flex items-center justify-between">
              <span className="text-sm text-gray-300">{item.label}</span>
              <span
                className={`rounded px-2 py-0.5 text-xs ${
                  item.active ? 'bg-green-900/50 text-green-400' : 'bg-gray-800 text-gray-500'
                }`}
              >
                {item.status}
              </span>
            </div>
          ))}
        </div>
      </div>

      {/* Add library */}
      <div className="rounded-lg bg-[#1a1a1a] p-4">
        <h3 className="mb-1 font-medium text-white">{t('summary.addLibrary')}</h3>
        <p className="mb-4 text-xs text-gray-500">{t('summary.addLibraryDescription')}</p>

        {libraryAdded ? (
          <div className="rounded bg-green-900/30 p-3 text-sm text-green-400">
            {t('summary.libraryAdded')}
          </div>
        ) : (
          <div className="space-y-3">
            <div className="grid grid-cols-2 gap-3">
              <div>
                <label className="mb-1 block text-xs text-gray-400">
                  {t('summary.libraryName')}
                </label>
                <input
                  type="text"
                  value={libraryForm.name}
                  onChange={(e) => setLibraryForm((prev) => ({ ...prev, name: e.target.value }))}
                  placeholder={t('summary.libraryNamePlaceholder')}
                  className="w-full rounded bg-[#0f0f0f] px-3 py-2 text-sm text-white placeholder-gray-500 outline-none ring-1 ring-gray-700 focus:ring-netflix-red"
                />
              </div>
              <div>
                <label className="mb-1 block text-xs text-gray-400">
                  {t('summary.libraryType')}
                </label>
                <Select
                  value={libraryForm.type}
                  onChange={(e) =>
                    setLibraryForm((prev) => ({
                      ...prev,
                      type: e.target.value as 'movie' | 'tvshow',
                    }))
                  }
                  className="w-full"
                >
                  <option value="movie">{t('summary.movie')}</option>
                  <option value="tvshow">{t('summary.tvshow')}</option>
                </Select>
              </div>
            </div>
            <div>
              <label className="mb-1 block text-xs text-gray-400">{t('summary.libraryPath')}</label>
              <div className="flex gap-2">
                <input
                  type="text"
                  value={libraryForm.path}
                  onChange={(e) => setLibraryForm((prev) => ({ ...prev, path: e.target.value }))}
                  placeholder={t('summary.libraryPathPlaceholder')}
                  className="flex-1 rounded bg-[#0f0f0f] px-3 py-2 text-sm text-white placeholder-gray-500 outline-none ring-1 ring-gray-700 focus:ring-netflix-red"
                />
                <button
                  type="button"
                  onClick={() => setShowPicker(true)}
                  className="rounded bg-gray-700 px-3 py-2 text-sm text-white transition-colors hover:bg-gray-600"
                >
                  Browse
                </button>
              </div>
            </div>
            {showPicker && (
              <DirectoryPicker
                onSelect={(path) => setLibraryForm((prev) => ({ ...prev, path }))}
                onClose={() => setShowPicker(false)}
              />
            )}
            <button
              onClick={handleAddLibrary}
              disabled={!libraryForm.name || !libraryForm.path || isCreating}
              className="rounded bg-netflix-red px-4 py-2 text-sm font-medium text-white transition-colors hover:bg-netflix-red-hover disabled:opacity-50"
            >
              {isCreating ? t('summary.adding') : t('summary.addLibraryButton')}
            </button>
          </div>
        )}
      </div>
    </div>
  )
}

// --- Main Wizard Page ---
const STEP_COMPONENTS: Record<Step, () => JSX.Element> = {
  metadata: MetadataStep,
  subtitles: SubtitlesStep,
  playback: PlaybackStep,
  pretranscode: PretranscodeStep,
  cinema: CinemaStep,
  summary: SummaryStep,
}

export function SetupWizardPage() {
  const navigate = useNavigate()
  const { t } = useTranslation('wizard')
  const { mutate: completeWizard } = useCompleteWizard()
  const { data: wizardStatus, isLoading: wizardLoading } = useWizardStatus()
  const [currentStep, setCurrentStep] = useState(0)

  // Redirect to home if wizard already completed
  if (!wizardLoading && wizardStatus?.completed) {
    return <Navigate to="/" replace />
  }

  const step = STEPS[currentStep]
  const StepComponent = STEP_COMPONENTS[step]
  const isFirst = currentStep === 0
  const isLast = currentStep === STEPS.length - 1

  const handleNext = () => {
    if (isLast) {
      completeWizard(undefined, {
        onSuccess: () => navigate('/', { replace: true }),
      })
    } else {
      setCurrentStep((prev) => prev + 1)
    }
  }

  const handleBack = () => {
    setCurrentStep((prev) => Math.max(0, prev - 1))
  }

  const handleSkipAll = () => {
    completeWizard(undefined, {
      onSuccess: () => navigate('/', { replace: true }),
    })
  }

  return (
    <div className="flex min-h-screen flex-col bg-netflix-black">
      {/* Background */}
      <div className="absolute inset-0 bg-gradient-to-b from-netflix-dark/50 via-netflix-black/80 to-netflix-black" />

      {/* Header */}
      <header className="relative z-10 flex items-center justify-between p-6">
        <Logo size="lg" />
        <button
          onClick={handleSkipAll}
          className="text-sm text-gray-400 transition-colors hover:text-white"
        >
          {t('buttons.skipAll')}
        </button>
      </header>

      {/* Content */}
      <main className="relative z-10 mx-auto flex w-full max-w-2xl flex-1 flex-col px-4 pb-8">
        {/* Progress */}
        <div className="mb-2">
          <ProgressBar current={currentStep} total={STEPS.length} />
        </div>
        <div className="mb-6 flex items-center justify-between text-xs text-gray-500">
          <span>
            {t(`steps.${step}`)} ({currentStep + 1}/{STEPS.length})
          </span>
        </div>

        {/* Step content */}
        <div className="flex-1">
          <StepComponent />
        </div>

        {/* Navigation */}
        <div className="mt-8 flex items-center justify-between">
          <button
            onClick={handleBack}
            disabled={isFirst}
            className="rounded px-6 py-2.5 text-sm text-gray-400 transition-colors hover:text-white disabled:invisible"
          >
            {t('buttons.back')}
          </button>

          <div className="flex gap-3">
            {!isLast && (
              <button
                onClick={() => setCurrentStep((prev) => prev + 1)}
                className="rounded px-6 py-2.5 text-sm text-gray-400 transition-colors hover:text-white"
              >
                {t('buttons.skip')}
              </button>
            )}
            <button
              onClick={handleNext}
              className="rounded bg-netflix-red px-8 py-2.5 text-sm font-semibold text-white transition-colors hover:bg-netflix-red-hover"
            >
              {isLast ? t('buttons.finish') : t('buttons.next')}
            </button>
          </div>
        </div>
      </main>
    </div>
  )
}
