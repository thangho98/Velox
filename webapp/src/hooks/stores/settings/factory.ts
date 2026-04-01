import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { api } from '@/lib/fetch'

const settingsKeys = {
  all: ['settings'] as const,
  key: (slug: string) => [...settingsKeys.all, slug] as const,
}

export { settingsKeys }

/**
 * Creates a pair of [useGet, useUpdate] hooks for an admin settings endpoint.
 *
 * Usage:
 *   const [useTMDbSettings, useUpdateTMDbSettings] = createSettingsHooks<TMDbSettings, TMDbUpdateRequest>('tmdb')
 */
export function createSettingsHooks<TData, TUpdate = Partial<TData>>(slug: string) {
  const queryKey = settingsKeys.key(slug)

  function useGet() {
    return useQuery({
      queryKey,
      queryFn: () => api.get<TData>(`/admin/settings/${slug}`),
      staleTime: 5 * 60 * 1000,
    })
  }

  function useUpdate() {
    const queryClient = useQueryClient()
    return useMutation({
      mutationFn: (data: TUpdate) => api.put<TData>(`/admin/settings/${slug}`, data),
      onSuccess: () => {
        queryClient.invalidateQueries({ queryKey })
      },
    })
  }

  return [useGet, useUpdate] as const
}
