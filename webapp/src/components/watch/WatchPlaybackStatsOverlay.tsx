import { memo, type RefObject } from 'react'
import { LuActivity, LuX } from 'react-icons/lu'
import type { PlaybackInfo } from '@/types/api'
import { formatChannelLayout } from './watchHelpers'

interface WatchPlaybackStatsOverlayProps {
  onClose: () => void
  playbackInfo: PlaybackInfo
  videoRef: RefObject<HTMLVideoElement | null>
}

function MethodBadge({ method }: { method: string }) {
  const colors: Record<string, string> = {
    DirectPlay: 'bg-green-500/20 text-green-400',
    DirectStream: 'bg-blue-500/20 text-blue-400',
    TranscodeAudio: 'bg-yellow-500/20 text-yellow-400',
    FullTranscode: 'bg-red-500/20 text-red-400',
  }
  const labels: Record<string, string> = {
    DirectPlay: 'Direct Play',
    DirectStream: 'Direct Stream',
    TranscodeAudio: 'Transcode Audio',
    FullTranscode: 'Full Transcode',
  }
  return (
    <span
      className={`rounded px-1.5 py-0.5 text-[10px] font-semibold ${colors[method] ?? 'bg-white/10 text-white/60'}`}
    >
      {labels[method] ?? method}
    </span>
  )
}

export const WatchPlaybackStatsOverlay = memo(function WatchPlaybackStatsOverlay({
  onClose,
  playbackInfo,
  videoRef,
}: WatchPlaybackStatsOverlayProps) {
  const selectedAudio =
    playbackInfo.audio_tracks?.find((t) => t.selected) ??
    playbackInfo.audio_tracks?.find((t) => t.is_default) ??
    playbackInfo.audio_tracks?.[0]

  const isTranscoding =
    playbackInfo.method === 'FullTranscode' || playbackInfo.method === 'TranscodeAudio'

  return (
    <div className="absolute left-4 top-20 z-30 w-80 overflow-hidden rounded-xl bg-black/70 backdrop-blur-md ring-1 ring-white/10">
      <button
        onClick={onClose}
        className="absolute right-2.5 top-2.5 rounded-lg p-1 text-white/40 hover:bg-white/10 hover:text-white"
      >
        <LuX size={16} />
      </button>

      <div className="space-y-0">
        {/* Playback Method */}
        <div className="border-b border-white/10 px-4 py-3">
          <div className="mb-2 flex items-center gap-2">
            <p className="text-sm font-bold text-white">Playback</p>
            <MethodBadge method={playbackInfo.method} />
          </div>
          <p className="font-mono text-xs leading-relaxed text-white/60">
            {playbackInfo.decision_reason}
          </p>
        </div>

        {/* Video */}
        <div className="border-b border-white/10 px-4 py-3">
          <div className="mb-2 flex items-center gap-2">
            <p className="text-sm font-bold text-white">Video</p>
            {playbackInfo.method === 'FullTranscode' ? (
              <span className="rounded bg-red-500/20 px-1.5 py-0.5 text-[10px] font-semibold text-red-400">
                Transcoding
              </span>
            ) : (
              <span className="rounded bg-green-500/20 px-1.5 py-0.5 text-[10px] font-semibold text-green-400">
                Direct
              </span>
            )}
          </div>
          <p className="font-mono text-xs leading-relaxed text-white/80">
            {playbackInfo.video_codec?.toUpperCase() || '—'}{' '}
            {playbackInfo.height > 0 && `${playbackInfo.width}×${playbackInfo.height}`}
            {playbackInfo.video_profile && ` ${playbackInfo.video_profile}`}
            {playbackInfo.video_level > 0 && ` L${playbackInfo.video_level}`}
          </p>
          <p className="font-mono text-xs leading-relaxed text-white/60">
            {playbackInfo.bitrate > 0 &&
              `${playbackInfo.bitrate >= 1000 ? `${(playbackInfo.bitrate / 1000).toFixed(1)} Mbps` : `${playbackInfo.bitrate} Kbps`}`}
            {playbackInfo.video_fps > 0 &&
              ` · ${Number.isInteger(playbackInfo.video_fps) ? playbackInfo.video_fps : playbackInfo.video_fps.toFixed(2)} fps`}
          </p>
          <p className="mt-1 font-mono text-xs text-white/60">
            Dropped:{' '}
            <span
              className={
                (videoRef.current?.getVideoPlaybackQuality?.()?.droppedVideoFrames ?? 0) > 0
                  ? 'text-yellow-400'
                  : 'text-white/60'
              }
            >
              {videoRef.current?.getVideoPlaybackQuality?.()?.droppedVideoFrames ?? 0}
            </span>
          </p>
        </div>

        {/* Audio */}
        {selectedAudio && (
          <div className="border-b border-white/10 px-4 py-3">
            <div className="mb-2 flex items-center gap-2">
              <p className="text-sm font-bold text-white">Audio</p>
              {isTranscoding ? (
                <span className="rounded bg-yellow-500/20 px-1.5 py-0.5 text-[10px] font-semibold text-yellow-400">
                  Transcoding
                </span>
              ) : (
                <span className="rounded bg-green-500/20 px-1.5 py-0.5 text-[10px] font-semibold text-green-400">
                  Direct
                </span>
              )}
            </div>
            <p className="font-mono text-xs leading-relaxed text-white/80">
              {selectedAudio.codec?.toUpperCase() || '—'}{' '}
              {formatChannelLayout(selectedAudio.channels)}
              {selectedAudio.language && ` · ${selectedAudio.language}`}
              {selectedAudio.is_default && ' (Default)'}
            </p>
            <p className="font-mono text-xs leading-relaxed text-white/60">
              {selectedAudio.bitrate > 0 &&
                `${selectedAudio.bitrate >= 1000 ? `${Math.round(selectedAudio.bitrate / 1000)} Kbps` : `${selectedAudio.bitrate} bps`}`}
              {selectedAudio.sample_rate > 0 && ` · ${selectedAudio.sample_rate} Hz`}
            </p>
          </div>
        )}

        {/* Stream Info */}
        <div className="px-4 py-3">
          <p className="mb-2 text-sm font-bold text-white">Stream</p>
          <p className="font-mono text-xs leading-relaxed text-white/80">
            {playbackInfo.method === 'DirectPlay' ? 'HTTP Range' : 'HLS'}
            {' · '}
            {playbackInfo.container?.toUpperCase() || '—'}
            {playbackInfo.file_size > 0 &&
              ` · ${(playbackInfo.file_size / (1024 * 1024 * 1024)).toFixed(1)} GB`}
          </p>
          {playbackInfo.estimated_bitrate > 0 && isTranscoding && (
            <p className="font-mono text-xs leading-relaxed text-white/60">
              Estimated:{' '}
              {playbackInfo.estimated_bitrate >= 1000
                ? `${(playbackInfo.estimated_bitrate / 1000).toFixed(1)} Mbps`
                : `${playbackInfo.estimated_bitrate} Kbps`}
            </p>
          )}
        </div>

        <div className="border-t border-white/10 px-4 py-2.5">
          <button
            onClick={onClose}
            className="flex items-center gap-1.5 text-xs text-white/60 transition-colors hover:text-white"
          >
            <LuActivity size={13} />
            Close
          </button>
        </div>
      </div>
    </div>
  )
})
