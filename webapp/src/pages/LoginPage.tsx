import { useState } from 'react'
import { useNavigate, Navigate } from 'react-router'
import { useLogin, useSetupStatus } from '@/hooks/stores/useAuth'
import { Logo } from '@/components/Logo'

export function LoginPage() {
  const navigate = useNavigate()
  const [username, setUsername] = useState('')
  const [password, setPassword] = useState('')
  const [error, setError] = useState('')

  const { data: setupStatus, isLoading: checkingSetup } = useSetupStatus()
  const { mutate: login, isPending } = useLogin()

  // Redirect to setup if not configured
  if (!checkingSetup && setupStatus && !setupStatus.configured) {
    return <Navigate to="/setup" replace />
  }

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setError('')

    if (!username || !password) {
      setError('Username and password are required')
      return
    }

    login(
      { username, password },
      {
        onSuccess: () => {
          navigate('/')
        },
        onError: (err: Error) => {
          setError(err.message || 'Login failed')
        },
      },
    )
  }

  return (
    <div className="flex min-h-screen flex-col bg-netflix-black">
      {/* Background gradient */}
      <div className="absolute inset-0 bg-gradient-to-b from-netflix-black/50 via-netflix-black/80 to-netflix-black" />

      {/* Header */}
      <header className="relative z-10 p-6">
        <Logo size="lg" />
      </header>

      {/* Content */}
      <main className="relative z-10 flex flex-1 items-center justify-center px-4">
        <div className="w-full max-w-md rounded-xl bg-black/75 p-8 backdrop-blur-sm md:p-12">
          <h1 className="mb-8 text-3xl font-bold text-white">Sign In</h1>

          {error && (
            <div className="mb-4 rounded bg-netflix-red/20 p-3 text-sm text-netflix-red">
              {error}
            </div>
          )}

          <form onSubmit={handleSubmit} className="space-y-4">
            <div>
              <input
                type="text"
                value={username}
                onChange={(e) => setUsername(e.target.value)}
                placeholder="Username"
                className="w-full rounded bg-netflix-gray px-4 py-3 text-white placeholder-gray-400 outline-none ring-1 ring-transparent transition-all focus:ring-netflix-red"
                disabled={isPending}
              />
            </div>

            <div>
              <input
                type="password"
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                placeholder="Password"
                className="w-full rounded bg-netflix-gray px-4 py-3 text-white placeholder-gray-400 outline-none ring-1 ring-transparent transition-all focus:ring-netflix-red"
                disabled={isPending}
              />
            </div>

            <button
              type="submit"
              disabled={isPending}
              className="w-full rounded bg-netflix-red py-3 font-semibold text-white transition-colors hover:bg-netflix-red-hover disabled:cursor-not-allowed disabled:opacity-50"
            >
              {isPending ? 'Signing in...' : 'Sign In'}
            </button>
          </form>

          <div className="mt-8 text-center">
            <p className="text-gray-400">
              New to Velox?{' '}
              <span className="text-white hover:underline cursor-pointer">
                Contact your administrator
              </span>
            </p>
          </div>
        </div>
      </main>

      {/* Footer */}
      <footer className="relative z-10 bg-netflix-black/90 p-8 text-sm text-gray-500">
        <div className="mx-auto max-w-4xl">
          <p className="mb-4">Questions? Contact your server administrator.</p>
          <div className="grid grid-cols-2 gap-4 md:grid-cols-4">
            <span className="hover:underline">Privacy</span>
            <span className="hover:underline">Terms of Service</span>
            <span className="hover:underline">Help Center</span>
          </div>
        </div>
      </footer>
    </div>
  )
}
