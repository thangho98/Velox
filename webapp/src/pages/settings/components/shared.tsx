/* eslint-disable react-refresh/only-export-components */
import { LuX } from 'react-icons/lu'

export const inputClass =
  'w-full rounded bg-netflix-gray px-4 py-2.5 text-sm text-white outline-none ring-1 ring-transparent transition-all placeholder:text-gray-500 focus:ring-netflix-red'
export const inputDisabled = 'w-full rounded bg-netflix-black px-4 py-2.5 text-sm text-gray-500'

export function SectionHeader({ title, description }: { title: string; description: string }) {
  return (
    <div>
      <h2 className="text-2xl font-bold text-white">{title}</h2>
      <p className="text-sm text-gray-400">{description}</p>
    </div>
  )
}

export function Field({
  label,
  compact,
  children,
}: {
  label: React.ReactNode
  compact?: boolean
  children: React.ReactNode
}) {
  return (
    <div>
      <label
        className={`mb-1.5 block text-sm font-medium text-gray-400 ${compact ? 'text-xs' : ''}`}
      >
        {label}
      </label>
      {children}
    </div>
  )
}

export function SaveButton({
  isPending,
  label = 'Save Changes',
}: {
  isPending: boolean
  label?: string
}) {
  return (
    <button
      type="submit"
      disabled={isPending}
      className="rounded bg-netflix-red px-6 py-2 text-sm font-semibold text-white transition-colors hover:bg-netflix-red-hover disabled:opacity-50"
    >
      {isPending ? 'Saving...' : label}
    </button>
  )
}

export function Spinner() {
  return (
    <div className="flex h-40 items-center justify-center">
      <div className="h-8 w-8 animate-spin rounded-full border-2 border-netflix-red border-t-transparent" />
    </div>
  )
}

export function ErrorMsg({ children }: { children: React.ReactNode }) {
  return <div className="rounded-lg bg-netflix-red/20 p-3 text-sm text-netflix-red">{children}</div>
}

export function SuccessMsg({ children }: { children: React.ReactNode }) {
  return <div className="rounded-lg bg-green-500/20 p-3 text-sm text-green-400">{children}</div>
}

export function Modal({
  title,
  onClose,
  children,
}: {
  title: string
  onClose: () => void
  children: React.ReactNode
}) {
  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/70 backdrop-blur-sm p-4">
      <div className="w-full max-w-lg rounded-xl bg-netflix-dark p-6 shadow-2xl">
        <div className="mb-5 flex items-center justify-between">
          <h2 className="text-lg font-bold text-white">{title}</h2>
          <button onClick={onClose} className="text-gray-400 hover:text-white">
            <LuX size={22} />
          </button>
        </div>
        {children}
      </div>
    </div>
  )
}

export function formatBytes(bytes: number): string {
  if (bytes === 0) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return parseFloat((bytes / Math.pow(k, i)).toFixed(1)) + ' ' + sizes[i]
}

export function timeAgo(dateStr: string): string {
  const seconds = Math.floor((Date.now() - new Date(dateStr).getTime()) / 1000)
  if (seconds < 60) return 'just now'
  const minutes = Math.floor(seconds / 60)
  if (minutes < 60) return `${minutes}m ago`
  const hours = Math.floor(minutes / 60)
  if (hours < 24) return `${hours}h ago`
  const days = Math.floor(hours / 24)
  return `${days}d ago`
}
