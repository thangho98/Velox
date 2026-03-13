import { useState } from 'react'
import {
  useLibraries,
  useCreateLibrary,
  useDeleteLibrary,
  useScanLibrary,
} from '@/hooks/stores/useMedia'
import { DirectoryPicker } from '@/components/DirectoryPicker'
import {
  LuFilm,
  LuTv,
  LuList,
  LuPlus,
  LuLibrary,
  LuRefreshCw,
  LuTrash2,
  LuX,
  LuFolder,
} from 'react-icons/lu'

// ── Library type definitions ──────────────────────────────────────────────────

interface LibraryTypeOption {
  value: string
  label: string
  description: string
  icon: React.ReactNode
  color: string
}

const LIBRARY_TYPES: LibraryTypeOption[] = [
  {
    value: 'movies',
    label: 'Movies',
    description: 'Feature films',
    color: 'blue',
    icon: <LuFilm size={20} />,
  },
  {
    value: 'tvshows',
    label: 'TV Shows',
    description: 'Series & episodes',
    color: 'purple',
    icon: <LuTv size={20} />,
  },
  {
    value: 'mixed',
    label: 'Mixed Content',
    description: 'Movies & TV combined',
    color: 'green',
    icon: <LuList size={20} />,
  },
]

const TYPE_COLORS: Record<string, string> = {
  movies: 'bg-blue-500/20 text-blue-400 border-blue-500',
  tvshows: 'bg-purple-500/20 text-purple-400 border-purple-500',
  mixed: 'bg-green-500/20 text-green-400 border-green-500',
}

const TYPE_ICON_BG: Record<string, string> = {
  movies: 'bg-blue-500/20 text-blue-400',
  tvshows: 'bg-purple-500/20 text-purple-400',
  mixed: 'bg-green-500/20 text-green-400',
}

// ── Form state ────────────────────────────────────────────────────────────────

interface LibraryFormData {
  name: string
  paths: string[]
  type: string
}

const DEFAULT_FORM: LibraryFormData = {
  name: '',
  paths: [''],
  type: 'movies',
}

// ── Component ─────────────────────────────────────────────────────────────────

