/**
 * Shared query keys for media hooks — breaks require cycle between useProgress and useContinueWatching
 */

import type {
  ContinueWatchingParams,
  NextUpParams,
  FavoritesListParams,
  RecentlyWatchedParams,
} from '../../types'

// Continue Watching Keys
const CONTINUE_WATCHING_ALL = ['continueWatching'] as const
export const continueWatchingKeys = {
  all: CONTINUE_WATCHING_ALL,
  list: (params: ContinueWatchingParams) => [...CONTINUE_WATCHING_ALL, 'list', params] as const,
}

// Next Up Keys
const NEXT_UP_ALL = ['nextUp'] as const
export const nextUpKeys = {
  all: NEXT_UP_ALL,
  list: (params: NextUpParams) => [...NEXT_UP_ALL, 'list', params] as const,
}

// User Data Keys
const USER_DATA_ALL = ['userData'] as const
export const userDataKeys = {
  all: USER_DATA_ALL,
  progress: (mediaId: number) => [...USER_DATA_ALL, 'progress', mediaId] as const,
  favorites: (params: FavoritesListParams) => [...USER_DATA_ALL, 'favorites', params] as const,
  recentlyWatched: (params: RecentlyWatchedParams) =>
    [...USER_DATA_ALL, 'recentlyWatched', params] as const,
}
