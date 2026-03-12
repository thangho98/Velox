import { useState, useRef, useEffect } from 'react'

interface HlsLevel {
  level: number
  height: number
  bitrate?: number
}

interface QualityPickerProps {
  levels: HlsLevel[]
  currentLevel: number // -1 = auto, 0+ = specific level index
  onChange: (level: number) => void
  currentHeight?: number // Actual height being played (for badge)
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

  // Close when clicking outside
  useEffect(() => {
    const handleClickOutside = (e: MouseEvent) => {
      if (containerRef.current && !containerRef.current.contains(e.target as Node)) {
        setIsOpen(false)
      }
    }
    if (isOpen) {
      document.addEventListener('mousedown', handleClickOutside)
    }
    return () => document.removeEventListener('mousedown', handleClickOutside)
  }, [isOpen])

  // Get current quality display
  const getCurrentQualityLabel = (): string => {
    if (currentLevel === -1) {
      // Auto mode - show current playing height if available
      if (currentHeight) {
        return QUALITY_LABELS[currentHeight] || `${currentHeight}p`
      }
      return 'Auto'
    }
    const level = levels.find((l) => l.level === currentLevel)
    if (level) {
      return QUALITY_LABELS[level.height] || `${level.height}p`
    }
    return 'Auto'
  }

  // Get badge for quality (HD, 4K)
  const getQualityBadge = (): string | null => {
    const height =
      currentLevel === -1 ? currentHeight : levels.find((l) => l.level === currentLevel)?.height
    if (height && height >= 2160) return '4K'
    if (height && height >= 1080) return 'HD'
    return null
  }

  const badge = getQualityBadge()
  const currentLabel = getCurrentQualityLabel()

  // Sort levels by height desc
  const sortedLevels = [...levels].sort((a, b) => b.height - a.height)

  return (
    <div ref={containerRef} className="relative">
      <button
        onClick={() => setIsOpen(!isOpen)}
        className={`flex items-center gap-1 rounded px-2 py-1.5 text-sm font-medium transition-colors ${
          isOpen ? 'bg-netflix-red text-white' : 'bg-white/10 text-white hover:bg-white/20'
        }`}
      >
        <svg className="h-4 w-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path
            strokeLinecap="round"
            strokeLinejoin="round"
            strokeWidth={2}
            d="M13 10V3L4 14h7v7l9-11h-7z"
          />
        </svg>
        <span>{currentLabel}</span>
        {badge && <span className="ml-1 rounded bg-netflix-red px-1 text-xs">{badge}</span>}
      </button>

      {isOpen && (
        <div className="absolute bottom-full right-0 mb-2 min-w-[160px] rounded-lg bg-black/90 p-2 shadow-xl">
          <p className="mb-1 px-3 text-xs text-gray-400">Quality</p>

          {/* Auto option */}
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

          {/* Original option - only show if no transcoding (single level or direct play) */}
          {levels.length === 0 && (
            <button
              onClick={() => {
                onChange(-2) // Special value for original
                setIsOpen(false)
              }}
              className={`w-full rounded px-3 py-1.5 text-left text-sm ${
                currentLevel === -2 ? 'bg-netflix-red text-white' : 'text-white hover:bg-white/10'
              }`}
            >
              Original
            </button>
          )}

          {/* Quality levels */}
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
              {level.height >= 1080 && <span className="text-xs opacity-75">HD</span>}
              {level.height >= 2160 && <span className="text-xs opacity-75">4K</span>}
            </button>
          ))}
        </div>
      )}
    </div>
  )
}
