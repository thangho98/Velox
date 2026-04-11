import { useState } from 'react'
import { LuSave, LuCheck, LuEye, LuEyeOff } from 'react-icons/lu'
import { Select } from '@/components/ui/Select'
import {
  useOpenSubsSettings,
  useUpdateOpenSubsSettings,
  useSubdlSettings,
  useUpdateSubdlSettings,
  useAITranslationSettings,
  useUpdateAITranslationSettings,
  useDeepLSettings,
  useUpdateDeepLSettings,
  useAutoSubSettings,
  useUpdateAutoSubSettings,
} from '@/hooks/stores/useSettings'
import { useTranslation } from '@/hooks/useTranslation'
import { SectionHeader, Field, Spinner, inputClass } from './shared'

export function SubtitlesSection() {
  const { t } = useTranslation('settings')
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
  const [showApiKey, setShowApiKey] = useState(false)
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
      <SectionHeader
        title={t('sections.subtitles.title')}
        description={t('sections.subtitles.description')}
      />

      {/* OpenSubtitles */}
      <div className="rounded-lg bg-netflix-dark p-5">
        <div className="mb-1 flex items-center gap-2">
          <h3 className="text-sm font-semibold text-white">{t('providers.opensubtitles.name')}</h3>
          {settings?.password_set && settings?.api_key && (
            <span className="rounded bg-green-500/20 px-2 py-0.5 text-[10px] font-medium text-green-400">
              {t('status.enabled')}
            </span>
          )}
        </div>
        <p className="mb-5 text-xs text-gray-400">{t('providers.opensubtitles.description')}</p>

        <form onSubmit={handleSave} className="space-y-5">
          {/* Step 1: API Key */}
          <div className="space-y-3">
            <div className="flex items-center gap-2">
              <span className="flex h-5 w-5 shrink-0 items-center justify-center rounded-full bg-netflix-gray text-[10px] font-bold text-white">
                1
              </span>
              <span className="text-xs font-medium text-gray-300">
                {t('providers.opensubtitles.apiKey')}
              </span>
            </div>
            <p className="pl-7 text-[11px] text-gray-500">{t('providers.opensubtitles.getKey')}</p>
            <div className="pl-7 relative">
              <input
                type={showApiKey ? 'text' : 'password'}
                value={apiKey}
                onChange={(e) => setApiKey(e.target.value)}
                placeholder={t('providers.opensubtitles.apiKey')}
                className={`${inputClass} pr-10`}
              />
              <button
                type="button"
                onClick={() => setShowApiKey(!showApiKey)}
                className="absolute right-3 top-1/2 -translate-y-1/2 text-gray-500 hover:text-gray-300"
              >
                {showApiKey ? <LuEyeOff size={16} /> : <LuEye size={16} />}
              </button>
            </div>
          </div>

          <div className="border-t border-netflix-gray/30" />

          {/* Step 2: Account */}
          <div className="space-y-3">
            <div className="flex items-center gap-2">
              <span className="flex h-5 w-5 shrink-0 items-center justify-center rounded-full bg-netflix-gray text-[10px] font-bold text-white">
                2
              </span>
              <span className="text-xs font-medium text-gray-300">
                {t('providers.opensubtitles.username')}
              </span>
            </div>
            <p className="pl-7 text-[11px] text-gray-500">
              {t('providers.opensubtitles.description')}
            </p>
            <div className="space-y-3 pl-7">
              <Field label={t('providers.opensubtitles.username')} compact>
                <input
                  type="text"
                  value={username}
                  onChange={(e) => setUsername(e.target.value)}
                  placeholder={t('providers.opensubtitles.username')}
                  className={inputClass}
                />
              </Field>
              <Field
                label={
                  <>
                    {t('providers.opensubtitles.password')}
                    {settings?.password_set && (
                      <span className="ml-2 text-xs text-green-400">
                        ({t('providers.opensubtitles.passwordSet')})
                      </span>
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
                        : t('providers.opensubtitles.password')
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
                  <LuCheck size={14} /> {t('actions.saved')}
                </>
              ) : (
                <>
                  <LuSave size={14} /> {isSaving ? t('actions.saving') : t('actions.save')}
                </>
              )}
            </button>
          </div>
        </form>
      </div>

      {/* Subdl */}
      <SubdlCard />

      {/* AI Translation */}
      <AITranslationCard />

      {/* DeepL Translation */}
      <DeepLCard />

      {/* Auto-Download */}
      <AutoSubCard />
    </div>
  )
}

