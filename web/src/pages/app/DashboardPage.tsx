import { useMemo } from 'react'
import { Link } from 'react-router-dom'
import { useOverview } from '@/hooks/queries'
import { useOrgScope } from '@/app/orgScope'
import { DataTable, ChartCard, StatCard, Pagination, type Column } from '@/components/ui/data'
import { CategoryBarChart, DistributionPieChart, HorizontalCountList, TerritorialBarChart } from '@/components/ui/charts'
import { Card, CardBody, CardHeader, ErrorState, EmptyState, Button, PageHeader } from '@/components/ui/primitives'
import { humanizeError } from '@/api'
import { formatNumber, formatPercent } from '@/lib/format'
import { activityTypeLabels } from '@/lib/dictionaries'
import { downloadPdf } from '@/lib/pdf'
import { buildPdf } from '@/lib/pdf'
import { useToast } from '@/hooks/useToast'
import { useState } from 'react'
import type { TerritorialDistribution } from '@/api'

const PER_PAGE = 8

export function DashboardPage() {
  const { organizationId, activeName } = useOrgScope()
  const query = useOverview(organizationId)
  const toast = useToast()
  const [page, setPage] = useState(1)

  const overview = query.data
  const updatedAt = query.dataUpdatedAt ? new Date(query.dataUpdatedAt) : null

  const territorial = useMemo(() => {
    if (!overview) return []
    return [...overview.territorial_distribution].sort((a, b) => b.count - a.count)
  }, [overview])

  const territorialPage = territorial.slice((page - 1) * PER_PAGE, page * PER_PAGE)

  const columns: Column<TerritorialDistribution>[] = [
    { key: 'comuna', header: 'Comuna', render: (row) => <strong>{row.comuna}</strong> },
    { key: 'region', header: 'Región', render: (row) => row.region || '—' },
    { key: 'count', header: 'Servicios', align: 'right', render: (row) => formatNumber(row.count) },
  ]

  if (query.isError) {
    return (
      <div className="content">
        <ErrorState
          description={humanizeError(query.error)}
          onRetry={() => void query.refetch()}
        />
      </div>
    )
  }

  const handleExportPdf = () => {
    if (!overview) return
    const blob = buildPdf({
      title: 'Resumen de indicadores institucionales',
      subtitle: `Ámbito: ${activeName}`,
      metadata: [
        { label: 'Organización', value: activeName },
        { label: 'Consultado', value: updatedAt ? updatedAt.toLocaleString('es-CL') : '—' },
        { label: 'Fuente', value: 'API institucional de Abudemy' },
      ],
      sections: [
        {
          title: 'Indicadores principales',
          columns: ['Indicador', 'Valor'],
          rows: [
            ['Servicios registrados', overview.services_total],
            ['Servicios activos', overview.services_active],
            ['Actividades registradas', overview.activities_total],
            ['Actividades publicadas', overview.activities_published],
            ['Actividades próximas', overview.activities_upcoming],
            ['Inscripciones totales', overview.enrollments_total],
            ['Inscripciones activas', overview.enrollments_active],
            ['Personas registradas', overview.users_total],
            ['Organizaciones', overview.organizations_total],
            ['Oportunidades abiertas', overview.opportunities_open],
            ['Solicitudes de asistencia', overview.assistance_total],
            ['Solicitudes pendientes', overview.assistance_pending],
            ['Ocupación promedio', `${(overview.average_occupancy * 100).toFixed(1)}%`],
          ],
        },
        {
          title: 'Servicios por categoría',
          columns: ['Categoría', 'Servicios'],
          rows: overview.services_by_category.map((item) => [item.label, item.count]),
        },
        {
          title: 'Actividades por tipo',
          columns: ['Tipo', 'Actividades'],
          rows: overview.activities_by_type.map((item) => [
            activityTypeLabels[item.label as keyof typeof activityTypeLabels] ?? item.label,
            item.count,
          ]),
        },
        {
          title: 'Distribución territorial',
          columns: ['Comuna', 'Región', 'Servicios'],
          rows: territorial.map((item) => [item.comuna, item.region, item.count]),
        },
      ],
    })
    downloadPdf(blob, `abudemy-resumen-${new Date().toISOString().slice(0, 10)}.pdf`)
    toast.success('Resumen generado', 'El PDF se descargó con los datos del panel.')
  }

  return (
    <>
      <PageHeader
        title="Panel institucional"
        description={`Indicadores disponibles para ${activeName}, según tu organización activa.`}
        actions={
          <>
            <Button variant="secondary" icon="refresh" onClick={() => void query.refetch()} loading={query.isFetching}>
              Actualizar
            </Button>
            <Button icon="download" onClick={handleExportPdf} disabled={!overview}>
              Resumen PDF
            </Button>
          </>
        }
      />

      <section className="stat-grid" aria-label="Indicadores principales">
        <StatCard
          label="Servicios registrados"
          value={formatNumber(overview?.services_total)}
          icon="services"
          tone="primary"
          description="Total de establecimientos y recursos de apoyo registrados en el catálogo de servicios."
          hint={overview ? `${formatNumber(overview.services_active)} activos` : undefined}
          updatedAt={updatedAt}
          loading={query.isLoading}
        />
        <StatCard
          label="Actividades registradas"
          value={formatNumber(overview?.activities_total)}
          icon="activities"
          tone="accent"
          description="Talleres, cursos, eventos y grupos de apoyo creados. Incluye borradores y actividades finalizadas."
          hint={overview ? `${formatNumber(overview.activities_published)} publicadas` : undefined}
          updatedAt={updatedAt}
          loading={query.isLoading}
        />
        <StatCard
          label="Actividades próximas"
          value={formatNumber(overview?.activities_upcoming)}
          icon="calendar"
          tone="success"
          description="Actividades publicadas cuya fecha de inicio es igual o posterior al momento de la consulta."
          updatedAt={updatedAt}
          loading={query.isLoading}
        />
        <StatCard
          label="Inscripciones"
          value={formatNumber(overview?.enrollments_total)}
          icon="users"
          tone="info"
          description="Total de inscripciones registradas en actividades. Incluye inscripciones canceladas."
          hint={overview ? `${formatNumber(overview.enrollments_active)} activas` : undefined}
          updatedAt={updatedAt}
          loading={query.isLoading}
        />
        <StatCard
          label="Ocupación promedio"
          value={formatPercent(overview?.average_occupancy, 1)}
          icon="trend"
          tone="primary"
          description="Promedio de la razón entre inscritos y cupo en las actividades que declaran capacidad."
          updatedAt={updatedAt}
          loading={query.isLoading}
        />
        <StatCard
          label="Solicitudes de asistencia"
          value={formatNumber(overview?.assistance_total)}
          icon="heart"
          tone="danger"
          description="Solicitudes de asistencia registradas. Cuando hay una organización activa, se cuentan las de sus integrantes."
          hint={overview ? `${formatNumber(overview.assistance_pending)} pendientes` : undefined}
          updatedAt={updatedAt}
          loading={query.isLoading}
        />
        <StatCard
          label="Personas registradas"
          value={formatNumber(overview?.users_total)}
          icon="user"
          tone="neutral"
          description="Cuentas creadas en la plataforma. No se muestran datos personales, solo el total."
          hint={overview ? `${formatNumber(overview.users_mobile)} en la app móvil` : undefined}
          updatedAt={updatedAt}
          loading={query.isLoading}
        />
        <StatCard
          label="Organizaciones"
          value={formatNumber(overview?.organizations_total)}
          icon="organizations"
          tone="neutral"
          description="Organizaciones registradas en la plataforma con al menos un integrante."
          updatedAt={updatedAt}
          loading={query.isLoading}
        />
        <StatCard
          label="Oportunidades abiertas"
          value={formatNumber(overview?.opportunities_open)}
          icon="trend"
          tone="accent"
          description="Oportunidades económicas, laborales o de voluntariado con estado «abierta»."
          updatedAt={updatedAt}
          loading={query.isLoading}
        />
      </section>

      <div className="chart-grid">
        <ChartCard
          title="Servicios por categoría"
          description="Distribución del catálogo de servicios según su categoría."
          updatedAt={updatedAt}
        >
          {overview && overview.services_by_category.length > 0 ? (
            <CategoryBarChart data={overview.services_by_category} />
          ) : (
            <EmptyState
              title="Sin datos por categoría"
              description="Aún no hay servicios clasificados. Crea servicios y asígnales una categoría."
              icon="services"
              action={
                <Link className="btn btn--secondary btn--sm" to="/app/servicios">
                  Ir a servicios
                </Link>
              }
            />
          )}
        </ChartCard>

        <ChartCard
          title="Actividades por tipo"
          description="Composición de las actividades registradas según su tipo."
          updatedAt={updatedAt}
        >
          {overview && overview.activities_by_type.length > 0 ? (
            <DistributionPieChart
              data={overview.activities_by_type.map((item) => ({
                label: activityTypeLabels[item.label as keyof typeof activityTypeLabels] ?? item.label,
                count: item.count,
              }))}
            />
          ) : (
            <EmptyState
              title="Sin actividades registradas"
              description="Publica talleres, cursos o eventos para ver su distribución."
              icon="activities"
              action={
                <Link className="btn btn--secondary btn--sm" to="/app/actividades">
                  Ir a actividades
                </Link>
              }
            />
          )}
        </ChartCard>
      </div>

      <div className="grid-2">
        <ChartCard
          title="Distribución territorial"
          description="Comunas con más servicios registrados (máximo 10 en el gráfico)."
          updatedAt={updatedAt}
        >
          {territorial.length > 0 ? (
            <TerritorialBarChart data={territorial} />
          ) : (
            <EmptyState
              title="Sin información territorial"
              description="Los servicios aún no tienen comuna declarada. Complétala para habilitar este análisis."
              icon="map"
            />
          )}
        </ChartCard>

        <Card>
          <CardHeader
            title="Resumen por comuna"
            description="Distribución territorial de los servicios registrados."
            actions={
              <Link className="btn btn--ghost btn--sm" to="/app/mapa">
                Ver mapa
              </Link>
            }
          />
          <CardBody>
            <DataTable
              columns={columns}
              rows={territorialPage}
              rowKey={(row) => `${row.comuna}-${row.region}`}
              loading={query.isLoading}
              empty={
                <EmptyState
                  title="Sin información territorial"
                  description="Registra la comuna de tus servicios para ver este resumen."
                  icon="map"
                />
              }
            />
            {territorial.length > PER_PAGE && (
              <Pagination
                page={page}
                perPage={PER_PAGE}
                total={territorial.length}
                totalPages={Math.ceil(territorial.length / PER_PAGE)}
                onPageChange={setPage}
              />
            )}
          </CardBody>
        </Card>
      </div>

      {overview && overview.services_by_category.length > 0 && (
        <Card>
          <CardHeader
            title="Detalle de servicios por categoría"
            description="Cantidad exacta de servicios por categoría."
          />
          <CardBody>
            <HorizontalCountList data={overview.services_by_category} unit="servicios" />
          </CardBody>
        </Card>
      )}

    </>
  )
}
