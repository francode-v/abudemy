import { useState } from 'react'
import { Link, useParams } from 'react-router-dom'
import { humanizeError, type Enrollment } from '@/api'
import { useActivity, useActivityEnrollments, useEnrollmentStatus } from '@/hooks/queries'
import { useAuth } from '@/auth/AuthProvider'
import { useToast } from '@/hooks/useToast'
import { Badge, Button, Callout, Card, CardBody, CardHeader, EmptyState, ErrorState, PageHeader, Spinner } from '@/components/ui/primitives'
import { DataTable, KeyValue, Pagination, StatCard, type Column } from '@/components/ui/data'
import { activityStatusLabels, activityTypeLabels, enrollmentStatusLabels, modalityLabels } from '@/lib/dictionaries'
import { activityStatusTone, enrollmentStatusTone } from '@/lib/status'
import { formatCurrency, formatDateTime, formatNumber, relativeTime } from '@/lib/format'
import { buildCsv, downloadCsv, timestampSuffix } from '@/lib/csv'
import { Icon } from '@/components/ui/Icon'

const PER_PAGE = 20

export function ActivityDetailPage() {
  const { id } = useParams<{ id: string }>()
  const { permissions } = useAuth()
  const toast = useToast()
  const [page, setPage] = useState(1)

  const activity = useActivity(id)
  const enrollments = useActivityEnrollments(id, page, PER_PAGE)
  const statusMutation = useEnrollmentStatus(id ?? '')

  const canManage = permissions.canManageAnyContent

  const columns: Column<Enrollment>[] = [
    {
      key: 'user',
      header: 'Persona',
      render: (row) => (
        <div>
          <strong>{row.user_full_name || 'Sin nombre'}</strong>
          <div className="muted text-sm">{row.user_email || 'Sin correo'}</div>
        </div>
      ),
    },
    {
      key: 'created',
      header: 'Inscripción',
      render: (row) => <span className="text-sm">{formatDateTime(row.created_at)}</span>,
      hideOnMobile: true,
    },
    {
      key: 'status',
      header: 'Estado',
      render: (row) =>
        canManage ? (
          <select
            className="input input--select"
            style={{ minHeight: 32, paddingTop: 2, paddingBottom: 2, width: 150 }}
            value={row.status}
            aria-label={`Estado de ${row.user_full_name ?? row.user_id}`}
            disabled={statusMutation.isPending}
            onChange={(event) =>
              statusMutation.mutate(
                { userId: row.user_id, status: event.target.value as Enrollment['status'] },
                {
                  onSuccess: () => toast.success('Inscripción actualizada'),
                  onError: (error) => toast.error('No se pudo actualizar', humanizeError(error)),
                },
              )
            }
          >
            <option value="enrolled">Inscrito</option>
            <option value="waitlist">Lista de espera</option>
            <option value="attended">Asistió</option>
            <option value="cancelled">Cancelada</option>
          </select>
        ) : (
          <Badge tone={enrollmentStatusTone(row.status)}>{enrollmentStatusLabels[row.status]}</Badge>
        ),
    },
  ]

  if (activity.isLoading) {
    return (
      <div className="content">
        <Spinner size="lg" />
      </div>
    )
  }

  if (activity.isError || !activity.data) {
    return (
      <div className="content">
        <ErrorState
          title="No pudimos cargar la actividad"
          description={humanizeError(activity.error)}
          onRetry={() => void activity.refetch()}
        />
      </div>
    )
  }

  const item = activity.data
  const occupancy = item.capacity > 0 ? item.enrolled_count / item.capacity : null

  const exportAttendance = () => {
    const rows = enrollments.data?.items ?? []
    if (rows.length === 0) {
      toast.warning('Sin inscripciones para exportar')
      return
    }
    const csv = buildCsv(rows, [
      { header: 'Nombre', value: (row) => row.user_full_name ?? '' },
      { header: 'Correo', value: (row) => row.user_email ?? '' },
      { header: 'Estado', value: (row) => enrollmentStatusLabels[row.status] },
      { header: 'Fecha de inscripción', value: (row) => row.created_at },
    ])
    downloadCsv(`abudemy-inscripciones-${timestampSuffix()}.csv`, csv)
    toast.success('Listado exportado', 'Corresponde a la página de inscripciones actualmente visible.')
  }

  return (
    <>
      <PageHeader
        title={item.title}
        breadcrumb={
          <Link to="/app/actividades">
            <Icon name="chevron-left" size={12} /> Volver a actividades
          </Link>
        }
        description={item.description || 'Sin descripción registrada.'}
        actions={
          <>
            <Button variant="secondary" icon="download" onClick={exportAttendance}>
              Exportar inscripciones
            </Button>
            <Link className="btn btn--secondary" to="/app/actividades">
              Ver todas
            </Link>
          </>
        }
      />

      <section className="stat-grid">
        <StatCard
          label="Estado de publicación"
          value={<Badge tone={activityStatusTone(item.status)}>{activityStatusLabels[item.status]}</Badge>}
          description="Estado actual de la actividad en la plataforma."
          updatedAt={item.updated_at}
        />
        <StatCard
          label="Inscritos"
          value={formatNumber(item.enrolled_count)}
          hint={item.capacity > 0 ? `Cupo: ${formatNumber(item.capacity)}` : 'Sin cupo máximo'}
          description="Personas con inscripción vigente: estados «inscrito» y «asistió»."
          updatedAt={item.updated_at}
        />
        <StatCard
          label="Ocupación"
          value={occupancy === null ? '—' : `${Math.round(occupancy * 100)}%`}
          description="Relación entre inscritos y cupo declarado. Se calcula sólo si hay cupo."
          tone="accent"
          updatedAt={item.updated_at}
        />
        <StatCard
          label="Inicio"
          value={formatDateTime(item.start_at)}
          hint={item.end_at ? `Término: ${formatDateTime(item.end_at)}` : undefined}
          description="Fecha y hora de inicio declarada por la organización."
          tone="info"
          updatedAt={item.updated_at}
        />
      </section>

      <div className="grid-2">
        <Card>
          <CardHeader title="Detalles de la actividad" icon="activities" />
          <CardBody>
            <KeyValue
              items={[
                { label: 'Tipo', value: activityTypeLabels[item.type] },
                { label: 'Modalidad', value: modalityLabels[item.modality] },
                { label: 'Lugar', value: item.location || '—' },
                { label: 'Comuna', value: item.comuna || '—' },
                { label: 'Región', value: item.region || '—' },
                { label: 'Precio', value: item.price > 0 ? formatCurrency(item.price) : 'Gratuita' },
                { label: 'Creada', value: formatDateTime(item.created_at) },
                { label: 'Última actualización', value: `${formatDateTime(item.updated_at)} (${relativeTime(item.updated_at)})` },
              ]}
            />
          </CardBody>
        </Card>

        <Card>
          <CardHeader
            title="Inscripciones"
            description="Personas inscritas en la actividad. Los datos personales se muestran únicamente a quien administra la actividad."
            icon="users"
          />
          <CardBody>
            {enrollments.isError ? (
              <ErrorState
                description={humanizeError(enrollments.error)}
                onRetry={() => void enrollments.refetch()}
              />
            ) : (
              <>
                <DataTable
                  columns={columns}
                  rows={enrollments.data?.items ?? []}
                  rowKey={(row) => row.id}
                  loading={enrollments.isLoading}
                  empty={
                    <EmptyState
                      title="Sin inscripciones"
                      description="Todavía no hay personas inscritas en esta actividad."
                      icon="users"
                    />
                  }
                />
                {enrollments.data && enrollments.data.pagination.total > PER_PAGE && (
                  <Pagination
                    page={enrollments.data.pagination.page}
                    perPage={enrollments.data.pagination.per_page}
                    total={enrollments.data.pagination.total}
                    totalPages={enrollments.data.pagination.total_pages}
                    onPageChange={setPage}
                  />
                )}
              </>
            )}
          </CardBody>
        </Card>
      </div>

      {!canManage && (
        <Callout tone="info" title="Permisos de inscripciones">
          Para cambiar el estado de una inscripción necesitas permisos de edición en la
          organización responsable de la actividad.
        </Callout>
      )}
    </>
  )
}
