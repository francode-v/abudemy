import type { TokenPort } from './client'
import type { TokenPair } from './types'

const ACCESS_KEY = 'abudemy.access_token'
const REFRESH_KEY = 'abudemy.refresh_token'
const SESSION_KEY = 'abudemy.session_id'

type Listener = () => void

/**
 * Persistencia de la sesión en el navegador.
 *
 * Nota de seguridad: el backend emite un *access token* de vida corta (15 min
 * por defecto) y un *refresh token* rotativo que es el que persiste. El access
 * token se guarda para sobrevivir a recargas; en despliegues con requisitos
 * estrictos puede moverse a memoria. La autorización real siempre la aplica el
 * backend, nunca este cliente.
 */
class BrowserTokenStore implements TokenPort {
  private listeners = new Set<Listener>()

  getAccessToken(): string | null {
    return safeGet(ACCESS_KEY)
  }

  getRefreshToken(): string | null {
    return safeGet(REFRESH_KEY)
  }

  getSessionId(): string | null {
    return safeGet(SESSION_KEY)
  }

  setTokens(pair: TokenPair): void {
    safeSet(ACCESS_KEY, pair.access_token)
    safeSet(REFRESH_KEY, pair.refresh_token)
    if (pair.session_id) safeSet(SESSION_KEY, pair.session_id)
    this.emit()
  }

  clear(): void {
    safeRemove(ACCESS_KEY)
    safeRemove(REFRESH_KEY)
    safeRemove(SESSION_KEY)
    this.emit()
  }

  hasSession(): boolean {
    return Boolean(this.getRefreshToken())
  }

  subscribe(listener: Listener): () => void {
    this.listeners.add(listener)
    return () => this.listeners.delete(listener)
  }

  private emit(): void {
    for (const listener of this.listeners) listener()
  }
}

function safeGet(key: string): string | null {
  try {
    return window.localStorage.getItem(key)
  } catch {
    return null
  }
}

function safeSet(key: string, value: string): void {
  try {
    window.localStorage.setItem(key, value)
  } catch {
    /* almacenamiento no disponible (modo privado): la sesión vive en memoria */
  }
}

function safeRemove(key: string): void {
  try {
    window.localStorage.removeItem(key)
  } catch {
    /* ignorar */
  }
}

export const tokenStore = new BrowserTokenStore()
