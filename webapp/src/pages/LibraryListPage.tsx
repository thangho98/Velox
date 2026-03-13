import { useState } from 'react'
import { Link } from 'react-router'
import { useLibraries, useCreateLibrary, useMediaList } from '@/hooks/stores/useMedia'
import { MediaCard } from '@/components/MediaCard'
import { useAuthStore } from '@/stores/auth'
import { useUIStore } from '@/stores/ui'
import {
  LuPlus,
  LuChevronRight,
  LuLibrary,
  LuFilm,
  LuLayoutGrid,
  LuList,
  LuX,
} from 'react-icons/lu'

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
            <LuPlus size={20} />
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
                <LuChevronRight size={20} className="text-gray-500" />
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
              <LuLibrary size={64} className="mb-4 text-gray-600" />
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
            <LuLayoutGrid size={16} />
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
            <LuList size={16} />
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
          <LuFilm size={48} className="mb-4 text-gray-600" />
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
                    <LuFilm size={24} className="text-gray-600" />
                  </div>
                )}
              </div>
              <div className="flex-1 min-w-0">
                <h3 className="truncate font-medium text-white">{item.title}</h3>
                <p className="truncate text-sm text-gray-400">{item.genres.join(', ')}</p>
              </div>
              <LuChevronRight size={20} className="text-gray-500" />
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
            <LuX size={24} />
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
