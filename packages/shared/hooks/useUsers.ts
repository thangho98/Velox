/**
 * User management hooks
 */

import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { api } from '../api'
import type { User } from '../types/auth'
import type { CreateUserRequest, UpdateUserRequest } from '../types/admin'

// --- Query Keys ---

export const userKeys = {
  all: ['users'] as const,
  list: () => [...userKeys.all, 'list'] as const,
  detail: (id: number) => [...userKeys.all, 'detail', id] as const,
  libraryAccess: (id: number) => [...userKeys.all, 'library-access', id] as const,
}

// --- Hooks ---

export function useUsers() {
  return useQuery({
    queryKey: userKeys.list(),
    queryFn: () => api.get<User[]>('/admin/users'),
  })
}

export function useCreateUser() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (data: CreateUserRequest) => api.post<User>('/admin/users', data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: userKeys.list() })
    },
  })
}

export function useUpdateUser() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: ({ id, data }: { id: number; data: UpdateUserRequest }) =>
      api.put<User>(`/admin/users/${id}`, data),
    onSuccess: (_, variables) => {
      queryClient.invalidateQueries({ queryKey: userKeys.list() })
      queryClient.invalidateQueries({ queryKey: userKeys.detail(variables.id) })
    },
  })
}

export function useDeleteUser() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (id: number) => api.delete(`/admin/users/${id}`),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: userKeys.list() })
    },
  })
}

export function useLibraryAccess(userId: number) {
  return useQuery({
    queryKey: userKeys.libraryAccess(userId),
    queryFn: () => api.get<{ library_id: number; access_type: string }[]>(`/admin/users/${userId}/libraries`),
  })
}

export function useSetLibraryAccess() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: ({ userId, libraryId, accessType }: { userId: number; libraryId: number; accessType: string }) =>
      api.put(`/admin/users/${userId}/libraries/${libraryId}`, { access_type: accessType }),
    onSuccess: (_, variables) => {
      queryClient.invalidateQueries({ queryKey: userKeys.libraryAccess(variables.userId) })
    },
  })
}
