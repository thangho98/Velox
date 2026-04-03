/**
 * TasksSection - Scheduled tasks list with run-now capability
 */

import { useState } from 'react'
import {
  View,
  Text,
  StyleSheet,
  TouchableOpacity,
  ActivityIndicator,
} from 'react-native'
import { useScheduledTasks, useRunTask } from '@velox/shared/hooks/admin'
import { CollapsibleSection } from './CollapsibleSection'
import { Toast } from '../Toast'

function timeAgo(dateStr: string): string {
  const date = new Date(dateStr)
  const now = new Date()
  const diffMs = now.getTime() - date.getTime()
  const diffMins = Math.floor(diffMs / 60000)

  if (diffMins < 1) return 'Just now'
  if (diffMins < 60) return `${diffMins}m ago`
  if (diffMins < 1440) return `${Math.floor(diffMins / 60)}h ago`
  if (diffMins < 10080) return `${Math.floor(diffMins / 1440)}d ago`
  return date.toLocaleDateString()
}

export function TasksSection() {
  const [runningTask, setRunningTask] = useState<string | null>(null)
  const hasRunning = runningTask !== null
  const { data: tasks, isLoading } = useScheduledTasks(hasRunning)
  const { mutate: runTask } = useRunTask()

  const handleRun = (name: string) => {
    setRunningTask(name)
    runTask(name, {
      onError: () => {
        setRunningTask(null)
        Toast.error(`Failed to run ${name}`)
      },
    })
  }

  // Clear local running state when server confirms task is done
  const serverTask = tasks?.find((t) => t.name === runningTask)
  if (runningTask && serverTask && !serverTask.running && serverTask.last_run) {
    setRunningTask(null)
  }

  const sortedTasks = tasks?.slice().sort((a, b) => a.name.localeCompare(b.name))
  const runningCount = tasks?.filter((t) => t.running).length ?? 0

  return (
    <CollapsibleSection
      title="Scheduled Tasks"
      subtitle={
        runningCount > 0
          ? `${runningCount} running`
          : `${tasks?.length || 0} tasks`
      }
    >
      {isLoading ? (
        <Text style={styles.loadingText}>Loading tasks...</Text>
      ) : !sortedTasks || sortedTasks.length === 0 ? (
        <Text style={styles.emptyText}>No scheduled tasks</Text>
      ) : (
        sortedTasks.map((task) => {
          const isTaskRunning = task.running || runningTask === task.name

          return (
            <View key={task.name} style={styles.taskItem}>
              <View style={styles.taskInfo}>
                <View style={styles.taskHeader}>
                  <Text style={styles.taskName}>{task.name}</Text>
                  {isTaskRunning ? (
                    <View style={styles.runningBadge}>
                      <ActivityIndicator size="small" color="#eab308" />
                      <Text style={styles.runningText}>Running</Text>
                    </View>
                  ) : (
                    <Text style={styles.idleText}>Idle</Text>
                  )}
                </View>
                <View style={styles.taskMeta}>
                  <Text style={styles.metaText}>Every {task.interval}</Text>
                  <Text style={styles.metaDot}> · </Text>
                  <Text style={styles.metaText}>
                    Last: {task.last_run ? timeAgo(task.last_run) : 'Never'}
                  </Text>
                </View>
                <Text style={styles.nextRunText}>
                  Next: {new Date(task.next_run).toLocaleString()}
                </Text>
              </View>

              <TouchableOpacity
                style={[styles.runBtn, isTaskRunning && styles.runBtnDisabled]}
                onPress={() => handleRun(task.name)}
                disabled={isTaskRunning}
              >
                {isTaskRunning ? (
                  <ActivityIndicator size="small" color="#888" />
                ) : (
                  <Text style={styles.runBtnText}>Run</Text>
                )}
              </TouchableOpacity>
            </View>
          )
        })
      )}
    </CollapsibleSection>
  )
}

const styles = StyleSheet.create({
  loadingText: {
    color: '#888',
    fontSize: 14,
  },
  emptyText: {
    color: '#666',
    fontSize: 14,
    textAlign: 'center',
    paddingVertical: 16,
  },
  taskItem: {
    flexDirection: 'row',
    alignItems: 'center',
    backgroundColor: '#2a2a2a',
    borderRadius: 8,
    padding: 12,
    marginBottom: 8,
  },
  taskInfo: {
    flex: 1,
    marginRight: 10,
  },
  taskHeader: {
    flexDirection: 'row',
    alignItems: 'center',
    justifyContent: 'space-between',
    marginBottom: 4,
  },
  taskName: {
    color: '#fff',
    fontSize: 14,
    fontWeight: '500',
    flex: 1,
  },
  runningBadge: {
    flexDirection: 'row',
    alignItems: 'center',
    gap: 4,
  },
  runningText: {
    color: '#eab308',
    fontSize: 11,
    fontWeight: '600',
  },
  idleText: {
    color: '#555',
    fontSize: 11,
  },
  taskMeta: {
    flexDirection: 'row',
    alignItems: 'center',
  },
  metaText: {
    color: '#888',
    fontSize: 12,
  },
  metaDot: {
    color: '#555',
    fontSize: 12,
  },
  nextRunText: {
    color: '#555',
    fontSize: 11,
    marginTop: 2,
  },
  runBtn: {
    backgroundColor: '#3a3a3a',
    borderRadius: 8,
    paddingHorizontal: 16,
    paddingVertical: 8,
    minWidth: 56,
    alignItems: 'center',
  },
  runBtnDisabled: {
    opacity: 0.5,
  },
  runBtnText: {
    color: '#fff',
    fontSize: 13,
    fontWeight: '600',
  },
})
