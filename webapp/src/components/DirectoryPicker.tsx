import { useState } from 'react'
import { LuX, LuFolder, LuFolderOpen, LuChevronLeft, LuChevronRight } from 'react-icons/lu'
import { useFsBrowse } from '@/hooks/stores/useMedia'

interface DirectoryPickerProps {
  onSelect: (path: string) => void
  onClose: () => void
}

export function DirectoryPicker({ onSelect, onClose }: DirectoryPickerProps) {
  const [currentPath, setCurrentPath] = useState('/')
  const { data, isLoading, isError } = useFsBrowse(currentPath)

  const handleSelect = () => {
    onSelect(currentPath)
    onClose()
  }

  return (
    <div className="fixed inset-0 z-[60] flex items-center justify-center bg-black/80 p-4">
      <div className="flex w-full max-w-lg flex-col rounded-lg bg-[#1a1a1a] shadow-2xl">
        {/* Header */}
        <div className="flex items-center justify-between border-b border-white/10 px-5 py-4">
          <h3 className="text-lg font-semibold text-white">Select Folder</h3>
          <button onClick={onClose} className="text-gray-400 hover:text-white transition-colors">
            <LuX size={20} />
          </button>
        </div>

        {/* Current path */}
        <div className="flex items-center gap-2 border-b border-white/10 px-5 py-3">
          <LuFolderOpen size={16} className="shrink-0 text-gray-400" />
          <span className="truncate font-mono text-sm text-white">{currentPath}</span>
        </div>

        {/* Directory listing */}
        <div className="h-72 overflow-y-auto">
          {isLoading && (
            <div className="flex h-full items-center justify-center">
              <div className="h-6 w-6 animate-spin rounded-full border-2 border-netflix-red border-t-transparent" />
            </div>
          )}
          {isError && (
            <div className="flex h-full items-center justify-center px-6 text-center text-sm text-red-400">
              Cannot read this directory. Check permissions.
            </div>
          )}
          {data && (
            <ul className="py-1">
              {data.parent !== undefined && data.parent !== '' && (
                <li>
                  <button
                    onClick={() => setCurrentPath(data.parent!)}
                    className="flex w-full items-center gap-3 px-5 py-2.5 text-left text-sm text-gray-300 transition-colors hover:bg-white/5"
                  >
                    <LuChevronLeft size={16} className="shrink-0 text-gray-500" />
                    <span className="text-gray-400">.. (up one level)</span>
                  </button>
                </li>
              )}
              {data.dirs.length === 0 && (
                <li className="px-5 py-4 text-sm text-gray-500">No subdirectories</li>
              )}
              {data.dirs.map((dir) => (
                <li key={dir.path}>
                  <button
                    onClick={() => setCurrentPath(dir.path)}
                    className="flex w-full items-center gap-3 px-5 py-2.5 text-left text-sm transition-colors hover:bg-white/5"
                  >
                    <LuFolder size={16} className="shrink-0 text-yellow-500/80" />
                    <span className="truncate text-white">{dir.name}</span>
                    <LuChevronRight size={16} className="ml-auto shrink-0 text-gray-600" />
                  </button>
                </li>
              ))}
            </ul>
          )}
        </div>

        {/* Footer */}
        <div className="flex items-center justify-between border-t border-white/10 px-5 py-4">
          <span className="truncate font-mono text-xs text-gray-500">{currentPath}</span>
          <div className="flex gap-3">
            <button
              onClick={onClose}
              className="rounded px-4 py-2 text-sm text-gray-400 transition-colors hover:text-white"
            >
              Cancel
            </button>
            <button
              onClick={handleSelect}
              className="rounded bg-netflix-red px-4 py-2 text-sm font-medium text-white transition-colors hover:bg-netflix-red-hover"
            >
              Select This Folder
            </button>
          </div>
        </div>
      </div>
    </div>
  )
}
