import type { User } from '@/api'
import { ageFrom } from './format'

/** Tramos de edad usados en el informe de personas registradas. */
export const AGE_BUCKETS: { label: string; min: number; max: number }[] = [
  { label: 'Menos de 60 años', min: 0, max: 59 },
  { label: '60 a 69 años', min: 60, max: 69 },
  { label: '70 a 79 años', min: 70, max: 79 },
  { label: '80 a 89 años', min: 80, max: 89 },
  { label: '90 años o más', min: 90, max: 200 },
]

/** Agrupa filas por una etiqueta y ordena de mayor a menor. */
export function distribution(rows: User[], key: (row: User) => string): [string, number][] {
  const counts = new Map<string, number>()
  for (const row of rows) {
    const label = key(row).trim() || 'Sin registrar'
    counts.set(label, (counts.get(label) ?? 0) + 1)
  }
  return [...counts.entries()].sort((a, b) => b[1] - a[1] || a[0].localeCompare(b[0], 'es'))
}

/** Cuenta personas por tramo de edad; suma un tramo aparte para quienes no tienen fecha. */
export function ageDistribution(rows: User[]): [string, number][] {
  const counts = AGE_BUCKETS.map(() => 0)
  let unknown = 0
  for (const row of rows) {
    const age = ageFrom(row.birth_date)
    if (age === null) {
      unknown += 1
      continue
    }
    const index = AGE_BUCKETS.findIndex((bucket) => age >= bucket.min && age <= bucket.max)
    if (index >= 0) counts[index] += 1
  }
  const result: [string, number][] = AGE_BUCKETS.map((bucket, index) => [bucket.label, counts[index]])
  if (unknown > 0) result.push(['Sin fecha de nacimiento', unknown])
  return result
}
