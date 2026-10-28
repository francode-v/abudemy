import { describe, expect, it } from 'vitest'
import type { User } from '@/api'
import { ageDistribution, distribution } from './people'

function birthDateForAge(age: number): string {
  const date = new Date()
  date.setDate(date.getDate() - 1)
  date.setFullYear(date.getFullYear() - age)
  const pad = (n: number) => String(n).padStart(2, '0')
  return `${date.getFullYear()}-${pad(date.getMonth() + 1)}-${pad(date.getDate())}`
}

const person = (fields: Partial<Pick<User, 'birth_date' | 'comuna' | 'sex'>>): User =>
  fields as unknown as User

describe('distribution', () => {
  it('agrupa, ordena de mayor a menor y usa una etiqueta para vacíos', () => {
    const rows = [
      person({ comuna: 'Concepción' }),
      person({ comuna: 'Concepción' }),
      person({ comuna: 'Talcahuano' }),
      person({ comuna: '' }),
    ]
    expect(distribution(rows, (row) => row.comuna ?? '')).toEqual([
      ['Concepción', 2],
      ['Sin registrar', 1],
      ['Talcahuano', 1],
    ])
  })

  it('ordena alfabéticamente cuando hay empate', () => {
    const rows = [person({ comuna: 'Talcahuano' }), person({ comuna: 'Concepción' })]
    expect(distribution(rows, (row) => row.comuna ?? '').map(([label]) => label)).toEqual([
      'Concepción',
      'Talcahuano',
    ])
  })
})

describe('ageDistribution', () => {
  it('cuenta los tramos etarios y separa a quienes no tienen fecha', () => {
    const rows = [
      person({ birth_date: birthDateForAge(35) }),
      person({ birth_date: birthDateForAge(60) }),
      person({ birth_date: birthDateForAge(69) }),
      person({ birth_date: birthDateForAge(75) }),
      person({ birth_date: birthDateForAge(89) }),
      person({ birth_date: birthDateForAge(90) }),
      person({}),
      person({ birth_date: 'no-es-fecha' }),
    ]
    expect(ageDistribution(rows)).toEqual([
      ['Menos de 60 años', 1],
      ['60 a 69 años', 2],
      ['70 a 79 años', 1],
      ['80 a 89 años', 1],
      ['90 años o más', 1],
      ['Sin fecha de nacimiento', 2],
    ])
  })

  it('mantiene los tramos vacíos en cero', () => {
    const result = ageDistribution([person({ birth_date: birthDateForAge(72) })])
    expect(result).toEqual([
      ['Menos de 60 años', 0],
      ['60 a 69 años', 0],
      ['70 a 79 años', 1],
      ['80 a 89 años', 0],
      ['90 años o más', 0],
    ])
  })
})
