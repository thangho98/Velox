// Generic API responses & filesystem browser

export interface ApiResponse<T> {
  data: T
}

export interface ApiErrorResponse {
  error: string
}

export interface FsDirEntry {
  name: string
  path: string
}

export interface FsBrowseResponse {
  current: string
  parent?: string
  dirs: FsDirEntry[]
}
