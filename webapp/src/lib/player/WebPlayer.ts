/**
 * WebPlayer — IPlayer over HTMLVideoElement.
 *
 * Thin wrapper. Does NOT own hls.js — WatchPage still constructs / destroys hls.js
 * itself for the web path (the integration is too entangled with React state to
 * cleanly extract right now). WebPlayer is purely a normalized facade over the
 * video element so WatchPage can call `player.play()` / `player.seek(s)` etc.
 * with the same surface as DesktopPlayer.
 */

import type Hls from 'hls.js'
import type {
  BufferRange,
  IPlayer,
  PlayerEvent,
  PlayerEventCallback,
  PlayerEventPayload,
  PlayerLoadOptions,
  PlayerSubtitleTrack,
  Unsubscribe,
  VideoSize,
} from './types'

type Listeners = {
  [K in PlayerEvent]: Set<PlayerEventCallback<K>>
}

const VIDEO_TO_PLAYER_EVENTS: Array<{
  videoEvent: keyof HTMLVideoElementEventMap
  playerEvent: PlayerEvent
}> = [
  { videoEvent: 'play', playerEvent: 'play' },
  { videoEvent: 'pause', playerEvent: 'pause' },
  { videoEvent: 'ended', playerEvent: 'ended' },
  { videoEvent: 'seeking', playerEvent: 'seeking' },
  { videoEvent: 'seeked', playerEvent: 'seeked' },
  { videoEvent: 'waiting', playerEvent: 'waiting' },
  { videoEvent: 'playing', playerEvent: 'playing' },
]

export class WebPlayer implements IPlayer {
  readonly kind = 'web' as const

  private video: HTMLVideoElement | null = null
  private hls: Hls | null = null
  private listeners: Listeners = {
    ready: new Set(),
    play: new Set(),
    pause: new Set(),
    ended: new Set(),
    error: new Set(),
    'time-update': new Set(),
    'buffer-update': new Set(),
    seeking: new Set(),
    seeked: new Set(),
    waiting: new Set(),
    playing: new Set(),
  }
  private cleanups: Array<() => void> = []
  private subtitleEl: HTMLTrackElement | null = null

  mount(container: HTMLElement): void {
    // Web path: WatchPage already renders the <video> element in JSX.
    // We expect the container to contain the video as its first child.
    const video = container.querySelector('video') as HTMLVideoElement | null
    if (!video) {
      throw new Error('WebPlayer.mount: container must contain a <video> child')
    }
    this.attachToElement(video)
  }

  /** Direct alternative to mount() — pass the existing <video> element. */
  attachToElement(video: HTMLVideoElement): void {
    if (this.video === video) return
    this.detachListeners()
    this.video = video
    this.attachListeners()
  }

  /** WatchPage builds hls.js itself (for now); call this so WebPlayer can
   *  expose it via hlsInstance() to consumers that need it. */
  setHlsInstance(hls: Hls | null): void {
    this.hls = hls
  }

  unmount(): void {
    this.detachListeners()
    this.video = null
  }

  destroy(): void {
    this.unmount()
    for (const set of Object.values(this.listeners) as Set<unknown>[]) set.clear()
  }

  // ── Lifecycle ───────────────────────────────────────────────────────────────

  async load(opts: PlayerLoadOptions): Promise<void> {
    if (!this.video) return
    this.video.src = opts.url
    if (opts.startTime != null) this.video.currentTime = opts.startTime
  }

  async unload(): Promise<void> {
    if (!this.video) return
    this.video.removeAttribute('src')
    this.video.load()
  }

  // ── Playback ────────────────────────────────────────────────────────────────

  async play(): Promise<void> {
    if (!this.video) return
    await this.video.play()
  }

  pause(): void {
    this.video?.pause()
  }

  seek(seconds: number): void {
    if (!this.video) return
    this.video.currentTime = seconds
  }

  setVolume(v: number): void {
    if (!this.video) return
    this.video.volume = Math.max(0, Math.min(1, v))
  }

  setMuted(m: boolean): void {
    if (!this.video) return
    this.video.muted = m
  }

  setRate(r: number): void {
    if (!this.video) return
    this.video.playbackRate = r
  }

  // ── Getters ─────────────────────────────────────────────────────────────────

