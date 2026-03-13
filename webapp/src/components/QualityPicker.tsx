import { useState, useRef, useEffect } from 'react'
import { LuZap } from 'react-icons/lu'

interface HlsLevel {
  level: number
  height: number
  bitrate?: number
}

interface QualityPickerProps {
  levels: HlsLevel[]
  currentLevel: number
  onChange: (level: number) => void
  currentHeight?: number
}

const QUALITY_LABELS: Record<number, string> = {
  2160: '4K',
  1440: '1440p',
  1080: '1080p',
  720: '720p',
  480: '480p',
  360: '360p',
  240: '240p',
}

export function QualityPicker({
  levels,
  currentLevel,
  onChange,
  currentHeight,
}: QualityPickerProps) {
  const [isOpen, setIsOpen] = useState(false)
  const containerRef = useRef<HTMLDivElement>(null)

  useEffect(() => {
    const handleClickOutside = (e: MouseEvent) => {
      if (containerRef.current && !containerRef.current.contains(e.target as Node)) {
        setIsOpen(false)
      }
    }
    if (isOpen) document.addEventListener('mousedown', handleClickOutside)
    return () => document.removeEventListener('mousedown', handleClickOutside)
  }, [isOpen])

  const getCurrentQualityLabel = (): string => {
    if (currentLevel === -1) {
      if (currentHeight) return QUALITY_LABELS[currentHeight] || `${currentHeight}p`
      return 'Auto'
    }
    const level = levels.find((l) => l.level === currentLevel)
    if (level) return QUALITY_LABELS[level.height] || `${level.height}p`
    return 'Auto'
  }

  const getQualityBadge = (): string | null => {
    const height =
      currentLevel === -1 ? currentHeight : levels.find((l) => l.level === currentLevel)?.height
    if (height && height >= 2160) return '4K'
    if (height && height >= 1080) return 'HD'
    return null
  }

  const badge = getQualityBadge()
  const sortedLevels = [...levels].sort((a, b) => b.height - a.height)

  return (
    <div ref={containerRef} className="relative">
      <button
        onClick={() => setIsOpen(!isOpen)}
        className={`flex items-center gap-1.5 rounded px-2 py-1.5 text-sm font-medium transition-colors ${
          isOpen ? 'bg-netflix-red text-white' : 'bg-white/10 text-white hover:bg-white/20'
        }`}
      >
        <LuZap size={14} />
        <span>{getCurrentQualityLabel()}</span>
        {badge && <span className="rounded bg-netflix-red/80 px-1 text-xs">{badge}</span>}
      </button>

      {isOpen && (
        <div className="absolute bottom-full right-0 mb-2 min-w-[160px] rounded-lg bg-black/90 p-2 shadow-xl backdrop-blur-sm">
          <p className="mb-1 px-3 text-xs text-gray-400">Quality</p>
          <button
            onClick={() => {
              onChange(-1)
              setIsOpen(false)
            }}
            className={`flex w-full items-center justify-between rounded px-3 py-1.5 text-left text-sm ${
              currentLevel === -1 ? 'bg-netflix-red text-white' : 'text-white hover:bg-white/10'
            }`}
          >
            <span>Auto</span>
            {currentLevel === -1 && currentHeight && (
              <span className="text-xs opacity-75">
                ({QUALITY_LABELS[currentHeight] || `${currentHeight}p`})
              </span>
            )}
          </button>
          {sortedLevels.map((level) => (
            <button
              key={level.level}
              onClick={() => {
                onChange(level.level)
                setIsOpen(false)
              }}
              className={`flex w-full items-center justify-between rounded px-3 py-1.5 text-left text-sm ${
                currentLevel === level.level
                  ? 'bg-netflix-red text-white'
                  : 'text-white hover:bg-white/10'
              }`}
            >
              <span>{QUALITY_LABELS[level.height] || `${level.height}p`}</span>
              {level.height >= 2160 && <span className="text-xs opacity-75">4K</span>}
              {level.height >= 1080 && level.height < 2160 && (
                <span className="text-xs opacity-75">HD</span>
              )}
            </button>
          ))}
        </div>
      )}
    </div>
  )
}
