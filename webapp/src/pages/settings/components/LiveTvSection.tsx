import { useState } from 'react'
import { useTranslation } from '@/hooks/useTranslation'
import { useLiveTVPlaylists } from '@/hooks/stores/useLiveTV'
import type { LivePlaylist } from '@/hooks/stores/useLiveTV'
import { useQueryClient, useMutation } from '@tanstack/react-query'
import { api } from '@velox/shared/api'
import { LuTv, LuPlus, LuRefreshCw, LuTrash2 } from 'react-icons/lu'

export function LiveTvSection() {
  const { t } = useTranslation('settings')
  const { data: playlists = [], isLoading } = useLiveTVPlaylists()
  const queryClient = useQueryClient()

  const [newName, setNewName] = useState('')
  const [sourceType, setSourceType] = useState<'url' | 'file'>('url')
  const [newUrl, setNewUrl] = useState('')
  const [newFile, setNewFile] = useState<File | null>(null)

  const createMutation = useMutation({
    mutationFn: async (
      payload: FormData | { name: string; url: string; sync_interval_hours: number },
    ) => {
      if (payload instanceof FormData) {
        return api.uploadFormData('/livetv/playlists', payload)
      }
      return api.post('/livetv/playlists', payload)
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['livetv', 'playlists'] })
      setNewName('')
      setNewUrl('')
      setNewFile(null)
    },
  })

  const syncMutation = useMutation({
    mutationFn: async (id: number) => {
      return api.post(`/livetv/playlists/${id}/sync`, {})
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['livetv', 'playlists'] })
    },
  })

  const deleteMutation = useMutation({
    mutationFn: async (id: number) => {
      return api.delete(`/livetv/playlists/${id}`)
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['livetv', 'playlists'] })
    },
  })

  const handleAdd = (e: React.FormEvent) => {
    e.preventDefault()
    if (!newName.trim()) return

    if (sourceType === 'url') {
      if (!newUrl.trim()) return
      createMutation.mutate({
        name: newName,
        url: newUrl,
        sync_interval_hours: 24,
      })
    } else {
      if (!newFile) return
      const formData = new FormData()
      formData.append('name', newName)
      formData.append('sync_interval_hours', '24')
      formData.append('file', newFile)
      createMutation.mutate(formData)
    }
  }

  return (
    <div className="space-y-8 max-w-4xl">
      <div>
        <h2 className="text-xl font-bold tracking-tight text-white flex items-center gap-2">
          <LuTv className="text-netflix-red" />
          Live TV Playlists
        </h2>
        <p className="mt-2 text-sm text-gray-400">
          Manage your external M3U IPTV playlists. Channels will be automatically parsed, grouped,
          and updated daily.
        </p>
      </div>

      {/* Add New Playlist */}
      <div className="rounded-lg border border-white/10 bg-netflix-gray/20 p-5">
        <h3 className="text-sm font-semibold text-white mb-4">Add M3U Playlist</h3>

        <div className="mb-4 flex gap-4">
          <label className="flex items-center gap-2 text-sm text-gray-300">
            <input
              type="radio"
              checked={sourceType === 'url'}
              onChange={() => setSourceType('url')}
              className="accent-netflix-red"
            />
            URL Source
          </label>
          <label className="flex items-center gap-2 text-sm text-gray-300">
            <input
              type="radio"
              checked={sourceType === 'file'}
              onChange={() => setSourceType('file')}
              className="accent-netflix-red"
            />
            Local M3U File
          </label>
        </div>

        <form onSubmit={handleAdd} className="flex flex-col sm:flex-row gap-3">
          <input
            type="text"
            placeholder="Name (e.g., iptv-org)"
            value={newName}
            onChange={(e) => setNewName(e.target.value)}
            className="flex-1 rounded bg-black/40 border border-white/10 px-3 py-2 text-sm text-white placeholder:text-gray-600 focus:border-netflix-red focus:outline-none transition-colors"
          />
          {sourceType === 'url' ? (
            <input
              type="url"
              placeholder="M3U URL (https://...)"
              value={newUrl}
              onChange={(e) => setNewUrl(e.target.value)}
              className="flex-[2] rounded bg-black/40 border border-white/10 px-3 py-2 text-sm text-white placeholder:text-gray-600 focus:border-netflix-red focus:outline-none transition-colors"
            />
          ) : (
            <input
              type="file"
              accept=".m3u,.m3u8"
              onChange={(e) => setNewFile(e.target.files?.[0] || null)}
              className="flex-[2] rounded bg-black/40 border border-white/10 px-3 py-1.5 text-sm text-white file:mr-4 file:rounded file:border-0 file:bg-white/10 file:px-4 file:py-1 file:text-sm file:font-semibold file:text-white hover:file:bg-white/20 transition-colors"
            />
          )}
          <button
            type="submit"
            disabled={
              createMutation.isPending || !newName || (sourceType === 'url' ? !newUrl : !newFile)
            }
            className="flex items-center justify-center gap-2 bg-netflix-red text-white px-5 py-2 rounded text-sm font-semibold hover:bg-red-700 disabled:opacity-50 transition-colors"
          >
            <LuPlus size={16} />
            {createMutation.isPending ? 'Adding...' : 'Add Playlist'}
          </button>
        </form>
      </div>

      {/* Playlists Table */}
      <div className="rounded-lg border border-white/10 overflow-hidden bg-black/20">
        <table className="w-full text-left text-sm text-gray-400">
          <thead className="bg-netflix-gray/40 text-gray-300">
            <tr>
              <th className="px-5 py-3 font-medium">Name</th>
              <th className="px-5 py-3 font-medium">Source URL</th>
              <th className="px-5 py-3 font-medium">Last Updated</th>
              <th className="px-5 py-3 font-medium text-right">Actions</th>
            </tr>
          </thead>
          <tbody className="divide-y divide-white/5">
            {isLoading ? (
              <tr>
                <td colSpan={4} className="px-5 py-8 text-center text-gray-500">
                  Loading playlists...
                </td>
              </tr>
            ) : playlists.length === 0 ? (
              <tr>
                <td colSpan={4} className="px-5 py-8 text-center text-gray-500">
                  No playlists added yet.
                </td>
              </tr>
            ) : (
              playlists.map((playlist: LivePlaylist) => (
                <tr key={playlist.id} className="hover:bg-netflix-gray/10 transition-colors">
                  <td className="px-5 py-4 font-medium text-white break-words">
                    {playlist.name}
                    {!playlist.isActive && (
                      <span className="ml-2 text-[10px] bg-gray-700 text-gray-300 px-1.5 py-0.5 rounded">
                        Disabled
                      </span>
                    )}
                  </td>
                  <td className="px-5 py-4 truncate max-w-[200px]" title={playlist.url}>
                    {playlist.url}
                  </td>
                  <td className="px-5 py-4 text-xs">
                    {playlist.lastSyncedAt
                      ? new Date(playlist.lastSyncedAt).toLocaleString()
                      : new Date(playlist.updatedAt).toLocaleString()}
                  </td>
                  <td className="px-5 py-4 text-right space-x-3">
                    <button
                      onClick={() => syncMutation.mutate(playlist.id)}
                      disabled={syncMutation.isPending && syncMutation.variables === playlist.id}
                      className="text-gray-400 hover:text-white transition-colors disabled:opacity-50"
                      title="Sync Channels Now"
                    >
                      <LuRefreshCw
                        size={16}
                        className={
                          syncMutation.isPending && syncMutation.variables === playlist.id
                            ? 'animate-spin'
                            : ''
                        }
                      />
                    </button>
                    <button
                      onClick={() => {
                        if (window.confirm('Delete this playlist and hide all its channels?')) {
                          deleteMutation.mutate(playlist.id)
                        }
                      }}
                      className="text-gray-400 hover:text-netflix-red transition-colors disabled:opacity-50"
                      title="Delete Playlist"
                    >
                      <LuTrash2 size={16} />
                    </button>
                  </td>
                </tr>
              ))
            )}
          </tbody>
        </table>
      </div>
    </div>
  )
}
