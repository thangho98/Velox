import type { PlaybackAudioTrack } from '@/types/api'

interface AudioPickerProps {
  tracks: PlaybackAudioTrack[]
  selectedLanguage: string | null
  onSelect: (language: string, trackId: number) => void
}

export function AudioPicker({ tracks, selectedLanguage, onSelect }: AudioPickerProps) {
  if (tracks.length === 0) return null

  return (
    <div className="min-w-[180px] rounded-lg bg-black/90 p-2 shadow-xl">
      <p className="mb-1 px-2 text-xs font-semibold text-gray-400">Audio Track</p>
      {tracks.map((track) => (
        <button
          key={track.id}
          onClick={() => onSelect(track.language, track.id)}
          className={`w-full rounded px-3 py-2 text-left text-sm ${
            selectedLanguage === track.language
              ? 'bg-netflix-red text-white'
              : 'text-white hover:bg-white/10'
          }`}
        >
          <span>{track.label}</span>
          {track.codec && (
            <span className="ml-2 text-xs opacity-50">
              {track.codec.toUpperCase()}
              {track.channels > 0 && ` ${track.channels}ch`}
            </span>
          )}
          {!track.is_default && selectedLanguage !== track.language && (
            <span className="ml-1 text-xs text-yellow-400 opacity-70">(HLS)</span>
          )}
        </button>
      ))}
    </div>
  )
}
