import { useState } from 'react'
import { LuCloud, LuPlus, LuRefreshCw, LuTrash2 } from 'react-icons/lu'
import { Select } from '@/components/ui/Select'
import {
  useCloudDrivers,
  useCreateCloudProvider,
  useDeleteCloudProvider,
  useRefreshCloudProvider,
  useStorageProviders,
} from '@/hooks/stores/useAdmin'
import type { ProviderHealth, StorageProvider } from '@velox/shared/types'
import { ErrorMsg, Field, Modal, SectionHeader, Spinner, inputClass } from './shared'

const HEALTH_STYLES: Record<ProviderHealth, { label: string; classes: string }> = {
  healthy: { label: '✅ Healthy', classes: 'bg-green-500/20 text-green-400' },
  expiring_soon: { label: '⚠️ Expiring', classes: 'bg-amber-500/20 text-amber-400' },
  expired: { label: '❌ Expired', classes: 'bg-red-500/20 text-red-400' },
  error: { label: '❌ Error', classes: 'bg-red-500/20 text-red-400' },
  unknown: { label: '❓ Unknown', classes: 'bg-gray-500/20 text-gray-400' },
}

export function StorageSection() {
  const { data: providers, isLoading } = useStorageProviders()
  const { data: drivers } = useCloudDrivers()
  const refreshMut = useRefreshCloudProvider()
  const deleteMut = useDeleteCloudProvider()

  const [showAdd, setShowAdd] = useState(false)

  return (
    <div className="space-y-6">
      <div className="flex items-start justify-between gap-4">
        <SectionHeader
          title="Storage Providers"
          description="Cloud accounts (fshare, future: Google Drive, OneDrive) used by cloud-backed libraries."
        />
        <button
          onClick={() => setShowAdd(true)}
          disabled={!drivers || drivers.length === 0}
          className="inline-flex items-center gap-2 rounded bg-netflix-red px-4 py-2 text-sm font-semibold text-white transition-colors hover:bg-netflix-red-hover disabled:opacity-50"
        >
          <LuPlus size={16} /> Add Provider
        </button>
      </div>

      {isLoading ? (
        <Spinner />
      ) : providers && providers.length > 0 ? (
        <div className="space-y-3">
          {providers.map((p) => (
            <ProviderCard
              key={p.id}
              provider={p}
              onRefresh={() => refreshMut.mutate(p.id)}
              onDelete={() => {
                if (
                  confirm(
                    `Delete provider "${p.display_name}"?\n\nLibraries linked to this provider will stop working until re-linked.`,
                  )
                ) {
                  deleteMut.mutate(p.id)
                }
              }}
              refreshing={refreshMut.isPending && refreshMut.variables === p.id}
            />
          ))}
        </div>
      ) : (
        <EmptyState />
      )}

      {showAdd && drivers && (
        <AddProviderModal drivers={drivers} onClose={() => setShowAdd(false)} />
      )}
    </div>
  )
}

function ProviderCard({
  provider,
  onRefresh,
  onDelete,
  refreshing,
}: {
  provider: StorageProvider
  onRefresh: () => void
  onDelete: () => void
  refreshing: boolean
}) {
  const health = HEALTH_STYLES[provider.health] ?? HEALTH_STYLES.unknown
  return (
    <div className="rounded-lg bg-netflix-gray p-4">
      <div className="flex items-start gap-3">
        <div className="rounded-md bg-netflix-red/20 p-2 text-netflix-red">
          <LuCloud size={22} />
        </div>
        <div className="flex-1 min-w-0">
          <div className="flex items-center gap-2">
            <h3 className="text-base font-semibold text-white truncate">{provider.display_name}</h3>
            <span className={`rounded px-2 py-0.5 text-xs ${health.classes}`}>{health.label}</span>
            {provider.account_type && (
              <span className="rounded bg-netflix-black px-2 py-0.5 text-xs text-gray-400">
                {provider.account_type}
              </span>
            )}
          </div>
          <p className="mt-1 text-sm text-gray-400 truncate">
            {provider.provider_type} · {provider.account_email}
          </p>
          {provider.last_error && (
            <p className="mt-2 rounded bg-red-500/10 px-2 py-1 text-xs text-red-400">
              {provider.last_error}
            </p>
          )}
        </div>
        <div className="flex gap-2">
          <button
            onClick={onRefresh}
            disabled={refreshing}
            className="rounded px-3 py-1.5 text-sm text-gray-300 hover:bg-netflix-black disabled:opacity-50"
            title="Refresh session"
          >
            <LuRefreshCw size={16} className={refreshing ? 'animate-spin' : ''} />
          </button>
          <button
            onClick={onDelete}
            className="rounded px-3 py-1.5 text-sm text-red-400 hover:bg-red-500/10"
            title="Delete provider"
          >
            <LuTrash2 size={16} />
          </button>
        </div>
      </div>
    </div>
  )
}

