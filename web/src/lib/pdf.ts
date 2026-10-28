/**
 * Exportación a PDF del lado del cliente.
 *
 * El backend documenta `format=pdf` como una capacidad futura y responde
 * `501 Not Implemented` (ver backend/docs/openapi.yaml). Para no depender de un
 * servicio inexistente, el resumen se genera en el navegador a partir de los
 * datos ya consultados a la API y se indica su procedencia en el documento.
 */
import jsPDF from 'jspdf'
import autoTable from 'jspdf-autotable'

export interface PdfSection {
  title: string
  columns: string[]
  rows: (string | number)[][]
}

export interface PdfDocumentInput {
  title: string
  subtitle?: string
  metadata: { label: string; value: string }[]
  sections: PdfSection[]
  footer?: string
}

const BRAND = { r: 20, g: 92, b: 107 }
const TEXT = { r: 27, g: 42, b: 51 }

export function buildPdf(input: PdfDocumentInput): Blob {
  const doc = new jsPDF({ unit: 'pt', format: 'a4' })
  const pageWidth = doc.internal.pageSize.getWidth()
  const marginX = 40

  doc.setFillColor(BRAND.r, BRAND.g, BRAND.b)
  doc.rect(0, 0, pageWidth, 74, 'F')
  doc.setTextColor(255, 255, 255)
  doc.setFont('helvetica', 'bold')
  doc.setFontSize(18)
  doc.text('Abudemy', marginX, 32)
  doc.setFont('helvetica', 'normal')
  doc.setFontSize(11)
  doc.text('Plataforma institucional', marginX, 50)

  doc.setTextColor(TEXT.r, TEXT.g, TEXT.b)
  doc.setFont('helvetica', 'bold')
  doc.setFontSize(14)
  doc.text(input.title, marginX, 102)

  let cursorY = 118
  if (input.subtitle) {
    doc.setFont('helvetica', 'normal')
    doc.setFontSize(10)
    doc.text(input.subtitle, marginX, cursorY)
    cursorY += 16
  }

  doc.setFontSize(9)
  for (const item of input.metadata) {
    doc.setFont('helvetica', 'bold')
    doc.text(`${item.label}:`, marginX, cursorY)
    doc.setFont('helvetica', 'normal')
    doc.text(item.value, marginX + 110, cursorY)
    cursorY += 13
  }

  for (const section of input.sections) {
    cursorY += 10
    doc.setFont('helvetica', 'bold')
    doc.setFontSize(11)
    doc.text(section.title, marginX, cursorY)
    autoTable(doc, {
      startY: cursorY + 8,
      head: [section.columns],
      body: section.rows.map((row) => row.map((cell) => String(cell))),
      margin: { left: marginX, right: marginX },
      styles: { fontSize: 8, cellPadding: 4 },
      headStyles: { fillColor: [BRAND.r, BRAND.g, BRAND.b] },
      alternateRowStyles: { fillColor: [244, 247, 249] },
    })
    const finalY = (doc as unknown as { lastAutoTable?: { finalY: number } }).lastAutoTable?.finalY
    cursorY = finalY ? finalY : cursorY
  }

  const pageCount = doc.getNumberOfPages()
  for (let page = 1; page <= pageCount; page += 1) {
    doc.setPage(page)
    doc.setFontSize(8)
    doc.setTextColor(120, 130, 138)
    doc.text(
      input.footer ??
        'Documento generado por la plataforma web de Abudemy. Los datos provienen de la API institucional.',
      marginX,
      doc.internal.pageSize.getHeight() - 24,
    )
    doc.text(`Página ${page} de ${pageCount}`, pageWidth - marginX, doc.internal.pageSize.getHeight() - 24, {
      align: 'right',
    })
  }

  return doc.output('blob')
}

export function downloadPdf(blob: Blob, filename: string): void {
  const url = URL.createObjectURL(blob)
  const anchor = document.createElement('a')
  anchor.href = url
  anchor.download = filename
  document.body.appendChild(anchor)
  anchor.click()
  anchor.remove()
  URL.revokeObjectURL(url)
}
