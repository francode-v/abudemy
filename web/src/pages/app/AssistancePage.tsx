import { useEffect, useMemo, useState } from 'react'
import { humanizeError, type AssistanceRequest } from '@/api'
import { useAssistanceRequests, useUpdateAssistanceStatus } from '@/hooks/queries'
import { useAuth } from '@/auth/AuthProvider'
import { useToast } from '@/hooks/useToast'
import { Badge, Button, Callout, Card, CardBody, EmptyState, ErrorState, PageHeader, SelectInput } from '@/components/ui/primitives'
import { DataTable, Pagination, StatCard, type Column } from '@/components/ui/data'
import { assistanceStatusLabels } from '@/lib/dictionaries'
import { assistanceStatusTone } from '@/lib/status'
import { formatDateTime, formatNumber, relativeTime } from '@/lib/format'

const PER_PAGE = 20

export function AssistancePage() {
  const { permissions } = useAuth()
  const toast = useToast()
  const [status, setStatus] = useState('')
  const [page, setPage] = useState(1)

  useEffect(() => {
    setPage(1)
  }, [status])

  const query = useMemo(
    () => ({ status: status || undefined, page, per_page: PER_PAGE }),
    [status, page],
  )

  const requests = useAssistanceRequests(query)
  const updateStatus = useUpdateAssistanceStatus()

  const items = requests.data?.items ?? []
  const pending = items.filter((item) => item.status === 'pending').length
  const resolved = items.filter((item) => item.status === 'resolved').length

  const columns: Column<AssistanceRequest>[] = [
    {
      key: 'type',
      header: 'Tipo',
      render: (row) => <strong>{row.type || 'General'}</strong>,
    },
    {
      key: 'description',
      header: 'Descripción',
      render: (row) => <span className="text-sm muted">{row.description || 'Sin detalle'}</span>,
      hideOnMobile: true,
    },
    {
      key: 'created',
      header: 'Registrada',
      render: (row) => <span className="text-sm">{formatDateTime(row.created_at)}</span>,
    },
    {
      key: 'notifications',
      header: 'Notificaciones',
      render: (row) =>
        row.notifications && row.notifications.length > 0 ? (
          <span className="text-sm">
            {row.notifications.filter((n) => n.status === 'sent').length} enviadas ·{' '}
            {row.notifications.filter((n) => n.status === 'skipped').length} omitidas
          </span>
        ) : (
          <span className="muted text-sm">Sin registros</span>
        ),
      hideOnMobile: true,
    },
    {
      key: 'status',
      header: 'Estado',
      render: (row) =>
        permissions.isAdmin ? (
          <select
            className="input input--select"
            style={{ minHeight: 32, paddingTop: 2, paddingBottom: 2, width: 150 }}
            value={row.status}
            aria-label={`Estado de la solicitud ${row.id}`}
            disabled={updateStatus.isPending}
            onChange={(event) =>
              updateStatus.mutate(
                { id: row.id, status: event.target.value as AssistanceRequest['status'] },
                {
                  onSuccess: () => toast.success('Solicitud actualizada'),
                  onError: (error) => toast.error('No se pudo actualizar', humanizeError(error)),
                },
              )
            }
          >
            <option value="pending">Pendiente</option>
            <option value="notified">Notificada</option>
            <option value="resolved">Resuelta</option>
            <option value="cancelled">Cancelada</option>
          </select>
        ) : (
          <Badge tone={assistanceStatusTone(row.status)}>{assistanceStatusLabels[row.status]}</Badge>
        ),
    },
  ]

  return (
    <>
      <PageHeader
        title="Solicitudes de asistencia"
        description="Registro de solicitudes de ayuda creadas desde la aplicación móvil."
        actions={
          <Button variant="secondary" icon="refresh" onClick={() => void requests.refetch()} loading={requests.isFetching}>
            Actualizar
          </Button>
        }
      />

      <Callout tone="info" title="Privacidad por diseño">
        {permissions.isAdmin
          ? 'Como administrador puedes ver y actualizar todas las solicitudes. Evita reproducir descripciones o ubicaciones fuera de la plataforma.'
          : 'Solo puedes ver tus propias solicitudes: las de otras personas nunca se muestran a clientes institucionales. Para conocer totales agregados usa el panel de indicadores, que cuenta sin mostrar datos personales.'}
      </Callout>

      <section className="stat-grid">
        <StatCard
          label="Solicitudes (página actual)"
          value={formatNumber(items.length)}
          description="Solicitudes listadas en esta página según el filtro aplicado."
          icon="heart"
          updatedAt={requests.dataUpdatedAt ? new Date(requests.dataUpdatedAt) : undefined}
        />
        <StatCard
          label="Pendientes"
          value={formatNumber(pending)}
          tone="warning"
          description="Solicitudes en estado «pendiente» dentro de la página actual."
          updatedAt={requests.dataUpdatedAt ? new Date(requests.dataUpdatedAt) : undefined}
        />
        <StatCard
          label="Resueltas"
          value={formatNumber(resolved)}
          tone="success"
          description="Solicitudes marcadas como resueltas dentro de la página actual."
          updatedAt={requests.dataUpdatedAt ? new Date(requests.dataUpdatedAt) : undefined}
        />
      </section>

      <Card>
        <div className="filters-bar">
          <SelectInput label="Estado" value={status} onChange={(event) => setStatus(event.target.value)}>
            <option value="">Todos</option>
            <option value="pending">Pendiente</option>
            <option value="notified">Notificada</option>
            <option value="resolved">Resuelta</option>
            <option value="cancelled">Cancelada</option>
          </SelectInput>
          <div className="toolbar__actions">
            <Button variant="ghost" size="sm" icon="close" onClick={() => setStatus('')}>
              Limpiar
            </Button>
          </div>
        </div>

        <CardBody>
          {requests.isError ? (
            <ErrorState description={humanizeError(requests.error)} onRetry={() => void requests.refetch()} />
          ) : (
            <>
              <DataTable
                columns={columns}
                rows={items}
                rowKey={(row) => row.id}
                loading={requests.isLoading}
                empty={
                  <EmptyState
                    title="No hay solicitudes"
                    description="No se registran solicitudes de asistencia con los filtros actuales."
                    icon="heart"
                  />
                }
              />
              {requests.data && requests.data.pagination.total > 0 && (
                <Pagination
                  page={requests.data.pagination.page}
                  perPage={requests.data.pagination.per_page}
                  total={requests.data.pagination.total}
                  totalPages={requests.data.pagination.total_pages}
                  onPageChange={setPage}
                />
              )}
              <p className="split-note" style={{ marginTop: 8 }}>
                Actualizado {relativeTime(requests.dataUpdatedAt)}.
              </p>
            </>
          )}
        </CardBody>
      </Card>
    </>
  )
}
