/**
 * IPlayer — platform-agnostic media player contract.
 *
 * Two implementations:
 *   - WebPlayer  (HTMLVideoElement + hls.js)
 *   - DesktopPlayer (Tauri commands → libmpv)
 *
 * Anything HLS.js-specific (LEVEL_SWITCHED, FRAG_BUFFERED) is exposed as
 * optional escape hatches via `videoElement()` / `hlsInstance()`. WatchPage's
 * quality-indicator + weak-network UI use those hatches and degrade gracefully
 * on desktop (where libmpv handles auto-quality and buffering internally).
 */

import type Hls from 'hls.js'

export type PlayerKind = 'web' | 'desktop'

export interface PlayerLoadOptions {
  url: string
  /** Resume position in seconds. */
  startTime?: number
  /** Authorization token; web puts it in the URL via existing build helpers,
   *  desktop forwards it via mpv `http-header-fields`. */
  authToken?: string
  /** Force-flag the URL as HLS so the WebPlayer routes through hls.js even
   *  when canPlayType('application/vnd.apple.mpegurl') is true. */
  isHls?: boolean
}

export interface PlayerSubtitleTrack {
  url: string
  label: string
  lang: string
}

export interface BufferRange {
  start: number
  end: number
}

export interface VideoSize {
  width: number
  height: number
}

export type PlayerEvent =
  | 'ready' // metadata loaded, duration is known
  | 'play'
  | 'pause'
  | 'ended'
  | 'error'
  | 'time-update'
  | 'buffer-update'
  | 'seeking'
  | 'seeked'
  | 'waiting'
  | 'playing'

export interface PlayerEventPayload {
  ready: { duration: number; videoSize: VideoSize | null }
  play: void
  pause: void
  ended: void
  error: { message: string; fatal: boolean }
  'time-update': { currentTime: number }
  'buffer-update': { ranges: BufferRange[] }
  seeking: void
  seeked: void
  waiting: void
  playing: void
}

export type PlayerEventCallback<E extends PlayerEvent> = (payload: PlayerEventPayload[E]) => void

export type Unsubscribe = () => void

export interface IPlayer {
  readonly kind: PlayerKind

  // ── Lifecycle ─────────────────────────────────────────────────────────────
  /** Attach to a host container. Web inserts a <video> child; desktop is a no-op
   *  (libmpv renders into its own NSOpenGLView mounted under the webview). */
  mount(container: HTMLElement): void
  unmount(): void

  load(opts: PlayerLoadOptions): Promise<void>
  /** Stop and release the current source. */
  unload(): Promise<void>
  destroy(): void

  // ── Playback ──────────────────────────────────────────────────────────────
  play(): Promise<void>
  pause(): void
  seek(seconds: number): void

  setVolume(v: number): void
  setMuted(m: boolean): void
  setRate(r: number): void

  // ── Synchronous getters (cheap; called from React render) ─────────────────
  currentTime(): number
  duration(): number
  paused(): boolean
  volume(): number
  muted(): boolean
  rate(): number
  buffered(): BufferRange[]
  videoSize(): VideoSize | null

  // ── Subtitles ─────────────────────────────────────────────────────────────
  /** Replace the current native subtitle. Web adds a <track>; desktop calls
   *  mpv `sub-add`. Pass null to remove. */
  setSubtitle(track: PlayerSubtitleTrack | null): void
  setSubtitleDelay(delayMs: number): void
  /** Whether the player's native subtitle render is shown.
   *  WatchPage uses DualSubtitleOverlay (React-rendered cues), so this is
   *  usually false — except for image subs (PGS/VobSub) where the overlay
   *  can't render and we let the player burn them. */
  setSubtitleVisible(visible: boolean): void

  // ── Audio tracks (multi-audio HLS) ────────────────────────────────────────
  /** Match by language and optional track name; web matches against
   *  hls.audioTracks, desktop calls mpv `aid` with track index. */
  setAudioTrackByLang(lang: string, name?: string): void

  // ── Events ────────────────────────────────────────────────────────────────
  on<E extends PlayerEvent>(event: E, cb: PlayerEventCallback<E>): Unsubscribe

  // ── Web-only escape hatches ───────────────────────────────────────────────
  /** Underlying HTMLVideoElement. Returns null on desktop. */
  videoElement(): HTMLVideoElement | null
  /** Underlying hls.js instance. Returns null when not using hls.js (native
   *  HLS playback, direct play, or desktop). */
  hlsInstance(): Hls | null
}
