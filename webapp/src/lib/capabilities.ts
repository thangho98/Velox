// Client capability detection using MediaSource API
// Results are cached in localStorage

const STORAGE_KEY = 'velox-client-capabilities'

export interface ClientCapabilities {
  browser: string
  platform: string
  isMobile: boolean
  videoCodecs: string[]
  audioCodecs: string[]
  containers: string[]
  supportsHLS: boolean
  supportsWebM: boolean
  maxResolution: {
    width: number
    height: number
  }
  detectedAt: string
}

// Test codec support using MediaSource.isTypeSupported
function testVideoCodec(mimeType: string): boolean {
  try {
    if (window.MediaSource && MediaSource.isTypeSupported) {
      return MediaSource.isTypeSupported(mimeType)
    }
  } catch {
    // ignore
  }
  // Fallback: try creating a video element
  const video = document.createElement('video')
  return video.canPlayType(mimeType) !== ''
}

function testAudioCodec(mimeType: string): boolean {
  try {
    const audio = document.createElement('audio')
    return audio.canPlayType(mimeType) !== ''
  } catch {
    return false
  }
}

// Detect browser
function detectBrowser(): string {
  const ua = navigator.userAgent.toLowerCase()

  if (ua.includes('edg')) return 'edge'
  if (ua.includes('chrome') && ua.includes('safari')) return 'chrome'
  if (ua.includes('firefox')) return 'firefox'
  if (ua.includes('safari') && !ua.includes('chrome')) return 'safari'

  return 'unknown'
}

// Detect platform
function detectPlatform(): { platform: string; isMobile: boolean } {
  const ua = navigator.userAgent.toLowerCase()

  if (/iphone|ipad|ipod/.test(ua)) return { platform: 'ios', isMobile: true }
  if (/android/.test(ua)) return { platform: 'android', isMobile: true }
  if (/windows/.test(ua)) return { platform: 'windows', isMobile: false }
  if (/macintosh|mac os x/.test(ua)) return { platform: 'macos', isMobile: false }
  if (/linux/.test(ua)) return { platform: 'linux', isMobile: false }

  return { platform: 'unknown', isMobile: false }
}

// Detect max resolution based on device
function detectMaxResolution(): { width: number; height: number } {
  // Get screen resolution
  const width = window.screen.width * window.devicePixelRatio
  const height = window.screen.height * window.devicePixelRatio

  // Cap to reasonable values
  if (width > 3840 || height > 2160) return { width: 3840, height: 2160 } // 4K
  if (width > 1920 || height > 1080) return { width: 1920, height: 1080 } // 1080p

  return { width: 1920, height: 1080 }
}

// Test HLS support
function testHLSSupport(): boolean {
  const video = document.createElement('video')
  return (
    video.canPlayType('application/vnd.apple.mpegurl') !== '' ||
    video.canPlayType('application/x-mpegURL') !== ''
  )
}

// Test WebM support
function testWebMSupport(): boolean {
  const video = document.createElement('video')
  return video.canPlayType('video/webm') !== ''
}

// Main detection function
export function detectCapabilities(): ClientCapabilities {
  const browser = detectBrowser()
  const { platform, isMobile } = detectPlatform()
  const maxResolution = detectMaxResolution()

  // Video codec tests
  const videoCodecs: string[] = []

  // H.264 (baseline, main, high profiles)
  if (testVideoCodec('video/mp4; codecs="avc1.42E01E"')) {
    videoCodecs.push('h264')
  }

  // H.265/HEVC
  if (testVideoCodec('video/mp4; codecs="hev1.1.6.L93.B0"')) {
    videoCodecs.push('hevc')
  }

  // VP9
  if (testVideoCodec('video/webm; codecs="vp9"')) {
    videoCodecs.push('vp9')
  }

  // VP8
  if (testVideoCodec('video/webm; codecs="vp8"')) {
    videoCodecs.push('vp8')
  }

  // AV1
  if (testVideoCodec('video/mp4; codecs="av01.0.00M.08"')) {
    videoCodecs.push('av1')
  }

  // Audio codec tests
  const audioCodecs: string[] = []

  if (testAudioCodec('audio/mp4; codecs="mp4a.40.2"')) {
    audioCodecs.push('aac')
  }
  if (testAudioCodec('audio/webm; codecs="opus"')) {
    audioCodecs.push('opus')
  }
  if (testAudioCodec('audio/mpeg')) {
    audioCodecs.push('mp3')
  }
  if (testAudioCodec('audio/flac')) {
    audioCodecs.push('flac')
  }
  if (testAudioCodec('audio/mp4; codecs="ac-3"')) {
    audioCodecs.push('ac3')
  }

  // Container support
  const containers: string[] = ['mp4']
  if (testWebMSupport()) containers.push('webm')
  if (testHLSSupport()) containers.push('hls')

  return {
    browser,
    platform,
    isMobile,
    videoCodecs,
    audioCodecs,
    containers,
    supportsHLS: testHLSSupport(),
    supportsWebM: testWebMSupport(),
    maxResolution,
    detectedAt: new Date().toISOString(),
  }
}

// Get cached capabilities or detect new
export function getCapabilities(): ClientCapabilities {
  try {
    const cached = localStorage.getItem(STORAGE_KEY)
    if (cached) {
      const parsed = JSON.parse(cached) as ClientCapabilities
      // Refresh if older than 7 days
      const detectedAt = new Date(parsed.detectedAt)
      const daysSince = (Date.now() - detectedAt.getTime()) / (1000 * 60 * 60 * 24)
      if (daysSince < 7) {
        return parsed
      }
    }
  } catch {
    // ignore parsing errors
  }

  // Detect and cache
  const caps = detectCapabilities()
  localStorage.setItem(STORAGE_KEY, JSON.stringify(caps))
  return caps
}

// Clear cached capabilities
export function clearCapabilities(): void {
  localStorage.removeItem(STORAGE_KEY)
}

// Check if specific codec is supported
export function supportsVideoCodec(codec: string): boolean {
  const caps = getCapabilities()
  return caps.videoCodecs.includes(codec.toLowerCase())
}

export function supportsAudioCodec(codec: string): boolean {
  const caps = getCapabilities()
  return caps.audioCodecs.includes(codec.toLowerCase())
}

// Check if can play without transcoding
export function canDirectPlay(videoCodec: string, audioCodec: string, container: string): boolean {
  const caps = getCapabilities()
  return (
    caps.videoCodecs.includes(videoCodec.toLowerCase()) &&
    caps.audioCodecs.includes(audioCodec.toLowerCase()) &&
    caps.containers.includes(container.toLowerCase())
  )
}

// Check if HLS is preferred for this client
export function shouldUseHLS(): boolean {
  const caps = getCapabilities()
  return caps.supportsHLS && (caps.isMobile || caps.browser === 'safari')
}