function SubdlCard() {
  const { t } = useTranslation('settings')
  const { data: settings, isLoading } = useSubdlSettings()
  const { mutate: updateSettings, isPending: isSaving } = useUpdateSubdlSettings()
  const [editedApiKey, setEditedApiKey] = useState<string | null>(null)
  const apiKey = editedApiKey ?? settings?.api_key ?? ''
  const [showApiKey, setShowApiKey] = useState(false)
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
        <h3 className="text-sm font-semibold text-white">{t('providers.subdl.name')}</h3>
        <span
          className={`rounded px-2 py-0.5 text-[10px] font-medium ${
            settings?.api_key
              ? 'bg-blue-500/20 text-blue-400'
              : settings?.has_builtin
                ? 'bg-green-500/20 text-green-400'
                : 'bg-gray-500/20 text-gray-400'
          }`}
        >
          {settings?.api_key
            ? t('status.customKey')
            : settings?.has_builtin
              ? t('status.envKey')
              : t('status.notConfigured')}
        </span>
      </div>
      <p className="mb-5 text-xs text-gray-400">
        {t('providers.subdl.description')}
        {settings?.has_builtin
          ? ' ' + t('providers.subdl.hasBuiltin')
          : ' ' + t('providers.subdl.noBuiltin')}{' '}
        <a
          href="https://subdl.com/panel/api"
          target="_blank"
          rel="noopener noreferrer"
          className="text-netflix-red hover:underline"
        >
          {t('actions.getFreeKey')}
        </a>
      </p>

      <form onSubmit={handleSave} className="space-y-4">
        <div className="relative">
          <input
            type={showApiKey ? 'text' : 'password'}
            value={apiKey}
            onChange={(e) => setEditedApiKey(e.target.value)}
            placeholder={
              settings?.has_builtin
                ? t('providers.subdl.placeholderOptional')
                : t('providers.subdl.placeholderRequired')
            }
            className={`${inputClass} pr-10`}
          />
          <button
            type="button"
            onClick={() => setShowApiKey(!showApiKey)}
            className="absolute right-3 top-1/2 -translate-y-1/2 text-gray-500 hover:text-gray-300"
          >
            {showApiKey ? <LuEyeOff size={16} /> : <LuEye size={16} />}
          </button>
        </div>
        <button
          type="submit"
          disabled={isSaving}
          className="flex items-center gap-2 rounded bg-netflix-red px-4 py-2 text-sm font-medium text-white transition-colors hover:bg-netflix-red-hover disabled:opacity-50"
        >
          {saved ? (
            <>
              <LuCheck size={14} /> {t('actions.saved')}
            </>
          ) : (
            <>
              <LuSave size={14} /> {isSaving ? t('actions.saving') : t('actions.save')}
            </>
          )}
        </button>
      </form>
    </div>
  )
}

