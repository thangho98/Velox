import { useState } from 'react'
import { LuSave, LuCheck, LuRefreshCw, LuEye, LuEyeOff, LuClipboardPaste } from 'react-icons/lu'
import {
  useAniListSettings,
  useUpdateAniListSettings,
  useTMDbSettings,
  useUpdateTMDbSettings,
  useOMDbSettings,
  useUpdateOMDbSettings,
  useTVDBSettings,
  useUpdateTVDBSettings,
  useFanartSettings,
  useUpdateFanartSettings,
  useBulkRefreshRatings,
  useBulkForceRefreshMetadata,
} from '@/hooks/stores/useSettings'
import { useTranslation } from '@/hooks/useTranslation'
import { SectionHeader, Spinner, inputClass } from './shared'

export function MetadataSection() {
  const { t } = useTranslation('settings')
  const { data: anilistSettings, isLoading: anilistLoading } = useAniListSettings()
  const { mutate: updateAniList, isPending: anilistSaving } = useUpdateAniListSettings()
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
  const {
    mutate: bulkForceRefresh,
    isPending: isForceRefreshing,
    data: forceRefreshResult,
    error: forceRefreshError,
  } = useBulkForceRefreshMetadata()

  const [anilistEdited, setAniListEdited] = useState<string | null>(null)
  const [anilistSaved, setAniListSaved] = useState(false)
  const [showAniList, setShowAniList] = useState(false)
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

  const anilistKey = anilistEdited ?? anilistSettings?.api_key ?? ''
  const tmdbKey = tmdbEdited ?? tmdbSettings?.api_key ?? ''
  const omdbKey = omdbEdited ?? omdbSettings?.api_key ?? ''

  const handleAniListSave = (e: React.FormEvent) => {
    e.preventDefault()
    updateAniList(
      { api_key: anilistKey },
      {
        onSuccess: () => {
          setAniListEdited(null)
          setAniListSaved(true)
          setTimeout(() => setAniListSaved(false), 2000)
        },
      },
    )
  }

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

  if (anilistLoading || tmdbLoading || omdbLoading || tvdbLoading || fanartLoading)
    return <Spinner />

  return (
    <div className="max-w-xl space-y-6">
      <SectionHeader
        title={t('sections.metadata.title')}
        description={t('sections.metadata.description')}
      />

      <div className="rounded-lg bg-netflix-dark p-5">
        <div className="mb-1 flex items-center gap-2">
          <h3 className="text-sm font-semibold text-white">{t('providers.anilist.name')}</h3>
          <span
            className={`rounded px-2 py-0.5 text-[10px] font-medium ${
              anilistSettings?.api_key
                ? 'bg-blue-500/20 text-blue-400'
                : anilistSettings?.has_builtin
                  ? 'bg-green-500/20 text-green-400'
                  : 'bg-gray-500/20 text-gray-400'
            }`}
          >
            {anilistSettings?.api_key
              ? t('providers.anilist.status.customToken')
              : anilistSettings?.has_builtin
                ? t('providers.anilist.status.envToken')
                : t('status.notConfigured')}
          </span>
        </div>
        <p className="mb-5 text-xs text-gray-400">
          {t('providers.anilist.description')}{' '}
          {anilistSettings?.has_builtin
            ? t('providers.anilist.hasBuiltin')
            : t('providers.anilist.noBuiltin')}{' '}
          <a
            href="https://anilist.gitbook.io/anilist-apiv2-docs"
            target="_blank"
            rel="noopener noreferrer"
            className="text-netflix-red hover:underline"
          >
            {t('actions.learnMore')}
          </a>
        </p>

        <form onSubmit={handleAniListSave} className="space-y-5">
          <div className="space-y-3">
            <span className="text-xs font-medium text-gray-300">
              {t('providers.anilist.token')}
            </span>
            <p className="text-[11px] text-gray-500">
              {anilistSettings?.has_builtin
                ? t('providers.anilist.optional')
                : t('providers.anilist.optionalNoBuiltin')}
            </p>
            <p className="text-[11px] text-gray-500">
              {t('providers.anilist.authNote')}{' '}
              <a
                href="https://project-s8tij.vercel.app"
                target="_blank"
                rel="noopener noreferrer"
                className="text-netflix-red hover:underline"
              >
                Connect AniList
              </a>{' '}
              to get your token.
            </p>
            <div className="relative">
              <input
                type={showAniList ? 'text' : 'password'}
                value={anilistKey}
                onChange={(e) => setAniListEdited(e.target.value)}
                placeholder={t('providers.anilist.placeholder')}
                className={`${inputClass} pr-20`}
              />
              <div className="absolute right-3 top-1/2 flex -translate-y-1/2 items-center gap-1">
                <button
                  type="button"
                  onClick={async () => {
                    const text = await navigator.clipboard.readText()
                    if (text) setAniListEdited(text.trim())
                  }}
                  className="text-gray-500 hover:text-gray-300"
                  title="Paste from clipboard"
                >
                  <LuClipboardPaste size={16} />
                </button>
                <button
                  type="button"
                  onClick={() => setShowAniList(!showAniList)}
                  className="text-gray-500 hover:text-gray-300"
                >
                  {showAniList ? <LuEyeOff size={16} /> : <LuEye size={16} />}
                </button>
              </div>
            </div>
          </div>

          <div className="pt-1">
            <button
              type="submit"
              disabled={anilistSaving}
              className="flex items-center gap-2 rounded bg-netflix-red px-4 py-2 text-sm font-medium text-white transition-colors hover:bg-netflix-red-hover disabled:opacity-50"
            >
              {anilistSaved ? (
                <>
                  <LuCheck size={14} /> {t('actions.saved')}
                </>
              ) : (
                <>
                  <LuSave size={14} /> {anilistSaving ? t('actions.saving') : t('actions.save')}
                </>
              )}
            </button>
          </div>
        </form>
      </div>

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
        <h3 className="mb-2 text-sm font-semibold text-white">{t('providers.refresh.title')}</h3>
        <p className="mb-5 text-xs text-gray-400">{t('providers.refresh.description')}</p>
        <div className="flex flex-col gap-3">
          <div className="flex items-center gap-3">
            <button
              onClick={() => bulkRefresh()}
              disabled={isRefreshing || isForceRefreshing}
              className="flex items-center gap-2 rounded bg-gray-600 px-4 py-2 text-sm font-medium text-white transition-colors hover:bg-gray-500 disabled:opacity-50"
            >
              {isRefreshing ? (
                <>
                  <LuRefreshCw size={14} className="animate-spin" /> {t('actions.refreshing')}
                </>
              ) : (
                <>
                  <LuRefreshCw size={14} /> {t('actions.refreshMissingMetadata')}
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
          <div className="flex items-center gap-3">
            <button
              onClick={() => {
                if (window.confirm(t('providers.refresh.confirmForce'))) {
                  bulkForceRefresh()
                }
              }}
              disabled={isRefreshing || isForceRefreshing}
              className="flex items-center gap-2 rounded bg-netflix-red px-4 py-2 text-sm font-medium text-white transition-colors hover:bg-netflix-red-hover disabled:opacity-50"
            >
              {isForceRefreshing ? (
                <>
                  <LuRefreshCw size={14} className="animate-spin" /> {t('actions.refreshing')}
                </>
              ) : (
                <>
                  <LuRefreshCw size={14} /> {t('actions.refreshAllMetadata')}
                </>
              )}
            </button>
            {forceRefreshResult && !isForceRefreshing && (
              <span className="text-xs text-green-400">
                {t('messages.updated', { count: forceRefreshResult.updated })}
              </span>
            )}
            {forceRefreshError && !isForceRefreshing && (
              <span className="text-xs text-red-400">
                {t('messages.error', { message: forceRefreshError.message })}
              </span>
            )}
          </div>
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
