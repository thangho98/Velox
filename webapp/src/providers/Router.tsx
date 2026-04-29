import { lazy, Suspense } from 'react'
import { BrowserRouter, Routes, Route, Navigate, Outlet, useLocation } from 'react-router'
import { useAuthStore, useAuthHydrated } from '@/stores/auth'
import { useTokenRefresh, useWizardStatus } from '@/hooks/stores/useAuth'
import { Layout } from '@/components/Layout'
import { ErrorBoundary } from '@/components/ErrorBoundary'
import { isTauri } from '@/platform'
import { isPaired } from '@/platform/desktop-adapter'
import { useDesktopBridge } from '@/lib/desktop/useDesktopBridge'

const UNAUTH_REDIRECT = isTauri() ? '/onboarding' : '/login'

const HomePage = lazy(() => import('@/pages/HomePage'))
const LoginPage = lazy(() => import('@/pages/LoginPage'))
const OnboardingPage = lazy(() => import('@/pages/OnboardingPage'))
const SetupPage = lazy(() => import('@/pages/SetupPage'))
const SetupWizardPage = lazy(() => import('@/pages/SetupWizardPage'))
const LibraryListPage = lazy(() => import('@/pages/LibraryListPage'))
const MediaDetailPage = lazy(() => import('@/pages/MediaDetailPage'))
const SeriesDetailPage = lazy(() => import('@/pages/SeriesDetailPage'))
const WatchPage = lazy(() => import('@/pages/WatchPage'))
const SettingsPage = lazy(() => import('@/pages/settings'))
const MoviesPage = lazy(() => import('@/pages/MoviesPage'))
const SeriesPage = lazy(() => import('@/pages/SeriesPage'))
const FavoritesPage = lazy(() => import('@/pages/FavoritesPage'))
const RecentlyWatchedPage = lazy(() => import('@/pages/RecentlyWatchedPage'))
const SearchPage = lazy(() => import('@/pages/SearchPage'))
const BrowsePage = lazy(() => import('@/pages/BrowsePage'))
const LiveTvPage = lazy(() => import('@/pages/LiveTvPage'))
const WatchLivePage = lazy(() => import('@/pages/WatchLivePage'))

// Auth guard component - wraps content with Layout
function RequireAuth() {
  const { isAuthenticated, user } = useAuthStore()
  useTokenRefresh()

  const { data: wizardStatus, isLoading: wizardLoading } = useWizardStatus()

  if (!isAuthenticated) {
    return <Navigate to={UNAUTH_REDIRECT} replace />
  }

  // Redirect admin to wizard if not completed
  if (!wizardLoading && user?.is_admin && wizardStatus && !wizardStatus.completed) {
    return <Navigate to="/setup/wizard" replace />
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
    return <Navigate to={UNAUTH_REDIRECT} replace />
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

function PageLoader() {
  return (
    <div className="flex min-h-screen items-center justify-center bg-netflix-black">
      <div className="h-8 w-8 animate-spin rounded-full border-2 border-gray-600 border-t-white" />
    </div>
  )
}

function HydrationGate({ children }: { children: React.ReactNode }) {
  const hydrated = useAuthHydrated()
  const location = useLocation()
  // Mount Tauri OS bridges (menu, deep-link, drag-drop). No-op on web.
  useDesktopBridge()
  if (!hydrated) return <PageLoader />
  // Tauri: if no server URL has been paired yet, force onboarding even if a
  // stale auth token exists — every API call would otherwise hit `/api`
  // (relative) and fail through Vite's proxy.
  if (isTauri() && !isPaired() && location.pathname !== '/onboarding') {
    return <Navigate to="/onboarding" replace />
  }
  return <>{children}</>
}

export function RouterProvider() {
  return (
    <BrowserRouter>
      <ErrorBoundary>
        <HydrationGate>
          <Suspense fallback={<PageLoader />}>
            <Routes>
              {/* Public routes */}
              <Route path="/login" element={<LoginPage />} />
              <Route path="/onboarding" element={<OnboardingPage />} />
              <Route path="/setup" element={<SetupPage />} />

              {/* Setup wizard (authenticated, no Layout) */}
              <Route element={<RequireAuthFullScreen />}>
                <Route path="/setup/wizard" element={<SetupWizardPage />} />
              </Route>

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
                <Route path="/livetv" element={<LiveTvPage />} />
                {/* Redirects for old/missing routes */}
                <Route
                  path="/notifications"
                  element={<Navigate to="/settings?section=activity" replace />}
                />
                <Route
                  path="/profile"
                  element={<Navigate to="/settings?section=profile" replace />}
                />
                <Route
                  path="/admin/libraries"
                  element={<Navigate to="/settings?section=libraries" replace />}
                />
                <Route
                  path="/admin/users"
                  element={<Navigate to="/settings?section=users" replace />}
                />
                <Route
                  path="/admin/settings"
                  element={<Navigate to="/settings?section=subtitles" replace />}
                />
              </Route>

              {/* Fullscreen routes (player) */}
              <Route element={<RequireAuthFullScreen />}>
                <Route path="/watch/:id" element={<WatchPage />} />
                <Route path="/livetv/watch/:channelId" element={<WatchLivePage />} />
              </Route>

              {/* 404 */}
              <Route path="*" element={<NotFoundPage />} />
            </Routes>
          </Suspense>
        </HydrationGate>
      </ErrorBoundary>
    </BrowserRouter>
  )
}
