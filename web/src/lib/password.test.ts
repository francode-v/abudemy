import { describe, expect, it } from 'vitest'
import {
  PASSWORD_MAX_LENGTH,
  PASSWORD_MIN_LENGTH,
  checkPassword,
  passwordByteLength,
  passwordError,
} from './password'

describe('política de contraseñas', () => {
  it('acepta una contraseña con letras y números', () => {
    expect(checkPassword('Segura123').valid).toBe(true)
    expect(passwordError('Segura123')).toBeNull()
  })

  it('exige el mínimo de caracteres', () => {
    const short = 'a1'.repeat(PASSWORD_MIN_LENGTH / 2 - 1)
    expect(checkPassword(short).minLength).toBe(false)
    expect(passwordError(short)).toContain(`al menos ${PASSWORD_MIN_LENGTH}`)
  })

  it('exige al menos una letra y un número', () => {
    expect(passwordError('12345678')).toBe('La contraseña debe incluir al menos una letra.')
    expect(passwordError('abcdefgh')).toBe('La contraseña debe incluir al menos un número.')
  })

  it('acepta letras acentuadas', () => {
    expect(checkPassword('contraseña1').valid).toBe(true)
  })

  it('respeta el límite de 72 bytes de bcrypt', () => {
    const atLimit = `${'a'.repeat(PASSWORD_MAX_LENGTH - 1)}1`
    expect(atLimit).toHaveLength(PASSWORD_MAX_LENGTH)
    expect(checkPassword(atLimit).valid).toBe(true)

    const overLimit = `${'a'.repeat(PASSWORD_MAX_LENGTH)}1`
    expect(checkPassword(overLimit).maxLength).toBe(false)
    expect(passwordError(overLimit)).toBe('La contraseña es demasiado larga.')
  })

  it('cuenta los bytes UTF-8 de los caracteres acentuados', () => {
    const accented = 'ñ'.repeat(36) + '1'
    expect([...accented]).toHaveLength(37)
    expect(passwordByteLength(accented)).toBe(73)
    expect(checkPassword(accented).maxLength).toBe(false)
  })
})
