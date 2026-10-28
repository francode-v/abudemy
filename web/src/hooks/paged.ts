import { useQueries } from '@tanstack/react-query'
import { api } from '@/api'
import type { Activity, Service } from '@/api'
import type { ActivityQuery, ServiceQuery } from '@/api/endpoints'
import { queryKeys, useActivities, useServices } from './queries'

/** Máximo de páginas recolectadas para capas de mapa / agregaciones locales. */
export const MAX_PAGES = 5
const PER_PAGE = 100

export interface Collected<T> {
  items: T[]
  total: number
  loaded: number
  truncated: boolean
  isLoading: boolean
  isError: boolean
  error: unknown
  refetch: () => void
}

/**
 * Recolecta varias páginas de servicios para alimentar el mapa y agregaciones
 * locales. Está acotado a `MAX_PAGES` para no degradar el rendimiento.
 */
export function useAllServices(filters: Omit<ServiceQuery, 'page' | 'per_page'>, enabled = true): Collected<Service> {
  const base = { ...filters, per_page: PER_PAGE }
  const first = useServices({ ...base, page: 1 }, enabled)
  const totalPages = Math.min(first.data?.pagination.total_pages ?? 1, MAX_PAGES)

  const rest = useQueries({
    queries: Array.from({ length: Math.max(0, totalPages - 1) }, (_, index) => {
      const page = index + 2
      const query = { ...base, page }
      return {
        queryKey: queryKeys.services(query),
        queryFn: () => api.services.list(query),
        enabled,
        staleTime: 30_000,
      }
    }),
  })

  const items = [
    ...(first.data?.items ?? []),
    ...rest.flatMap((result) => result.data?.items ?? []),
  ]

  const total = first.data?.pagination.total ?? items.length
  const serversTotalPages = first.data?.pagination.total_pages ?? 1

  return {
    items,
    total,
    loaded: items.length,
    truncated: serversTotalPages > MAX_PAGES,
    isLoading: first.isLoading || rest.some((result) => result.isLoading),
    isError: first.isError || rest.some((result) => result.isError),
    error: first.error ?? rest.find((result) => result.error)?.error,
    refetch: () => {
      void first.refetch()
      for (const result of rest) void result.refetch()
    },
  }
}

export function useAllActivities(
  filters: Omit<ActivityQuery, 'page' | 'per_page'>,
  enabled = true,
): Collected<Activity> {
  const base = { ...filters, per_page: PER_PAGE }
  const first = useActivities({ ...base, page: 1 }, enabled)
  const totalPages = Math.min(first.data?.pagination.total_pages ?? 1, MAX_PAGES)

  const rest = useQueries({
    queries: Array.from({ length: Math.max(0, totalPages - 1) }, (_, index) => {
      const page = index + 2
      const query = { ...base, page }
      return {
        queryKey: queryKeys.activities(query),
        queryFn: () => api.activities.list(query),
        enabled,
        staleTime: 30_000,
      }
    }),
  })

  const items = [
    ...(first.data?.items ?? []),
    ...rest.flatMap((result) => result.data?.items ?? []),
  ]

  const total = first.data?.pagination.total ?? items.length
  const serversTotalPages = first.data?.pagination.total_pages ?? 1

  return {
    items,
    total,
    loaded: items.length,
    truncated: serversTotalPages > MAX_PAGES,
    isLoading: first.isLoading || rest.some((result) => result.isLoading),
    isError: first.isError || rest.some((result) => result.isError),
    error: first.error ?? rest.find((result) => result.error)?.error,
    refetch: () => {
      void first.refetch()
      for (const result of rest) void result.refetch()
    },
  }
}
