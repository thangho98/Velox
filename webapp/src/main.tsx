/**
 * Initialize platform adapter BEFORE any other imports
 * This ensures getPlatform() is available for all modules
 */
import { initPlatform } from '@velox/shared/platform'
import { webPlatform } from './platform/web-adapter'
initPlatform(webPlatform)

import { setAuthStateGetter } from '@velox/shared/hooks/auth'
import { useAuthStore } from './stores/auth'
setAuthStateGetter(() => useAuthStore.getState())

import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import './index.css'
import './i18n' // Initialize i18n before App
import { App } from './App'

createRoot(document.getElementById('root')!).render(
  <StrictMode>
    <App />
  </StrictMode>,
)
