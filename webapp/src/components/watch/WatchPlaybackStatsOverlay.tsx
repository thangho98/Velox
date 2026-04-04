import { memo, useEffect, useState, type RefObject } from 'react'
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
    PreTranscode: 'bg-cyan-500/20 text-cyan-400',
    TranscodeAudio: 'bg-yellow-500/20 text-yellow-400',
    FullTranscode: 'bg-red-500/20 text-red-400',
  }
  const labels: Record<string, string> = {
    DirectPlay: 'Direct Play',
    DirectStream: 'Direct Stream',
    PreTranscode: 'PreTranscode',
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
  const isPreTranscode = playbackInfo.method === 'PreTranscode'

  // Use pretranscode details when available, otherwise original file info
  const videoCodec =
    isPreTranscode && playbackInfo.pt_video_codec
      ? playbackInfo.pt_video_codec
      : playbackInfo.video_codec
  const videoHeight =
    isPreTranscode && playbackInfo.pt_height ? playbackInfo.pt_height : playbackInfo.height
  const videoWidth =
    isPreTranscode && playbackInfo.pt_height
      ? Math.round((playbackInfo.width / playbackInfo.height) * playbackInfo.pt_height)
      : playbackInfo.width
  const videoBitrate =
    isPreTranscode && playbackInfo.pt_video_bitrate
      ? playbackInfo.pt_video_bitrate
      : playbackInfo.bitrate

  // Read dropped frames from videoRef via effect to avoid accessing refs during render
  const [droppedFrames, setDroppedFrames] = useState(0)
  useEffect(() => {
    const update = () => {
      const quality = videoRef.current?.getVideoPlaybackQuality?.()
      setDroppedFrames(quality?.droppedVideoFrames ?? 0)
    }
    update()
    const id = setInterval(update, 1000)
    return () => clearInterval(id)
  }, [videoRef])

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
            ) : isPreTranscode ? (
              <span className="rounded bg-cyan-500/20 px-1.5 py-0.5 text-[10px] font-semibold text-cyan-400">
                PreTranscode
              </span>
            ) : (
              <span className="rounded bg-green-500/20 px-1.5 py-0.5 text-[10px] font-semibold text-green-400">
                Direct
              </span>
            )}
          </div>
          <p className="font-mono text-xs leading-relaxed text-white/80">
            {videoCodec?.toUpperCase() || '—'} {videoHeight > 0 && `${videoWidth}×${videoHeight}`}
            {!isPreTranscode && playbackInfo.video_profile && ` ${playbackInfo.video_profile}`}
            {!isPreTranscode && playbackInfo.video_level > 0 && ` L${playbackInfo.video_level}`}
          </p>
          <p className="font-mono text-xs leading-relaxed text-white/60">
            {videoBitrate > 0 &&
              `${videoBitrate >= 1000 ? `${(videoBitrate / 1000).toFixed(1)} Mbps` : `${videoBitrate} Kbps`}`}
            {playbackInfo.video_fps > 0 &&
              ` · ${Number.isInteger(playbackInfo.video_fps) ? playbackInfo.video_fps : playbackInfo.video_fps.toFixed(2)} fps`}
          </p>
          <p className="mt-1 font-mono text-xs text-white/60">
            Dropped:{' '}
            <span className={droppedFrames > 0 ? 'text-yellow-400' : 'text-white/60'}>
              {droppedFrames}
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
              ) : isPreTranscode ? (
                <span className="rounded bg-cyan-500/20 px-1.5 py-0.5 text-[10px] font-semibold text-cyan-400">
                  PreTranscode
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
            {playbackInfo.method === 'DirectPlay'
              ? 'HTTP Range'
              : isPreTranscode
                ? 'HTTP Range (PreTranscode)'
                : 'HLS'}
            {' · '}
            {isPreTranscode ? 'MP4' : playbackInfo.container?.toUpperCase() || '—'}
            {!isPreTranscode &&
              playbackInfo.file_size > 0 &&
              ` · ${(playbackInfo.file_size / (1024 * 1024 * 1024)).toFixed(1)} GB`}
          </p>
          {playbackInfo.estimated_bitrate > 0 && (isTranscoding || isPreTranscode) && (
            <p className="font-mono text-xs leading-relaxed text-white/60">
              Estimated:{' '}
              {playbackInfo.estimated_bitrate >= 1000
                ? `${(playbackInfo.estimated_bitrate / 1000).toFixed(1)} Mbps`
                : `${playbackInfo.estimated_bitrate} Kbps`}
            </p>
          )}
        </div>
      </div>
    </div>
  )
})
