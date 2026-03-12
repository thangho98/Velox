import { useState } from 'react'
import { Link } from 'react-router'
import { useUsers, useCreateUser, useUpdateUser, useDeleteUser } from '@/hooks/stores/useUsers'
import { useAuthStore } from '@/stores/auth'
import type { User } from '@/types/api'

export function AdminUsersPage() {
  const { user: currentUser } = useAuthStore()
  const { data: users, isLoading } = useUsers()
  const [isCreateModalOpen, setIsCreateModalOpen] = useState(false)
  const [editingUser, setEditingUser] = useState<User | null>(null)

  if (!currentUser?.is_admin) {
    return (
      <div className="flex h-screen flex-col items-center justify-center">
        <p className="mb-4 text-xl text-white">Access denied</p>
        <Link to="/" className="text-netflix-red hover:underline">
          Go home
        </Link>
      </div>
    )
  }

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold text-white">User Management</h1>
          <p className="text-gray-400">Manage user accounts and permissions</p>
        </div>
        <button
          onClick={() => setIsCreateModalOpen(true)}
          className="flex items-center gap-2 rounded bg-netflix-red px-4 py-2 font-semibold text-white transition-colors hover:bg-netflix-red-hover"
        >
          <svg className="h-5 w-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 4v16m8-8H4" />
          </svg>
          Add User
        </button>
      </div>

      {isLoading ? (
        <div className="flex h-64 items-center justify-center">
          <div className="h-8 w-8 animate-spin rounded-full border-2 border-netflix-red border-t-transparent" />
        </div>
      ) : (
        <div className="overflow-hidden rounded-xl bg-netflix-dark">
          <table className="w-full">
            <thead className="border-b border-netflix-gray bg-netflix-black/50">
              <tr>
                <th className="px-4 py-3 text-left text-sm font-medium text-gray-400">User</th>
                <th className="px-4 py-3 text-left text-sm font-medium text-gray-400">Role</th>
                <th className="px-4 py-3 text-left text-sm font-medium text-gray-400">Created</th>
                <th className="px-4 py-3 text-left text-sm font-medium text-gray-400">Actions</th>
              </tr>
            </thead>
            <tbody>
              {users?.map((user) => (
                <tr
                  key={user.id}
                  className="border-b border-netflix-gray/50 last:border-b-0 hover:bg-netflix-gray/30"
                >
                  <td className="px-4 py-3">
                    <div className="flex items-center gap-3">
                      <div className="flex h-10 w-10 items-center justify-center rounded-full bg-netflix-gray">
                        <span className="text-sm font-medium text-white">
                          {(user.display_name || user.username).charAt(0).toUpperCase()}
                        </span>
                      </div>
                      <div>
                        <p className="font-medium text-white">
                          {user.display_name || user.username}
                        </p>
                        <p className="text-sm text-gray-500">{user.username}</p>
                      </div>
                    </div>
                  </td>
                  <td className="px-4 py-3">
                    <span
                      className={`rounded px-2 py-1 text-xs font-medium ${
                        user.is_admin
                          ? 'bg-purple-500/20 text-purple-400'
                          : 'bg-blue-500/20 text-blue-400'
                      }`}
                    >
                      {user.is_admin ? 'Admin' : 'User'}
                    </span>
                  </td>
                  <td className="px-4 py-3 text-sm text-gray-400">
                    {new Date(user.created_at).toLocaleDateString()}
                  </td>
                  <td className="px-4 py-3">
                    <div className="flex gap-2">
                      <button
                        onClick={() => setEditingUser(user)}
                        className="rounded bg-netflix-gray px-3 py-1.5 text-sm text-white transition-colors hover:bg-gray-600"
                      >
                        Edit
                      </button>
                      {user.id !== currentUser.id && <DeleteUserButton userId={user.id} />}
                    </div>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
          {users?.length === 0 && (
            <div className="p-8 text-center">
              <p className="text-gray-400">No users found</p>
            </div>
          )}
        </div>
      )}

      {isCreateModalOpen && <CreateUserModal onClose={() => setIsCreateModalOpen(false)} />}
      {editingUser && <EditUserModal user={editingUser} onClose={() => setEditingUser(null)} />}
    </div>
  )
}

function CreateUserModal({ onClose }: { onClose: () => void }) {
  const { mutate: createUser, isPending } = useCreateUser()
  const [form, setForm] = useState({
    username: '',
    password: '',
    display_name: '',
    is_admin: false,
  })
  const [error, setError] = useState('')

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault()
    setError('')

    if (!form.username || !form.password) {
      setError('Username and password are required')
      return
    }

    if (form.password.length < 8) {
      setError('Password must be at least 8 characters')
      return
    }

    createUser(form, {
      onSuccess: () => onClose(),
      onError: (err: Error) => setError(err.message),
    })
  }

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/70 backdrop-blur-sm">
      <div className="w-full max-w-md rounded-xl bg-netflix-dark p-6 shadow-2xl">
        <div className="mb-6 flex items-center justify-between">
          <h2 className="text-xl font-bold text-white">Add User</h2>
          <button onClick={onClose} className="text-gray-400 hover:text-white">
            <svg className="h-6 w-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                strokeWidth={2}
                d="M6 18L18 6M6 6l12 12"
              />
            </svg>
          </button>
        </div>

        {error && (
          <div className="mb-4 rounded-lg bg-netflix-red/20 p-3 text-sm text-netflix-red">
            {error}
          </div>
        )}

        <form onSubmit={handleSubmit} className="space-y-4">
          <div>
            <label className="mb-2 block text-sm font-medium text-gray-400">Username *</label>
            <input
              type="text"
              value={form.username}
              onChange={(e) => setForm({ ...form, username: e.target.value })}
              className="w-full rounded-lg bg-netflix-gray px-4 py-3 text-white outline-none ring-1 ring-transparent transition-all focus:ring-netflix-red"
              required
            />
          </div>

          <div>
            <label className="mb-2 block text-sm font-medium text-gray-400">Display Name</label>
            <input
              type="text"
              value={form.display_name}
              onChange={(e) => setForm({ ...form, display_name: e.target.value })}
              className="w-full rounded-lg bg-netflix-gray px-4 py-3 text-white outline-none ring-1 ring-transparent transition-all focus:ring-netflix-red"
            />
          </div>

          <div>
            <label className="mb-2 block text-sm font-medium text-gray-400">Password *</label>
            <input
              type="password"
              value={form.password}
              onChange={(e) => setForm({ ...form, password: e.target.value })}
              className="w-full rounded-lg bg-netflix-gray px-4 py-3 text-white outline-none ring-1 ring-transparent transition-all focus:ring-netflix-red"
              required
              minLength={8}
            />
          </div>

          <div className="flex items-center gap-3">
            <input
              type="checkbox"
              id="is_admin"
              checked={form.is_admin}
              onChange={(e) => setForm({ ...form, is_admin: e.target.checked })}
              className="h-4 w-4 rounded border-netflix-gray bg-netflix-gray text-netflix-red focus:ring-netflix-red"
            />
            <label htmlFor="is_admin" className="text-sm text-gray-300">
              Administrator
            </label>
          </div>

          <div className="flex justify-end gap-3 pt-4">
            <button
              type="button"
              onClick={onClose}
              className="rounded-lg px-4 py-2 text-gray-300 transition-colors hover:bg-netflix-gray hover:text-white"
            >
              Cancel
            </button>
            <button
              type="submit"
              disabled={isPending}
              className="rounded-lg bg-netflix-red px-4 py-2 font-semibold text-white transition-colors hover:bg-netflix-red-hover disabled:opacity-50"
            >
              {isPending ? 'Creating...' : 'Create'}
            </button>
          </div>
        </form>
      </div>
    </div>
  )
}

