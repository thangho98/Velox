import { useState, type FormEvent } from 'react'
import { LuX } from 'react-icons/lu'
import { useCreateLibrary } from '@/hooks/stores/useMedia'

interface CreateLibraryModalProps {
  onClose: () => void
}

export function CreateLibraryModal({ onClose }: CreateLibraryModalProps) {
  const [name, setName] = useState('')
  const [path, setPath] = useState('')
  const [type, setType] = useState('movies')
  const [error, setError] = useState('')

  const { mutate: createLibrary, isPending } = useCreateLibrary()

  const handleSubmit = (e: FormEvent) => {
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
