import { useMemo, useState } from 'react'
import { api, humanizeError, isApiError, type ReportType } from '@/api'
import { useOverview, useReportPreview, useReportsCatalog } from '@/hooks/queries'
import { useAllActivities } from '@/hooks/paged'
import { useAuth } from '@/auth/AuthProvider'
import { useOrgScope } from '@/app/orgScope'
import { useToast } from '@/hooks/useToast'
import { Button, Callout, Card, CardBody, CardHeader, EmptyState, ErrorState, PageHeader, SelectInput, TextInput } from '@/components/ui/primitives'
import { ChartCard, DataTable, StatCard, type Column } from '@/components/ui/data'
import { CategoryBarChart, DistributionPieChart, TerritorialBarChart, TrendLineChart, type TrendPoint } from '@/components/ui/charts'
import { activityTypeLabels, reportTypeLabels } from '@/lib/dictionaries'
import { downloadBlob, timestampSuffix } from '@/lib/csv'
import { buildPdf, downloadPdf } from '@/lib/pdf'
import { formatNumber, formatPercent } from '@/lib/format'
import type { TerritorialDistribution } from '@/api'

const REPORTS: ReportType[] = ['services', 'activities', 'enrollments', 'assistance', 'opportunities', 'territorial']

const REPORT_COLUMN_LABELS: Record<string, string> = {
  id: 'ID',
  name: 'Servicio',
  organization: 'Organización',
  category: 'Categoría',
  type: 'Tipo',
  status: 'Estado',
  comuna: 'Comuna',
  region: 'Región',
  phone: 'Teléfono',
  email: 'Correo',
  title: 'Título',
  modality: 'Modalidad',
  start_at: 'Inicio',
  capacity: 'Cupo',
  price: 'Precio',
  enrolled: 'Inscritos',
  activity: 'Actividad',
  user_email: 'Correo de la persona',
  user_name: 'Nombre de la persona',
  created_at: 'Creado',
  resolved_at: 'Resuelto',
  salary_min: 'Salario mínimo',
  salary_max: 'Salario máximo',
  services: 'Servicios',
}

