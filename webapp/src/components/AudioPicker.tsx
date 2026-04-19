import type { PlaybackAudioTrack } from '@/types/api'

interface AudioPickerProps {
  tracks: PlaybackAudioTrack[]
  selectedTrackId: number | null
  onSelect: (language: string, trackId: number) => void
}

export function AudioPicker({ tracks, selectedTrackId, onSelect }: AudioPickerProps) {
  if (tracks.length === 0) return null

  return (
    <div className="w-full rounded-lg bg-black/90 p-2 shadow-xl sm:w-auto sm:min-w-[180px]">
      <p className="mb-1 px-2 text-xs font-semibold text-gray-400">Audio Track</p>
      <div className="max-h-[300px] overflow-y-auto w-full max-w-[280px] sm:max-w-[400px]">
        {tracks.map((track) => (
          <button
            key={track.id}
            title={track.label}
            onClick={() => onSelect(track.language, track.id)}
            className={`w-full flex items-center rounded px-3 py-2 text-left text-sm transition-colors ${
              selectedTrackId === track.id
                ? 'bg-netflix-red text-white'
                : 'text-white hover:bg-white/10'
            }`}
          >
            <span className="truncate flex-1" title={track.label}>
              {track.label}
            </span>
            {track.codec && (
              <span className="ml-2 shrink-0 text-xs opacity-50">
                {track.codec.toUpperCase()}
                {track.channels > 0 && ` ${track.channels}ch`}
              </span>
            )}
            {!track.is_default && selectedTrackId !== track.id && (
              <span className="ml-1 shrink-0 text-xs text-yellow-400 opacity-70">(HLS)</span>
            )}
          </button>
        ))}
      </div>
    </div>
  )
}
