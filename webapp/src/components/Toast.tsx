import { LuCircleCheck, LuCircleX, LuInfo, LuX } from 'react-icons/lu'
import { useUIStore } from '@/stores/ui'
import type { Toast, ToastType } from '@/stores/ui'

function toastIcon(type: ToastType) {
  if (type === 'success') return <LuCircleCheck size={20} className="shrink-0 text-green-400" />
  if (type === 'error') return <LuCircleX size={20} className="shrink-0 text-netflix-red" />
  return <LuInfo size={20} className="shrink-0 text-blue-400" />
}

function ToastItem({ toast }: { toast: Toast }) {
  const { removeToast } = useUIStore()
  return (
    <div
      className="flex items-start gap-3 rounded-lg bg-netflix-dark border border-netflix-gray px-4 py-3 shadow-xl min-w-[280px] max-w-sm"
      role="alert"
    >
      {toastIcon(toast.type)}
      <p className="flex-1 text-sm text-white">{toast.message}</p>
      <button
        onClick={() => removeToast(toast.id)}
        className="shrink-0 text-gray-500 hover:text-white transition-colors"
        aria-label="Dismiss"
      >
        <LuX size={16} />
      </button>
    </div>
  )
}

export function Toaster() {
  const { toasts } = useUIStore()
  if (toasts.length === 0) return null
  return (
    <div className="fixed bottom-6 right-6 z-[200] flex flex-col gap-2">
      {toasts.map((toast) => (
        <ToastItem key={toast.id} toast={toast} />
      ))}
    </div>
  )
}

export function useToast() {
  const { addToast } = useUIStore()
  return {
    success: (message: string) => addToast(message, 'success'),
    error: (message: string) => addToast(message, 'error'),
    info: (message: string) => addToast(message, 'info'),
  }
}
