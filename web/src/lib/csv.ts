/** Utilidades de exportación de datos tabulares. */

export interface CsvColumn<T> {
  header: string
  value: (row: T) => string | number | null | undefined
}

export function buildCsv<T>(rows: T[], columns: CsvColumn<T>[]): string {
  const escape = (value: string | number | null | undefined): string => {
    const text = value === null || value === undefined ? '' : String(value)
    if (/[",\n;]/.test(text)) {
      return `"${text.replace(/"/g, '""')}"`
    }
    return text
  }
  const header = columns.map((c) => escape(c.header)).join(',')
  const body = rows.map((row) => columns.map((c) => escape(c.value(row))).join(','))
  // BOM para que Excel abra correctamente los acentos.
  return `\uFEFF${[header, ...body].join('\r\n')}`
}

export function downloadBlob(blob: Blob, filename: string): void {
  const url = URL.createObjectURL(blob)
  const anchor = document.createElement('a')
  anchor.href = url
  anchor.download = filename
  document.body.appendChild(anchor)
  anchor.click()
  anchor.remove()
  URL.revokeObjectURL(url)
}

export function downloadCsv(filename: string, content: string): void {
  downloadBlob(new Blob([content], { type: 'text/csv;charset=utf-8' }), filename)
}

export function timestampSuffix(date = new Date()): string {
  const pad = (n: number) => String(n).padStart(2, '0')
  return `${date.getFullYear()}${pad(date.getMonth() + 1)}${pad(date.getDate())}-${pad(date.getHours())}${pad(date.getMinutes())}`
}
