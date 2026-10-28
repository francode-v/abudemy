import { afterEach, describe, expect, it, vi } from 'vitest'
import {
  ageFrom,
  formatNumber,
  formatPercent,
  formatDate,
  initials,
  pluralize,
  toDateInputValue,
} from './format'

describe('formatNumber', () => {
  it('formatea con separadores locales', () => {
    expect(formatNumber(1234)).toBe('1.234')
    expect(formatNumber(0)).toBe('0')
  })

  it('devuelve un guion para valores ausentes', () => {
    expect(formatNumber(undefined)).toBe('—')
    expect(formatNumber(null)).toBe('—')
    expect(formatNumber(Number.NaN)).toBe('—')
  })
})

describe('formatPercent', () => {
  it('convierte una razón en porcentaje', () => {
    expect(formatPercent(0.5, 1)).toBe('50.0%')
    expect(formatPercent(0.256)).toBe('26%')
  })

  it('devuelve un guion para valores ausentes', () => {
    expect(formatPercent(undefined)).toBe('—')
  })
})

describe('formatDate', () => {
  it('devuelve un guion para fechas inválidas o ausentes', () => {
    expect(formatDate(undefined)).toBe('—')
    expect(formatDate('no-es-fecha')).toBe('—')
  })

  it('formatea una fecha ISO válida', () => {
    expect(formatDate('2024-03-05T10:00:00Z')).toMatch(/2024/)
  })
})

describe('toDateInputValue', () => {
  it('convierte una fecha ISO a YYYY-MM-DD', () => {
    expect(toDateInputValue('2024-03-05T10:00:00Z')).toBe('2024-03-05')
  })

  it('devuelve cadena vacía para valores ausentes', () => {
    expect(toDateInputValue(undefined)).toBe('')
    expect(toDateInputValue('invalida')).toBe('')
  })
})

describe('initials', () => {
  it('toma las iniciales de las dos primeras palabras', () => {
    expect(initials('Ana María Pérez')).toBe('AM')
    expect(initials('Pedro')).toBe('P')
  })

  it('devuelve un guion cuando no hay nombre', () => {
    expect(initials(undefined)).toBe('—')
  })
})

describe('pluralize', () => {
  it('usa el singular para uno y el plural para el resto', () => {
    expect(pluralize(1, 'servicio', 'servicios')).toBe('1 servicio')
    expect(pluralize(3, 'servicio', 'servicios')).toBe('3 servicios')
  })
})

describe('ageFrom', () => {
  afterEach(() => {
    vi.useRealTimers()
  })

  it('calcula los años cumplidos', () => {
    const year = new Date().getFullYear()
    expect(ageFrom(`${year - 40}-01-01`)).toBe(40)
  })

  it('resta un año cuando el cumpleaños aún no llega', () => {
    vi.useFakeTimers()
    vi.setSystemTime(new Date('2026-06-15T12:00:00'))
    expect(ageFrom('1996-08-20')).toBe(29)
    expect(ageFrom('1996-05-20')).toBe(30)
  })

  it('devuelve null para fechas ausentes o inválidas', () => {
    expect(ageFrom(undefined)).toBeNull()
    expect(ageFrom('')).toBeNull()
    expect(ageFrom('no-es-fecha')).toBeNull()
  })
})