function AITranslationCard() {
  const { t } = useTranslation('settings')
  const { data: settings, isLoading } = useAITranslationSettings()
  const { mutate: updateSettings, isPending: isSaving } = useUpdateAITranslationSettings()
  const [editedProvider, setEditedProvider] = useState<string | null>(null)
  const [editedApiKey, setEditedApiKey] = useState<string | null>(null)
  const [editedBaseUrl, setEditedBaseUrl] = useState<string | null>(null)
  const [editedModel, setEditedModel] = useState<string | null>(null)
  const [showApiKey, setShowApiKey] = useState(false)
  const [saved, setSaved] = useState(false)

  const provider = editedProvider ?? settings?.provider ?? ''
  const defaultBaseUrl =
    provider === 'openai_compatible'
      ? 'https://api.openai.com/v1'
      : provider === 'gemini_compatible'
        ? 'https://generativelanguage.googleapis.com/v1beta'
        : provider === 'anthropic_compatible'
          ? 'https://api.anthropic.com'
          : ''
  const apiKey = editedApiKey ?? settings?.api_key ?? ''
  const baseUrl = editedBaseUrl ?? settings?.base_url ?? defaultBaseUrl
  const model = editedModel ?? settings?.model ?? ''
  const isEnabled = provider !== ''

  const providerOptions = [
    { value: '', label: t('providers.aiTranslation.disabled') },
    { value: 'openai_compatible', label: t('providers.aiTranslation.providers.openai') },
    { value: 'gemini_compatible', label: t('providers.aiTranslation.providers.gemini') },
    { value: 'anthropic_compatible', label: t('providers.aiTranslation.providers.anthropic') },
  ]

  const handleSave = (e: React.FormEvent) => {
    e.preventDefault()
    updateSettings(
      {
        provider,
        api_key: apiKey,
        base_url: baseUrl,
        model,
      },
      {
        onSuccess: () => {
          setSaved(true)
          setTimeout(() => setSaved(false), 2000)
        },
      },
    )
  }

  const handleProviderChange = (nextProvider: string) => {
    setEditedProvider(nextProvider)
    setEditedBaseUrl(
      nextProvider === ''
        ? ''
        : nextProvider === 'openai_compatible'
          ? 'https://api.openai.com/v1'
          : nextProvider === 'gemini_compatible'
            ? 'https://generativelanguage.googleapis.com/v1beta'
            : 'https://api.anthropic.com',
    )
  }

  if (isLoading) return <Spinner />

  return (
    <div className="rounded-lg bg-netflix-dark p-5">
      <div className="mb-1 flex items-center gap-2">
        <h3 className="text-sm font-semibold text-white">{t('providers.aiTranslation.name')}</h3>
        <span
          className={`rounded px-2 py-0.5 text-[10px] font-medium ${
            isEnabled ? 'bg-emerald-500/20 text-emerald-400' : 'bg-gray-500/20 text-gray-400'
          }`}
        >
          {isEnabled
            ? providerOptions.find((option) => option.value === provider)?.label
            : t('providers.aiTranslation.disabled')}
        </span>
      </div>
      <p className="mb-5 text-xs text-gray-400">{t('providers.aiTranslation.description')}</p>

      <form onSubmit={handleSave} className="space-y-4">
        <Field label={t('providers.aiTranslation.provider')} compact>
          <Select
            value={provider}
            onChange={(e) => handleProviderChange(e.target.value)}
            className="w-full"
          >
            {providerOptions.map((option) => (
              <option key={option.value || 'disabled'} value={option.value}>
                {option.label}
              </option>
            ))}
          </Select>
        </Field>

        <Field label={t('providers.aiTranslation.apiKey')} compact>
          <div className="relative">
            <input
              type={showApiKey ? 'text' : 'password'}
              value={apiKey}
              onChange={(e) => setEditedApiKey(e.target.value)}
              disabled={!isEnabled}
              placeholder={t('providers.aiTranslation.apiKeyPlaceholder')}
              className={`${inputClass} pr-10 disabled:opacity-50`}
            />
            <button
              type="button"
              onClick={() => setShowApiKey(!showApiKey)}
              className="absolute right-3 top-1/2 -translate-y-1/2 text-gray-500 hover:text-gray-300"
            >
              {showApiKey ? <LuEyeOff size={16} /> : <LuEye size={16} />}
            </button>
          </div>
        </Field>

        <Field label={t('providers.aiTranslation.baseUrl')} compact>
          <input
            type="text"
            value={baseUrl}
            onChange={(e) => setEditedBaseUrl(e.target.value)}
            disabled={!isEnabled}
            placeholder={defaultBaseUrl}
            className={`${inputClass} disabled:opacity-50`}
          />
        </Field>

        <Field label={t('providers.aiTranslation.model')} compact>
          <input
            type="text"
            value={model}
            onChange={(e) => setEditedModel(e.target.value)}
            disabled={!isEnabled}
            placeholder={t('providers.aiTranslation.modelPlaceholder')}
            className={`${inputClass} disabled:opacity-50`}
          />
        </Field>

        <p className="text-[11px] text-gray-500">{t('providers.aiTranslation.validationHint')}</p>

        <button
          type="submit"
          disabled={isSaving}
          className="flex items-center gap-2 rounded bg-netflix-red px-4 py-2 text-sm font-medium text-white transition-colors hover:bg-netflix-red-hover disabled:opacity-50"
        >
          {saved ? (
            <>
              <LuCheck size={14} /> {t('actions.saved')}
            </>
          ) : (
            <>
              <LuSave size={14} /> {isSaving ? t('actions.saving') : t('actions.save')}
            </>
          )}
        </button>
      </form>
    </div>
  )
}

