import { useState, useEffect } from 'react'
import { useQueryClient } from '@tanstack/react-query'
import { useActivity } from '@/hooks/stores/useAdmin'
import { Select } from '@/components/ui/Select'
import { useTranslation } from '@/hooks/useTranslation'
import { SectionHeader, Spinner, inputClass } from './shared'

const ACTION_BADGES: Record<string, string> = {
  login: 'bg-blue-500/20 text-blue-400',
  play_start: 'bg-green-500/20 text-green-400',
  play_stop: 'bg-yellow-500/20 text-yellow-400',
  library_scan: 'bg-purple-500/20 text-purple-400',
  media_added: 'bg-teal-500/20 text-teal-400',
}

export function ActivitySection() {
  const { t } = useTranslation('settings')
  const queryClient = useQueryClient()
  const [action, setAction] = useState('')
  const [dateFrom, setDateFrom] = useState('')
  const [dateTo, setDateTo] = useState('')
  const [limit, setLimit] = useState('25')

  // Force fresh data every time section becomes visible
  useEffect(() => {
    queryClient.invalidateQueries({ queryKey: ['admin', 'activity'] })
  }, [queryClient])

  const filters: Record<string, string> = {}
  if (action) filters.action = action
  if (dateFrom) filters.from = dateFrom
  if (dateTo) filters.to = dateTo
  if (limit) filters.limit = limit

  const { data: logs, isLoading } = useActivity(filters)

  return (
    <div className="max-w-4xl">
      <SectionHeader title={t('sections.activity.title')} description={t('sections.activity.description')} />

      {/* Filters */}
      <div className="mt-6 flex flex-wrap items-end gap-3">
        <div>
          <label className="mb-1 block text-xs text-gray-400">{t('activity.action')}</label>
          <Select
            value={action}
            onChange={(e) => setAction(e.target.value)}
            className="min-w-[140px]"
          >
            <option value="">{t('activity.allActions')}</option>
            <option value="login">{t('activity.actions.login')}</option>
            <option value="play_start">{t('activity.actions.play_start')}</option>
            <option value="play_stop">{t('activity.actions.play_stop')}</option>
            <option value="library_scan">{t('activity.actions.library_scan')}</option>
            <option value="media_added">{t('activity.actions.media_added')}</option>
          </Select>
        </div>
        <div>
          <label className="mb-1 block text-xs text-gray-400">{t('activity.from')}</label>
          <input
            type="date"
            value={dateFrom}
            onChange={(e) => setDateFrom(e.target.value)}
            className={inputClass + ' !w-auto'}
          />
        </div>
        <div>
          <label className="mb-1 block text-xs text-gray-400">{t('activity.to')}</label>
          <input
            type="date"
            value={dateTo}
            onChange={(e) => setDateTo(e.target.value)}
            className={inputClass + ' !w-auto'}
          />
        </div>
        <div>
          <label className="mb-1 block text-xs text-gray-400">{t('activity.limit')}</label>
          <Select value={limit} onChange={(e) => setLimit(e.target.value)} className="min-w-[80px]">
            <option value="25">25</option>
            <option value="50">50</option>
            <option value="100">100</option>
          </Select>
        </div>
      </div>

      {/* Activity Table */}
      {isLoading ? (
        <Spinner />
      ) : (
        <div className="mt-4 overflow-hidden rounded-xl bg-netflix-dark">
          <div className="overflow-x-auto">
            <table className="w-full">
              <thead className="border-b border-netflix-gray bg-netflix-black/50">
                <tr>
                  <th className="px-4 py-3 text-left text-xs font-medium text-gray-400">{t('activity.table.time')}</th>
                  <th className="px-4 py-3 text-left text-xs font-medium text-gray-400">{t('activity.table.user')}</th>
                  <th className="px-4 py-3 text-left text-xs font-medium text-gray-400">{t('activity.table.action')}</th>
                  <th className="px-4 py-3 text-left text-xs font-medium text-gray-400">{t('activity.table.media')}</th>
                  <th className="px-4 py-3 text-left text-xs font-medium text-gray-400">{t('activity.table.ip')}</th>
                </tr>
              </thead>
              <tbody>
                {logs?.map((log) => (
                  <tr
                    key={log.id}
                    className="border-b border-netflix-gray/50 last:border-b-0 hover:bg-netflix-gray/30"
                  >
                    <td className="whitespace-nowrap px-4 py-3 text-xs text-gray-400">
                      {new Date(log.created_at).toLocaleString()}
                    </td>
                    <td className="px-4 py-3 text-sm text-white">{log.username ?? t('activity.system')}</td>
                    <td className="px-4 py-3">
                      <span
                        className={`rounded px-2 py-0.5 text-xs font-medium ${ACTION_BADGES[log.action] ?? 'bg-gray-500/20 text-gray-400'}`}
                      >
                        {t(`activity.actions.${log.action}`, { defaultValue: log.action })}
                      </span>
                    </td>
                    <td className="max-w-[200px] truncate px-4 py-3 text-sm text-gray-300">
                      {log.media_title ?? '-'}
                    </td>
                    <td className="whitespace-nowrap px-4 py-3 font-mono text-xs text-gray-500">
                      {log.ip_address}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
          {(!logs || logs.length === 0) && (
            <p className="py-8 text-center text-sm text-gray-400">{t('activity.noActivity')}</p>
          )}
        </div>
      )}
    </div>
  )
}
