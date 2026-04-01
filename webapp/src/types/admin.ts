// Admin, Activity, Webhook, Task types

export interface ActivityLog {
  id: number
  user_id?: number
  username?: string
  action: string
  media_id?: number
  media_title?: string
  details: string
  ip_address: string
  created_at: string
}

export interface PlaybackStats {
  most_watched: Array<{ media_id: number; title: string; play_count: number }>
  most_active_users: Array<{ user_id: number; username: string; play_count: number }>
  plays_today: number
  plays_this_week: number
  plays_this_month: number
  total_plays: number
}

export interface ServerInfo {
  version: string
  uptime: string
  go_version: string
  os: string
  arch: string
  ffmpeg_version: string
  database: string
  hw_accel: string
  media_count: number
  series_count: number
  user_count: number
  total_size_bytes: number
}

export interface LibraryStatsItem {
  id: number
  name: string
  type: string
  item_count: number
  file_count: number
  total_size_bytes: number
  last_scanned?: string
}

export interface Webhook {
  id: number
  url: string
  events: string
  active: boolean
  created_at: string
  updated_at: string
}

export interface ScheduledTask {
  name: string
  interval: string
  last_run?: string
  next_run: string
  running: boolean
}
