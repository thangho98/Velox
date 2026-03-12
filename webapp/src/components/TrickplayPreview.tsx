import { useState, useEffect, useRef, useCallback } from 'react'
import { useAuthStore } from '@/stores/auth'

interface VttCue {
  startTime: number
  endTime: number
  imageUrl: string
  x: number
  y: number
  width: number
  height: number
}

interface TrickplayPreviewProps {
  mediaId: number
  currentHoverTime: number // seconds
  visible: boolean
  positionX: number // pixel position on timeline for centering
}

export function TrickplayPreview({
  mediaId,
  currentHoverTime,
  visible,
  positionX,
}: TrickplayPreviewProps) {
  const [cues, setCues] = useState<VttCue[]>([])
  const [currentCue, setCurrentCue] = useState<VttCue | null>(null)
  const loadedRef = useRef(false)
  const preloadedSprites = useRef<Set<string>>(new Set())
  const { accessToken } = useAuthStore()

  // Parse VTT time string (HH:MM:SS.mmm or MM:SS.mmm)
  const parseTime = (timeStr: string): number => {
    const parts = timeStr.split(':').map(Number)
    if (parts.length === 3) {
      // HH:MM:SS.mmm
      return parts[0] * 3600 + parts[1] * 60 + parts[2]
    } else if (parts.length === 2) {
      // MM:SS.mmm
      return parts[0] * 60 + parts[1]
    }
    return 0
  }

  // Parse VTT content
  const parseVtt = useCallback(
    (vttText: string, baseUrl: string, token: string | null): VttCue[] => {
      const lines = vttText.split('\n')
      const parsed: VttCue[] = []
      let i = 0

      // Skip WEBVTT header
      while (i < lines.length && !lines[i].includes('-->')) {
        i++
      }

      while (i < lines.length) {
        const line = lines[i].trim()

        // Look for timestamp line
        if (line.includes('-->')) {
          const timeMatch = line.match(
            /(\d{2}:\d{2}:\d{2}\.\d{3})\s*-->\s*(\d{2}:\d{2}:\d{2}\.\d{3})/,
          )
          if (timeMatch && i + 1 < lines.length) {
            const startTime = parseTime(timeMatch[1])
            const endTime = parseTime(timeMatch[2])
            const imageLine = lines[i + 1].trim()

            // Parse image reference: sprite_001.jpg#xywh=0,0,320,180
            const imgMatch = imageLine.match(/([^#]+)#xywh=(\d+),(\d+),(\d+),(\d+)/)
            if (imgMatch) {
              const spriteFile = imgMatch[1]
              const x = parseInt(imgMatch[2], 10)
              const y = parseInt(imgMatch[3], 10)
              const width = parseInt(imgMatch[4], 10)
              const height = parseInt(imgMatch[5], 10)

              // Construct full URL: backend may return absolute (/api/...) or bare filename.
              const rawImageUrl =
                spriteFile.startsWith('http') || spriteFile.startsWith('/')
                  ? spriteFile
                  : `${baseUrl}/${spriteFile}`
              // Append auth token so <img> requests are authorized.
              const imageUrl = token
                ? rawImageUrl +
                  (rawImageUrl.includes('?') ? '&' : '?') +
                  'token=' +
                  encodeURIComponent(token)
                : rawImageUrl

              parsed.push({
                startTime,
                endTime,
                imageUrl,
                x,
                y,
                width,
                height,
              })
            }
            i += 2
            continue
          }
        }
        i++
      }

      return parsed
    },
    [],
  )

  // Fetch VTT manifest, polling if backend returns 202 (still generating).
  useEffect(() => {
    if (loadedRef.current) return

    let cancelled = false
    let retryTimer: ReturnType<typeof setTimeout> | null = null

    const vttUrl = `/api/media/${mediaId}/trickplay/manifest.vtt`
    const baseUrl = vttUrl.substring(0, vttUrl.lastIndexOf('/') + 1)

    const fetchVtt = async () => {
      try {
        const headers: Record<string, string> = {}
        if (accessToken) headers['Authorization'] = `Bearer ${accessToken}`
        const response = await fetch(vttUrl, { headers })
        if (cancelled) return

        if (response.status === 202) {
          // Backend is generating sprites asynchronously — retry in 5s.
          retryTimer = setTimeout(fetchVtt, 5000)
          return
        }

        if (response.ok) {
          const vttText = await response.text()
          const parsed = parseVtt(vttText, baseUrl, accessToken)
          setCues(parsed)
          loadedRef.current = true

          // Preload sprite images to avoid flicker on first hover.
          const uniqueSprites = new Set(parsed.map((c) => c.imageUrl))
          uniqueSprites.forEach((spriteUrl) => {
            if (!preloadedSprites.current.has(spriteUrl)) {
              const img = new Image()
              img.src = spriteUrl
              preloadedSprites.current.add(spriteUrl)
            }
          })
          return
        }

        // 404 (disabled) or other error — trickplay not available.
        loadedRef.current = true
      } catch {
        // Network error — give up.
        loadedRef.current = true
      }
    }

    fetchVtt()

    return () => {
      cancelled = true
      if (retryTimer) clearTimeout(retryTimer)
    }
  }, [mediaId, parseVtt, accessToken])

  // Find current cue based on hover time
  useEffect(() => {
    if (!visible || cues.length === 0) {
      setCurrentCue(null)
      return
    }

    const cue = cues.find((c) => currentHoverTime >= c.startTime && currentHoverTime < c.endTime)
    setCurrentCue(cue || null)
  }, [currentHoverTime, cues, visible])

  if (!visible || !currentCue) {
    return null
  }

  // Calculate thumbnail position (centered above timeline)
  const thumbnailWidth = 160
  const thumbnailHeight = 90
  const left = Math.max(
    10,
    Math.min(positionX - thumbnailWidth / 2, window.innerWidth - thumbnailWidth - 10),
  )

  return (
    <div
      className="absolute bottom-16 z-20 rounded border border-white/20 bg-black/80 shadow-xl"
      style={{
        left: `${left}px`,
        width: `${thumbnailWidth}px`,
        height: `${thumbnailHeight}px`,
      }}
    >
      {/* Thumbnail using CSS clip to show sprite region */}
      <div
        className="h-full w-full overflow-hidden"
        style={{
          backgroundImage: `url(${currentCue.imageUrl})`,
          backgroundPosition: `-${currentCue.x}px -${currentCue.y}px`,
          backgroundRepeat: 'no-repeat',
        }}
      />
      {/* Time display */}
      <div className="absolute bottom-1 left-1/2 -translate-x-1/2 rounded bg-black/60 px-2 py-0.5 text-xs text-white">
        {formatTime(currentHoverTime)}
      </div>
    </div>
  )
}

function formatTime(seconds: number): string {
  const mins = Math.floor(seconds / 60)
  const secs = Math.floor(seconds % 60)
  const hours = Math.floor(mins / 60)
  const displayMins = mins % 60

  if (hours > 0) {
    return `${hours}:${displayMins.toString().padStart(2, '0')}:${secs.toString().padStart(2, '0')}`
  }
  return `${mins}:${secs.toString().padStart(2, '0')}`
}
