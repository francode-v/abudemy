import { describe, expect, it } from 'vitest'
import { buildCsv } from './csv'

interface Row {
  name: string
  value: number | null
}

describe('buildCsv', () => {
  it('incluye encabezados y filas', () => {
    const csv = buildCsv<Row>([{ name: 'Ana', value: 10 }], [
      { header: 'Nombre', value: (row) => row.name },
      { header: 'Valor', value: (row) => row.value },
    ])
    expect(csv).toContain('Nombre,Valor')
    expect(csv).toContain('Ana,10')
  })

  it('escapa valores con comas y comillas', () => {
    const csv = buildCsv<Row>([{ name: 'Pérez, Ana "La Jefa"', value: null }], [
      { header: 'Nombre', value: (row) => row.name },
      { header: 'Valor', value: (row) => row.value },
    ])
    expect(csv).toContain('"Pérez, Ana ""La Jefa"""')
  })

  it('antepone un BOM para que Excel respete los acentos', () => {
    const csv = buildCsv<Row>([], [{ header: 'Nombre', value: (row) => row.name }])
    expect(csv.charCodeAt(0)).toBe(0xfeff)
  })
})
