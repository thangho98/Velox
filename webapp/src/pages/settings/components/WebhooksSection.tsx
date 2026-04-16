import { useTranslation } from '@/hooks/useTranslation'
import { useState } from 'react'
import { LuPlus, LuGlobe, LuPause, LuPlay, LuTrash2 } from 'react-icons/lu'
import {
  useWebhooks,
  useCreateWebhook,
  useUpdateWebhook,
  useDeleteWebhook,
} from '@/hooks/stores/useAdmin'
import type { Webhook } from '@/types/api'
import { SectionHeader, Field, Spinner, ErrorMsg, Modal, inputClass } from './shared'

const WEBHOOK_EVENTS = [
  'scan_complete',
  'transcode_complete',
  'transcode_failed',
  'library_watcher',
]

function parseWebhookEvents(raw: string): string[] {
  try {
    const parsed = JSON.parse(raw)
    if (Array.isArray(parsed)) return parsed as string[]
  } catch {
    // Legacy CSV row (pre-migration) — degrade gracefully
    if (raw && raw !== '[]')
      return raw
        .split(',')
        .map((s) => s.trim())
        .filter(Boolean)
  }
  return []
}

export function WebhooksSection() {
  const { t } = useTranslation('settings')

  const { data: webhooks, isLoading } = useWebhooks()
  const { mutate: createWebhook, isPending: isCreating } = useCreateWebhook()
  const { mutate: updateWebhook } = useUpdateWebhook()
  const { mutate: deleteWebhook } = useDeleteWebhook()
  const [showAddModal, setShowAddModal] = useState(false)
  const [formUrl, setFormUrl] = useState('')
  const [formEvents, setFormEvents] = useState<string[]>([])
  const [formError, setFormError] = useState('')

  const toggleEvent = (event: string) => {
    setFormEvents((prev) =>
      prev.includes(event) ? prev.filter((e) => e !== event) : [...prev, event],
    )
  }

  const handleCreate = (e: React.FormEvent) => {
    e.preventDefault()
    setFormError('')
    if (!formUrl.trim()) {
      setFormError('URL is required')
      return
    }
    if (formEvents.length === 0) {
      setFormError('Select at least one event')
      return
    }
    createWebhook(
      { url: formUrl.trim(), events: JSON.stringify(formEvents), active: true },
      {
        onSuccess: () => {
          setShowAddModal(false)
          setFormUrl('')
          setFormEvents([])
        },
        onError: (err: Error) => setFormError(err.message),
      },
    )
  }

  const handleToggleActive = (webhook: Webhook) => {
    updateWebhook({ id: webhook.id, data: { active: !webhook.active } })
  }

  const handleDelete = (id: number) => {
    if (confirm('Delete this webhook?')) deleteWebhook(id)
  }

  return (
    <div className="max-w-3xl">
      <div className="flex items-center justify-between">
        <SectionHeader
          title="Webhooks"
          description={`${webhooks?.length || 0} webhook${(webhooks?.length || 0) !== 1 ? 's' : ''} configured`}
        />
        <button
          onClick={() => setShowAddModal(true)}
          className="flex items-center gap-2 rounded bg-netflix-red px-4 py-2 text-sm font-semibold text-white transition-colors hover:bg-netflix-red-hover"
        >
          <LuPlus size={16} />
          Add Webhook
        </button>
      </div>

      {isLoading ? (
        <Spinner />
      ) : webhooks?.length === 0 ? (
        <div className="mt-6 flex h-40 flex-col items-center justify-center rounded-lg bg-netflix-dark">
          <LuGlobe size={36} className="text-gray-600" />
          <p className="mt-2 text-sm text-gray-400">No webhooks configured</p>
        </div>
      ) : (
        <div className="mt-6 space-y-3">
          {webhooks?.map((wh) => (
            <div
              key={wh.id}
              className="flex items-center justify-between rounded-lg bg-netflix-dark p-4"
            >
              <div className="min-w-0 flex-1">
                <div className="flex items-center gap-2">
                  <LuGlobe size={14} className="shrink-0 text-gray-400" />
                  <p className="truncate font-mono text-sm text-white">{wh.url}</p>
                </div>
                <div className="mt-1.5 flex flex-wrap gap-1.5">
                  {parseWebhookEvents(wh.events).map((ev) => (
                    <span
                      key={ev}
                      className="rounded bg-netflix-gray px-2 py-0.5 text-[10px] font-medium text-gray-300"
                    >
                      {ev}
                    </span>
                  ))}
                </div>
              </div>
              <div className="ml-4 flex shrink-0 items-center gap-2">
                <button
                  onClick={() => handleToggleActive(wh)}
                  className={`flex items-center gap-1.5 rounded px-3 py-1.5 text-xs transition-colors ${
                    wh.active
                      ? 'bg-green-500/20 text-green-400 hover:bg-green-500/30'
                      : 'bg-netflix-gray text-gray-400 hover:bg-gray-600'
                  }`}
                >
                  {wh.active ? (
                    <>
                      <LuPause size={12} /> Active
                    </>
                  ) : (
                    <>
                      <LuPlay size={12} /> Inactive
                    </>
                  )}
                </button>
                <button
                  onClick={() => handleDelete(wh.id)}
                  className="rounded bg-netflix-gray p-1.5 text-white transition-colors hover:bg-red-600"
                >
                  <LuTrash2 size={13} />
                </button>
              </div>
            </div>
          ))}
        </div>
      )}

      {showAddModal && (
        <Modal title="Add Webhook" onClose={() => setShowAddModal(false)}>
          <form onSubmit={handleCreate} className="space-y-5">
            {formError && <ErrorMsg>{formError}</ErrorMsg>}
            <Field label="URL">
              <input
                type="url"
                value={formUrl}
                onChange={(e) => setFormUrl(e.target.value)}
                placeholder="https://example.com/webhook"
                className={inputClass}
                required
              />
            </Field>
            <Field label="Events">
              <div className="space-y-2">
                {WEBHOOK_EVENTS.map((event) => (
                  <label key={event} className="flex items-center gap-3">
                    <input
                      type="checkbox"
                      checked={formEvents.includes(event)}
                      onChange={() => toggleEvent(event)}
                      className="h-4 w-4 rounded"
                    />
                    <span className="text-sm text-gray-300">{event}</span>
                  </label>
                ))}
              </div>
            </Field>
            <div className="flex gap-3 pt-2">
              <button
                type="button"
                onClick={() => {
                  setShowAddModal(false)
                  setFormUrl('')
                  setFormEvents([])
                  setFormError('')
                }}
                className="flex-1 rounded bg-netflix-gray px-4 py-2.5 font-medium text-white hover:bg-gray-600"
              >
                Cancel
              </button>
              <button
                type="submit"
                disabled={isCreating}
                className="flex-1 rounded bg-netflix-red px-4 py-2.5 font-medium text-white hover:bg-netflix-red-hover disabled:opacity-50"
              >
                {isCreating ? 'Creating...' : 'Create Webhook'}
              </button>
            </div>
          </form>
        </Modal>
      )}
    </div>
  )
}
