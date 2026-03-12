import { useState } from 'react'
import { Link } from 'react-router'
import { useLibraries, useCreateLibrary, useMediaList } from '@/hooks/stores/useMedia'
import { MediaCard } from '@/components/MediaCard'
import { useAuthStore } from '@/stores/auth'
import { useUIStore } from '@/stores/ui'

export function LibraryListPage() {
  const { user } = useAuthStore()
  const { data: libraries } = useLibraries()
  const [selectedLibrary, setSelectedLibrary] = useState<number | null>(null)
  const [isCreateModalOpen, setIsCreateModalOpen] = useState(false)

  const selectedLib = libraries?.find((l) => l.id === selectedLibrary)

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <h1 className="text-3xl font-bold text-white">Libraries</h1>
        {user?.is_admin && (
          <button
            onClick={() => setIsCreateModalOpen(true)}
            className="flex items-center gap-2 rounded bg-netflix-red px-4 py-2 font-semibold text-white transition-colors hover:bg-netflix-red-hover"
          >
            <svg className="h-5 w-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                strokeWidth={2}
                d="M12 4v16m8-8H4"
              />
            </svg>
            Add Library
          </button>
        )}
      </div>

      <div className="grid grid-cols-1 gap-6 lg:grid-cols-3">
        {/* Library List */}
        <div className="space-y-2">
          <h2 className="mb-4 text-lg font-semibold text-gray-300">Select a library</h2>
          {libraries?.map((lib) => (
            <button
              key={lib.id}
              onClick={() => setSelectedLibrary(lib.id)}
              className={`w-full rounded-lg p-4 text-left transition-all ${
                selectedLibrary === lib.id
                  ? 'bg-netflix-red/20 ring-1 ring-netflix-red'
                  : 'bg-netflix-dark hover:bg-netflix-gray'
              }`}
            >
              <div className="flex items-start justify-between">
                <div>
                  <h3 className="font-medium text-white">{lib.name}</h3>
                  <p className="text-sm capitalize text-gray-400">{lib.type}</p>
                </div>
                <svg
                  className="h-5 w-5 text-gray-500"
                  fill="none"
                  stroke="currentColor"
                  viewBox="0 0 24 24"
                >
                  <path
                    strokeLinecap="round"
                    strokeLinejoin="round"
                    strokeWidth={2}
                    d="M9 5l7 7-7 7"
                  />
                </svg>
              </div>
              <p className="mt-2 truncate text-xs text-gray-500">{lib.paths?.[0]}</p>
            </button>
          ))}
          {libraries?.length === 0 && (
            <div className="rounded-lg bg-netflix-dark p-6 text-center">
              <p className="text-gray-400">No libraries yet</p>
              {user?.is_admin && (
                <button
                  onClick={() => setIsCreateModalOpen(true)}
                  className="mt-2 text-netflix-red hover:underline"
                >
                  Create your first library
                </button>
              )}
            </div>
          )}
        </div>

        {/* Library Content */}
        <div className="lg:col-span-2">
          {selectedLib ? (
            <LibraryContent libraryId={selectedLib.id} libraryName={selectedLib.name} />
          ) : (
            <div className="flex h-96 flex-col items-center justify-center rounded-lg bg-netflix-dark">
              <svg
                className="mb-4 h-16 w-16 text-gray-600"
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
              <p className="text-gray-400">Select a library to view contents</p>
            </div>
          )}
        </div>
      </div>

      {isCreateModalOpen && <CreateLibraryModal onClose={() => setIsCreateModalOpen(false)} />}
    </div>
  )
}