function EmptyState() {
  return (
    <div className="rounded-lg border border-dashed border-netflix-gray/50 p-8 text-center">
      <LuCloud size={32} className="mx-auto mb-2 text-gray-500" />
      <p className="text-sm text-gray-400">
        No storage providers yet. Add one to use cloud-backed libraries.
      </p>
    </div>
  )
}

function AddProviderModal({
  drivers,
  onClose,
}: {
  drivers: Array<{ type: string; display_name: string; auth_flow: string }>
  onClose: () => void
}) {
  const [providerType, setProviderType] = useState(drivers[0]?.type ?? '')
  const driver = drivers.find((d) => d.type === providerType)
  const [displayName, setDisplayName] = useState(`My ${driver?.display_name ?? ''}`)
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [error, setError] = useState('')

  const createMut = useCreateCloudProvider()

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault()
    setError('')
    if (!email.trim() || !password.trim()) {
      setError('Email and password are required')
      return
    }
    createMut.mutate(
      {
        provider_type: providerType,
        display_name: displayName.trim(),
        auth: { flow: 'password', email: email.trim(), password },
      },
      {
        onSuccess: () => onClose(),
        onError: (err: Error) => setError(err.message || 'Failed to create provider'),
      },
    )
  }

  return (
    <Modal title="Add Storage Provider" onClose={onClose}>
      <form onSubmit={handleSubmit} className="space-y-4">
        <Field label="Provider">
          <Select
            value={providerType}
            onChange={(e) => {
              setProviderType(e.target.value)
              const d = drivers.find((drv) => drv.type === e.target.value)
              if (d) setDisplayName(`My ${d.display_name}`)
            }}
            className="w-full"
          >
            {drivers.map((d) => (
              <option key={d.type} value={d.type}>
                {d.display_name}
              </option>
            ))}
          </Select>
        </Field>

        <Field label="Display name">
          <input
            type="text"
            value={displayName}
            onChange={(e) => setDisplayName(e.target.value)}
            className={inputClass}
            placeholder="My Fshare"
          />
        </Field>

        {driver?.auth_flow === 'password' && (
          <>
            <Field label="Email">
              <input
                type="email"
                value={email}
                onChange={(e) => setEmail(e.target.value)}
                className={inputClass}
                autoComplete="off"
                required
              />
            </Field>
            <Field label="Password">
              <input
                type="password"
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                className={inputClass}
                autoComplete="off"
                required
              />
            </Field>
            {providerType === 'fshare' && (
              <p className="rounded bg-amber-500/10 p-2 text-xs text-amber-400">
                ⚠️ Requires a fshare <strong>VIP</strong> account. Free tier cannot stream.
              </p>
            )}
          </>
        )}

        {driver?.auth_flow === 'oauth' && (
          <p className="rounded bg-blue-500/10 p-2 text-xs text-blue-400">
            OAuth sign-in is not yet supported. Pick a password-based provider.
          </p>
        )}

        {error && <ErrorMsg>{error}</ErrorMsg>}

        <div className="flex justify-end gap-2 pt-2">
          <button
            type="button"
            onClick={onClose}
            className="rounded px-4 py-2 text-sm text-gray-300 hover:bg-netflix-gray"
          >
            Cancel
          </button>
          <button
            type="submit"
            disabled={createMut.isPending || driver?.auth_flow !== 'password'}
            className="rounded bg-netflix-red px-4 py-2 text-sm font-semibold text-white hover:bg-netflix-red-hover disabled:opacity-50"
          >
            {createMut.isPending ? 'Testing...' : 'Test & Save'}
          </button>
        </div>
      </form>
    </Modal>
  )
}
