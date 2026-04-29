/**
 * DesktopPlayer — IPlayer over Tauri commands → libmpv.
 *
 * Forwards every play/pause/seek/sub/audio call to a `player_*` Tauri command,
 * and translates the `player-event` event stream into IPlayer events. Keeps a
 * shadow copy of currentTime/duration/paused so the synchronous getters can
 * answer without round-tripping to Rust.
 */

import { invoke } from '@tauri-apps/api/core'
import { listen, type UnlistenFn } from '@tauri-apps/api/event'
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

type RawEvent = {
  kind: string
  detail?: string | null
  f?: number | null
  b?: boolean | null
}

type Listeners = {
  [K in PlayerEvent]: Set<PlayerEventCallback<K>>
}

export class DesktopPlayer implements IPlayer {
  readonly kind = 'desktop' as const

  private _currentTime = 0
  private _duration = 0
  private _paused = true
  private _volume = 1
  private _muted = false
  private _rate = 1
  private _videoSize: VideoSize | null = null
  private _ready = false

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
  private unlisten: UnlistenFn | null = null

  constructor() {
    void this.attachEventStream()
  }

  // ── Lifecycle ───────────────────────────────────────────────────────────────

  mount(_container: HTMLElement): void {
    // libmpv renders into a native NSOpenGLView mounted by Rust during setup.
    // Nothing to do here; the container should be transparent so the view shows.
  }

  unmount(): void {
    // Keep mpv alive across mounts/unmounts — only `unload` releases the source.
  }

  async load(opts: PlayerLoadOptions): Promise<void> {
    this._ready = false
    this._currentTime = opts.startTime ?? 0
    this._duration = 0
    this._videoSize = null
    await invoke('player_load', { url: opts.url, startTime: opts.startTime ?? null })
  }

  async unload(): Promise<void> {
    await invoke('player_stop').catch(() => {})
    this._ready = false
    this._currentTime = 0
    this._duration = 0
    this._paused = true
  }

  destroy(): void {
    this.unlisten?.()
    this.unlisten = null
    for (const set of Object.values(this.listeners) as Set<unknown>[]) set.clear()
  }

  // ── Playback ────────────────────────────────────────────────────────────────

  async play(): Promise<void> {
    await invoke('player_play')
  }
  pause(): void {
    void invoke('player_pause')
  }
  seek(seconds: number): void {
    this._currentTime = seconds
    this.emit('seeking', undefined)
    void invoke('player_seek', { seconds }).then(() => this.emit('seeked', undefined))
  }
  setVolume(v: number): void {
    this._volume = Math.max(0, Math.min(1, v))
    void invoke('player_set_volume', { volume: this._volume })
  }
  setMuted(m: boolean): void {
    this._muted = m
    void invoke('player_set_muted', { muted: m })
  }
  setRate(r: number): void {
    this._rate = r
    void invoke('player_set_rate', { rate: r })
  }

  // ── Getters ─────────────────────────────────────────────────────────────────

  currentTime(): number {
    return this._currentTime
  }
  duration(): number {
    return this._duration
  }
  paused(): boolean {
    return this._paused
  }
  volume(): number {
    return this._volume
  }
  muted(): boolean {
    return this._muted
  }
  rate(): number {
    return this._rate
  }
  buffered(): BufferRange[] {
    // libmpv doesn't expose easily-mappable buffer ranges; return what we know
    // — i.e. the area we've played through. WatchPage uses this only for the
    // seekbar's "downloaded" overlay; an approximation is fine.
    if (this._duration <= 0 || this._currentTime <= 0) return []
    return [{ start: 0, end: this._currentTime + 5 }]
  }
  videoSize(): VideoSize | null {
    return this._videoSize
  }

  // ── Subtitles ───────────────────────────────────────────────────────────────

  setSubtitle(track: PlayerSubtitleTrack | null): void {
    if (!track) {
      void invoke('player_sub_remove').catch(() => {})
      return
    }
    void invoke('player_sub_add', {
      url: track.url,
      lang: track.lang,
      label: track.label,
    }).catch(() => {})
  }
  setSubtitleDelay(delayMs: number): void {
    void invoke('player_sub_delay', { delayMs })
  }
  setSubtitleVisible(visible: boolean): void {
    void invoke('player_sub_visible', { visible })
  }

  // ── Audio tracks ────────────────────────────────────────────────────────────

  setAudioTrackByLang(lang: string, name?: string): void {
    void invoke('player_audio_set_lang', { lang, name: name ?? null }).catch(() => {})
  }

  // ── Events ──────────────────────────────────────────────────────────────────

  on<E extends PlayerEvent>(event: E, cb: PlayerEventCallback<E>): Unsubscribe {
    this.listeners[event].add(cb as PlayerEventCallback<PlayerEvent>) as never
    return () => {
      this.listeners[event].delete(cb as PlayerEventCallback<PlayerEvent>) as never
    }
  }

  // ── Escape hatches (none on desktop) ────────────────────────────────────────

  videoElement(): HTMLVideoElement | null {
    return null
  }
  hlsInstance(): null {
    return null
  }

  // ── Internals ───────────────────────────────────────────────────────────────

  private emit<E extends PlayerEvent>(event: E, payload: PlayerEventPayload[E]): void {
    for (const cb of this.listeners[event]) (cb as PlayerEventCallback<E>)(payload)
  }

  private async attachEventStream(): Promise<void> {
    this.unlisten = await listen<RawEvent>('player-event', (e) => {
      this.handleRawEvent(e.payload)
    })
  }

  private handleRawEvent(raw: RawEvent): void {
    switch (raw.kind) {
      case 'time-pos':
        if (raw.f != null) {
          this._currentTime = raw.f
          this.emit('time-update', { currentTime: raw.f })
        }
        break
      case 'duration':
        if (raw.f != null) {
          this._duration = raw.f
          if (!this._ready) {
            this._ready = true
            this.emit('ready', { duration: raw.f, videoSize: this._videoSize })
          }
        }
        break
      case 'pause':
        if (raw.b != null) {
          const wasPaused = this._paused
          this._paused = raw.b
          if (raw.b && !wasPaused) this.emit('pause', undefined)
          else if (!raw.b && wasPaused) this.emit('play', undefined)
        }
        break
      case 'buffering':
        if (raw.b != null) this.emit(raw.b ? 'waiting' : 'playing', undefined)
        break
      case 'file-loaded':
        // Width/height available after file-loaded; pull them lazily.
        void this.refreshVideoSize()
        break
      case 'end-file':
        this.emit('ended', undefined)
        break
      case 'shutdown':
        this.emit('error', { message: 'mpv shutdown', fatal: true })
        break
      default:
        break
    }
  }

  private async refreshVideoSize(): Promise<void> {
    try {
      const info = await invoke<{
        width?: string | null
        height?: string | null
      }>('color_info')
      const w = info.width ? parseInt(info.width, 10) : NaN
      const h = info.height ? parseInt(info.height, 10) : NaN
      if (!isNaN(w) && !isNaN(h)) {
        this._videoSize = { width: w, height: h }
        if (this._ready) {
          this.emit('ready', { duration: this._duration, videoSize: this._videoSize })
        }
      }
    } catch {
      // best-effort
    }
  }
}
