/**
 * WebhooksSection - CRUD for webhook management
 */

import { useState } from 'react'
import {
  View,
  Text,
  StyleSheet,
  TouchableOpacity,
  TextInput,
  Switch,
  Alert,
  ActivityIndicator,
} from 'react-native'
import {
  useWebhooks,
  useCreateWebhook,
  useUpdateWebhook,
  useDeleteWebhook,
} from '@velox/shared/hooks/admin'
import type { Webhook } from '@velox/shared/types'
import { CollapsibleSection } from './CollapsibleSection'
import { Toast } from '../Toast'

const WEBHOOK_EVENTS = [
  'scan_complete',
  'transcode_complete',
  'transcode_failed',
  'media_added',
  'subtitle_downloaded',
]

function parseWebhookEvents(raw: string): string[] {
  try {
    const parsed = JSON.parse(raw)
    if (Array.isArray(parsed)) return parsed as string[]
  } catch {
    if (raw && raw !== '[]')
      return raw
        .split(',')
        .map((s) => s.trim())
        .filter(Boolean)
  }
  return []
}

export function WebhooksSection() {
  const { data: webhooks, isLoading } = useWebhooks()
  const { mutate: createWebhook, isPending: isCreating } = useCreateWebhook()
  const { mutate: updateWebhook } = useUpdateWebhook()
  const { mutate: deleteWebhook } = useDeleteWebhook()

  const [showForm, setShowForm] = useState(false)
  const [formUrl, setFormUrl] = useState('')
  const [formEvents, setFormEvents] = useState<string[]>([])

  const toggleEvent = (event: string) => {
    setFormEvents((prev) =>
      prev.includes(event) ? prev.filter((e) => e !== event) : [...prev, event],
    )
  }

  const handleCreate = () => {
    if (!formUrl.trim()) {
      Toast.error('URL is required')
      return
    }
    if (formEvents.length === 0) {
      Toast.error('Select at least one event')
      return
    }
    createWebhook(
      { url: formUrl.trim(), events: JSON.stringify(formEvents), active: true },
      {
        onSuccess: () => {
          setShowForm(false)
          setFormUrl('')
          setFormEvents([])
          Toast.success('Webhook created')
        },
        onError: () => Toast.error('Failed to create webhook'),
      },
    )
  }

  const handleToggleActive = (webhook: Webhook) => {
    updateWebhook(
      { id: webhook.id, data: { active: !webhook.active } },
      {
        onSuccess: () =>
          Toast.success(webhook.active ? 'Webhook deactivated' : 'Webhook activated'),
      },
    )
  }

  const handleDelete = (id: number) => {
    Alert.alert('Delete Webhook', 'Are you sure you want to delete this webhook?', [
      { text: 'Cancel', style: 'cancel' },
      {
        text: 'Delete',
        style: 'destructive',
        onPress: () =>
          deleteWebhook(id, {
            onSuccess: () => Toast.success('Webhook deleted'),
            onError: () => Toast.error('Failed to delete'),
          }),
      },
    ])
  }

  return (
    <CollapsibleSection
      title="Webhooks"
      subtitle={`${webhooks?.length || 0} configured`}
    >
      {isLoading ? (
        <Text style={styles.loadingText}>Loading webhooks...</Text>
      ) : (
        <>
          {/* Existing Webhooks */}
          {webhooks?.length === 0 ? (
            <Text style={styles.emptyText}>No webhooks configured</Text>
          ) : (
            webhooks?.map((wh) => (
              <View key={wh.id} style={styles.webhookItem}>
                <View style={styles.webhookInfo}>
                  <Text style={styles.webhookUrl} numberOfLines={1}>
                    {wh.url}
                  </Text>
                  <View style={styles.eventsRow}>
                    {parseWebhookEvents(wh.events).map((ev) => (
                      <View key={ev} style={styles.eventTag}>
                        <Text style={styles.eventTagText}>{ev}</Text>
                      </View>
                    ))}
                  </View>
                </View>
                <View style={styles.webhookActions}>
                  <TouchableOpacity
                    style={[
                      styles.statusBtn,
                      wh.active ? styles.statusBtnActive : styles.statusBtnInactive,
                    ]}
                    onPress={() => handleToggleActive(wh)}
                  >
                    <Text
                      style={[
                        styles.statusBtnText,
                        wh.active ? styles.statusTextActive : styles.statusTextInactive,
                      ]}
                    >
                      {wh.active ? 'Active' : 'Off'}
                    </Text>
                  </TouchableOpacity>
                  <TouchableOpacity
                    style={styles.deleteBtn}
                    onPress={() => handleDelete(wh.id)}
                  >
                    <Text style={styles.deleteBtnText}>X</Text>
                  </TouchableOpacity>
                </View>
              </View>
            ))
          )}

          {/* Add Webhook Form */}
          {showForm ? (
            <View style={styles.formCard}>
              <Text style={styles.formTitle}>New Webhook</Text>
              <TextInput
                style={styles.input}
                value={formUrl}
                onChangeText={setFormUrl}
                placeholder="https://example.com/webhook"
                placeholderTextColor="#666"
                autoCapitalize="none"
                autoCorrect={false}
                keyboardType="url"
              />

              <Text style={styles.formLabel}>Events</Text>
              {WEBHOOK_EVENTS.map((event) => (
                <TouchableOpacity
                  key={event}
                  style={styles.checkboxRow}
                  onPress={() => toggleEvent(event)}
                >
                  <View
                    style={[
                      styles.checkbox,
                      formEvents.includes(event) && styles.checkboxChecked,
                    ]}
                  >
                    {formEvents.includes(event) && (
                      <Text style={styles.checkmark}>✓</Text>
                    )}
                  </View>
                  <Text style={styles.checkboxLabel}>{event}</Text>
                </TouchableOpacity>
              ))}

              <View style={styles.formActions}>
                <TouchableOpacity
                  style={styles.cancelBtn}
                  onPress={() => {
                    setShowForm(false)
                    setFormUrl('')
                    setFormEvents([])
                  }}
                >
                  <Text style={styles.cancelBtnText}>Cancel</Text>
                </TouchableOpacity>
                <TouchableOpacity
                  style={[styles.createBtn, isCreating && { opacity: 0.5 }]}
                  onPress={handleCreate}
                  disabled={isCreating}
                >
                  {isCreating ? (
                    <ActivityIndicator size="small" color="#fff" />
                  ) : (
                    <Text style={styles.createBtnText}>Create</Text>
                  )}
                </TouchableOpacity>
              </View>
            </View>
          ) : (
            <TouchableOpacity
              style={styles.addBtn}
              onPress={() => setShowForm(true)}
            >
              <Text style={styles.addBtnText}>+ Add Webhook</Text>
            </TouchableOpacity>
          )}
        </>
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
  webhookItem: {
    flexDirection: 'row',
    alignItems: 'center',
    backgroundColor: '#2a2a2a',
    borderRadius: 8,
    padding: 12,
    marginBottom: 8,
  },
  webhookInfo: {
    flex: 1,
    marginRight: 10,
  },
  webhookUrl: {
    color: '#fff',
    fontSize: 13,
    fontFamily: 'monospace',
  },
  eventsRow: {
    flexDirection: 'row',
    flexWrap: 'wrap',
    gap: 4,
    marginTop: 6,
  },
  eventTag: {
    backgroundColor: '#3a3a3a',
    borderRadius: 4,
    paddingHorizontal: 6,
    paddingVertical: 2,
  },
  eventTagText: {
    color: '#ccc',
    fontSize: 10,
    fontWeight: '500',
  },
  webhookActions: {
    flexDirection: 'row',
    gap: 6,
    alignItems: 'center',
  },
  statusBtn: {
    paddingHorizontal: 10,
    paddingVertical: 5,
    borderRadius: 6,
  },
  statusBtnActive: {
    backgroundColor: 'rgba(34, 197, 94, 0.2)',
  },
  statusBtnInactive: {
    backgroundColor: '#3a3a3a',
  },
  statusBtnText: {
    fontSize: 11,
    fontWeight: '600',
  },
  statusTextActive: {
    color: '#22c55e',
  },
  statusTextInactive: {
    color: '#888',
  },
  deleteBtn: {
    width: 28,
    height: 28,
    borderRadius: 6,
    backgroundColor: '#3a3a3a',
    alignItems: 'center',
    justifyContent: 'center',
  },
  deleteBtnText: {
    color: '#ef4444',
    fontSize: 12,
    fontWeight: '700',
  },
  formCard: {
    backgroundColor: '#2a2a2a',
    borderRadius: 10,
    padding: 14,
    marginTop: 8,
  },
  formTitle: {
    color: '#fff',
    fontSize: 15,
    fontWeight: '600',
    marginBottom: 12,
  },
  formLabel: {
    color: '#888',
    fontSize: 11,
    fontWeight: '600',
    textTransform: 'uppercase',
    marginTop: 12,
    marginBottom: 8,
  },
  input: {
    backgroundColor: '#1a1a1a',
    borderRadius: 8,
    paddingHorizontal: 14,
    paddingVertical: 12,
    color: '#fff',
    fontSize: 14,
  },
  checkboxRow: {
    flexDirection: 'row',
    alignItems: 'center',
    paddingVertical: 6,
  },
  checkbox: {
    width: 20,
    height: 20,
    borderRadius: 4,
    borderWidth: 2,
    borderColor: '#555',
    marginRight: 10,
    alignItems: 'center',
    justifyContent: 'center',
  },
  checkboxChecked: {
    backgroundColor: '#e50914',
    borderColor: '#e50914',
  },
  checkmark: {
    color: '#fff',
    fontSize: 12,
    fontWeight: '700',
  },
  checkboxLabel: {
    color: '#ccc',
    fontSize: 13,
  },
  formActions: {
    flexDirection: 'row',
    gap: 8,
    marginTop: 16,
  },
  cancelBtn: {
    flex: 1,
    backgroundColor: '#3a3a3a',
    borderRadius: 8,
    paddingVertical: 12,
    alignItems: 'center',
  },
  cancelBtnText: {
    color: '#fff',
    fontSize: 14,
    fontWeight: '500',
  },
  createBtn: {
    flex: 1,
    backgroundColor: '#e50914',
    borderRadius: 8,
    paddingVertical: 12,
    alignItems: 'center',
  },
  createBtnText: {
    color: '#fff',
    fontSize: 14,
    fontWeight: '600',
  },
  addBtn: {
    backgroundColor: '#e50914',
    borderRadius: 8,
    paddingVertical: 12,
    alignItems: 'center',
    marginTop: 8,
  },
  addBtnText: {
    color: '#fff',
    fontSize: 14,
    fontWeight: '600',
  },
})