export function ReportsPage() {
  const { organizationId, activeName, options } = useOrgScope()
  const { permissions } = useAuth()
  const toast = useToast()

  const [from, setFrom] = useState('')
  const [to, setTo] = useState('')
  const [reportType, setReportType] = useState<ReportType>('services')
  const [exporting, setExporting] = useState(false)

  const overview = useOverview(organizationId)
  const catalog = useReportsCatalog()

  const activities = useAllActivities({
    organization_id: organizationId ?? undefined,
    from: from ? new Date(`${from}T00:00:00`).toISOString() : undefined,
    to: to ? new Date(`${to}T23:59:59`).toISOString() : undefined,
  })

  const trend: TrendPoint[] = useMemo(() => {
    const buckets = new Map<string, number>()
    for (const activity of activities.items) {
      const date = new Date(activity.start_at)
      if (Number.isNaN(date.getTime())) continue
      const key = `${date.getFullYear()}-${String(date.getMonth() + 1).padStart(2, '0')}`
      buckets.set(key, (buckets.get(key) ?? 0) + 1)
    }
    return [...buckets.entries()]
      .sort(([a], [b]) => a.localeCompare(b))
      .slice(-12)
      .map(([key, value]) => {
        const [year, month] = key.split('-')
        const label = new Date(Number(year), Number(month) - 1, 1).toLocaleDateString('es-CL', {
          month: 'short',
          year: '2-digit',
        })
        return { label, value }
      })
  }, [activities.items])

  const territorialColumns: Column<TerritorialDistribution>[] = [
    { key: 'comuna', header: 'Comuna', render: (row) => <strong>{row.comuna}</strong> },
    { key: 'region', header: 'Región', render: (row) => row.region || '—' },
    { key: 'count', header: 'Servicios', align: 'right', render: (row) => formatNumber(row.count) },
  ]

  const canExportOrg = useMemo(() => {
    if (permissions.isAdmin) return true
    if (organizationId) return true
    return options.length === 1
  }, [permissions.isAdmin, organizationId, options.length])

  const preview = useReportPreview(reportType, organizationId, canExportOrg)

  const previewColumns: Column<string[]>[] = useMemo(() => {
    const headers = preview.data?.headers ?? []
    return headers.map((header, index) => ({
      key: `${index}-${header}`,
      header: REPORT_COLUMN_LABELS[header] ?? header,
      render: (row) => row[index] || '—',
    }))
  }, [preview.data])

  const handleCsv = async () => {
    if (!canExportOrg) {
      toast.warning(
        'Selecciona una organización',
        'Indica la organización cuyo informe quieres consultar.',
      )
      return
    }
    setExporting(true)
    try {
      const blob = await api.analytics.exportCsv(reportType, organizationId ?? undefined)
      downloadBlob(blob, `abudemy-${reportType}-${timestampSuffix()}.csv`)
      toast.success('CSV descargado', `Informe «${reportTypeLabels[reportType]}» exportado.`)
    } catch (error) {
      if (isApiError(error) && error.isNotImplemented) {
        toast.warning('Exportación no disponible', humanizeError(error))
      } else {
        toast.error('No se pudo exportar', humanizeError(error))
      }
    } finally {
      setExporting(false)
    }
  }

  const handlePdf = () => {
    if (!overview.data) {
      toast.warning('Sin datos para el informe')
      return
    }
    const data = overview.data
    const blob = buildPdf({
      title: `Informe de gestión — ${activeName}`,
      subtitle: `Período: ${from || 'sin inicio'} a ${to || 'sin término'}`,
      metadata: [
        { label: 'Organización', value: activeName },
        { label: 'Generado', value: new Date().toLocaleString('es-CL') },
        { label: 'Fuente de datos', value: 'API institucional de Abudemy' },
        { label: 'Filas territoriales', value: String(data.territorial_distribution.length) },
      ],
      sections: [
        {
          title: 'Indicadores del período',
          columns: ['Indicador', 'Valor'],
          rows: [
            ['Servicios registrados', data.services_total],
            ['Servicios activos', data.services_active],
            ['Actividades registradas', data.activities_total],
            ['Actividades publicadas', data.activities_published],
            ['Inscripciones totales', data.enrollments_total],
            ['Inscripciones activas', data.enrollments_active],
            ['Solicitudes de asistencia', data.assistance_total],
            ['Solicitudes pendientes', data.assistance_pending],
            ['Ocupación promedio', `${(data.average_occupancy * 100).toFixed(1)}%`],
          ],
        },
        {
          title: 'Territorio',
          columns: ['Comuna', 'Región', 'Servicios'],
          rows: data.territorial_distribution.map((row) => [row.comuna, row.region, row.count]),
        },
        {
          title: 'Evolución mensual de actividades',
          columns: ['Mes', 'Actividades'],
          rows: trend.map((point) => [point.label, point.value]),
        },
      ],
      footer: 'Generado por la plataforma web de Abudemy.',
    })
    downloadPdf(blob, `abudemy-informe-${timestampSuffix()}.pdf`)
    toast.success('Informe PDF generado')
  }

  return (
    <>
      <PageHeader
        title="Informes"
        description="Consulta indicadores por territorio y período, revisa resúmenes y exporta la información."
        actions={
          <>
            <Button variant="secondary" icon="refresh" onClick={() => void overview.refetch()} loading={overview.isFetching}>
              Actualizar
            </Button>
            <Button variant="secondary" icon="download" onClick={handleCsv} loading={exporting}>
              Exportar CSV
            </Button>
            <Button icon="download" onClick={handlePdf} disabled={!overview.data}>
              Generar PDF
            </Button>
          </>
        }
      />

      <Card>
        <div className="filters-bar">
          <SelectInput
            label="Tipo de informe"
            value={reportType}
            onChange={(event) => setReportType(event.target.value as ReportType)}
          >
            {(catalog.data?.items as ReportType[] | undefined ?? REPORTS).map((type) => (
              <option key={type} value={type}>
                {reportTypeLabels[type] ?? type}
              </option>
            ))}
          </SelectInput>
          <TextInput label="Período desde" type="date" value={from} onChange={(event) => setFrom(event.target.value)} />
          <TextInput label="Período hasta" type="date" value={to} onChange={(event) => setTo(event.target.value)} />
          <div className="toolbar__actions">
            <Button
              variant="ghost"
              size="sm"
              icon="close"
              onClick={() => {
                setFrom('')
                setTo('')
              }}
            >
              Limpiar período
            </Button>
          </div>
        </div>
      </Card>

      <Card>
        <CardHeader
          title={`Vista previa · ${reportTypeLabels[reportType] ?? reportType}`}
          description={`Contenido del informe para ${activeName}. Cambia el tipo de informe para actualizar la tabla.`}
        />
        <CardBody>
          {!canExportOrg ? (
            <Callout tone="warning" title="Selecciona una organización">
              Administras más de una organización. Elige una en el selector superior para consultar
              este informe.
            </Callout>
          ) : preview.isError ? (
            <ErrorState
              description={humanizeError(preview.error)}
              onRetry={() => void preview.refetch()}
            />
          ) : (
            <DataTable
              columns={previewColumns}
              rows={preview.data?.rows ?? []}
              rowKey={(row) => row[0] || row.join('|')}
              loading={preview.isLoading}
              empty={
                <EmptyState
                  title="Sin registros"
                  description="El informe seleccionado no tiene datos en este ámbito."
                  icon="reports"
                />
              }
            />
          )}
          {preview.data && (
            <p className="split-note" style={{ marginTop: 8 }}>
              {preview.data.truncated
                ? `Vista previa con las primeras ${formatNumber(preview.data.rows.length)} de ${formatNumber(
                    preview.data.total,
                  )} filas. Descarga el CSV para ver el informe completo.`
                : `${formatNumber(preview.data.rows.length)} filas y ${
                    preview.data.headers.length
                  } columnas.`}
            </p>
          )}
        </CardBody>
      </Card>

      {overview.isError ? (
        <ErrorState description={humanizeError(overview.error)} onRetry={() => void overview.refetch()} />
      ) : (
        <>
          <section className="stat-grid">
            <StatCard
              label="Servicios activos"
              value={formatNumber(overview.data?.services_active)}
              hint={`${formatNumber(overview.data?.services_total)} registrados`}
              description="Servicios en estado activo dentro del ámbito seleccionado."
              updatedAt={overview.dataUpdatedAt ? new Date(overview.dataUpdatedAt) : undefined}
              loading={overview.isLoading}
            />
            <StatCard
              label="Actividades publicadas"
              value={formatNumber(overview.data?.activities_published)}
              hint={`${formatNumber(overview.data?.activities_total)} registradas`}
              description="Actividades publicadas y visibles para la comunidad."
              tone="accent"
              updatedAt={overview.dataUpdatedAt ? new Date(overview.dataUpdatedAt) : undefined}
              loading={overview.isLoading}
            />
            <StatCard
              label="Inscripciones activas"
              value={formatNumber(overview.data?.enrollments_active)}
              hint={`${formatNumber(overview.data?.enrollments_total)} totales`}
              description="Inscripciones en estado «inscrito» o «asistió»."
              tone="success"
              updatedAt={overview.dataUpdatedAt ? new Date(overview.dataUpdatedAt) : undefined}
              loading={overview.isLoading}
            />
            <StatCard
              label="Ocupación promedio"
              value={formatPercent(overview.data?.average_occupancy, 1)}
              description="Ocupación media de las actividades con cupo declarado."
              tone="info"
              updatedAt={overview.dataUpdatedAt ? new Date(overview.dataUpdatedAt) : undefined}
              loading={overview.isLoading}
            />
          </section>

          <div className="chart-grid">
            <ChartCard title="Servicios por categoría" updatedAt={overview.dataUpdatedAt ? new Date(overview.dataUpdatedAt) : undefined}>
              {overview.data && overview.data.services_by_category.length > 0 ? (
                <CategoryBarChart data={overview.data.services_by_category} />
              ) : (
                <EmptyState title="Sin datos" description="No hay servicios categorizados en este ámbito." icon="services" />
              )}
            </ChartCard>

            <ChartCard title="Actividades por tipo" updatedAt={overview.dataUpdatedAt ? new Date(overview.dataUpdatedAt) : undefined}>
              {overview.data && overview.data.activities_by_type.length > 0 ? (
                <DistributionPieChart
                  data={overview.data.activities_by_type.map((item) => ({
                    label: activityTypeLabels[item.label as keyof typeof activityTypeLabels] ?? item.label,
                    count: item.count,
                  }))}
                />
              ) : (
                <EmptyState title="Sin datos" description="No hay actividades registradas en este ámbito." icon="activities" />
              )}
            </ChartCard>
          </div>

          <ChartCard
            title="Evolución mensual de actividades"
            description="Cantidad de actividades según su fecha de inicio, dentro del período y ámbito seleccionados."
            updatedAt={activities.isLoading ? undefined : new Date()}
          >
            {trend.length > 0 ? (
              <TrendLineChart data={trend} label="Actividades" />
            ) : (
              <EmptyState
                title="Sin serie temporal"
                description="No hay actividades con fecha de inicio en el período seleccionado."
                icon="trend"
              />
            )}
          </ChartCard>

          <div className="grid-2">
            <ChartCard title="Distribución territorial" updatedAt={overview.dataUpdatedAt ? new Date(overview.dataUpdatedAt) : undefined}>
              {overview.data && overview.data.territorial_distribution.length > 0 ? (
                <TerritorialBarChart data={overview.data.territorial_distribution} />
              ) : (
                <EmptyState title="Sin información territorial" icon="map" />
              )}
            </ChartCard>

            <Card>
              <CardHeader title="Tabla territorial" description="Base del informe exportable." />
              <CardBody>
                <DataTable
                  columns={territorialColumns}
                  rows={overview.data?.territorial_distribution ?? []}
                  rowKey={(row) => `${row.comuna}-${row.region}`}
                  loading={overview.isLoading}
                  empty={<EmptyState title="Sin comunas registradas" icon="map" />}
                />
              </CardBody>
            </Card>
          </div>

          {activities.truncated && (
            <Callout tone="warning" title="Serie parcial">
              La serie mensual se calcula sobre las primeras {formatNumber(activities.loaded)}{' '}
              actividades (hasta 5 páginas). Acota el período para mayor precisión.
            </Callout>
          )}

        </>
      )}
    </>
  )
}
