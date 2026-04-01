import { useServerInfo, useLibraryStats } from '@/hooks/stores/useAdmin'
import { SectionHeader, Spinner, formatBytes, timeAgo } from './shared'

export function GeneralSection() {
  const { data: serverInfo, isLoading: serverLoading } = useServerInfo()
  const { data: libraryStats, isLoading: statsLoading } = useLibraryStats()

  if (serverLoading || statsLoading) return <Spinner />

  return (
    <div className="max-w-3xl space-y-6">
      <SectionHeader title="Dashboard" description="Server information and status" />

      {/* Stats Cards */}
      <div className="grid grid-cols-2 gap-4 lg:grid-cols-4">
        <div className="rounded-lg bg-netflix-dark p-4">
          <p className="text-xs text-gray-400">Total Media</p>
          <p className="mt-1 text-2xl font-bold text-white">{serverInfo?.media_count ?? 0}</p>
        </div>
        <div className="rounded-lg bg-netflix-dark p-4">
          <p className="text-xs text-gray-400">Series</p>
          <p className="mt-1 text-2xl font-bold text-white">{serverInfo?.series_count ?? 0}</p>
        </div>
        <div className="rounded-lg bg-netflix-dark p-4">
          <p className="text-xs text-gray-400">Users</p>
          <p className="mt-1 text-2xl font-bold text-white">{serverInfo?.user_count ?? 0}</p>
        </div>
        <div className="rounded-lg bg-netflix-dark p-4">
          <p className="text-xs text-gray-400">Total Size</p>
          <p className="mt-1 text-2xl font-bold text-white">
            {formatBytes(serverInfo?.total_size_bytes ?? 0)}
          </p>
        </div>
      </div>

      {/* Server Info */}
      <div className="rounded-lg bg-netflix-dark p-5">
        <h3 className="mb-3 text-sm font-semibold text-white">Server Information</h3>
        <div className="space-y-0">
          <InfoRow label="Version" value={serverInfo?.version ?? 'Unknown'} />
          <InfoRow label="Uptime" value={serverInfo?.uptime ?? 'Unknown'} />
          <InfoRow label="Go Version" value={serverInfo?.go_version ?? 'Unknown'} />
          <InfoRow
            label="OS / Arch"
            value={`${serverInfo?.os ?? '?'} / ${serverInfo?.arch ?? '?'}`}
          />
          <InfoRow label="FFmpeg" value={serverInfo?.ffmpeg_version ?? 'Unknown'} />
          <InfoRow label="HW Acceleration" value={serverInfo?.hw_accel || 'None'} />
          <InfoRow label="Database" value={serverInfo?.database ?? 'SQLite'} />
        </div>
      </div>

      {/* Library Stats */}
      {libraryStats && libraryStats.length > 0 && (
        <div className="overflow-hidden rounded-lg bg-netflix-dark">
          <div className="px-5 pt-5 pb-3">
            <h3 className="text-sm font-semibold text-white">Library Statistics</h3>
          </div>
          <div className="overflow-x-auto">
            <table className="w-full">
              <thead className="border-b border-netflix-gray bg-netflix-black/50">
                <tr>
                  <th className="px-5 py-2.5 text-left text-xs font-medium text-gray-400">
                    Library
                  </th>
                  <th className="px-5 py-2.5 text-left text-xs font-medium text-gray-400">Type</th>
                  <th className="px-5 py-2.5 text-right text-xs font-medium text-gray-400">
                    Items
                  </th>
                  <th className="px-5 py-2.5 text-right text-xs font-medium text-gray-400">
                    Files
                  </th>
                  <th className="px-5 py-2.5 text-right text-xs font-medium text-gray-400">Size</th>
                  <th className="px-5 py-2.5 text-right text-xs font-medium text-gray-400">
                    Last Scanned
                  </th>
                </tr>
              </thead>
              <tbody>
                {libraryStats.map((lib) => (
                  <tr
                    key={lib.id}
                    className="border-b border-netflix-gray/50 last:border-b-0 hover:bg-netflix-gray/30"
                  >
                    <td className="px-5 py-3 text-sm font-medium text-white">{lib.name}</td>
                    <td className="px-5 py-3">
                      <span
                        className={`rounded px-2 py-0.5 text-xs font-medium ${
                          lib.type === 'movies'
                            ? 'bg-blue-500/20 text-blue-400'
                            : lib.type === 'tvshows'
                              ? 'bg-purple-500/20 text-purple-400'
                              : 'bg-green-500/20 text-green-400'
                        }`}
                      >
                        {lib.type}
                      </span>
                    </td>
                    <td className="px-5 py-3 text-right text-sm text-gray-300">{lib.item_count}</td>
                    <td className="px-5 py-3 text-right text-sm text-gray-300">{lib.file_count}</td>
                    <td className="px-5 py-3 text-right text-sm text-gray-300">
                      {formatBytes(lib.total_size_bytes)}
                    </td>
                    <td className="px-5 py-3 text-right text-sm text-gray-400">
                      {lib.last_scanned ? timeAgo(lib.last_scanned) : 'Never'}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>
      )}
    </div>
  )
}

function InfoRow({ label, value }: { label: string; value: string }) {
  return (
    <div className="flex items-center justify-between gap-4 border-b border-netflix-gray/30 py-3 last:border-b-0">
      <span className="shrink-0 text-sm text-gray-400">{label}</span>
      <span className="truncate text-sm font-medium text-white">{value}</span>
    </div>
  )
}