function LibraryContent({ libraryId, libraryName }: { libraryId: number; libraryName: string }) {
  const { libraryViewMode, setLibraryViewMode } = useUIStore()
  const { data: media, isLoading } = useMediaList({ library_id: libraryId, limit: 50 })

  return (
    <div>
      <div className="mb-4 flex items-center justify-between">
        <h2 className="text-lg font-semibold text-white">{libraryName}</h2>
        <div className="flex gap-2">
          <button
            onClick={() => setLibraryViewMode('grid')}
            className={`flex items-center gap-2 rounded px-3 py-1.5 text-sm transition-colors ${
              libraryViewMode === 'grid'
                ? 'bg-netflix-red text-white'
                : 'bg-netflix-gray text-gray-300 hover:bg-gray-700'
            }`}
          >
            <svg className="h-4 w-4" fill="currentColor" viewBox="0 0 20 20">
              <path d="M5 3a2 2 0 00-2 2v2a2 2 0 002 2h2a2 2 0 002-2V5a2 2 0 00-2-2H5zM5 11a2 2 0 00-2 2v2a2 2 0 002 2h2a2 2 0 002-2v-2a2 2 0 00-2-2H5zM11 5a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2h-2a2 2 0 01-2-2V5zM11 13a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2h-2a2 2 0 01-2-2v-2z" />
            </svg>
            Grid
          </button>
          <button
            onClick={() => setLibraryViewMode('list')}
            className={`flex items-center gap-2 rounded px-3 py-1.5 text-sm transition-colors ${
              libraryViewMode === 'list'
                ? 'bg-netflix-red text-white'
                : 'bg-netflix-gray text-gray-300 hover:bg-gray-700'
            }`}
          >
            <svg className="h-4 w-4" fill="currentColor" viewBox="0 0 20 20">
              <path
                fillRule="evenodd"
                d="M3 4a1 1 0 011-1h12a1 1 0 110 2H4a1 1 0 01-1-1zm0 4a1 1 0 011-1h12a1 1 0 110 2H4a1 1 0 01-1-1zm0 4a1 1 0 011-1h12a1 1 0 110 2H4a1 1 0 01-1-1zm0 4a1 1 0 011-1h12a1 1 0 110 2H4a1 1 0 01-1-1z"
                clipRule="evenodd"
              />
            </svg>
            List
          </button>
        </div>
      </div>

      {isLoading ? (
        <div className="flex h-64 items-center justify-center">
          <div className="h-8 w-8 animate-spin rounded-full border-2 border-netflix-red border-t-transparent" />
        </div>
      ) : media?.length === 0 ? (
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
              d="M7 4v16M17 4v16M3 8h4m10 0h4M3 12h18M3 16h4m10 0h4M4 20h16a1 1 0 001-1V5a1 1 0 00-1-1H4a1 1 0 00-1 1v14a1 1 0 001 1z"
            />
          </svg>
          <p className="text-gray-400">No media found in this library.</p>
        </div>
      ) : libraryViewMode === 'grid' ? (
        <div className="grid grid-cols-2 gap-4 sm:grid-cols-3 lg:grid-cols-4">
          {media?.map((item) => (
            <MediaCard
              key={item.id}
              id={item.id}
              title={item.title}
              posterPath={item.poster_path}
              type={item.media_type === 'episode' ? 'series' : 'movie'}
              year={item.release_date ? new Date(item.release_date).getFullYear() : undefined}
              rating={item.rating}
            />
          ))}
        </div>
      ) : (
        <div className="space-y-2">
          {media?.map((item) => (
            <Link
              key={item.id}
              to={`/media/${item.id}`}
              className="flex items-center gap-4 rounded-lg bg-netflix-dark p-3 transition-colors hover:bg-netflix-gray"
            >
              <div className="h-16 w-12 flex-shrink-0 overflow-hidden rounded bg-netflix-gray">
                {item.poster_path ? (
                  <img
                    src={item.poster_path}
                    alt={item.title}
                    className="h-full w-full object-cover"
                  />
                ) : (
                  <div className="flex h-full w-full items-center justify-center">
                    <svg
                      className="h-6 w-6 text-gray-600"
                      fill="none"
                      stroke="currentColor"
                      viewBox="0 0 24 24"
                    >
                      <path
                        strokeLinecap="round"
                        strokeLinejoin="round"
                        strokeWidth={1.5}
                        d="M7 4v16M17 4v16M3 8h4m10 0h4M3 12h18M3 16h4m10 0h4M4 20h16a1 1 0 001-1V5a1 1 0 00-1-1H4a1 1 0 00-1 1v14a1 1 0 001 1z"
                      />
                    </svg>
                  </div>
                )}
              </div>
              <div className="flex-1 min-w-0">
                <h3 className="truncate font-medium text-white">{item.title}</h3>
                <p className="truncate text-sm text-gray-400">{item.genres.join(', ')}</p>
              </div>
              <svg
                className="h-5 w-5 text-gray-500"
                fill="none"
                stroke="currentColor"
                viewBox="0 0 24 24"
              >
                <path
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  strokeWidth={2}
                  d="M9 5l7 7-7 7"
                />
              </svg>
            </Link>
          ))}
        </div>
      )}
    </div>
  )
}

