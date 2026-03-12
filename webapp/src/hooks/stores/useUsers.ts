import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { api } from '@/lib/fetch'
import type { User } from '@/types/api'

interface CreateUserRequest {
  username: string
  password: string
  display_name: string
  is_admin: boolean
}

interface UpdateUserRequest {
  display_name?: string
  is_admin?: boolean
  password?: string
}

// User Management API Functions (Admin only)
const userApi = {
  list: () => api.get<User[]>('/users'),
  create: (data: CreateUserRequest) => api.post<User>('/users', data),
  update: (id: number, data: UpdateUserRequest) => api.patch<User>(`/users/${id}`, data),
  delete: (id: number) => api.delete(`/users/${id}`),
  setLibraryAccess: (id: number, libraryIds: number[]) =>
    api.put<void>(`/users/${id}/library-access`, { library_ids: libraryIds }),
}

// Query Keys
export const userKeys = {
  all: ['users'] as const,
  list: () => [...userKeys.all, 'list'] as const,
  detail: (id: number) => [...userKeys.all, 'detail', id] as const,
}

// React Query Hooks
export function useUsers() {
  return useQuery({
    queryKey: userKeys.list(),
    queryFn: userApi.list,
    staleTime: 5 * 60 * 1000, // 5 minutes
  })
}

export function useCreateUser() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: userApi.create,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: userKeys.list() })
    },
  })
}

export function useUpdateUser() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: ({ id, data }: { id: number; data: UpdateUserRequest }) => userApi.update(id, data),
    onSuccess: (_, variables) => {
      queryClient.invalidateQueries({ queryKey: userKeys.list() })
      queryClient.invalidateQueries({ queryKey: userKeys.detail(variables.id) })
    },
  })
}

export function useDeleteUser() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: userApi.delete,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: userKeys.list() })
    },
  })
}

export function useSetLibraryAccess() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: ({ id, libraryIds }: { id: number; libraryIds: number[] }) =>
      userApi.setLibraryAccess(id, libraryIds),
    onSuccess: (_, variables) => {
      queryClient.invalidateQueries({ queryKey: userKeys.detail(variables.id) })
    },
  })
}
