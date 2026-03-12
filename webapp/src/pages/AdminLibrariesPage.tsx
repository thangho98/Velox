import { useState } from 'react'
import {
  useLibraries,
  useCreateLibrary,
  useDeleteLibrary,
  useScanLibrary,
} from '@/hooks/stores/useMedia'

interface LibraryFormData {
  name: string
  path: string
  type: 'movies' | 'series'
}

export function AdminLibrariesPage() {
  const { data: libraries, isLoading } = useLibraries()
  const { mutate: createLibrary, isPending: isCreating } = useCreateLibrary()
  const { mutate: deleteLibrary } = useDeleteLibrary()
  const { mutate: scanLibrary } = useScanLibrary()

  const [showAddModal, setShowAddModal] = useState(false)
  const [formData, setFormData] = useState<LibraryFormData>({
    name: '',
    path: '',
    type: 'movies',
  })
  const [formError, setFormError] = useState('')
  const [scanningId, setScanningId] = useState<number | null>(null)

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault()
    setFormError('')

    if (!formData.name.trim()) {
      setFormError('Library name is required')
      return
    }
    if (!formData.path.trim()) {
      setFormError('Library path is required')
      return
    }

    createLibrary(
      {
        name: formData.name.trim(),
        path: formData.path.trim(),
        type: formData.type,
      },
      {
        onSuccess: () => {
          setShowAddModal(false)
          setFormData({ name: '', path: '', type: 'movies' })
        },
        onError: (err: Error) => {
          setFormError(err.message || 'Failed to create library')
        },
      },
    )
  }

  const handleDelete = (id: number, name: string) => {
    if (confirm(`Are you sure you want to delete "${name}"? This action cannot be undone.`)) {
      deleteLibrary(id)
    }
  }

  const handleScan = (id: number) => {
    setScanningId(id)
    scanLibrary(id, {
      onSettled: () => setScanningId(null),
    })
  }

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
          <svg className="h-5 w-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 4v16m8-8H4" />
          </svg>
          Add Library
        </button>
      </div>

      {/* Libraries List */}
      {isLoading ? (
        <div className="flex h-64 items-center justify-center">
          <div className="h-8 w-8 animate-spin rounded-full border-2 border-netflix-red border-t-transparent" />
        </div>
      ) : libraries?.length === 0 ? (
        <div className="flex h-64 flex-col items-center justify-center rounded-lg bg-netflix-dark">
          <svg
            className="mb-4 h-12 w-12 text-gray-600"
            fill="none"
            stroke="currentColor"
            viewBox="0 0 24 24"
          >
            <path
              strokeLinecap="round"
              strokeLinejoin="round"
              strokeWidth={1.5}
              d="M19 11H5m14 0a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2v-6a2 2 0 012-2m14 0V9a2 2 0 00-2-2M5 11V9a2 2 0 012-2m0 0V5a2 2 0 012-2h6a2 2 0 012 2v2M7 7h10"
            />
          </svg>
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
          {libraries?.map((lib) => (
            <div
              key={lib.id}
              className="flex items-center justify-between rounded-lg bg-netflix-dark p-4 transition-colors hover:bg-netflix-gray"
            >
              <div className="flex items-center gap-4">
                <div
                  className={`flex h-12 w-12 items-center justify-center rounded-lg ${
                    lib.type === 'movies' ? 'bg-blue-500/20' : 'bg-purple-500/20'
                  }`}
                >
                  <svg
                    className={`h-6 w-6 ${lib.type === 'movies' ? 'text-blue-500' : 'text-purple-500'}`}
                    fill="none"
                    stroke="currentColor"
                    viewBox="0 0 24 24"
                  >
                    <path
                      strokeLinecap="round"
                      strokeLinejoin="round"
                      strokeWidth={2}
                      d="M7 4v16M17 4v16M3 8h4m10 0h4M3 12h18M3 16h4m10 0h4M4 20h16a1 1 0 001-1V5a1 1 0 00-1-1H4a1 1 0 00-1 1v14a1 1 0 001 1z"
                    />
                  </svg>
                </div>
                <div>
                  <h3 className="font-semibold text-white">{lib.name}</h3>
                  <p className="text-sm text-gray-400">{lib.path}</p>
                  <div className="mt-1 flex items-center gap-2">
                    <span
                      className={`rounded px-2 py-0.5 text-xs ${
                        lib.type === 'movies'
                          ? 'bg-blue-500/20 text-blue-400'
                          : 'bg-purple-500/20 text-purple-400'
                      }`}
                    >
                      {lib.type}
                    </span>
                    <span className="text-xs text-gray-500">
                      Created {new Date(lib.created_at).toLocaleDateString()}
                    </span>
                  </div>
                </div>
              </div>

              <div className="flex items-center gap-2">
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
                      <svg
                        className="h-4 w-4"
                        fill="none"
                        stroke="currentColor"
                        viewBox="0 0 24 24"
                      >
                        <path
                          strokeLinecap="round"
                          strokeLinejoin="round"
                          strokeWidth={2}
                          d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15"
                        />
                      </svg>
                      Scan
                    </>
                  )}
                </button>
                <button
                  onClick={() => handleDelete(lib.id, lib.name)}
                  className="rounded bg-netflix-gray px-3 py-1.5 text-sm text-white transition-colors hover:bg-red-600"
                >
                  <svg className="h-4 w-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path
                      strokeLinecap="round"
                      strokeLinejoin="round"
                      strokeWidth={2}
                      d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16"
                    />
                  </svg>
                </button>
              </div>
            </div>
          ))}
        </div>
      )}

      {/* Add Library Modal */}
      {showAddModal && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/70 p-4">
          <div className="w-full max-w-md rounded-lg bg-netflix-dark p-6">
            <div className="mb-6 flex items-center justify-between">
              <h2 className="text-xl font-bold text-white">Add New Library</h2>
              <button
                onClick={() => setShowAddModal(false)}
                className="text-gray-400 hover:text-white"
              >
                <svg className="h-6 w-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path
                    strokeLinecap="round"
                    strokeLinejoin="round"
                    strokeWidth={2}
                    d="M6 18L18 6M6 6l12 12"
                  />
                </svg>
              </button>
            </div>

            <form onSubmit={handleSubmit} className="space-y-4">
              {formError && (
                <div className="rounded-lg bg-netflix-red/20 p-3 text-sm text-netflix-red">
                  {formError}
                </div>
              )}

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

              <div>
                <label className="mb-2 block text-sm font-medium text-gray-400">Folder Path</label>
                <input
                  type="text"
                  value={formData.path}
                  onChange={(e) => setFormData({ ...formData, path: e.target.value })}
                  placeholder="e.g., /media/movies"
                  className="w-full rounded bg-netflix-gray px-4 py-3 text-white outline-none ring-1 ring-transparent transition-all focus:ring-netflix-red"
                  required
                />
                <p className="mt-1 text-xs text-gray-500">
                  Absolute path to the folder containing your media files
                </p>
              </div>

              <div>
                <label className="mb-2 block text-sm font-medium text-gray-400">Library Type</label>
                <div className="flex gap-4">
                  <button
                    type="button"
                    onClick={() => setFormData({ ...formData, type: 'movies' })}
                    className={`flex flex-1 items-center justify-center gap-2 rounded-lg border-2 px-4 py-3 transition-colors ${
                      formData.type === 'movies'
                        ? 'border-blue-500 bg-blue-500/20 text-blue-400'
                        : 'border-netflix-gray bg-netflix-gray text-gray-400'
                    }`}
                  >
                    <svg className="h-5 w-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path
                        strokeLinecap="round"
                        strokeLinejoin="round"
                        strokeWidth={2}
                        d="M7 4v16M17 4v16M3 8h4m10 0h4M3 12h18M3 16h4m10 0h4M4 20h16a1 1 0 001-1V5a1 1 0 00-1-1H4a1 1 0 00-1 1v14a1 1 0 001 1z"
                      />
                    </svg>
                    Movies
                  </button>
                  <button
                    type="button"
                    onClick={() => setFormData({ ...formData, type: 'series' })}
                    className={`flex flex-1 items-center justify-center gap-2 rounded-lg border-2 px-4 py-3 transition-colors ${
                      formData.type === 'series'
                        ? 'border-purple-500 bg-purple-500/20 text-purple-400'
                        : 'border-netflix-gray bg-netflix-gray text-gray-400'
                    }`}
                  >
                    <svg className="h-5 w-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path
                        strokeLinecap="round"
                        strokeLinejoin="round"
                        strokeWidth={2}
                        d="M9.75 17L9 20l-1 1h8l-1-1-.75-3M3 13h18M5 17h14a2 2 0 002-2V5a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z"
                      />
                    </svg>
                    Series
                  </button>
                </div>
              </div>

              <div className="flex gap-3 pt-4">
                <button
                  type="button"
                  onClick={() => setShowAddModal(false)}
                  className="flex-1 rounded bg-netflix-gray px-4 py-2 font-medium text-white transition-colors hover:bg-gray-600"
                >
                  Cancel
                </button>
                <button
                  type="submit"
                  disabled={isCreating}
                  className="flex-1 rounded bg-netflix-red px-4 py-2 font-medium text-white transition-colors hover:bg-netflix-red-hover disabled:opacity-50"
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
