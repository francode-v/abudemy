/** Política de contraseñas compartida por registro, restablecimiento y perfil. */
export const PASSWORD_MIN_LENGTH = 8
/** Límite en caracteres que se muestra en la interfaz. */
export const PASSWORD_MAX_LENGTH = 72
/** bcrypt (backend) rechaza entradas de más de 72 bytes. */
const PASSWORD_MAX_BYTES = 72

export const PASSWORD_REQUIREMENTS =
  'Entre 8 y 72 caracteres, con al menos una letra y un número.'

const encoder = new TextEncoder()

export interface PasswordChecks {
  minLength: boolean
  maxLength: boolean
  hasLetter: boolean
  hasDigit: boolean
  valid: boolean
}

/** Longitud en bytes UTF-8, que es lo que valida bcrypt en el backend. */
export function passwordByteLength(value: string): number {
  return encoder.encode(value).length
}

export function checkPassword(value: string): PasswordChecks {
  const length = [...value].length
  const minLength = length >= PASSWORD_MIN_LENGTH
  // El backend usa bcrypt, que solo acepta hasta 72 bytes. Comprobamos ambos
  // límites para que una contraseña válida en la web nunca falle al guardarse.
  const maxLength = length <= PASSWORD_MAX_LENGTH && passwordByteLength(value) <= PASSWORD_MAX_BYTES
  const hasLetter = /\p{L}/u.test(value)
  const hasDigit = /\d/.test(value)
  return {
    minLength,
    maxLength,
    hasLetter,
    hasDigit,
    valid: minLength && maxLength && hasLetter && hasDigit,
  }
}

/** Devuelve el primer problema de la contraseña, o `null` si cumple la política. */
export function passwordError(value: string): string | null {
  if (!value) return 'La contraseña es obligatoria.'
  const checks = checkPassword(value)
  if (!checks.minLength) return `La contraseña debe tener al menos ${PASSWORD_MIN_LENGTH} caracteres.`
  if (!checks.maxLength) return 'La contraseña es demasiado larga.'
  if (!checks.hasLetter) return 'La contraseña debe incluir al menos una letra.'
  if (!checks.hasDigit) return 'La contraseña debe incluir al menos un número.'
  return null
}
