/**
 * Cloud storage hooks — Plan W.
 *
 * Provider drivers are listed at GET /api/admin/cloud/drivers; provider rows
 * live at /api/admin/cloud/providers. These hooks wrap the admin UI flow.
 */

import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { api } from '../api'
import type {
  CloudDriverMeta,
  CreateProviderPasswordRequest,
  StorageProvider,
  ValidateFolderURLResponse,
} from '../types/cloud'

const cloudKeys = {
  all: ['cloud'] as const,
  drivers: () => [...cloudKeys.all, 'drivers'] as const,
  providers: () => [...cloudKeys.all, 'providers'] as const,
}

export function useCloudDrivers() {
  return useQuery({
    queryKey: cloudKeys.drivers(),
    queryFn: () => api.get<CloudDriverMeta[]>('/admin/cloud/drivers'),
    // Drivers are registered at startup; almost never change within a session.
    staleTime: 5 * 60 * 1000,
  })
}

export function useStorageProviders() {
  return useQuery({
    queryKey: cloudKeys.providers(),
    queryFn: () => api.get<StorageProvider[]>('/admin/cloud/providers'),
    // Health derives from token_expires_at; refetch once a minute so
    // "expiring soon" badges flip live without a page reload.
    refetchInterval: 60 * 1000,
  })
}

export function useCreateCloudProvider() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (payload: CreateProviderPasswordRequest) =>
      api.post<StorageProvider>('/admin/cloud/providers', payload),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: cloudKeys.providers() })
    },
  })
}

export function useRefreshCloudProvider() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (id: number) =>
      api.post<StorageProvider>(`/admin/cloud/providers/${id}/refresh`),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: cloudKeys.providers() })
    },
  })
}

export function useDeleteCloudProvider() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (id: number) =>
      api.delete<void>(`/admin/cloud/providers/${id}`),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: cloudKeys.providers() })
    },
  })
}

export function useValidateCloudFolderURL() {
  return useMutation({
    mutationFn: (args: { providerID: number; url: string }) =>
      api.post<ValidateFolderURLResponse>(
        `/admin/cloud/providers/${args.providerID}/validate-url`,
        { url: args.url },
      ),
  })
}
