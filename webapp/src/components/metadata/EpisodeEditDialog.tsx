import { useState } from 'react'
import { LuX, LuSave } from 'react-icons/lu'
import type { Episode, EpisodeMetadataEditRequest } from '@/types/api'

interface EpisodeEditDialogProps {
  episode: Episode
  onSave: (req: EpisodeMetadataEditRequest) => void
  isSaving: boolean
  onClose: () => void
}

export function EpisodeEditDialog({ episode, onSave, isSaving, onClose }: EpisodeEditDialogProps) {
  const [title, setTitle] = useState(episode.title)
  const [overview, setOverview] = useState(episode.overview)
  const [airDate, setAirDate] = useState(episode.air_date ?? '')
  const [episodeNumber, setEpisodeNumber] = useState(episode.episode_number)

  function handleSave() {
    onSave({
      title,
      overview,
      air_date: airDate || undefined,
      episode_number: episodeNumber !== episode.episode_number ? episodeNumber : undefined,
      metadata_locked: true,
    })
  }

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center">
      <div className="absolute inset-0 bg-black/60" onClick={onClose} />
      <div className="relative w-full max-w-lg rounded-xl bg-[#1a1a1a] shadow-2xl">
        {/* Header */}
        <div className="flex items-center justify-between border-b border-gray-700 px-6 py-4">
          <h2 className="text-lg font-semibold text-white">
            Edit Episode {episode.episode_number}
          </h2>
          <button onClick={onClose} className="text-gray-400 hover:text-white">
            <LuX size={20} />
          </button>
        </div>

        {/* Body */}
        <div className="space-y-4 p-6">
          <div className="grid grid-cols-[1fr_80px] gap-3">
            <div>
              <label className="mb-1 block text-sm text-gray-300">Title</label>
              <input
                type="text"
                value={title}
                onChange={(e) => setTitle(e.target.value)}
                className="w-full rounded-lg bg-[#2a2a2a] px-3 py-2 text-white outline-none focus:ring-1 focus:ring-blue-500"
              />
            </div>
            <div>
              <label className="mb-1 block text-sm text-gray-300">Ep #</label>
              <input
                type="number"
                value={episodeNumber}
                onChange={(e) => setEpisodeNumber(Number(e.target.value))}
                min={1}
                className="w-full rounded-lg bg-[#2a2a2a] px-3 py-2 text-white outline-none focus:ring-1 focus:ring-blue-500"
              />
            </div>
          </div>

          <div>
            <label className="mb-1 block text-sm text-gray-300">Overview</label>
            <textarea
              value={overview}
              onChange={(e) => setOverview(e.target.value)}
              rows={4}
              className="w-full rounded-lg bg-[#2a2a2a] px-3 py-2 text-white outline-none focus:ring-1 focus:ring-blue-500"
            />
          </div>

          <div>
            <label className="mb-1 block text-sm text-gray-300">Air Date</label>
            <input
              type="date"
              value={airDate}
              onChange={(e) => setAirDate(e.target.value)}
              className="w-full rounded-lg bg-[#2a2a2a] px-3 py-2 text-white outline-none focus:ring-1 focus:ring-blue-500"
            />
          </div>
        </div>

        {/* Footer */}
        <div className="flex items-center justify-end gap-3 border-t border-gray-700 px-6 py-4">
          <button
            onClick={onClose}
            className="rounded-lg px-4 py-2 text-sm text-gray-400 hover:text-white"
          >
            Cancel
          </button>
          <button
            onClick={handleSave}
            disabled={isSaving || !title.trim()}
            className="flex items-center gap-2 rounded-lg bg-blue-600 px-5 py-2 text-sm font-medium text-white hover:bg-blue-500 disabled:opacity-50"
          >
            <LuSave size={16} />
            {isSaving ? 'Saving...' : 'Save'}
          </button>
        </div>
      </div>
    </div>
  )
}
