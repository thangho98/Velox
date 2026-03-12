import { BrowserRouter, Routes, Route, Navigate, Outlet } from 'react-router'
import { useAuthStore } from '@/stores/auth'
import { useTokenRefresh } from '@/hooks/stores/useAuth'
import { Layout } from '@/components/Layout'
import { HomePage } from '@/pages/HomePage'
import { LoginPage } from '@/pages/LoginPage'
import { SetupPage } from '@/pages/SetupPage'
import { LibraryListPage } from '@/pages/LibraryListPage'
import { MediaDetailPage } from '@/pages/MediaDetailPage'
import { WatchPage } from '@/pages/WatchPage'
import { ProfilePage } from '@/pages/ProfilePage'
import { AdminUsersPage } from '@/pages/AdminUsersPage'
import { AdminLibrariesPage } from '@/pages/AdminLibrariesPage'
import { MoviesPage } from '@/pages/MoviesPage'
import { SeriesPage } from '@/pages/SeriesPage'
import { FavoritesPage } from '@/pages/FavoritesPage'
import { RecentlyWatchedPage } from '@/pages/RecentlyWatchedPage'
import { SearchPage } from '@/pages/SearchPage'

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

// Admin guard component
function RequireAdmin() {
  const { user, isAuthenticated } = useAuthStore()
  useTokenRefresh()

  if (!isAuthenticated) {
    return <Navigate to="/login" replace />
  }

  if (!user?.is_admin) {
    return <Navigate to="/" replace />
  }

  return (
    <Layout>
      <Outlet />
    </Layout>
  )
}

// Fullscreen player layout (no sidebar)
function RequireAuthFullScreen() {
  const { isAuthenticated } = useAuthStore()
  useTokenRefresh()

  if (!isAuthenticated) {
    return <Navigate to="/login" replace />
  }

  return (
    <Layout fullWidth>
      <Outlet />
    </Layout>
  )
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
          <Route path="/series" element={<SeriesPage />} />
          <Route path="/favorites" element={<FavoritesPage />} />
          <Route path="/recently-watched" element={<RecentlyWatchedPage />} />
          <Route path="/libraries" element={<LibraryListPage />} />
          <Route path="/media/:id" element={<MediaDetailPage />} />
          <Route path="/profile" element={<ProfilePage />} />
          <Route path="/settings" element={<ProfilePage />} />
          <Route path="/search" element={<SearchPage />} />
        </Route>

        {/* Fullscreen routes (player) */}
        <Route element={<RequireAuthFullScreen />}>
          <Route path="/watch/:id" element={<WatchPage />} />
        </Route>

        {/* Admin routes */}
        <Route element={<RequireAdmin />}>
          <Route path="/admin/libraries" element={<AdminLibrariesPage />} />
          <Route path="/admin/users" element={<AdminUsersPage />} />
        </Route>

        {/* 404 */}
        <Route path="*" element={<NotFoundPage />} />
      </Routes>
    </BrowserRouter>
  )
}
