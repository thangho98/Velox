import { useState } from 'react'
import { useLibraries } from '@/hooks/stores/useMedia'
import { CreateLibraryModal } from '@/components/library/CreateLibraryModal'
import { LibraryContent } from '@/components/library/LibraryContent'
import { useAuthStore } from '@/stores/auth'
import { LuPlus, LuChevronRight, LuLibrary } from 'react-icons/lu'

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
