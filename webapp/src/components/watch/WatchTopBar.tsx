import { memo } from 'react'
import { LuChevronLeft, LuVolume1, LuVolume2, LuVolumeX, LuCast } from 'react-icons/lu'
import { useTranslation } from '@/hooks/useTranslation'

interface WatchTopBarProps {
  isMuted: boolean
  onBack: () => void
  onMuteToggle: () => void
  onVolumeChange: (value: number) => void
  volume: number
  castAvailable?: boolean
  castConnected?: boolean
  casting?: boolean
  onCastClick?: () => void
}

export const WatchTopBar = memo(function WatchTopBar({
  isMuted,
  onBack,
  onMuteToggle,
  onVolumeChange,
  volume,
  castAvailable,
  castConnected,
  casting,
  onCastClick,
}: WatchTopBarProps) {
  const { t } = useTranslation('navigation')
  const VolumeIcon = isMuted || volume === 0 ? LuVolumeX : volume < 0.5 ? LuVolume1 : LuVolume2

  return (
    <div
      className="flex items-center justify-between px-5 py-4"
      style={{ background: 'linear-gradient(to bottom, rgba(0,0,0,0.7) 0%, transparent 100%)' }}
    >
      <button
        onClick={onBack}
        className="flex items-center gap-1.5 text-white/80 transition-colors hover:text-white"
      >
        <LuChevronLeft size={22} />
        <span className="text-sm font-medium">{t('back')}</span>
      </button>

      <div className="flex items-center gap-3">
        <div className="flex items-center gap-2">
          <input
            type="range"
            min={0}
            max={1}
            step={0.05}
            value={isMuted ? 0 : volume}
            onChange={(e) => onVolumeChange(Number(e.target.value))}
            className="h-0.5 w-28 cursor-pointer accent-white"
          />
          <button
            onClick={onMuteToggle}
            className="text-white/80 transition-colors hover:text-white"
          >
            <VolumeIcon size={20} />
          </button>
        </div>

        {castAvailable && onCastClick && (
          <button
            onClick={onCastClick}
            className={`transition-colors ${
              casting
                ? 'text-blue-400 hover:text-blue-300'
                : castConnected
                  ? 'text-white hover:text-blue-400'
                  : 'text-white/60 hover:text-white'
            }`}
            title={
              casting ? 'Stop casting' : castConnected ? 'Cast to device' : 'Connect to Chromecast'
            }
          >
            <LuCast size={20} />
          </button>
        )}
      </div>
    </div>
  )
})
