/**
 * Desktop-only onboarding: enter server URL + credentials in one shot.
 * Web users go straight to /login (which assumes same-origin /api).
 */

import { useEffect, useState } from 'react'
import { useNavigate, Navigate } from 'react-router'
import { useAuthStore } from '@/stores/auth'
import { useLogin } from '@/hooks/stores/useAuth'
import { isTauri } from '@/platform'
import { Logo } from '@/components/Logo'

const URL_PATTERN = /^https?:\/\/[^/]+/i

function normalizeUrl(input: string): string {
  let url = input.trim()
  if (!url) return ''
  if (!/^https?:\/\//i.test(url)) {
    url = `http://${url}`
  }
  // Strip trailing slash + /api suffix the user may have pasted in.
  return url.replace(/\/+$/, '').replace(/\/api$/, '')
}

export default function OnboardingPage() {
  const navigate = useNavigate()
  const { isAuthenticated } = useAuthStore()

  const [serverUrl, setServerUrl] = useState('')
  const [username, setUsername] = useState('')
  const [password, setPassword] = useState('')
  const [error, setError] = useState('')
  const [savingUrl, setSavingUrl] = useState(false)

  const { mutate: login, isPending: loggingIn } = useLogin()

  // Pre-fill if user previously set a URL but hasn't logged in yet.
  useEffect(() => {
    if (!isTauri()) return
    let mounted = true
    ;(async () => {
      const { getServerUrl } = await import('@/platform/desktop-adapter')
      const existing = await getServerUrl()
      if (mounted && existing) setServerUrl(existing)
    })()
    return () => {
      mounted = false
    }
  }, [])

  // Desktop-only screen — web hits /login directly.
  if (!isTauri()) return <Navigate to="/login" replace />
  if (isAuthenticated) return <Navigate to="/" replace />

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setError('')

    const url = normalizeUrl(serverUrl)
    if (!URL_PATTERN.test(url)) {
      setError('Server URL phải có dạng http://host:port')
      return
    }
    if (!username || !password) {
      setError('Nhập username + password')
      return
    }

    setSavingUrl(true)
    try {
      const { setServerUrl: persistUrl } = await import('@/platform/desktop-adapter')
      await persistUrl(url)

      const ping = await fetch(`${url}/api/health`).catch(() => null)
      if (!ping || !ping.ok) {
        setError(`Không kết nối được ${url}. Kiểm tra IP/port + NAS đang chạy.`)
        setSavingUrl(false)
        return
      }
    } catch (err) {
      setError(`Lỗi lưu server URL: ${(err as Error).message}`)
      setSavingUrl(false)
      return
    }
    setSavingUrl(false)

    login(
      { username, password },
      {
        onSuccess: () => navigate('/'),
        onError: (err: Error) => setError(err.message || 'Đăng nhập thất bại'),
      },
    )
  }

  const isPending = savingUrl || loggingIn

  return (
    <div className="flex min-h-screen flex-col bg-netflix-black">
      <div className="absolute inset-0 bg-gradient-to-b from-netflix-black/50 via-netflix-black/80 to-netflix-black" />

      <header className="relative z-10 p-6">
        <Logo size="lg" />
      </header>

      <main className="relative z-10 flex flex-1 items-center justify-center px-4">
        <div className="w-full max-w-md rounded-xl bg-black/75 p-8 backdrop-blur-sm md:p-12">
          <h1 className="mb-2 text-3xl font-bold text-white">Connect to Velox</h1>
          <p className="mb-8 text-sm text-gray-400">
            Nhập địa chỉ NAS server + tài khoản để bắt đầu
          </p>

          {error && (
            <div className="mb-4 rounded bg-netflix-red/20 p-3 text-sm text-netflix-red">
              {error}
            </div>
          )}

          <form onSubmit={handleSubmit} className="space-y-4">
            <input
              type="text"
              value={serverUrl}
              onChange={(e) => setServerUrl(e.target.value)}
              placeholder="http://192.168.1.10:8098"
              className="w-full rounded bg-netflix-gray px-4 py-3 text-white placeholder-gray-400 outline-none ring-1 ring-transparent transition-all focus:ring-netflix-red"
              disabled={isPending}
              autoFocus
            />

            <input
              type="text"
              value={username}
              onChange={(e) => setUsername(e.target.value)}
              placeholder="Username"
              className="w-full rounded bg-netflix-gray px-4 py-3 text-white placeholder-gray-400 outline-none ring-1 ring-transparent transition-all focus:ring-netflix-red"
              disabled={isPending}
              autoComplete="username"
            />

            <input
              type="password"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              placeholder="Password"
              className="w-full rounded bg-netflix-gray px-4 py-3 text-white placeholder-gray-400 outline-none ring-1 ring-transparent transition-all focus:ring-netflix-red"
              disabled={isPending}
              autoComplete="current-password"
            />

            <button
              type="submit"
              disabled={isPending}
              className="w-full rounded bg-netflix-red py-3 font-semibold text-white transition-colors hover:bg-netflix-red-hover disabled:cursor-not-allowed disabled:opacity-50"
            >
              {savingUrl ? 'Connecting…' : loggingIn ? 'Signing in…' : 'Connect & Sign In'}
            </button>
          </form>
        </div>
      </main>
    </div>
  )
}