function CreateLibraryModal({ onClose }: { onClose: () => void }) {
  const [name, setName] = useState('')
  const [path, setPath] = useState('')
  const [type, setType] = useState('movies')
  const [error, setError] = useState('')

  const { mutate: createLibrary, isPending } = useCreateLibrary()

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault()
    setError('')

    if (!name || !path) {
      setError('Name and path are required')
      return
    }

    createLibrary(
      { name, paths: [path], type },
      {
        onSuccess: () => {
          onClose()
        },
        onError: (err: Error) => {
          setError(err.message)
        },
      },
    )
  }

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/70 backdrop-blur-sm">
      <div className="w-full max-w-md rounded-xl bg-netflix-dark p-6 shadow-2xl">
        <div className="mb-6 flex items-center justify-between">
          <h2 className="text-xl font-bold text-white">Add Library</h2>
          <button onClick={onClose} className="text-gray-400 hover:text-white">
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

        {error && (
          <div className="mb-4 rounded-lg bg-netflix-red/20 p-3 text-sm text-netflix-red">
            {error}
          </div>
        )}

        <form onSubmit={handleSubmit} className="space-y-4">
          <div>
            <label className="mb-2 block text-sm font-medium text-gray-400">Library Name</label>
            <input
              type="text"
              value={name}
              onChange={(e) => setName(e.target.value)}
              placeholder="e.g., Movies, TV Shows"
              className="w-full rounded-lg bg-netflix-gray px-4 py-3 text-white placeholder-gray-500 outline-none ring-1 ring-transparent transition-all focus:ring-netflix-red"
              required
            />
          </div>

          <div>
            <label className="mb-2 block text-sm font-medium text-gray-400">Path</label>
            <input
              type="text"
              value={path}
              onChange={(e) => setPath(e.target.value)}
              placeholder="/path/to/media"
              className="w-full rounded-lg bg-netflix-gray px-4 py-3 text-white placeholder-gray-500 outline-none ring-1 ring-transparent transition-all focus:ring-netflix-red"
              required
            />
            <p className="mt-1 text-xs text-gray-500">Absolute path to the media folder</p>
          </div>

          <div>
            <label className="mb-2 block text-sm font-medium text-gray-400">Type</label>
            <select
              value={type}
              onChange={(e) => setType(e.target.value)}
              className="w-full rounded-lg bg-netflix-gray px-4 py-3 text-white outline-none ring-1 ring-transparent transition-all focus:ring-netflix-red"
            >
              <option value="movies">Movies</option>
              <option value="tv">TV Shows</option>
              <option value="mixed">Mixed</option>
            </select>
          </div>

          <div className="flex justify-end gap-3 pt-4">
            <button
              type="button"
              onClick={onClose}
              className="rounded-lg px-4 py-2 text-gray-300 transition-colors hover:bg-netflix-gray hover:text-white"
            >
              Cancel
            </button>
            <button
              type="submit"
              disabled={isPending}
              className="rounded-lg bg-netflix-red px-4 py-2 font-semibold text-white transition-colors hover:bg-netflix-red-hover disabled:opacity-50"
            >
              {isPending ? 'Creating...' : 'Create Library'}
            </button>
          </div>
        </form>
      </div>
    </div>
  )
}
