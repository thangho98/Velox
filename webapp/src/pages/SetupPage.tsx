import { useState } from 'react'
import { useNavigate, Navigate } from 'react-router'
import { useSetupStatus, useSetup } from '@/hooks/stores/useAuth'
import { Logo } from '@/components/Logo'

export function SetupPage() {
  const navigate = useNavigate()
  const [username, setUsername] = useState('')
  const [password, setPassword] = useState('')
  const [confirmPassword, setConfirmPassword] = useState('')
  const [displayName, setDisplayName] = useState('')
  const [error, setError] = useState('')

  const { data: setupStatus, isLoading: checkingSetup } = useSetupStatus()
  const { mutate: setup, isPending } = useSetup()

  // Redirect to home if already configured
  if (!checkingSetup && setupStatus?.configured) {
    return <Navigate to="/" replace />
  }

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setError('')

    if (!username || !password) {
      setError('Username and password are required')
      return
    }

    if (password !== confirmPassword) {
      setError('Passwords do not match')
      return
    }

    if (password.length < 8) {
      setError('Password must be at least 8 characters')
      return
    }

    setup(
      { username, password, display_name: displayName || username },
      {
        onSuccess: () => {
          navigate('/login')
        },
        onError: (err: Error) => {
          setError(err.message || 'Setup failed')
        },
      },
    )
  }

  return (
    <div className="flex min-h-screen flex-col bg-netflix-black">
      {/* Background gradient */}
      <div className="absolute inset-0 bg-gradient-to-b from-netflix-dark/50 via-netflix-black/80 to-netflix-black" />

      {/* Header */}
      <header className="relative z-10 p-6">
        <Logo size="lg" />
      </header>

      {/* Content */}
      <main className="relative z-10 flex flex-1 items-center justify-center px-4">
        <div className="w-full max-w-md rounded-xl bg-black/75 p-8 backdrop-blur-sm md:p-12">
          <h1 className="mb-2 text-3xl font-bold text-white">Welcome to Velox</h1>
          <p className="mb-8 text-gray-400">Create your admin account to get started</p>

          {error && (
            <div className="mb-4 rounded bg-netflix-red/20 p-3 text-sm text-netflix-red">
              {error}
            </div>
          )}

          <form onSubmit={handleSubmit} className="space-y-4">
            <div>
              <label className="mb-2 block text-sm text-gray-400">Username</label>
              <input
                type="text"
                value={username}
                onChange={(e) => setUsername(e.target.value)}
                placeholder="Enter username"
                className="w-full rounded bg-netflix-gray px-4 py-3 text-white placeholder-gray-400 outline-none ring-1 ring-transparent transition-all focus:ring-netflix-red"
                disabled={isPending}
                required
              />
            </div>

            <div>
              <label className="mb-2 block text-sm text-gray-400">Display Name (optional)</label>
              <input
                type="text"
                value={displayName}
                onChange={(e) => setDisplayName(e.target.value)}
                placeholder="How should we call you?"
                className="w-full rounded bg-netflix-gray px-4 py-3 text-white placeholder-gray-400 outline-none ring-1 ring-transparent transition-all focus:ring-netflix-red"
                disabled={isPending}
              />
            </div>

            <div>
              <label className="mb-2 block text-sm text-gray-400">Password</label>
              <input
                type="password"
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                placeholder="Min 8 characters"
                className="w-full rounded bg-netflix-gray px-4 py-3 text-white placeholder-gray-400 outline-none ring-1 ring-transparent transition-all focus:ring-netflix-red"
                disabled={isPending}
                required
                minLength={8}
              />
            </div>

            <div>
              <label className="mb-2 block text-sm text-gray-400">Confirm Password</label>
              <input
                type="password"
                value={confirmPassword}
                onChange={(e) => setConfirmPassword(e.target.value)}
                placeholder="Confirm your password"
                className="w-full rounded bg-netflix-gray px-4 py-3 text-white placeholder-gray-400 outline-none ring-1 ring-transparent transition-all focus:ring-netflix-red"
                disabled={isPending}
                required
              />
            </div>

            <button
              type="submit"
              disabled={isPending}
              className="w-full rounded bg-netflix-red py-3 font-semibold text-white transition-colors hover:bg-netflix-red-hover disabled:cursor-not-allowed disabled:opacity-50"
            >
              {isPending ? 'Creating...' : 'Create Admin Account'}
            </button>
          </form>

          <div className="mt-6 text-xs text-gray-500">
            <p>By creating an account, you agree to manage this Velox media server.</p>
          </div>
        </div>
      </main>

      {/* Footer */}
      <footer className="relative z-10 bg-netflix-black/90 p-8 text-sm text-gray-500">
        <div className="mx-auto max-w-4xl">
          <p>Velox - Self-hosted Media Server</p>
        </div>
      </footer>
    </div>
  )
}
