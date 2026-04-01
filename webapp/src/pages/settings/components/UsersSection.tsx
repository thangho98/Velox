import { useState } from 'react'
import { LuPlus } from 'react-icons/lu'
import { useAuthStore } from '@/stores/auth'
import { useUsers, useCreateUser, useUpdateUser, useDeleteUser } from '@/hooks/stores/useUsers'
import { SectionHeader, Field, Modal, ErrorMsg, Spinner, inputClass, inputDisabled } from './shared'
import type { User } from '@/types/api'

export function UsersSection() {
  const { user: currentUser } = useAuthStore()
  const { data: users, isLoading } = useUsers()
  const [isCreateOpen, setIsCreateOpen] = useState(false)
  const [editingUser, setEditingUser] = useState<User | null>(null)

  return (
    <div className="max-w-3xl">
      <div className="flex items-center justify-between">
        <SectionHeader
          title="Users"
          description={`${users?.length || 0} ${(users?.length || 0) === 1 ? 'user' : 'users'}`}
        />
        <button
          onClick={() => setIsCreateOpen(true)}
          className="flex items-center gap-2 rounded bg-netflix-red px-4 py-2 text-sm font-semibold text-white transition-colors hover:bg-netflix-red-hover"
        >
          <LuPlus size={16} />
          Add User
        </button>
      </div>

      {isLoading ? (
        <Spinner />
      ) : (
        <div className="mt-6 overflow-hidden rounded-xl bg-netflix-dark">
          <div className="overflow-x-auto">
            <table className="w-full">
              <thead className="border-b border-netflix-gray bg-netflix-black/50">
                <tr>
                  <th className="px-4 py-3 text-left text-xs font-medium text-gray-400">User</th>
                  <th className="px-4 py-3 text-left text-xs font-medium text-gray-400">Role</th>
                  <th className="px-4 py-3 text-left text-xs font-medium text-gray-400">Created</th>
                  <th className="px-4 py-3 text-left text-xs font-medium text-gray-400">Actions</th>
                </tr>
              </thead>
              <tbody>
                {users?.map((u) => (
                  <tr
                    key={u.id}
                    className="border-b border-netflix-gray/50 last:border-b-0 hover:bg-netflix-gray/30"
                  >
                    <td className="px-4 py-3">
                      <div className="flex items-center gap-3">
                        <div className="flex h-8 w-8 items-center justify-center rounded-full bg-netflix-gray text-sm font-medium text-white">
                          {(u.display_name || u.username).charAt(0).toUpperCase()}
                        </div>
                        <div>
                          <p className="text-sm font-medium text-white">
                            {u.display_name || u.username}
                          </p>
                          <p className="text-xs text-gray-500">{u.username}</p>
                        </div>
                      </div>
                    </td>
                    <td className="px-4 py-3">
                      <span
                        className={`rounded px-2 py-0.5 text-xs font-medium ${
                          u.is_admin
                            ? 'bg-purple-500/20 text-purple-400'
                            : 'bg-blue-500/20 text-blue-400'
                        }`}
                      >
                        {u.is_admin ? 'Admin' : 'User'}
                      </span>
                    </td>
                    <td className="px-4 py-3 text-xs text-gray-400">
                      {new Date(u.created_at).toLocaleDateString()}
                    </td>
                    <td className="px-4 py-3">
                      <div className="flex gap-2">
                        <button
                          onClick={() => setEditingUser(u)}
                          className="rounded bg-netflix-gray px-3 py-1 text-xs text-white hover:bg-gray-600"
                        >
                          Edit
                        </button>
                        {currentUser && u.id !== currentUser.id && <DeleteUserBtn userId={u.id} />}
                      </div>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
          {users?.length === 0 && (
            <p className="py-8 text-center text-sm text-gray-400">No users found</p>
          )}
        </div>
      )}

      {isCreateOpen && <CreateUserModal onClose={() => setIsCreateOpen(false)} />}
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
    createUser(form, { onSuccess: onClose, onError: (err: Error) => setError(err.message) })
  }

  return (
    <Modal title="Add User" onClose={onClose}>
      {error && <ErrorMsg>{error}</ErrorMsg>}
      <form onSubmit={handleSubmit} className="space-y-4">
        <Field label="Username *">
          <input
            type="text"
            value={form.username}
            onChange={(e) => setForm({ ...form, username: e.target.value })}
            className={inputClass}
            required
          />
        </Field>
        <Field label="Display Name">
          <input
            type="text"
            value={form.display_name}
            onChange={(e) => setForm({ ...form, display_name: e.target.value })}
            className={inputClass}
          />
        </Field>
        <Field label="Password *">
          <input
            type="password"
            value={form.password}
            onChange={(e) => setForm({ ...form, password: e.target.value })}
            className={inputClass}
            required
            minLength={8}
          />
        </Field>
        <div className="flex items-center gap-3">
          <input
            type="checkbox"
            id="new_is_admin"
            checked={form.is_admin}
            onChange={(e) => setForm({ ...form, is_admin: e.target.checked })}
            className="h-4 w-4 rounded"
          />
          <label htmlFor="new_is_admin" className="text-sm text-gray-300">
            Administrator
          </label>
        </div>
        <div className="flex justify-end gap-3 pt-2">
          <button
            type="button"
            onClick={onClose}
            className="rounded px-4 py-2 text-gray-300 hover:bg-netflix-gray hover:text-white"
          >
            Cancel
          </button>
          <button
            type="submit"
            disabled={isPending}
            className="rounded bg-netflix-red px-4 py-2 font-semibold text-white hover:bg-netflix-red-hover disabled:opacity-50"
          >
            {isPending ? 'Creating...' : 'Create'}
          </button>
        </div>
      </form>
    </Modal>
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
      { onSuccess: onClose, onError: (err: Error) => setError(err.message) },
    )
  }

  return (
    <Modal title="Edit User" onClose={onClose}>
      {error && <ErrorMsg>{error}</ErrorMsg>}
      <form onSubmit={handleSubmit} className="space-y-4">
        <Field label="Username">
          <input type="text" value={user.username} disabled className={inputDisabled} />
        </Field>
        <Field label="Display Name">
          <input
            type="text"
            value={form.display_name}
            onChange={(e) => setForm({ ...form, display_name: e.target.value })}
            className={inputClass}
          />
        </Field>
        <Field label="New Password (optional)">
          <input
            type="password"
            value={form.password}
            onChange={(e) => setForm({ ...form, password: e.target.value })}
            placeholder="Leave blank to keep current"
            className={inputClass}
          />
        </Field>
        <div className="flex items-center gap-3">
          <input
            type="checkbox"
            id="edit_is_admin"
            checked={form.is_admin}
            onChange={(e) => setForm({ ...form, is_admin: e.target.checked })}
            className="h-4 w-4 rounded"
          />
          <label htmlFor="edit_is_admin" className="text-sm text-gray-300">
            Administrator
          </label>
        </div>
        <div className="flex justify-end gap-3 pt-2">
          <button
            type="button"
            onClick={onClose}
            className="rounded px-4 py-2 text-gray-300 hover:bg-netflix-gray hover:text-white"
          >
            Cancel
          </button>
          <button
            type="submit"
            disabled={isPending}
            className="rounded bg-netflix-red px-4 py-2 font-semibold text-white hover:bg-netflix-red-hover disabled:opacity-50"
          >
            {isPending ? 'Saving...' : 'Save'}
          </button>
        </div>
      </form>
    </Modal>
  )
}

function DeleteUserBtn({ userId }: { userId: number }) {
  const { mutate: deleteUser, isPending } = useDeleteUser()
  const [confirm, setConfirm] = useState(false)

  if (confirm) {
    return (
      <div className="flex gap-1">
        <button
          onClick={() => deleteUser(userId)}
          disabled={isPending}
          className="rounded bg-netflix-red px-2.5 py-1 text-xs text-white hover:bg-netflix-red-hover"
        >
          Confirm
        </button>
        <button
          onClick={() => setConfirm(false)}
          className="rounded bg-netflix-gray px-2.5 py-1 text-xs text-gray-300 hover:bg-gray-600"
        >
          Cancel
        </button>
      </div>
    )
  }

  return (
    <button
      onClick={() => setConfirm(true)}
      className="rounded bg-red-500/20 px-3 py-1 text-xs text-red-400 hover:bg-red-500/30"
    >
      Delete
    </button>
  )
}
