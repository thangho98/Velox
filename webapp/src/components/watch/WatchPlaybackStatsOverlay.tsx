import { memo, type RefObject } from 'react'
import { LuActivity, LuX } from 'react-icons/lu'
import type { PlaybackInfo } from '@/types/api'
import { formatChannelLayout } from './watchHelpers'

interface WatchPlaybackStatsOverlayProps {
  onClose: () => void
  playbackInfo: PlaybackInfo
  videoRef: RefObject<HTMLVideoElement | null>
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

  return (
    <div className="absolute left-4 top-20 z-30 w-80 overflow-hidden rounded-xl bg-black/70 backdrop-blur-md ring-1 ring-white/10">
      <button
        onClick={onClose}
        className="absolute right-2.5 top-2.5 rounded-lg p-1 text-white/40 hover:bg-white/10 hover:text-white"
      >
        <LuX size={16} />
      </button>

      <div className="space-y-0">
        <div className="border-b border-white/10 px-4 py-3">
          <p className="mb-2 text-sm font-bold text-white">Stream</p>
          <p className="font-mono text-xs leading-relaxed text-white/80">
            {playbackInfo.container?.toUpperCase() || '—'}
            {playbackInfo.bitrate > 0 &&
              ` (${playbackInfo.bitrate >= 1000 ? `${(playbackInfo.bitrate / 1000).toFixed(1)} mbps` : `${playbackInfo.bitrate} kbps`})`}
          </p>
          {playbackInfo.method === 'DirectPlay' && (
            <p className="font-mono text-xs leading-relaxed text-white/80">
              <span className="text-white/50">→ </span>Direct Play
            </p>
          )}
          {playbackInfo.method !== 'DirectPlay' && (
            <p className="font-mono text-xs leading-relaxed text-white/80">
              <span className="text-white/50">→ </span>
              HLS
              {playbackInfo.estimated_bitrate > 0
                ? ` (${playbackInfo.estimated_bitrate >= 1000 ? `${(playbackInfo.estimated_bitrate / 1000).toFixed(1)} mbps` : `${playbackInfo.estimated_bitrate} kbps`})`
                : playbackInfo.bitrate > 0
                  ? ` (${playbackInfo.bitrate >= 1000 ? `${(playbackInfo.bitrate / 1000).toFixed(1)} mbps` : `${playbackInfo.bitrate} kbps`})`
                  : ''}
            </p>
          )}
          {playbackInfo.method === 'TranscodeAudio' && (
            <p className="mt-1 text-[11px] text-white/50">Converting audio to compatible codec</p>
          )}
        </div>

        <div className="border-b border-white/10 px-4 py-3">
          <p className="mb-2 text-sm font-bold text-white">Video</p>
          <p className="font-mono text-xs leading-relaxed text-white/80">
            {playbackInfo.height > 0 ? `${playbackInfo.height}p` : '—'}{' '}
            {playbackInfo.video_codec?.toUpperCase() || ''}
          </p>
          <p className="font-mono text-xs leading-relaxed text-white/80">
            {playbackInfo.video_profile && `${playbackInfo.video_profile} `}
            {playbackInfo.video_level > 0 && `${playbackInfo.video_level} `}
            {playbackInfo.bitrate > 0 &&
              `${playbackInfo.bitrate >= 1000 ? `${(playbackInfo.bitrate / 1000).toFixed(0)} mbps` : `${playbackInfo.bitrate} kbps`} `}
            {(() => {
              if (playbackInfo.video_fps > 0) {
                return `${Number.isInteger(playbackInfo.video_fps) ? playbackInfo.video_fps : playbackInfo.video_fps.toFixed(3)} fps`
              }
              const video = videoRef.current
              if (video && 'getVideoPlaybackQuality' in video && video.currentTime > 2) {
                const quality = video.getVideoPlaybackQuality()
                if (quality.totalVideoFrames > 0) {
                  return `${(quality.totalVideoFrames / video.currentTime).toFixed(3)} fps`
                }
              }
              return ''
            })()}
          </p>
          {playbackInfo.method === 'FullTranscode' && (
            <p className="font-mono text-xs leading-relaxed text-white/80">
              <span className="text-white/50">→ </span>
              Transcode ({playbackInfo.video_codec?.toUpperCase() || 'H264'}
              {playbackInfo.estimated_bitrate > 0 &&
                ` ${playbackInfo.estimated_bitrate >= 1000 ? `${(playbackInfo.estimated_bitrate / 1000).toFixed(0)} mbps` : `${playbackInfo.estimated_bitrate} kbps`}`}
              )
            </p>
          )}
          {(playbackInfo.method === 'DirectPlay' ||
            playbackInfo.method === 'DirectStream' ||
            playbackInfo.method === 'TranscodeAudio') && (
            <p className="font-mono text-xs leading-relaxed text-white/80">
              <span className="text-white/50">→ </span>Direct Play
            </p>
          )}
          <p className="mt-1.5 font-mono text-xs text-white/80">
            Dropped Frames{' '}
            <span
              className={(() => {
                const dropped =
                  videoRef.current?.getVideoPlaybackQuality?.()?.droppedVideoFrames ?? 0
                return dropped > 0 ? 'text-yellow-400' : 'text-white/80'
              })()}
            >
              {videoRef.current?.getVideoPlaybackQuality?.()?.droppedVideoFrames ?? 0}
            </span>
          </p>
        </div>

        {selectedAudio && (
          <div className="px-4 py-3">
            <p className="mb-2 text-sm font-bold text-white">Audio</p>
            <p className="font-mono text-xs leading-relaxed text-white/80">
              {selectedAudio.language || 'Unknown'} {selectedAudio.codec?.toUpperCase() || ''}{' '}
              {formatChannelLayout(selectedAudio.channels)}
              {selectedAudio.is_default && ' (Default)'}
            </p>
            <p className="font-mono text-xs leading-relaxed text-white/80">
              {selectedAudio.bitrate > 0 &&
                `${selectedAudio.bitrate >= 1000 ? `${Math.round(selectedAudio.bitrate / 1000)} kbps` : `${selectedAudio.bitrate} bps`} `}
              {selectedAudio.sample_rate > 0 && `${selectedAudio.sample_rate} Hz`}
            </p>
            {(playbackInfo.method === 'FullTranscode' ||
              playbackInfo.method === 'TranscodeAudio') && (
              <p className="mt-1 text-[11px] text-white/50">Audio is being transcoded</p>
            )}
          </div>
        )}

        <div className="border-t border-white/10 px-4 py-2.5">
          <button
            onClick={onClose}
            className="flex items-center gap-1.5 text-xs text-white/60 transition-colors hover:text-white"
          >
            <LuActivity size={13} />
            Close Playback Info
          </button>
        </div>
      </div>
    </div>
  )
})
