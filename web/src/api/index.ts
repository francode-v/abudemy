import { ApiClient } from './client'
import { demoAdapter } from './demo'
import { Api } from './endpoints'
import { tokenStore } from './tokenStore'

const env = import.meta.env

export const DEMO_MODE = env.VITE_DEMO_MODE === 'true'
export const API_BASE_URL = (env.VITE_API_BASE_URL ?? '').replace(/\/+$/, '')

type ExpiredListener = () => void
const expiredListeners = new Set<ExpiredListener>()

/** Se notifica cuando el refresh token deja de ser válido. */
export const sessionEvents = {
  onExpired(listener: ExpiredListener): () => void {
    expiredListeners.add(listener)
    return () => expiredListeners.delete(listener)
  },
  emitExpired(): void {
    for (const listener of expiredListeners) listener()
  },
}

export const apiClient = new ApiClient({
  baseUrl: API_BASE_URL,
  tokens: tokenStore,
  adapter: DEMO_MODE ? demoAdapter : null,
  onSessionExpired: () => {
    tokenStore.clear()
    sessionEvents.emitExpired()
  },
})

export const api = new Api(apiClient)

export { Api } from './endpoints'
export { ApiError, humanizeError, isApiError } from './errors'
export { tokenStore } from './tokenStore'
export * from './types'
