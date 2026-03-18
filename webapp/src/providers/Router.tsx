import { BrowserRouter, Routes, Route, Navigate, Outlet } from 'react-router'
import { useAuthStore } from '@/stores/auth'
import { useTokenRefresh } from '@/hooks/stores/useAuth'
import { Layout } from '@/components/Layout'
import { HomePage } from '@/pages/HomePage'
import { LoginPage } from '@/pages/LoginPage'
import { SetupPage } from '@/pages/SetupPage'
import { LibraryListPage } from '@/pages/LibraryListPage'
import { MediaDetailPage } from '@/pages/MediaDetailPage'
import { SeriesDetailPage } from '@/pages/SeriesDetailPage'
import { WatchPage } from '@/pages/WatchPage'
import { SettingsPage } from '@/pages/SettingsPage'
import { MoviesPage } from '@/pages/MoviesPage'
import { SeriesPage } from '@/pages/SeriesPage'
import { FavoritesPage } from '@/pages/FavoritesPage'
import { RecentlyWatchedPage } from '@/pages/RecentlyWatchedPage'
import { SearchPage } from '@/pages/SearchPage'
import { BrowsePage } from '@/pages/BrowsePage'

// Auth guard component - wraps content with Layout
function RequireAuth() {
  const { isAuthenticated } = useAuthStore()
  useTokenRefresh()

  if (!isAuthenticated) {
    return <Navigate to="/login" replace />
  }

  return (
    <Layout>
      <Outlet />
    </Layout>
  )
}

// Fullscreen player — no Layout wrapper at all
function RequireAuthFullScreen() {
  const { isAuthenticated } = useAuthStore()
  useTokenRefresh()

  if (!isAuthenticated) {
    return <Navigate to="/login" replace />
  }

  return <Outlet />
}

function NotFoundPage() {
  return (
    <div className="flex min-h-screen flex-col items-center justify-center bg-netflix-black text-center">
      <h1 className="mb-4 text-6xl font-bold text-netflix-red">404</h1>
      <p className="mb-8 text-xl text-gray-400">Page not found</p>
      <a href="/" className="text-netflix-blue hover:underline">
        Go back home
      </a>
    </div>
  )
}

export function RouterProvider() {
  return (
    <BrowserRouter>
      <Routes>
        {/* Public routes */}
        <Route path="/login" element={<LoginPage />} />
        <Route path="/setup" element={<SetupPage />} />

        {/* Protected routes with Layout */}
        <Route element={<RequireAuth />}>
          <Route path="/" element={<HomePage />} />
          <Route path="/movies" element={<MoviesPage />} />
          <Route path="/movies/:id" element={<MediaDetailPage />} />
          <Route path="/series" element={<SeriesPage />} />
          <Route path="/series/:seriesId" element={<SeriesDetailPage />} />
          <Route path="/favorites" element={<FavoritesPage />} />
          <Route path="/recently-watched" element={<RecentlyWatchedPage />} />
          <Route path="/libraries" element={<LibraryListPage />} />
          <Route path="/settings" element={<SettingsPage />} />
          <Route path="/search" element={<SearchPage />} />
          <Route path="/browse" element={<BrowsePage />} />
          {/* Redirects for old routes */}
          <Route path="/profile" element={<Navigate to="/settings?section=profile" replace />} />
          <Route
            path="/admin/libraries"
            element={<Navigate to="/settings?section=libraries" replace />}
          />
          <Route path="/admin/users" element={<Navigate to="/settings?section=users" replace />} />
          <Route
            path="/admin/settings"
            element={<Navigate to="/settings?section=subtitles" replace />}
          />
        </Route>

        {/* Fullscreen routes (player) */}
        <Route element={<RequireAuthFullScreen />}>
          <Route path="/watch/:id" element={<WatchPage />} />
        </Route>

        {/* 404 */}
        <Route path="*" element={<NotFoundPage />} />
      </Routes>
    </BrowserRouter>
  )
}
