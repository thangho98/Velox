import { useTranslation } from '@/hooks/useTranslation'
import { useState } from 'react'
import { Select } from '@/components/ui/Select'
import {
  LuPlus,
  LuRefreshCw,
  LuTrash2,
  LuX,
  LuFolder,
  LuFilm,
  LuTv,
  LuList,
  LuLibrary,
} from 'react-icons/lu'
import {
  useLibraries,
  useCreateLibrary,
  useDeleteLibrary,
  useScanLibrary,
} from '@/hooks/stores/useMedia'
import { useStorageProviders } from '@/hooks/stores/useAdmin'
import { DirectoryPicker } from '@/components/DirectoryPicker'
import { SectionHeader, Spinner, Modal, Field, ErrorMsg, inputClass } from './shared'

const LIBRARY_TYPES = [
  { value: 'movies', label: 'Movies', description: 'Feature films', icon: <LuFilm size={20} /> },
  {
    value: 'tvshows',
    label: 'TV Shows',
    description: 'Series & episodes',
    icon: <LuTv size={20} />,
  },
  {
    value: 'anime',
    label: 'Anime',
    description: 'AniList-powered anime',
    icon: <LuTv size={20} />,
  },
  { value: 'mixed', label: 'Mixed', description: 'Movies & TV', icon: <LuList size={20} /> },
]

const TYPE_ICON_BG: Record<string, string> = {
  movies: 'bg-blue-500/20 text-blue-400',
  tvshows: 'bg-purple-500/20 text-purple-400',
  anime: 'bg-amber-500/20 text-amber-400',
  mixed: 'bg-green-500/20 text-green-400',
}

const TYPE_COLORS: Record<string, string> = {
  movies: 'bg-blue-500/20 text-blue-400 border-blue-500',
  tvshows: 'bg-purple-500/20 text-purple-400 border-purple-500',
  anime: 'bg-amber-500/20 text-amber-400 border-amber-500',
  mixed: 'bg-green-500/20 text-green-400 border-green-500',
}

type SourceKind = 'local' | 'cloud'

interface LibraryFormData {
  name: string
  paths: string[]
  type: string
  source: SourceKind
  storageProviderID?: number
  sourceURL: string
}

const DEFAULT_LIB_FORM: LibraryFormData = {
  name: '',
  paths: [''],
  type: 'movies',
  source: 'local',
  sourceURL: '',
}

export function LibrariesSection() {
  const { t } = useTranslation('settings')

  const { data: libraries, isLoading } = useLibraries()
  const { mutate: createLibrary, isPending: isCreating } = useCreateLibrary()
  const { mutate: deleteLibrary } = useDeleteLibrary()
  const { mutate: scanLibrary } = useScanLibrary()
  const { data: providers } = useStorageProviders()

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

    if (formData.source === 'cloud') {
      if (!formData.storageProviderID) {
        setFormError('Select a storage provider')
        return
      }
      if (!formData.sourceURL.trim()) {
        setFormError('Folder URL is required')
        return
      }
      createLibrary(
        {
          name: formData.name.trim(),
          type: formData.type,
          storage_provider_id: formData.storageProviderID,
          source_url: formData.sourceURL.trim(),
        },
        {
          onSuccess: () => {
            setShowAddModal(false)
            setFormData(DEFAULT_LIB_FORM)
          },
          onError: (err: Error) => setFormError(err.message || 'Failed to create library'),
        },
      )
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
            <Field label="Source">
              <div className="grid grid-cols-2 gap-2">
                <button
                  type="button"
                  onClick={() => setFormData({ ...formData, source: 'local' })}
                  className={`rounded border-2 px-3 py-2 text-sm ${
                    formData.source === 'local'
                      ? 'border-netflix-red bg-netflix-red/10 text-white'
                      : 'border-white/10 bg-netflix-gray text-gray-400 hover:border-white/20'
                  }`}
                >
                  📁 Local Filesystem
                </button>
                <button
                  type="button"
                  onClick={() => setFormData({ ...formData, source: 'cloud' })}
                  disabled={!providers || providers.length === 0}
                  className={`rounded border-2 px-3 py-2 text-sm disabled:opacity-40 ${
                    formData.source === 'cloud'
                      ? 'border-netflix-red bg-netflix-red/10 text-white'
                      : 'border-white/10 bg-netflix-gray text-gray-400 hover:border-white/20'
                  }`}
                  title={
                    !providers || providers.length === 0
                      ? 'Add a storage provider in Settings → Storage first'
                      : undefined
                  }
                >
                  ☁️ Cloud Storage
                </button>
              </div>
            </Field>

            {formData.source === 'local' ? (
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
            ) : (
              <>
                <Field label="Storage provider">
                  <Select
                    value={formData.storageProviderID ?? ''}
                    onChange={(e) =>
                      setFormData({
                        ...formData,
                        storageProviderID: Number(e.target.value) || undefined,
                      })
                    }
                    className="w-full"
                    required
                  >
                    <option value="">Choose provider…</option>
                    {providers?.map((p) => (
                      <option
                        key={p.id}
                        value={p.id}
                        disabled={p.health === 'expired' || p.health === 'error'}
                      >
                        {p.display_name} · {p.account_email}
                        {p.health !== 'healthy' ? ` (${p.health})` : ''}
                      </option>
                    ))}
                  </Select>
                </Field>
                <Field label="Folder URL">
                  <input
                    type="url"
                    value={formData.sourceURL}
                    onChange={(e) => setFormData({ ...formData, sourceURL: e.target.value })}
                    placeholder="https://www.fshare.vn/folder/XZWCPAZV3J71"
                    className={inputClass}
                    required
                  />
                  <p className="mt-1 text-xs text-gray-500">
                    Paste a fshare folder URL from your browser.
                  </p>
                </Field>
              </>
            )}
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
