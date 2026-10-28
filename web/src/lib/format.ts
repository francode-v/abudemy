const dateTimeFormatter = new Intl.DateTimeFormat('es-CL', {
  dateStyle: 'medium',
  timeStyle: 'short',
})

const dateFormatter = new Intl.DateTimeFormat('es-CL', { dateStyle: 'medium' })
const timeFormatter = new Intl.DateTimeFormat('es-CL', { timeStyle: 'short' })
const numberFormatter = new Intl.NumberFormat('es-CL')
const currencyFormatter = new Intl.NumberFormat('es-CL', {
  style: 'currency',
  currency: 'CLP',
  maximumFractionDigits: 0,
})

export function formatDateTime(value?: string | null): string {
  if (!value) return '—'
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return '—'
  return dateTimeFormatter.format(date)
}

export function formatDate(value?: string | null): string {
  if (!value) return '—'
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return '—'
  return dateFormatter.format(date)
}

export function formatTime(value?: string | null): string {
  if (!value) return '—'
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return '—'
  return timeFormatter.format(date)
}

export function formatNumber(value?: number | null): string {
  if (value === undefined || value === null || Number.isNaN(value)) return '—'
  return numberFormatter.format(value)
}

export function formatCurrency(value?: number | null): string {
  if (value === undefined || value === null || Number.isNaN(value)) return '—'
  return currencyFormatter.format(value)
}

export function formatPercent(value?: number | null, digits = 0): string {
  if (value === undefined || value === null || Number.isNaN(value)) return '—'
  return `${(value * 100).toFixed(digits)}%`
}

export function formatCoordinate(value?: number | null): string {
  if (value === undefined || value === null) return '—'
  return value.toFixed(5)
}

/** Convierte una fecha ISO en el valor `YYYY-MM-DD` para <input type="date">. */
export function toDateInputValue(value?: string | null): string {
  if (!value) return ''
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return ''
  return date.toISOString().slice(0, 10)
}

/** Convierte una fecha ISO en el valor `YYYY-MM-DDTHH:mm` para <input datetime-local>. */
export function toDateTimeInputValue(value?: string | null): string {
  if (!value) return ''
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return ''
  const pad = (n: number) => String(n).padStart(2, '0')
  return `${date.getFullYear()}-${pad(date.getMonth() + 1)}-${pad(date.getDate())}T${pad(date.getHours())}:${pad(date.getMinutes())}`
}

/** Texto relativo simple ("hace 3 días"), usado para frescura de datos. */
export function relativeTime(value?: string | number | Date | null): string {
  if (value === undefined || value === null) return '—'
  const date = value instanceof Date ? value : new Date(value)
  if (Number.isNaN(date.getTime())) return '—'
  const diff = Date.now() - date.getTime()
  const minutes = Math.round(diff / 60_000)
  if (Math.abs(minutes) < 1) return 'hace instantes'
  if (Math.abs(minutes) < 60) return `hace ${minutes} min`
  const hours = Math.round(minutes / 60)
  if (Math.abs(hours) < 24) return `hace ${hours} h`
  const days = Math.round(hours / 24)
  if (Math.abs(days) < 30) return `hace ${days} d`
  return formatDate(date.toISOString())
}

export function pluralize(count: number, singular: string, plural: string): string {
  return count === 1 ? `${count} ${singular}` : `${count} ${plural}`
}

export function initials(fullName?: string | null): string {
  if (!fullName) return '—'
  return fullName
    .split(/\s+/)
    .filter(Boolean)
    .slice(0, 2)
    .map((part) => part[0]?.toUpperCase() ?? '')
    .join('')
}

/**
 * Edad en años cumplidos a partir de una fecha de nacimiento `YYYY-MM-DD`.
 * Se parsea manualmente para no desplazar un día por la zona horaria.
 */
export function ageFrom(birthDate?: string | null): number | null {
  if (!birthDate) return null
  const match = /^(\d{4})-(\d{2})-(\d{2})/.exec(birthDate)
  if (!match) return null
  const year = Number(match[1])
  const month = Number(match[2])
  const day = Number(match[3])
  const today = new Date()
  let age = today.getFullYear() - year
  const monthDiff = today.getMonth() + 1 - month
  if (monthDiff < 0 || (monthDiff === 0 && today.getDate() < day)) age -= 1
  return age >= 0 && age <= 130 ? age : null
}
