declare module 'react-native-google-cast' {
  export enum CastState {
    NO_DEVICES_AVAILABLE = 'NO_DEVICES_AVAILABLE',
    NOT_CONNECTED = 'NOT_CONNECTED',
    CONNECTED = 'CONNECTED',
  }

  export enum MediaPlayerState {
    PLAYING = 'PLAYING',
    PAUSED = 'PAUSED',
    BUFFERING = 'BUFFERING',
  }

  export interface CastImage {
    url: string
  }

  export interface CastMetadata {
    type?: string
    title?: string
    images?: CastImage[]
  }

  export interface LoadMediaRequest {
    mediaInfo: {
      contentUrl: string
      contentType: string
      metadata?: CastMetadata
      streamDuration?: number
    }
    startTime?: number
    autoplay?: boolean
  }

  export interface RemoteMediaClient {
    loadMedia(request: LoadMediaRequest): Promise<void>
    pause(): Promise<void>
    play(): Promise<void>
    seek(options: { position: number }): Promise<void>
    stop(): Promise<void>
  }

  export interface CastSession {
    device?: {
      friendlyName?: string | null
    }
  }

  export interface MediaStatus {
    playerState?: MediaPlayerState
    streamPosition?: number
    mediaInfo?: {
      streamDuration?: number
    }
  }

  export function useRemoteMediaClient(): RemoteMediaClient | null
  export function useCastState(): CastState
  export function useCastSession(): CastSession | null
  export function useMediaStatus(): MediaStatus | null

  const GoogleCast: {
    showCastDialog(): void
    sessionManager: {
      endCurrentSession(stopCasting: boolean): Promise<void>
    }
  }

  export default GoogleCast
}