  currentTime(): number {
    return this.video?.currentTime ?? 0
  }
  duration(): number {
    const d = this.video?.duration
    return d != null && isFinite(d) ? d : 0
  }
  paused(): boolean {
    return this.video?.paused ?? true
  }
  volume(): number {
    return this.video?.volume ?? 1
  }
  muted(): boolean {
    return this.video?.muted ?? false
  }
  rate(): number {
    return this.video?.playbackRate ?? 1
  }
  buffered(): BufferRange[] {
    if (!this.video) return []
    const ranges: BufferRange[] = []
    const b = this.video.buffered
    for (let i = 0; i < b.length; i++) ranges.push({ start: b.start(i), end: b.end(i) })
    return ranges
  }
  videoSize(): VideoSize | null {
    const v = this.video
    if (!v || !v.videoWidth || !v.videoHeight) return null
    return { width: v.videoWidth, height: v.videoHeight }
  }

  // ── Subtitles ───────────────────────────────────────────────────────────────

  setSubtitle(track: PlayerSubtitleTrack | null): void {
    if (!this.video) return
    if (this.subtitleEl) {
      this.subtitleEl.remove()
      this.subtitleEl = null
    }
    if (!track) return
    const el = document.createElement('track')
    el.kind = 'subtitles'
    el.src = track.url
    el.label = track.label
    el.srclang = track.lang
    el.default = true
    this.video.appendChild(el)
    this.subtitleEl = el
  }

  setSubtitleDelay(_delayMs: number): void {
    // Web subtitles flow through DualSubtitleOverlay which applies its own
    // offset state; the native <track> element doesn't expose a delay knob.
  }

  setSubtitleVisible(visible: boolean): void {
    if (!this.video) return
    for (let i = 0; i < this.video.textTracks.length; i++) {
      this.video.textTracks[i].mode = visible ? 'showing' : 'hidden'
    }
  }

  // ── Audio tracks ────────────────────────────────────────────────────────────

  setAudioTrackByLang(lang: string, name?: string): void {
    if (this.hls && this.hls.audioTracks.length > 1) {
      const idx = this.hls.audioTracks.findIndex(
        (t) =>
          t.lang === lang ||
          (name && t.name?.toLowerCase() === name.toLowerCase()) ||
          t.name?.toLowerCase() === lang.toLowerCase(),
      )
      if (idx >= 0 && idx !== this.hls.audioTrack) {
        this.hls.audioTrack = idx
      }
      return
    }
    // Native HTMLMediaElement audioTracks (Safari)
    const v = this.video as HTMLVideoElement & {
      audioTracks?: { length: number; [k: number]: { language: string; enabled: boolean } }
    }
    if (!v?.audioTracks) return
    for (let i = 0; i < v.audioTracks.length; i++) {
      v.audioTracks[i].enabled = v.audioTracks[i].language === lang
    }
  }

  // ── Events ──────────────────────────────────────────────────────────────────

  on<E extends PlayerEvent>(event: E, cb: PlayerEventCallback<E>): Unsubscribe {
    this.listeners[event].add(cb as PlayerEventCallback<PlayerEvent>) as never
    return () => {
      this.listeners[event].delete(cb as PlayerEventCallback<PlayerEvent>) as never
    }
  }

  // ── Escape hatches ──────────────────────────────────────────────────────────

  videoElement(): HTMLVideoElement | null {
    return this.video
  }
  hlsInstance(): Hls | null {
    return this.hls
  }

  // ── Internals ───────────────────────────────────────────────────────────────

  private emit<E extends PlayerEvent>(event: E, payload: PlayerEventPayload[E]): void {
    for (const cb of this.listeners[event]) (cb as PlayerEventCallback<E>)(payload)
  }

  private attachListeners(): void {
    const v = this.video
    if (!v) return

    for (const { videoEvent, playerEvent } of VIDEO_TO_PLAYER_EVENTS) {
      const handler = () => this.emit(playerEvent, undefined as never)
      v.addEventListener(videoEvent, handler)
      this.cleanups.push(() => v.removeEventListener(videoEvent, handler))
    }

    const onTime = () => this.emit('time-update', { currentTime: v.currentTime })
    v.addEventListener('timeupdate', onTime)
    this.cleanups.push(() => v.removeEventListener('timeupdate', onTime))

    const onProgress = () => this.emit('buffer-update', { ranges: this.buffered() })
    v.addEventListener('progress', onProgress)
    this.cleanups.push(() => v.removeEventListener('progress', onProgress))

    const onLoaded = () =>
      this.emit('ready', { duration: this.duration(), videoSize: this.videoSize() })
    v.addEventListener('loadedmetadata', onLoaded)
    this.cleanups.push(() => v.removeEventListener('loadedmetadata', onLoaded))

    const onError = () => {
      const msg = v.error?.message || `media error code=${v.error?.code ?? '?'}`
      this.emit('error', { message: msg, fatal: true })
    }
    v.addEventListener('error', onError)
    this.cleanups.push(() => v.removeEventListener('error', onError))
  }

  private detachListeners(): void {
    for (const fn of this.cleanups) fn()
    this.cleanups = []
  }
}
