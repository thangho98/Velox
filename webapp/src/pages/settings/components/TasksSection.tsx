import { useState } from 'react'
import { LuPlay } from 'react-icons/lu'
import { useScheduledTasks, useRunTask } from '@/hooks/stores/useAdmin'
import { SectionHeader, Spinner, timeAgo } from './shared'

export function TasksSection() {
  const [runningTask, setRunningTask] = useState<string | null>(null)
  // Poll every 2s while any task is running to get live status updates
  const hasRunning = runningTask !== null
  const { data: tasks, isLoading } = useScheduledTasks(hasRunning)
  const { mutate: runTask } = useRunTask()

  const handleRun = (name: string) => {
    setRunningTask(name)
    runTask(name)
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
      <SectionHeader title="Scheduled Tasks" description="Background tasks and maintenance jobs" />

      {isLoading ? (
        <Spinner />
      ) : (
        <div className="mt-6 overflow-hidden rounded-xl bg-netflix-dark">
          <div className="overflow-x-auto">
            <table className="w-full">
              <thead className="border-b border-netflix-gray bg-netflix-black/50">
                <tr>
                  <th className="px-4 py-3 text-left text-xs font-medium text-gray-400">Task</th>
                  <th className="px-4 py-3 text-left text-xs font-medium text-gray-400">
                    Interval
                  </th>
                  <th className="px-4 py-3 text-left text-xs font-medium text-gray-400">
                    Last Run
                  </th>
                  <th className="px-4 py-3 text-left text-xs font-medium text-gray-400">
                    Next Run
                  </th>
                  <th className="px-4 py-3 text-left text-xs font-medium text-gray-400">Status</th>
                  <th className="px-4 py-3 text-left text-xs font-medium text-gray-400">Actions</th>
                </tr>
              </thead>
              <tbody>
                {sortedTasks?.map((task) => {
                  const isTaskRunning = task.running || runningTask === task.name
                  return (
                    <tr
                      key={task.name}
                      className="border-b border-netflix-gray/50 last:border-b-0 hover:bg-netflix-gray/30"
                    >
                      <td className="px-4 py-3 text-sm font-medium text-white">{task.name}</td>
                      <td className="px-4 py-3 text-sm text-gray-300">{task.interval}</td>
                      <td className="px-4 py-3 text-xs text-gray-400">
                        {task.last_run ? timeAgo(task.last_run) : 'Never'}
                      </td>
                      <td className="px-4 py-3 text-xs text-gray-400">
                        {new Date(task.next_run).toLocaleString()}
                      </td>
                      <td className="px-4 py-3">
                        {isTaskRunning ? (
                          <span className="flex items-center gap-1.5 text-xs text-yellow-400">
                            <div className="h-3 w-3 animate-spin rounded-full border-2 border-yellow-400 border-t-transparent" />
                            Running
                          </span>
                        ) : (
                          <span className="text-xs text-gray-500">Idle</span>
                        )}
                      </td>
                      <td className="px-4 py-3">
                        <button
                          onClick={() => handleRun(task.name)}
                          disabled={isTaskRunning}
                          className="flex items-center gap-1.5 rounded bg-netflix-gray px-3 py-1.5 text-xs text-white transition-colors hover:bg-blue-600 disabled:opacity-50"
                        >
                          <LuPlay size={12} />
                          {isTaskRunning ? 'Running...' : 'Run Now'}
                        </button>
                      </td>
                    </tr>
                  )
                })}
              </tbody>
            </table>
          </div>
          {(!tasks || tasks.length === 0) && (
            <p className="py-8 text-center text-sm text-gray-400">No scheduled tasks</p>
          )}
        </div>
      )}
    </div>
  )
}