function DeepLCard() {
  const { t } = useTranslation('settings')
  const { data: settings, isLoading } = useDeepLSettings()
  const { mutate: updateSettings, isPending: isSaving } = useUpdateDeepLSettings()
  const [editedApiKey, setEditedApiKey] = useState<string | null>(null)
  const apiKey = editedApiKey ?? settings?.api_key ?? ''
  const [showApiKey, setShowApiKey] = useState(false)
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
        <h3 className="text-sm font-semibold text-white">{t('providers.deepl.name')}</h3>
        <span className="rounded bg-blue-500/20 px-2 py-0.5 text-[10px] font-medium text-blue-400">
          {settings?.api_key ? 'DeepL' : 'Google Translate'}
        </span>
      </div>
      <p className="mb-5 text-xs text-gray-400">{t('providers.deepl.description')}</p>

      <form onSubmit={handleSave} className="space-y-4">
        <div>
          <span className="mb-1.5 block text-xs font-medium text-gray-300">
            {t('providers.deepl.getKey')}
          </span>
          <div className="relative">
            <input
              type={showApiKey ? 'text' : 'password'}
              value={apiKey}
              onChange={(e) => setEditedApiKey(e.target.value)}
              placeholder="Leave empty to use Google Translate"
              className={`${inputClass} pr-10`}
            />
            <button
              type="button"
              onClick={() => setShowApiKey(!showApiKey)}
              className="absolute right-3 top-1/2 -translate-y-1/2 text-gray-500 hover:text-gray-300"
            >
              {showApiKey ? <LuEyeOff size={16} /> : <LuEye size={16} />}
            </button>
          </div>
          <p className="mt-1.5 text-[11px] text-gray-500">{t('providers.deepl.getKey')}</p>
        </div>
        <button
          type="submit"
          disabled={isSaving}
          className="flex items-center gap-2 rounded bg-netflix-red px-4 py-2 text-sm font-medium text-white transition-colors hover:bg-netflix-red-hover disabled:opacity-50"
        >
          {saved ? (
            <>
              <LuCheck size={14} /> {t('actions.saved')}
            </>
          ) : (
            <>
              <LuSave size={14} /> {isSaving ? t('actions.saving') : t('actions.save')}
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
  const { t } = useTranslation('settings')
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
        <h3 className="text-sm font-semibold text-white">{t('subtitles.autoDownload')}</h3>
        {selected.length > 0 && (
          <span className="rounded bg-blue-500/20 px-2 py-0.5 text-[10px] font-medium text-blue-400">
            {selected.length} {t('fields.languages')}
          </span>
        )}
      </div>
      <p className="mb-4 text-xs text-gray-400">{t('subtitles.autoDownloadDescription')}</p>

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
          {t('subtitles.targetLanguagesDescription')}
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
            <LuCheck size={14} /> {t('actions.saved')}
          </>
        ) : (
          <>
            <LuSave size={14} /> {isSaving ? t('actions.saving') : t('actions.save')}
          </>
        )}
      </button>
    </div>
  )
}
