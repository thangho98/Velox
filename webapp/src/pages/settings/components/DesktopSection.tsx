/**
 * DesktopSection — Tauri-only settings: libmpv/ffmpeg versions, log/config
 * folder shortcuts, server URL re-pair. Visible only when running inside the
 * desktop shell (isTauri()).
 */

import { useEffect, useState } from 'react'
import { LuExternalLink, LuRefreshCw, LuLogOut, LuDownload } from 'react-icons/lu'
import { invoke } from '@tauri-apps/api/core'
import { openPath } from '@tauri-apps/plugin-opener'
import { useAuthStore } from '@/stores/auth'
import { useLogout } from '@velox/shared/hooks/auth'
import { SectionHeader, Field } from './shared'

type VersionInfo = {
  mpv: string
  ffmpeg: string
  placebo: string | null
  hwdec_codecs: string
}

export function DesktopSection() {
  const [versions, setVersions] = useState<VersionInfo | null>(null)
  const [versionError, setVersionError] = useState<string | null>(null)
  const [appVersion, setAppVersion] = useState<string>('')
  const [updateStatus, setUpdateStatus] = useState<string>('')
  const [checkingUpdate, setCheckingUpdate] = useState(false)
  const { user } = useAuthStore()
  const { mutate: logout } = useLogout()

  useEffect(() => {
    void invoke<VersionInfo>('get_versions')
      .then(setVersions)
      .catch((e) => {
        setVersionError(String(e))
      })
    void import('@tauri-apps/api/app').then(({ getVersion }) => {
      void getVersion()
        .then(setAppVersion)
        .catch(() => {})
    })
  }, [])

  const handleSwitchServer = () => {
    if (
      window.confirm(
        'Sign out and switch to a different Velox server? Your local preferences will be kept.',
      )
    ) {
      logout()
    }
  }

  const openConfigFolder = async () => {
    try {
      const { appConfigDir } = await import('@tauri-apps/api/path')
      const dir = await appConfigDir()
      await openPath(dir)
    } catch (err) {
      window.alert(`Could not open config folder: ${err}`)
    }
  }

  const openLogFolder = async () => {
    try {
      const { appLogDir } = await import('@tauri-apps/api/path')
      const dir = await appLogDir()
      await openPath(dir)
    } catch (err) {
      window.alert(`Could not open log folder: ${err}`)
    }
  }

  const checkForUpdates = async () => {
    setCheckingUpdate(true)
    setUpdateStatus('Checking…')
    try {
      const { check } = await import('@tauri-apps/plugin-updater')
      const update = await check()
      if (!update) {
        setUpdateStatus('You are on the latest version.')
        return
      }
      setUpdateStatus(`Update ${update.version} available — downloading…`)
      await update.downloadAndInstall()
      setUpdateStatus('Installed. Relaunch to apply.')
    } catch (err) {
      setUpdateStatus(`Update failed: ${err}`)
    } finally {
      setCheckingUpdate(false)
    }
  }

  return (
    <div className="space-y-6">
      <SectionHeader
        title="Desktop App"
        description="libmpv-powered native player. These options apply only when running the Velox desktop app."
      />

      <div className="rounded-lg border border-netflix-gray/40 bg-netflix-dark p-5">
        <h3 className="mb-3 text-sm font-semibold uppercase tracking-wider text-gray-400">
          Version
        </h3>
        <div className="grid grid-cols-1 gap-3 text-sm sm:grid-cols-2">
          <Field label="App">
            <div className="text-white">Velox Desktop {appVersion || '0.1.0'}</div>
          </Field>
          <Field label="libmpv">
            <div className="text-white">{versions?.mpv ?? '—'}</div>
          </Field>
          <Field label="FFmpeg">
            <div className="text-white">{versions?.ffmpeg ?? '—'}</div>
          </Field>
          <Field label="libplacebo">
            <div className="text-white">{versions?.placebo ?? 'unavailable'}</div>
          </Field>
          <Field label="HW decoders">
            <div className="break-all text-white">{versions?.hwdec_codecs || '—'}</div>
          </Field>
        </div>
        {versionError && (
          <p className="mt-3 text-xs text-netflix-red">Version probe failed: {versionError}</p>
        )}
      </div>

      <div className="rounded-lg border border-netflix-gray/40 bg-netflix-dark p-5">
        <h3 className="mb-3 text-sm font-semibold uppercase tracking-wider text-gray-400">
          Server
        </h3>
        <Field label="Signed in as">
          <div className="flex items-center justify-between gap-4">
            <div>
              <div className="text-white">{user?.username ?? '—'}</div>
              <div className="text-xs text-gray-500">
                {user?.is_admin ? 'Administrator' : 'Standard user'}
              </div>
            </div>
            <button
              onClick={handleSwitchServer}
              className="inline-flex items-center gap-2 rounded bg-netflix-gray/60 px-3 py-2 text-xs font-medium text-white transition-colors hover:bg-netflix-gray"
            >
              <LuLogOut size={14} /> Switch server / sign out
            </button>
          </div>
        </Field>
      </div>

      <div className="rounded-lg border border-netflix-gray/40 bg-netflix-dark p-5">
        <h3 className="mb-3 text-sm font-semibold uppercase tracking-wider text-gray-400">Files</h3>
        <div className="flex flex-wrap gap-3">
          <button
            onClick={openConfigFolder}
            className="inline-flex items-center gap-2 rounded bg-netflix-gray/60 px-4 py-2 text-xs font-medium text-white transition-colors hover:bg-netflix-gray"
          >
            <LuExternalLink size={14} /> Open config folder
          </button>
          <button
            onClick={openLogFolder}
            className="inline-flex items-center gap-2 rounded bg-netflix-gray/60 px-4 py-2 text-xs font-medium text-white transition-colors hover:bg-netflix-gray"
          >
            <LuExternalLink size={14} /> Open log folder
          </button>
          <button
            onClick={() => window.location.reload()}
            className="inline-flex items-center gap-2 rounded bg-netflix-gray/60 px-4 py-2 text-xs font-medium text-white transition-colors hover:bg-netflix-gray"
          >
            <LuRefreshCw size={14} /> Reload UI
          </button>
        </div>
      </div>

      <div className="rounded-lg border border-netflix-gray/40 bg-netflix-dark p-5">
        <h3 className="mb-3 text-sm font-semibold uppercase tracking-wider text-gray-400">
          Updates
        </h3>
        <div className="flex flex-wrap items-center gap-3">
          <button
            onClick={checkForUpdates}
            disabled={checkingUpdate}
            className="inline-flex items-center gap-2 rounded bg-netflix-gray/60 px-4 py-2 text-xs font-medium text-white transition-colors hover:bg-netflix-gray disabled:opacity-50"
          >
            <LuDownload size={14} /> {checkingUpdate ? 'Checking…' : 'Check for updates'}
          </button>
          {updateStatus && <span className="text-xs text-gray-400">{updateStatus}</span>}
        </div>
      </div>
    </div>
  )
}
