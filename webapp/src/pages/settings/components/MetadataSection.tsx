import { useState } from 'react'
import { LuSave, LuCheck, LuRefreshCw, LuEye, LuEyeOff } from 'react-icons/lu'
import {
  useTMDbSettings,
  useUpdateTMDbSettings,
  useOMDbSettings,
  useUpdateOMDbSettings,
  useTVDBSettings,
  useUpdateTVDBSettings,
  useFanartSettings,
  useUpdateFanartSettings,
  useBulkRefreshRatings,
} from '@/hooks/stores/useSettings'
import { useTranslation } from '@/hooks/useTranslation'
import { SectionHeader, Spinner, inputClass } from './shared'

export function MetadataSection() {
  const { t } = useTranslation('settings')
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
  const [showTmdb, setShowTmdb] = useState(false)
  const [omdbEdited, setOmdbEdited] = useState<string | null>(null)
  const [omdbSaved, setOmdbSaved] = useState(false)
  const [showOmdb, setShowOmdb] = useState(false)
  const [tvdbEdited, setTvdbEdited] = useState<string | null>(null)
  const [tvdbSaved, setTvdbSaved] = useState(false)
  const [showTvdb, setShowTvdb] = useState(false)
  const [fanartEdited, setFanartEdited] = useState<string | null>(null)
  const [fanartSaved, setFanartSaved] = useState(false)
  const [showFanart, setShowFanart] = useState(false)

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
          <h3 className="text-sm font-semibold text-white">{t('providers.tmdb.name')}</h3>
          <span
            className={`rounded px-2 py-0.5 text-[10px] font-medium ${
              tmdbSettings?.api_key
                ? 'bg-blue-500/20 text-blue-400'
                : tmdbSettings?.has_builtin
                  ? 'bg-green-500/20 text-green-400'
                  : 'bg-gray-500/20 text-gray-400'
            }`}
          >
            {tmdbSettings?.api_key
              ? t('status.customKey')
              : tmdbSettings?.has_builtin
                ? t('status.envKey')
                : t('status.notConfigured')}
          </span>
        </div>
        <p className="mb-5 text-xs text-gray-400">
          {t('providers.tmdb.description')}{' '}
          {tmdbSettings?.has_builtin
            ? t('providers.tmdb.hasBuiltin')
            : t('providers.tmdb.noBuiltin')}{' '}
          <a
            href="https://www.themoviedb.org/settings/api"
            target="_blank"
            rel="noopener noreferrer"
            className="text-netflix-red hover:underline"
          >
            {t('actions.getFreeKey')}
          </a>
        </p>

        <form onSubmit={handleTmdbSave} className="space-y-5">
          <div className="space-y-3">
            <span className="text-xs font-medium text-gray-300">{t('providers.tmdb.v4Token')}</span>
            <p className="text-[11px] text-gray-500">
              {tmdbSettings?.has_builtin
                ? t('providers.tmdb.optional')
                : t('providers.tmdb.required')}
            </p>
            <div className="relative">
              <input
                type={showTmdb ? 'text' : 'password'}
                value={tmdbKey}
                onChange={(e) => setTmdbEdited(e.target.value)}
                placeholder={
                  tmdbSettings?.has_builtin
                    ? t('providers.tmdb.placeholderOptional')
                    : t('providers.tmdb.placeholderRequired')
                }
                className={`${inputClass} pr-10`}
              />
              <button
                type="button"
                onClick={() => setShowTmdb(!showTmdb)}
                className="absolute right-3 top-1/2 -translate-y-1/2 text-gray-500 hover:text-gray-300"
              >
                {showTmdb ? <LuEyeOff size={16} /> : <LuEye size={16} />}
              </button>
            </div>
          </div>

          <div className="pt-1">
            <button
              type="submit"
              disabled={tmdbSaving}
              className="flex items-center gap-2 rounded bg-netflix-red px-4 py-2 text-sm font-medium text-white transition-colors hover:bg-netflix-red-hover disabled:opacity-50"
            >
              {tmdbSaved ? (
                <>
                  <LuCheck size={14} /> {t('actions.saved')}
                </>
              ) : (
                <>
                  <LuSave size={14} /> {tmdbSaving ? t('actions.saving') : t('actions.save')}
                </>
              )}
            </button>
          </div>
        </form>
      </div>

      <div className="rounded-lg bg-netflix-dark p-5">
        <div className="mb-1 flex items-center gap-2">
          <h3 className="text-sm font-semibold text-white">{t('providers.omdb.name')}</h3>
          <span
            className={`rounded px-2 py-0.5 text-[10px] font-medium ${
              omdbSettings?.api_key
                ? 'bg-blue-500/20 text-blue-400'
                : omdbSettings?.has_builtin
                  ? 'bg-green-500/20 text-green-400'
                  : 'bg-gray-500/20 text-gray-400'
            }`}
          >
            {omdbSettings?.api_key
              ? t('status.customKey')
              : omdbSettings?.has_builtin
                ? t('status.envKey')
                : t('status.notConfigured')}
          </span>
        </div>
        <p className="mb-5 text-xs text-gray-400">
          {t('providers.omdb.description')}{' '}
          {omdbSettings?.has_builtin
            ? t('providers.omdb.hasBuiltin')
            : t('providers.omdb.noBuiltin')}{' '}
          <a
            href="https://www.omdbapi.com/apikey.aspx"
            target="_blank"
            rel="noopener noreferrer"
            className="text-netflix-red hover:underline"
          >
            {t('actions.getFreeKey')}
          </a>
        </p>

        <form onSubmit={handleOmdbSave} className="space-y-5">
          <div className="space-y-3">
            <span className="text-xs font-medium text-gray-300">{t('fields.customApiKey')}</span>
            <p className="text-[11px] text-gray-500">
              {omdbSettings?.has_builtin
                ? t('providers.omdb.optional')
                : t('providers.omdb.required')}
            </p>
            <div className="relative">
              <input
                type={showOmdb ? 'text' : 'password'}
                value={omdbKey}
                onChange={(e) => setOmdbEdited(e.target.value)}
                placeholder={t('providers.omdb.placeholder')}
                className={`${inputClass} pr-10`}
              />
              <button
                type="button"
                onClick={() => setShowOmdb(!showOmdb)}
                className="absolute right-3 top-1/2 -translate-y-1/2 text-gray-500 hover:text-gray-300"
              >
                {showOmdb ? <LuEyeOff size={16} /> : <LuEye size={16} />}
              </button>
            </div>
          </div>

          <div className="pt-1">
            <button
              type="submit"
              disabled={omdbSaving}
              className="flex items-center gap-2 rounded bg-netflix-red px-4 py-2 text-sm font-medium text-white transition-colors hover:bg-netflix-red-hover disabled:opacity-50"
            >
              {omdbSaved ? (
                <>
                  <LuCheck size={14} /> {t('actions.saved')}
                </>
              ) : (
                <>
                  <LuSave size={14} /> {omdbSaving ? t('actions.saving') : t('actions.save')}
                </>
              )}
            </button>
          </div>
        </form>
      </div>

      <div className="rounded-lg bg-netflix-dark p-5">
        <div className="mb-1 flex items-center gap-2">
          <h3 className="text-sm font-semibold text-white">{t('providers.tvdb.name')}</h3>
          <span
            className={`rounded px-2 py-0.5 text-[10px] font-medium ${
              tvdbSettings?.api_key
                ? 'bg-blue-500/20 text-blue-400'
                : tvdbSettings?.has_builtin
                  ? 'bg-green-500/20 text-green-400'
                  : 'bg-gray-500/20 text-gray-400'
            }`}
          >
            {tvdbSettings?.api_key
              ? t('status.customKey')
              : tvdbSettings?.has_builtin
                ? t('status.envKey')
                : t('status.notConfigured')}
          </span>
        </div>
        <p className="mb-5 text-xs text-gray-400">
          {t('providers.tvdb.description')}{' '}
          {tvdbSettings?.has_builtin
            ? t('providers.tvdb.hasBuiltin')
            : t('providers.tvdb.noBuiltin')}{' '}
          <a
            href="https://thetvdb.com/api-information"
            target="_blank"
            rel="noopener noreferrer"
            className="text-netflix-red hover:underline"
          >
            {t('actions.getFreeKey')}
          </a>
        </p>

        <form onSubmit={handleTvdbSave} className="space-y-5">
          <div className="space-y-3">
            <span className="text-xs font-medium text-gray-300">{t('fields.customApiKey')}</span>
            <p className="text-[11px] text-gray-500">
              {tvdbSettings?.has_builtin
                ? t('providers.tvdb.optional')
                : t('providers.tvdb.required')}
            </p>
            <div className="relative">
              <input
                type={showTvdb ? 'text' : 'password'}
                value={tvdbKey}
                onChange={(e) => setTvdbEdited(e.target.value)}
                placeholder={t('providers.tvdb.placeholder')}
                className={`${inputClass} pr-10`}
              />
              <button
                type="button"
                onClick={() => setShowTvdb(!showTvdb)}
                className="absolute right-3 top-1/2 -translate-y-1/2 text-gray-500 hover:text-gray-300"
              >
                {showTvdb ? <LuEyeOff size={16} /> : <LuEye size={16} />}
              </button>
            </div>
          </div>

          <div className="pt-1">
            <button
              type="submit"
              disabled={tvdbSaving}
              className="flex items-center gap-2 rounded bg-netflix-red px-4 py-2 text-sm font-medium text-white transition-colors hover:bg-netflix-red-hover disabled:opacity-50"
            >
              {tvdbSaved ? (
                <>
                  <LuCheck size={14} /> {t('actions.saved')}
                </>
              ) : (
                <>
                  <LuSave size={14} /> {tvdbSaving ? t('actions.saving') : t('actions.save')}
                </>
              )}
            </button>
          </div>
        </form>
      </div>

      <div className="rounded-lg bg-netflix-dark p-5">
        <div className="mb-1 flex items-center gap-2">
          <h3 className="text-sm font-semibold text-white">{t('providers.fanart.name')}</h3>
          <span
            className={`rounded px-2 py-0.5 text-[10px] font-medium ${
              fanartSettings?.api_key
                ? 'bg-blue-500/20 text-blue-400'
                : fanartSettings?.has_builtin
                  ? 'bg-green-500/20 text-green-400'
                  : 'bg-gray-500/20 text-gray-400'
            }`}
          >
            {fanartSettings?.api_key
              ? t('status.customKey')
              : fanartSettings?.has_builtin
                ? t('status.envKey')
                : t('status.notConfigured')}
          </span>
        </div>
        <p className="mb-5 text-xs text-gray-400">
          {t('providers.fanart.description')}{' '}
          {fanartSettings?.has_builtin
            ? t('providers.fanart.hasBuiltin')
            : t('providers.fanart.noBuiltin')}{' '}
          <a
            href="https://fanart.tv/get-an-api-key/"
            target="_blank"
            rel="noopener noreferrer"
            className="text-netflix-red hover:underline"
          >
            {t('actions.getFreeKey')}
          </a>
        </p>

        <form onSubmit={handleFanartSave} className="space-y-5">
          <div className="space-y-3">
            <span className="text-xs font-medium text-gray-300">{t('fields.customApiKey')}</span>
            <p className="text-[11px] text-gray-500">
              {fanartSettings?.has_builtin
                ? t('providers.fanart.optional')
                : t('providers.fanart.required')}
            </p>
            <div className="relative">
              <input
                type={showFanart ? 'text' : 'password'}
                value={fanartKey}
                onChange={(e) => setFanartEdited(e.target.value)}
                placeholder={t('providers.fanart.placeholder')}
                className={`${inputClass} pr-10`}
              />
              <button
                type="button"
                onClick={() => setShowFanart(!showFanart)}
                className="absolute right-3 top-1/2 -translate-y-1/2 text-gray-500 hover:text-gray-300"
              >
                {showFanart ? <LuEyeOff size={16} /> : <LuEye size={16} />}
              </button>
            </div>
          </div>

          <div className="pt-1">
            <button
              type="submit"
              disabled={fanartSaving}
              className="flex items-center gap-2 rounded bg-netflix-red px-4 py-2 text-sm font-medium text-white transition-colors hover:bg-netflix-red-hover disabled:opacity-50"
            >
              {fanartSaved ? (
                <>
                  <LuCheck size={14} /> {t('actions.saved')}
                </>
              ) : (
                <>
                  <LuSave size={14} /> {fanartSaving ? t('actions.saving') : t('actions.save')}
                </>
              )}
            </button>
          </div>
        </form>
      </div>

      <div className="rounded-lg bg-netflix-dark p-5">
        <div className="mb-1 flex items-center gap-2">
          <h3 className="text-sm font-semibold text-white">{t('actions.refreshAllMetadata')}</h3>
        </div>
        <p className="mb-4 text-xs text-gray-400">{t('messages.howItWorksDescription')}</p>
        <div className="flex items-center gap-3">
          <button
            onClick={() => bulkRefresh()}
            disabled={isRefreshing}
            className="flex items-center gap-2 rounded bg-netflix-red px-4 py-2 text-sm font-medium text-white transition-colors hover:bg-netflix-red-hover disabled:opacity-50"
          >
            {isRefreshing ? (
              <>
                <LuRefreshCw size={14} className="animate-spin" /> {t('actions.refreshing')}
              </>
            ) : (
              <>
                <LuRefreshCw size={14} /> {t('actions.refreshAllMetadata')}
              </>
            )}
          </button>
          {refreshResult && !isRefreshing && (
            <span className="text-xs text-green-400">
              {t('messages.updated', { count: refreshResult.updated })}
            </span>
          )}
          {refreshError && !isRefreshing && (
            <span className="text-xs text-red-400">
              {t('messages.error', { message: refreshError.message })}
            </span>
          )}
        </div>
      </div>

      <div className="rounded-lg bg-netflix-dark p-5">
        <h3 className="mb-1 text-sm font-semibold text-white">{t('messages.howItWorks')}</h3>
        <ul className="space-y-1 text-xs text-gray-400">
          <li>• {t('messages.howItWorksDescription')}</li>
        </ul>
      </div>
    </div>
  )
}
