import type { ApiErrorBody } from './types'

/**
 * ApiError normaliza el envelope de error del backend:
 *   { "error": { "code": "...", "message": "...", "fields": [...] } }
 * Los mensajes de error del backend están en inglés; la UI los muestra tal
 * cual (nunca se inventan) y añade traducciones para los códigos conocidos.
 */
export class ApiError extends Error {
  readonly status: number
  readonly code: string
  readonly fields: { field: string; message: string }[]
  readonly requestId?: string

  constructor(
    status: number,
    code: string,
    message: string,
    fields: { field: string; message: string }[] = [],
    requestId?: string,
  ) {
    super(message)
    this.name = 'ApiError'
    this.status = status
    this.code = code
    this.fields = fields
    this.requestId = requestId
  }

  /** Error de red o de cancelación: no hubo respuesta del servidor. */
  static network(message: string): ApiError {
    return new ApiError(0, 'network_error', message)
  }

  /** El cliente abortó la petición (cambio de vista, navegación). */
  static aborted(): ApiError {
    return new ApiError(0, 'aborted', 'La petición fue cancelada.')
  }

  get isAborted(): boolean {
    return this.code === 'aborted'
  }

  get isUnauthorized(): boolean {
    return this.status === 401
  }

  get isForbidden(): boolean {
    return this.status === 403
  }

  get isNotFound(): boolean {
    return this.status === 404
  }

  get isValidation(): boolean {
    return this.code === 'validation_error' || this.status === 400
  }

  get isConflict(): boolean {
    return this.status === 409
  }

  get isNotImplemented(): boolean {
    return this.status === 501
  }

  /** Mensaje del primer campo inválido, útil para formularios. */
  fieldError(field: string): string | undefined {
    return this.fields.find((f) => f.field === field)?.message
  }
}

export function isApiError(err: unknown): err is ApiError {
  return err instanceof ApiError
}

/** Traduce los códigos de error del backend a un texto legible en español. */
export function humanizeError(err: unknown): string {
  if (!isApiError(err)) {
    return 'Ocurrió un error inesperado.'
  }
  switch (err.code) {
    case 'network_error':
      return 'No se pudo conectar con el servidor de Abudemy. Verifica tu conexión e inténtalo nuevamente.'
    case 'aborted':
      return ''
    case 'unauthorized':
      return 'Tu sesión expiró. Vuelve a iniciar sesión.'
    case 'forbidden':
      return 'No tienes permisos para realizar esta acción.'
    case 'not_found':
      return 'El recurso solicitado no existe o fue eliminado.'
    case 'conflict':
      return 'La operación entra en conflicto con el estado actual.'
    case 'validation_error':
      return err.message || 'Revisa los datos ingresados.'
    case 'not_implemented':
      return 'Esta funcionalidad todavía no está disponible en el backend de Abudemy.'
    case 'internal_error':
      return 'El servidor de Abudemy reportó un error. Inténtalo más tarde.'
    default:
      return err.message || 'Ocurrió un error inesperado.'
  }
}

export function parseErrorBody(status: number, body: unknown, requestId?: string): ApiError {
  const parsed = body as Partial<ApiErrorBody> | undefined
  const error = parsed?.error
  if (error?.code) {
    return new ApiError(status, error.code, error.message || 'Error', error.fields ?? [], requestId)
  }
  return new ApiError(status, `http_${status}`, `La solicitud falló con estado ${status}.`, [], requestId)
}
