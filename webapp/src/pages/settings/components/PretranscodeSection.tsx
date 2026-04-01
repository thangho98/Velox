import { useState } from 'react'
import { LuPlay, LuPause, LuSquare, LuTrash2 } from 'react-icons/lu'
import {
  usePretranscodeSettings,
  useUpdatePretranscodeSettings,
  usePretranscodeStatus,
  usePretranscodeProfiles,
  useTogglePretranscodeProfile,
  usePretranscodeEstimate,
  useStartPretranscode,
  useStopPretranscode,
  useResumePretranscode,
  useCleanupPretranscode,
} from '@/hooks/stores/useSettings'
import { useLibraries } from '@/hooks/stores/useMedia'
import { Toggle } from '@/components/ui/Toggle'
import { Select } from '@/components/ui/Select'
import { SectionHeader, Field, Modal, formatBytes } from './shared'

export function PretranscodeSection() {
  const { data: settings } = usePretranscodeSettings()
  const { data: status } = usePretranscodeStatus()
  const { data: profiles } = usePretranscodeProfiles()
  const { data: libraries } = useLibraries()
  const updateSettings = useUpdatePretranscodeSettings()
  const toggleProfile = useTogglePretranscodeProfile()
  const startEncode = useStartPretranscode()
  const stopEncode = useStopPretranscode()
  const resumeEncode = useResumePretranscode()
  const cleanupFiles = useCleanupPretranscode()
  const [selectedLibrary, setSelectedLibrary] = useState(0)
  const { data: estimate } = usePretranscodeEstimate(selectedLibrary)
  const [showCleanupConfirm, setShowCleanupConfirm] = useState(false)

  // Auto-select first library
  if (selectedLibrary === 0 && libraries && libraries.length > 0) {
    setSelectedLibrary(libraries[0].id)
  }

  const isEncoding = status && (status.encoding > 0 || status.queued > 0) && !status.paused
  const progressPercent =
    status && status.total > 0 ? Math.round((status.done / status.total) * 100) : 0

  return (
    <div className="space-y-8">
      <SectionHeader
        title="Pre-transcode"
        description="Encode media in advance for instant playback — no buffering, no waiting."
      />

      {/* Enable Toggle */}
      <div className="rounded-lg bg-netflix-black p-6 ring-1 ring-white/10">
        <div className="flex items-center justify-between">
          <div>
            <h3 className="text-lg font-semibold text-white">Offline Encoding</h3>
            <p className="text-sm text-gray-400">
              Pre-encode your library into browser-compatible H.264+AAC MP4 files. Like Netflix —
              instant playback, zero transcoding delay.
            </p>
          </div>
          <Toggle
            enabled={settings?.enabled ?? false}
            onChange={(v) => updateSettings.mutate({ enabled: v })}
          />
        </div>
      </div>

      {settings?.enabled && (
        <>
          {/* Quality Profiles */}
          <div className="rounded-lg bg-netflix-black p-6 ring-1 ring-white/10">
            <h3 className="mb-4 text-lg font-semibold text-white">Quality Profiles</h3>
            <div className="space-y-3">
              {profiles?.map((p) => (
                <label
                  key={p.id}
                  className="flex items-center gap-3 rounded-lg bg-netflix-gray p-3 transition-colors hover:bg-white/10"
                >
                  <input
                    type="checkbox"
                    checked={p.enabled}
                    onChange={() => toggleProfile.mutate({ id: p.id, enabled: !p.enabled })}
                    className="h-4 w-4 accent-netflix-red"
                  />
                  <div className="flex-1">
                    <span className="font-medium text-white">{p.name}</span>
                    <span className="ml-2 text-sm text-gray-400">
                      ({p.height}p, {p.video_bitrate / 1000}Mbps video + {p.audio_bitrate}kbps
                      audio)
                    </span>
                  </div>
                  <span className="text-xs text-gray-500">
                    ~{(((p.video_bitrate + p.audio_bitrate) * 5400) / 8 / 1024 / 1024).toFixed(1)}{' '}
                    GB/film
                  </span>
                </label>
              ))}
            </div>
          </div>

          {/* Schedule & Concurrency */}
          <div className="rounded-lg bg-netflix-black p-6 ring-1 ring-white/10">
            <h3 className="mb-4 text-lg font-semibold text-white">Schedule</h3>
            <div className="grid gap-4 sm:grid-cols-2">
              <Field label="Encode time">
                <Select
                  className="w-full"
                  value={settings.schedule || 'always'}
                  onChange={(e) => updateSettings.mutate({ schedule: e.target.value })}
                >
                  <option value="always">Always (fastest)</option>
                  <option value="night">Night only (00:00–06:00)</option>
                  <option value="idle">When idle (no one watching)</option>
                </Select>
              </Field>
              <Field label="Concurrent jobs">
                <Select
                  className="w-full"
                  value={settings.concurrency || '1'}
                  onChange={(e) => updateSettings.mutate({ concurrency: e.target.value })}
                >
                  <option value="1">1 (NAS-friendly)</option>
                  <option value="2">2</option>
                  <option value="3">3</option>
                  <option value="4">4 (powerful server)</option>
                </Select>
              </Field>
            </div>
          </div>

          {/* Storage Estimation */}
          <div className="rounded-lg bg-netflix-black p-6 ring-1 ring-white/10">
            <h3 className="mb-4 text-lg font-semibold text-white">Storage Estimation</h3>
            <Field label="Library">
              <Select
                className="w-full"
                value={selectedLibrary}
                onChange={(e) => setSelectedLibrary(Number(e.target.value))}
              >
                {libraries?.map((lib) => (
                  <option key={lib.id} value={lib.id}>
                    {lib.name}
                  </option>
                ))}
              </Select>
            </Field>
            {estimate && (
              <div className="mt-4 space-y-2">
                {estimate.profiles?.map((p) => (
                  <div key={p.profile_id} className="flex items-center justify-between text-sm">
                    <span className="text-gray-300">
                      {p.profile_name} ({p.file_count} files)
                    </span>
                    <span className="text-white font-medium">{p.estimated_gb.toFixed(1)} GB</span>
                  </div>
                ))}
                <div className="border-t border-white/10 pt-2 flex items-center justify-between text-sm font-semibold">
                  <span className="text-gray-300">Total</span>
                  <span className="text-white">{formatBytes(estimate.total_bytes)}</span>
                </div>
                <div className="flex items-center justify-between text-sm">
                  <span className="text-gray-400">Disk free</span>
                  <span
                    className={
                      estimate.disk_free_bytes > estimate.total_bytes
                        ? 'text-green-400'
                        : 'text-red-400'
                    }
                  >
                    {formatBytes(estimate.disk_free_bytes)}
                    {estimate.disk_free_bytes > estimate.total_bytes
                      ? ' ✓ Enough'
                      : ' ✗ Not enough'}
                  </span>
                </div>
              </div>
            )}
          </div>

          {/* Progress Dashboard */}
          {status && status.total > 0 && (
            <div className="rounded-lg bg-netflix-black p-6 ring-1 ring-white/10">
              <h3 className="mb-4 text-lg font-semibold text-white">Progress</h3>

              {/* Progress bar */}
              <div className="mb-3">
                <div className="flex items-center justify-between text-sm text-gray-400 mb-1">
                  <span>
                    {status.done}/{status.total} files
                  </span>
                  <span>{progressPercent}%</span>
                </div>
                <div className="h-3 w-full overflow-hidden rounded-full bg-gray-700">
                  <div
                    className="h-full rounded-full bg-netflix-red transition-all duration-500"
                    style={{ width: `${progressPercent}%` }}
                  />
                </div>
              </div>

              {/* Current file */}
              {status.current_file && (
                <div className="mb-3 text-sm">
                  <span className="text-gray-400">Encoding: </span>
                  <span className="text-white">{status.current_file}</span>
                  {status.speed && <span className="ml-2 text-gray-500">({status.speed})</span>}
                </div>
              )}

              {/* Stats */}
              <div className="flex gap-6 text-sm">
                <span className="text-green-400">✓ Done: {status.done}</span>
                {status.failed > 0 && (
                  <span className="text-red-400">✗ Failed: {status.failed}</span>
                )}
                <span className="text-gray-400">Queued: {status.queued}</span>
                <span className="text-gray-400">Disk: {formatBytes(status.disk_used)}</span>
              </div>

              {/* Controls */}
              <div className="mt-4 flex gap-3">
                {status.paused ? (
                  <button
                    onClick={() => resumeEncode.mutate()}
                    className="flex items-center gap-2 rounded-lg bg-green-600 px-4 py-2 text-sm font-medium text-white hover:bg-green-500"
                  >
                    <LuPlay size={16} /> Resume
                  </button>
                ) : isEncoding ? (
                  <button
                    onClick={() => stopEncode.mutate()}
                    className="flex items-center gap-2 rounded-lg bg-yellow-600 px-4 py-2 text-sm font-medium text-white hover:bg-yellow-500"
                  >
                    <LuPause size={16} /> Pause
                  </button>
                ) : null}
                <button
                  onClick={() => stopEncode.mutate()}
                  className="flex items-center gap-2 rounded-lg bg-gray-600 px-4 py-2 text-sm font-medium text-white hover:bg-gray-500"
                >
                  <LuSquare size={16} /> Stop
                </button>
              </div>
            </div>
          )}

          {/* Actions */}
          <div className="flex flex-wrap gap-3">
            {!isEncoding && (
              <button
                onClick={() => startEncode.mutate()}
                disabled={startEncode.isPending}
                className="rounded-lg bg-netflix-red px-6 py-2.5 text-sm font-semibold text-white transition-colors hover:bg-netflix-red-hover disabled:opacity-50"
              >
                {startEncode.isPending ? 'Starting...' : 'Start Encoding'}
              </button>
            )}
            <button
              onClick={() => setShowCleanupConfirm(true)}
              className="rounded-lg bg-gray-700 px-4 py-2.5 text-sm font-medium text-white hover:bg-gray-600"
            >
              <LuTrash2 className="mr-1.5 inline" size={14} />
              Delete All Pre-transcode Files
            </button>
          </div>

          {/* Cleanup confirmation */}
          {showCleanupConfirm && (
            <Modal
              title="Delete All Pre-transcode Files"
              onClose={() => setShowCleanupConfirm(false)}
            >
              <p className="mb-4 text-sm text-gray-300">
                This will permanently delete all pre-encoded files and disable pre-transcode. Your
                original media files are NOT affected.
              </p>
              <div className="flex justify-end gap-3">
                <button
                  onClick={() => setShowCleanupConfirm(false)}
                  className="rounded-lg bg-gray-700 px-4 py-2 text-sm text-white hover:bg-gray-600"
                >
                  Cancel
                </button>
                <button
                  onClick={() => {
                    cleanupFiles.mutate()
                    setShowCleanupConfirm(false)
                  }}
                  className="rounded-lg bg-red-600 px-4 py-2 text-sm font-medium text-white hover:bg-red-500"
                >
                  Delete All
                </button>
              </div>
            </Modal>
          )}
        </>
      )}
    </div>
  )
}