function EditUserModal({ user, onClose }: { user: User; onClose: () => void }) {
  const { mutate: updateUser, isPending } = useUpdateUser()
  const [form, setForm] = useState({
    display_name: user.display_name || '',
    is_admin: user.is_admin,
    password: '',
  })
  const [error, setError] = useState('')

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault()
    setError('')

    const data: { display_name?: string; is_admin?: boolean; password?: string } = {
      display_name: form.display_name,
      is_admin: form.is_admin,
    }

    if (form.password) {
      if (form.password.length < 8) {
        setError('Password must be at least 8 characters')
        return
      }
      data.password = form.password
    }

    updateUser(
      { id: user.id, data },
      {
        onSuccess: () => onClose(),
        onError: (err: Error) => setError(err.message),
      },
    )
  }

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/70 backdrop-blur-sm">
      <div className="w-full max-w-md rounded-xl bg-netflix-dark p-6 shadow-2xl">
        <div className="mb-6 flex items-center justify-between">
          <h2 className="text-xl font-bold text-white">Edit User</h2>
          <button onClick={onClose} className="text-gray-400 hover:text-white">
            <svg className="h-6 w-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                strokeWidth={2}
                d="M6 18L18 6M6 6l12 12"
              />
            </svg>
          </button>
        </div>

        {error && (
          <div className="mb-4 rounded-lg bg-netflix-red/20 p-3 text-sm text-netflix-red">
            {error}
          </div>
        )}

        <form onSubmit={handleSubmit} className="space-y-4">
          <div>
            <label className="mb-2 block text-sm font-medium text-gray-400">Username</label>
            <input
              type="text"
              value={user.username}
              disabled
              className="w-full rounded-lg bg-netflix-black px-4 py-3 text-gray-500"
            />
          </div>

          <div>
            <label className="mb-2 block text-sm font-medium text-gray-400">Display Name</label>
            <input
              type="text"
              value={form.display_name}
              onChange={(e) => setForm({ ...form, display_name: e.target.value })}
              className="w-full rounded-lg bg-netflix-gray px-4 py-3 text-white outline-none ring-1 ring-transparent transition-all focus:ring-netflix-red"
            />
          </div>

          <div>
            <label className="mb-2 block text-sm font-medium text-gray-400">
              New Password (optional)
            </label>
            <input
              type="password"
              value={form.password}
              onChange={(e) => setForm({ ...form, password: e.target.value })}
              placeholder="Leave blank to keep current"
              className="w-full rounded-lg bg-netflix-gray px-4 py-3 text-white placeholder-gray-500 outline-none ring-1 ring-transparent transition-all focus:ring-netflix-red"
            />
          </div>

          <div className="flex items-center gap-3">
            <input
              type="checkbox"
              id="edit_is_admin"
              checked={form.is_admin}
              onChange={(e) => setForm({ ...form, is_admin: e.target.checked })}
              className="h-4 w-4 rounded border-netflix-gray bg-netflix-gray text-netflix-red focus:ring-netflix-red"
            />
            <label htmlFor="edit_is_admin" className="text-sm text-gray-300">
              Administrator
            </label>
          </div>

          <div className="flex justify-end gap-3 pt-4">
            <button
              type="button"
              onClick={onClose}
              className="rounded-lg px-4 py-2 text-gray-300 transition-colors hover:bg-netflix-gray hover:text-white"
            >
              Cancel
            </button>
            <button
              type="submit"
              disabled={isPending}
              className="rounded-lg bg-netflix-red px-4 py-2 font-semibold text-white transition-colors hover:bg-netflix-red-hover disabled:opacity-50"
            >
              {isPending ? 'Saving...' : 'Save'}
            </button>
          </div>
        </form>
      </div>
    </div>
  )
}

function DeleteUserButton({ userId }: { userId: number }) {
  const { mutate: deleteUser, isPending } = useDeleteUser()
  const [showConfirm, setShowConfirm] = useState(false)

  if (showConfirm) {
    return (
      <div className="flex gap-1">
        <button
          onClick={() => deleteUser(userId)}
          disabled={isPending}
          className="rounded bg-netflix-red px-3 py-1.5 text-sm text-white transition-colors hover:bg-netflix-red-hover"
        >
          Confirm
        </button>
        <button
          onClick={() => setShowConfirm(false)}
          className="rounded bg-netflix-gray px-3 py-1.5 text-sm text-gray-300 transition-colors hover:bg-gray-600"
        >
          Cancel
        </button>
      </div>
    )
  }

  return (
    <button
      onClick={() => setShowConfirm(true)}
      className="rounded bg-red-500/20 px-3 py-1.5 text-sm text-red-400 transition-colors hover:bg-red-500/30"
    >
      Delete
    </button>
  )
}