export function AdminLibrariesPage() {
  const { data: libraries, isLoading } = useLibraries()
  const { mutate: createLibrary, isPending: isCreating } = useCreateLibrary()
  const { mutate: deleteLibrary } = useDeleteLibrary()
  const { mutate: scanLibrary } = useScanLibrary()

  const [showAddModal, setShowAddModal] = useState(false)
  const [dirPickerIndex, setDirPickerIndex] = useState<number | null>(null)
  const [formData, setFormData] = useState<LibraryFormData>(DEFAULT_FORM)
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
          setFormData(DEFAULT_FORM)
        },
        onError: (err: Error) => {
          setFormError(err.message || 'Failed to create library')
        },
      },
    )
  }

  const handleDelete = (id: number, name: string) => {
    if (confirm(`Delete "${name}"? This cannot be undone.`)) {
      deleteLibrary(id)
    }
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
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold text-white">Library Management</h1>
          <p className="text-gray-400">
            {libraries?.length || 0} {libraries?.length === 1 ? 'library' : 'libraries'} configured
          </p>
        </div>
        <button
          onClick={() => setShowAddModal(true)}
          className="flex items-center gap-2 rounded bg-netflix-red px-4 py-2 font-semibold text-white transition-colors hover:bg-netflix-red-hover"
        >
          <LuPlus size={20} />
          Add Library
        </button>
      </div>

      {/* Library List */}
      {isLoading ? (
        <div className="flex h-64 items-center justify-center">
          <div className="h-8 w-8 animate-spin rounded-full border-2 border-netflix-red border-t-transparent" />
        </div>
      ) : libraries?.length === 0 ? (
        <div className="flex h-64 flex-col items-center justify-center rounded-lg bg-netflix-dark">
          <LuLibrary size={48} className="mb-4 text-gray-600" />
          <p className="text-gray-400">No libraries configured</p>
          <p className="text-sm text-gray-500">
            Add your first library to start organizing your media
          </p>
          <button
            onClick={() => setShowAddModal(true)}
            className="mt-4 rounded bg-netflix-red px-4 py-2 text-sm font-medium text-white transition-colors hover:bg-netflix-red-hover"
          >
            Add Library
          </button>
        </div>
      ) : (
        <div className="space-y-4">
          {libraries?.map((lib) => {
            const opt = typeOption(lib.type)
            return (
              <div
                key={lib.id}
                className="flex items-center justify-between rounded-lg bg-netflix-dark p-4 transition-colors hover:bg-netflix-gray"
              >
                <div className="flex min-w-0 items-center gap-4">
                  <div
                    className={`flex h-12 w-12 shrink-0 items-center justify-center rounded-lg ${TYPE_ICON_BG[lib.type] ?? 'bg-gray-500/20 text-gray-400'}`}
                  >
                    {opt?.icon}
                  </div>
                  <div className="min-w-0">
                    <h3 className="font-semibold text-white">{lib.name}</h3>
                    {/* Paths list */}
                    <div className="mt-0.5 space-y-0.5">
                      {lib.paths?.map((p) => (
                        <p key={p} className="truncate font-mono text-xs text-gray-400">
                          {p}
                        </p>
                      ))}
                    </div>
                    <div className="mt-1 flex items-center gap-2">
                      <span
                        className={`rounded px-2 py-0.5 text-xs border ${TYPE_COLORS[lib.type] ?? 'bg-gray-500/20 text-gray-400 border-gray-500'}`}
                      >
                        {opt?.label ?? lib.type}
                      </span>
                      <span className="text-xs text-gray-500">
                        {lib.paths?.length > 1 ? `${lib.paths.length} folders` : '1 folder'} ·{' '}
                        {new Date(lib.created_at).toLocaleDateString()}
                      </span>
                    </div>
                  </div>
                </div>

                <div className="flex shrink-0 items-center gap-2">
                  <button
                    onClick={() => handleScan(lib.id)}
                    disabled={scanningId === lib.id}
                    className="flex items-center gap-1 rounded bg-netflix-gray px-3 py-1.5 text-sm text-white transition-colors hover:bg-blue-600 disabled:opacity-50"
                  >
                    {scanningId === lib.id ? (
                      <>
                        <div className="h-4 w-4 animate-spin rounded-full border-2 border-white border-t-transparent" />
                        Scanning...
                      </>
                    ) : (
                      <>
                        <LuRefreshCw size={16} />
                        Scan
                      </>
                    )}
                  </button>
                  <button
                    onClick={() => handleScan(lib.id, true)}
                    disabled={scanningId === lib.id}
                    className="flex items-center gap-1 rounded bg-netflix-gray px-3 py-1.5 text-sm text-white transition-colors hover:bg-amber-600 disabled:opacity-50"
                    title="Re-parse all filenames and update titles"
                  >
                    <LuRefreshCw size={16} />
                    Force Rescan
                  </button>
                  <button
                    onClick={() => handleDelete(lib.id, lib.name)}
                    className="rounded bg-netflix-gray px-3 py-1.5 text-sm text-white transition-colors hover:bg-red-600"
                  >
                    <LuTrash2 size={16} />
                  </button>
                </div>
              </div>
            )
          })}
        </div>
      )}

      {/* Directory Picker overlay */}
      {dirPickerIndex !== null && (
        <DirectoryPicker
          onSelect={(path) => {
            setPath(dirPickerIndex, path)
            setDirPickerIndex(null)
          }}
          onClose={() => setDirPickerIndex(null)}
        />
      )}

      {/* Add Library Modal */}
      {showAddModal && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/70 p-4">
          <div className="w-full max-w-lg rounded-lg bg-netflix-dark p-6">
            <div className="mb-6 flex items-center justify-between">
              <h2 className="text-xl font-bold text-white">Add New Library</h2>
              <button
                onClick={() => setShowAddModal(false)}
                className="text-gray-400 hover:text-white"
              >
                <LuX size={24} />
              </button>
            </div>

            <form onSubmit={handleSubmit} className="space-y-5">
              {formError && (
                <div className="rounded-lg bg-netflix-red/20 p-3 text-sm text-netflix-red">
                  {formError}
                </div>
              )}

              {/* Name */}
              <div>
                <label className="mb-2 block text-sm font-medium text-gray-400">Library Name</label>
                <input
                  type="text"
                  value={formData.name}
                  onChange={(e) => setFormData({ ...formData, name: e.target.value })}
                  placeholder="e.g., My Movies"
                  className="w-full rounded bg-netflix-gray px-4 py-3 text-white outline-none ring-1 ring-transparent transition-all focus:ring-netflix-red"
                  required
                />
              </div>

              {/* Content Type */}
              <div>
                <label className="mb-2 block text-sm font-medium text-gray-400">Content Type</label>
                <div className="grid grid-cols-3 gap-2">
                  {LIBRARY_TYPES.map((t) => {
                    const isSelected = formData.type === t.value
                    const colors: Record<string, string> = {
                      movies: isSelected
                        ? 'border-blue-500 bg-blue-500/15 text-blue-300'
                        : 'border-white/10 bg-netflix-gray text-gray-400 hover:border-white/20',
                      tvshows: isSelected
                        ? 'border-purple-500 bg-purple-500/15 text-purple-300'
                        : 'border-white/10 bg-netflix-gray text-gray-400 hover:border-white/20',
                      mixed: isSelected
                        ? 'border-green-500 bg-green-500/15 text-green-300'
                        : 'border-white/10 bg-netflix-gray text-gray-400 hover:border-white/20',
                    }
                    return (
                      <button
                        key={t.value}
                        type="button"
                        onClick={() => setFormData({ ...formData, type: t.value })}
                        className={`flex flex-col items-center gap-1.5 rounded-lg border-2 px-3 py-3 text-center transition-colors ${colors[t.value]}`}
                      >
                        {t.icon}
                        <span className="text-sm font-medium">{t.label}</span>
                        <span className="text-xs opacity-70">{t.description}</span>
                      </button>
                    )
                  })}
                </div>
              </div>

              {/* Folder Paths */}
              <div>
                <div className="mb-2 flex items-center justify-between">
                  <label className="text-sm font-medium text-gray-400">Folders</label>
                  <button
                    type="button"
                    onClick={addPath}
                    className="flex items-center gap-1 text-xs text-gray-400 transition-colors hover:text-white"
                  >
                    <LuPlus size={14} />
                    Add folder
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
                        className="min-w-0 flex-1 rounded bg-netflix-gray px-4 py-2.5 font-mono text-sm text-white outline-none ring-1 ring-transparent transition-all focus:ring-netflix-red"
                      />
                      <button
                        type="button"
                        onClick={() => setDirPickerIndex(idx)}
                        className="shrink-0 rounded bg-netflix-gray px-3 py-2.5 text-gray-300 transition-colors hover:bg-gray-600 hover:text-white"
                        title="Browse server folders"
                      >
                        <LuFolder size={16} />
                      </button>
                      {formData.paths.length > 1 && (
                        <button
                          type="button"
                          onClick={() => removePath(idx)}
                          className="shrink-0 rounded bg-netflix-gray px-3 py-2.5 text-gray-500 transition-colors hover:bg-red-600/20 hover:text-red-400"
                          title="Remove folder"
                        >
                          <LuX size={16} />
                        </button>
                      )}
                    </div>
                  ))}
                </div>
                <p className="mt-1.5 text-xs text-gray-500">
                  Type a path or click the folder icon to browse. Add multiple folders to the same
                  library.
                </p>
              </div>

              <div className="flex gap-3 pt-2">
                <button
                  type="button"
                  onClick={() => {
                    setShowAddModal(false)
                    setFormData(DEFAULT_FORM)
                  }}
                  className="flex-1 rounded bg-netflix-gray px-4 py-2.5 font-medium text-white transition-colors hover:bg-gray-600"
                >
                  Cancel
                </button>
                <button
                  type="submit"
                  disabled={isCreating}
                  className="flex-1 rounded bg-netflix-red px-4 py-2.5 font-medium text-white transition-colors hover:bg-netflix-red-hover disabled:opacity-50"
                >
                  {isCreating ? 'Creating...' : 'Create Library'}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  )
}
