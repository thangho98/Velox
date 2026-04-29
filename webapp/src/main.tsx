/**
 * Async bootstrap: pick platform adapter (web vs desktop) before any
 * module that calls getPlatform() runs. The desktop adapter needs to load
 * cached server URL/device name from disk before being usable.
 */

import { initPlatform } from '@velox/shared/platform'
import { loadPlatform, isTauri } from './platform'

async function bootstrap() {
  const adapter = await loadPlatform()
  initPlatform(adapter)

  // Tauri-only: if a previous session left a stale auth token in the
  // keychain but the server URL was never persisted (older builds), clear
  // the token so the user is forced through onboarding instead of landing
  // on a broken home page where every fetch hits Vite's `/api` proxy.
  if (isTauri()) {
    const { getServerUrl } = await import('./platform/desktop-adapter')
    const serverUrl = await getServerUrl()
    if (!serverUrl) {
      const platform = (await import('@velox/shared/platform')).getPlatform()
      await platform.secureStorage.removeItem('velox-auth').catch(() => {})
    }
  }

  const [{ setAuthStateGetter }, { useAuthStore }] = await Promise.all([
    import('@velox/shared/hooks/auth'),
    import('./stores/auth'),
  ])
  setAuthStateGetter(() => useAuthStore.getState())

  const [{ StrictMode }, { createRoot }, { App }] = await Promise.all([
    import('react'),
    import('react-dom/client'),
    import('./App'),
    import('./index.css'),
    import('./i18n'),
  ])

  createRoot(document.getElementById('root')!).render(
    <StrictMode>
      <App />
    </StrictMode>,
  )
}

bootstrap().catch((err) => {
  console.error('[Bootstrap] init failed:', err)
  document.body.innerHTML =
    '<pre style="color:#fff;background:#a00;padding:24px;font:14px monospace">' +
    'Failed to initialize Velox:\n\n' +
    String(err) +
    '</pre>'
})
