import { useState } from 'react'
import { LuPlay, LuPencil, LuX } from 'react-icons/lu'
import { useScheduledTasks, useRunTask, useUpdateTaskInterval } from '@/hooks/stores/useAdmin'
import { Select } from '@/components/ui/Select'
import { Trans } from 'react-i18next'
import { useTranslation } from '@/hooks/useTranslation'
import { SectionHeader, Spinner, timeAgo } from './shared'

const getIntervalOptions = (t: any) => [
  { label: t('tasks.intervalOptions.30m'), value: '30m' },
  { label: t('tasks.intervalOptions.1h'), value: '1h' },
  { label: t('tasks.intervalOptions.12h'), value: '12h' },
  { label: t('tasks.intervalOptions.24h'), value: '24h' },
  { label: t('tasks.intervalOptions.168h'), value: '168h' },
  { label: t('tasks.intervalOptions.custom'), value: 'custom' },
]

function cleanInterval(val: string) {
  if (!val) return val
  return val.replace(/0m0s$/, '').replace(/0s$/, '')
}

function EditIntervalModal({
  taskName,
  currentInterval,
  onClose,
  onSave,
}: {
  taskName: string
  currentInterval: string
  onClose: () => void
  onSave: (val: string) => void
}) {
  const { t } = useTranslation('settings')
  const options = getIntervalOptions(t)
  const isPreset = options.some((o) => o.value === currentInterval)
  const [selectedPreset, setSelectedPreset] = useState(isPreset ? currentInterval : 'custom')
  const [customVal, setCustomVal] = useState(!isPreset ? currentInterval : '')

  const handleSave = () => {
    let finalVal = selectedPreset === 'custom' ? customVal : selectedPreset
    if (!finalVal) return
    onSave(finalVal)
  }

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 p-4 backdrop-blur-sm">
      <div className="w-full max-w-sm rounded-xl border border-netflix-gray/50 bg-netflix-dark p-6 shadow-2xl">
        <div className="flex items-center justify-between mb-4">
          <h3 className="text-lg font-semibold text-white">{t('tasks.editInterval')}</h3>
          <button onClick={onClose} className="text-gray-400 hover:text-white transition">
            <LuX size={20} />
          </button>
        </div>
        <p className="text-sm text-gray-400 mb-4">
          <Trans
            i18nKey="tasks.editIntervalDesc"
            t={t}
            values={{ taskName }}
            components={{ 1: <span className="font-semibold text-gray-200" /> }}
          />
        </p>

        <div className="mb-6 space-y-4">
          <Select
            value={selectedPreset}
            onChange={(e) => setSelectedPreset(e.target.value)}
            className="w-full"
          >
            {options.map((opt) => (
              <option key={opt.value} value={opt.value}>
                {opt.label}
              </option>
            ))}
          </Select>

          {selectedPreset === 'custom' && (
            <input
              type="text"
              value={customVal}
              onChange={(e) => setCustomVal(e.target.value)}
              placeholder={t('tasks.customPlaceholder')}
              className="w-full rounded-lg border border-netflix-gray bg-netflix-black px-4 py-2.5 text-white focus:border-blue-500 focus:outline-none focus:ring-1 focus:ring-blue-500"
              autoFocus
              onKeyDown={(e) => {
                if (e.key === 'Enter') handleSave()
                else if (e.key === 'Escape') onClose()
              }}
            />
          )}
        </div>

        <div className="flex justify-end gap-3">
          <button
            onClick={onClose}
            className="rounded-lg px-4 py-2 text-sm font-medium text-gray-300 hover:bg-netflix-gray/50 transition"
          >
            {t('actions.cancel')}
          </button>
          <button
            onClick={handleSave}
            disabled={selectedPreset === 'custom' && !customVal.trim()}
            className="rounded-lg bg-blue-600 px-4 py-2 text-sm font-medium text-white hover:bg-blue-500 transition disabled:opacity-50 disabled:cursor-not-allowed"
          >
            {t('actions.save')}
          </button>
        </div>
      </div>
    </div>
  )
}

