import { ApiError, parseErrorBody } from './errors'
import type { TokenPair } from './types'

/** Puerto que el cliente usa para leer/escribir los tokens de sesión. */
export interface TokenPort {
  getAccessToken(): string | null
  getRefreshToken(): string | null
  setTokens(pair: TokenPair): void
  clear(): void
}

export interface RequestOptions {
  /** Parámetros de consulta; los valores vacíos/nulos se omiten. */
  query?: Record<string, string | number | boolean | null | undefined>
  /** Cuerpo JSON. */
  body?: unknown
  signal?: AbortSignal
  /** Tipo de respuesta esperada. Por defecto `json`. */
  responseType?: 'json' | 'text' | 'blob' | 'none'
  /** Permite reintentar manualmente la petición sin refresh automático. */
  skipAuthRefresh?: boolean
  /** Encabezados adicionales. */
  headers?: Record<string, string>
  /** Tiempo máximo en ms. Por defecto 20s. */
  timeoutMs?: number
}

export interface ApiClientConfig {
  baseUrl: string
  tokens: TokenPort
  /** Se invoca cuando la sesión deja de ser válida (refresh fallido). */
  onSessionExpired?: () => void
  /** Adaptador opcional (modo demostración con datos ficticios). */
  adapter?: RequestAdapter | null
  /** Registra cada petición (diagnóstico). */
  onRequest?: (info: { method: string; url: string }) => void
}

/**
 * Adaptador de transporte alternativo. Solo se usa en el "modo demostración",
 * que devuelve datos ficticios claramente identificados en la interfaz.
 */
export type RequestAdapter = (
  method: string,
  path: string,
  options: RequestOptions,
) => Promise<unknown>

const DEFAULT_TIMEOUT = 20_000

export class ApiClient {
  private readonly cfg: ApiClientConfig
  private refreshPromise: Promise<boolean> | null = null

  constructor(cfg: ApiClientConfig) {
    this.cfg = cfg
  }

  get baseUrl(): string {
    return this.cfg.baseUrl
  }

  get demoMode(): boolean {
    return Boolean(this.cfg.adapter)
  }

  buildUrl(path: string, query?: RequestOptions['query']): string {
    const base = this.cfg.baseUrl.replace(/\/+$/, '')
    const url = new URL(`${base}${path}`, base ? undefined : window.location.origin)
    if (query) {
      for (const [key, value] of Object.entries(query)) {
        if (value === undefined || value === null || value === '') continue
        url.searchParams.set(key, String(value))
      }
    }
    return base ? url.toString() : `${url.pathname}${url.search}`
  }

  async request<T>(method: string, path: string, options: RequestOptions = {}): Promise<T> {
    if (this.cfg.adapter) {
      return (await this.cfg.adapter(method, path, options)) as T
    }
    return this.execute<T>(method, path, options, false)
  }

  private async execute<T>(
    method: string,
    path: string,
    options: RequestOptions,
    isRetry: boolean,
  ): Promise<T> {
    const url = this.buildUrl(path, options.query)
    const controller = new AbortController()
    const timeout = window.setTimeout(() => controller.abort(), options.timeoutMs ?? DEFAULT_TIMEOUT)
    const signal = options.signal
      ? AbortSignal.any([options.signal, controller.signal])
      : controller.signal

    const headers: Record<string, string> = { Accept: 'application/json', ...options.headers }
    const token = this.cfg.tokens.getAccessToken()
    if (token) headers.Authorization = `Bearer ${token}`

    let body: BodyInit | undefined
    if (options.body !== undefined) {
      headers['Content-Type'] = 'application/json'
      body = JSON.stringify(options.body)
    }

    this.cfg.onRequest?.({ method, url })

    let response: Response
    try {
      response = await fetch(url, { method, headers, body, signal, credentials: 'omit' })
    } catch (err) {
      window.clearTimeout(timeout)
      if (options.signal?.aborted) throw ApiError.aborted()
      if ((err as Error)?.name === 'AbortError') {
        throw ApiError.network('La solicitud tardó demasiado. Inténtalo nuevamente.')
      }
      throw ApiError.network('No se pudo conectar con el servidor de Abudemy.')
    } finally {
      window.clearTimeout(timeout)
    }

    // 401 -> intenta refrescar el token una sola vez y repite la petición.
    if (response.status === 401 && !isRetry && !options.skipAuthRefresh && this.hasRefreshToken()) {
      const refreshed = await this.refreshTokens()
      if (refreshed) {
        return this.execute<T>(method, path, options, true)
      }
      this.cfg.onSessionExpired?.()
    }

    const requestId = response.headers.get('X-Request-Id') ?? undefined

    if (!response.ok) {
      const parsed = await this.readErrorBody(response)
      throw parseErrorBody(response.status, parsed, requestId)
    }

    const type = options.responseType ?? 'json'
    if (type === 'none' || response.status === 204) return undefined as T
    if (type === 'blob') return (await response.blob()) as T
    if (type === 'text') return (await response.text()) as T
    if (response.status === 204) return undefined as T
    const text = await response.text()
    if (!text) return undefined as T
    return JSON.parse(text) as T
  }

  private async readErrorBody(response: Response): Promise<unknown> {
    try {
      const text = await response.text()
      return text ? JSON.parse(text) : undefined
    } catch {
      return undefined
    }
  }

  private hasRefreshToken(): boolean {
    return Boolean(this.cfg.tokens.getRefreshToken())
  }

  /** Refresca los tokens (single-flight: una sola rotación concurrente). */
  refreshTokens(): Promise<boolean> {
    if (this.refreshPromise) return this.refreshPromise
    const refreshToken = this.cfg.tokens.getRefreshToken()
    if (!refreshToken) return Promise.resolve(false)

    this.refreshPromise = (async () => {
      try {
        const url = this.buildUrl('/api/v1/auth/refresh')
        const response = await fetch(url, {
          method: 'POST',
          headers: { 'Content-Type': 'application/json', Accept: 'application/json' },
          body: JSON.stringify({ refresh_token: refreshToken }),
          credentials: 'omit',
        })
        if (!response.ok) {
          this.cfg.tokens.clear()
          return false
        }
        const pair = (await response.json()) as TokenPair
        this.cfg.tokens.setTokens(pair)
        return true
      } catch {
        return false
      } finally {
        this.refreshPromise = null
      }
    })()

    return this.refreshPromise
  }

  get<T>(path: string, options?: RequestOptions) {
    return this.request<T>('GET', path, options)
  }
  post<T>(path: string, body?: unknown, options?: RequestOptions) {
    return this.request<T>('POST', path, { ...options, body })
  }
  put<T>(path: string, body?: unknown, options?: RequestOptions) {
    return this.request<T>('PUT', path, { ...options, body })
  }
  patch<T>(path: string, body?: unknown, options?: RequestOptions) {
    return this.request<T>('PATCH', path, { ...options, body })
  }
  del<T>(path: string, options?: RequestOptions) {
    return this.request<T>('DELETE', path, options)
  }
}