export function TasksSection() {
  const { t } = useTranslation('settings')
  const [runningTask, setRunningTask] = useState<string | null>(null)
  const [editingTask, setEditingTask] = useState<{ name: string; interval: string } | null>(null)

  // Poll every 2s while any task is running to get live status updates
  const hasRunning = runningTask !== null
  const { data: tasks, isLoading } = useScheduledTasks(hasRunning)
  const { mutate: runTask } = useRunTask()
  const { mutate: updateInterval } = useUpdateTaskInterval()

  const handleRun = (name: string) => {
    setRunningTask(name)
    runTask(name)
  }

  const handleSaveInterval = (newVal: string) => {
    if (editingTask && newVal !== editingTask.interval) {
      updateInterval({ name: editingTask.name, interval: newVal })
    }
    setEditingTask(null)
  }

  // Clear local running state when server confirms task is done
  const serverTask = tasks?.find((t) => t.name === runningTask)
  if (runningTask && serverTask && !serverTask.running && serverTask.last_run) {
    setRunningTask(null)
  }

  // Sort tasks by name to prevent row jumping on re-render
  const sortedTasks = tasks?.slice().sort((a, b) => a.name.localeCompare(b.name))

  return (
    <div className="max-w-3xl">
      <SectionHeader title={t('sections.tasks.title')} description={t('sections.tasks.description')} />

      {editingTask && (
        <EditIntervalModal
          taskName={editingTask.name}
          currentInterval={editingTask.interval}
          onClose={() => setEditingTask(null)}
          onSave={handleSaveInterval}
        />
      )}

      {isLoading ? (
        <Spinner />
      ) : (
        <div className="mt-6 overflow-hidden rounded-xl bg-netflix-dark">
          <div className="overflow-x-auto">
            <table className="w-full">
              <thead className="border-b border-netflix-gray bg-netflix-black/50">
                <tr>
                  <th className="px-4 py-3 text-left text-xs font-medium text-gray-400">{t('tasks.table.task')}</th>
                  <th className="px-4 py-3 text-left text-xs font-medium text-gray-400">
                    {t('tasks.table.interval')}
                  </th>
                  <th className="px-4 py-3 text-left text-xs font-medium text-gray-400">
                    {t('tasks.table.lastRun')}
                  </th>
                  <th className="px-4 py-3 text-left text-xs font-medium text-gray-400">
                    {t('tasks.table.nextRun')}
                  </th>
                  <th className="px-4 py-3 text-left text-xs font-medium text-gray-400">{t('tasks.table.status')}</th>
                  <th className="px-4 py-3 text-left text-xs font-medium text-gray-400">{t('tasks.table.actions')}</th>
                </tr>
              </thead>
              <tbody>
                {sortedTasks?.map((task) => {
                  const cleanedInterval = cleanInterval(task.interval)
                  const isTaskRunning = task.running || runningTask === task.name
                  return (
                    <tr
                      key={task.name}
                      className="border-b border-netflix-gray/50 last:border-b-0 hover:bg-netflix-gray/30 transition-colors"
                    >
                      <td className="px-4 py-3 text-sm font-medium text-white">{task.name}</td>
                      <td className="px-4 py-3 text-sm text-gray-300">
                        <div className="flex items-center gap-2 group">
                          {cleanedInterval}
                          <button
                            onClick={() =>
                              setEditingTask({ name: task.name, interval: cleanedInterval })
                            }
                            className="text-gray-500 opacity-0 group-hover:opacity-100 hover:text-white transition-all outline-none"
                            title={t('tasks.action.editInterval')}
                          >
                            <LuPencil size={14} />
                          </button>
                        </div>
                      </td>
                      <td className="px-4 py-3 text-xs text-gray-400">
                        {task.last_run ? timeAgo(task.last_run) : t('tasks.never')}
                      </td>
                      <td className="px-4 py-3 text-xs text-gray-400">
                        {new Date(task.next_run).toLocaleString()}
                      </td>
                      <td className="px-4 py-3">
                        {isTaskRunning ? (
                          <span className="flex items-center gap-1.5 text-xs text-yellow-400">
                            <div className="h-3 w-3 animate-spin rounded-full border-2 border-yellow-400 border-t-transparent" />
                            {t('tasks.status.running')}
                          </span>
                        ) : (
                          <span className="text-xs text-gray-500">{t('tasks.status.idle')}</span>
                        )}
                      </td>
                      <td className="px-4 py-3">
                        <button
                          onClick={() => handleRun(task.name)}
                          disabled={isTaskRunning}
                          className="flex items-center gap-1.5 rounded bg-netflix-gray px-3 py-1.5 text-xs font-medium text-white transition-colors hover:bg-blue-600 disabled:opacity-50"
                        >
                          <LuPlay size={14} />
                          {isTaskRunning ? t('tasks.action.running') : t('tasks.action.runNow')}
                        </button>
                      </td>
                    </tr>
                  )
                })}
              </tbody>
            </table>
          </div>
          {(!tasks || tasks.length === 0) && (
            <p className="py-8 text-center text-sm text-gray-400">{t('tasks.noTasks')}</p>
          )}
        </div>
      )}
    </div>
  )
}
