import { useEffect, useMemo, useState } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import { api, humanizeError, isApiError, type Activity, type ActivityInput } from '@/api'
import { useActivities, useActivityStatus, useCategories, useSaveActivity } from '@/hooks/queries'
import { useOrgScope } from '@/app/orgScope'
import { useAuth } from '@/auth/AuthProvider'
import { useDebouncedValue } from '@/hooks/useDebouncedValue'
import { useToast } from '@/hooks/useToast'
import { Badge, Button, Card, CardBody, EmptyState, ErrorState, PageHeader, SelectInput, TextInput } from '@/components/ui/primitives'
import { DataTable, Pagination, type Column } from '@/components/ui/data'
import { Modal } from '@/components/ui/overlays'
import { ActivityForm } from '@/components/forms/ActivityForm'
import { activityStatusLabels, activityTypeLabels, modalityLabels } from '@/lib/dictionaries'
import { activityStatusTone } from '@/lib/status'
import { formatDateTime, formatNumber, relativeTime } from '@/lib/format'
import { downloadBlob } from '@/lib/csv'

const PER_PAGE = 20

export function ActivitiesPage() {
  const { organizationId } = useOrgScope()
  const { permissions } = useAuth()
  const toast = useToast()
  const navigate = useNavigate()
  const categories = useCategories()

  const [search, setSearch] = useState('')
  const [type, setType] = useState('')
  const [status, setStatus] = useState('')
  const [modality, setModality] = useState('')
  const [comuna, setComuna] = useState('')
  const [from, setFrom] = useState('')
  const [to, setTo] = useState('')
  const [page, setPage] = useState(1)

  const [createOpen, setCreateOpen] = useState(false)
  const [editing, setEditing] = useState<Activity | null>(null)
  const [exporting, setExporting] = useState(false)

  const debouncedSearch = useDebouncedValue(search)

  useEffect(() => {
    setPage(1)
  }, [debouncedSearch, type, status, modality, comuna, from, to, organizationId])

  const query = useMemo(
    () => ({
      search: debouncedSearch || undefined,
      type: type || undefined,
      status: status || undefined,
      modality: modality || undefined,
      comuna: comuna || undefined,
      from: from ? new Date(`${from}T00:00:00`).toISOString() : undefined,
      to: to ? new Date(`${to}T23:59:59`).toISOString() : undefined,
      organization_id: organizationId ?? undefined,
      page,
      per_page: PER_PAGE,
    }),
    [debouncedSearch, type, status, modality, comuna, from, to, organizationId, page],
  )

  const activities = useActivities(query)
  const save = useSaveActivity()
  const statusMutation = useActivityStatus()

  const canManage = permissions.canManageAnyContent

  const closeForm = () => {
    setCreateOpen(false)
    setEditing(null)
    save.reset()
  }

  const columns: Column<Activity>[] = [
    {
      key: 'title',
      header: 'Actividad',
      render: (row) => (
        <div>
          <strong>{row.title}</strong>
          <div className="muted text-sm">
            {activityTypeLabels[row.type]} · {modalityLabels[row.modality]}
          </div>
        </div>
      ),
    },
    {
      key: 'start',
      header: 'Inicio',
      render: (row) => <span className="text-sm">{formatDateTime(row.start_at)}</span>,
    },
    {
      key: 'location',
      header: 'Lugar',
      render: (row) => row.comuna || row.location || <span className="muted">Sin lugar</span>,
      hideOnMobile: true,
    },
    {
      key: 'occupancy',
      header: 'Inscritos',
      align: 'right',
      render: (row) =>
        row.capacity > 0 ? (
          <span>
            {formatNumber(row.enrolled_count)}
            <span className="muted"> / {formatNumber(row.capacity)}</span>
          </span>
        ) : (
          <span>{formatNumber(row.enrolled_count)}</span>
        ),
    },
    {
      key: 'status',
      header: 'Estado',
      render: (row) =>
        canManage ? (
          <select
            className="input input--select"
            style={{ minHeight: 32, paddingTop: 2, paddingBottom: 2, width: 178 }}
            value={row.status}
            aria-label={`Estado de ${row.title}`}
            disabled={statusMutation.isPending}
            onChange={(event) =>
              statusMutation.mutate(
                { id: row.id, status: event.target.value as Activity['status'] },
                {
                  onSuccess: () => toast.success('Estado actualizado', row.title),
                  onError: (error) => toast.error('No se pudo cambiar el estado', humanizeError(error)),
                },
              )
            }
          >
            <option value="draft">Borrador / en revisión</option>
            <option value="published">Publicada</option>
            <option value="cancelled">Cancelada</option>
            <option value="finished">Finalizada</option>
          </select>
        ) : (
          <Badge tone={activityStatusTone(row.status)}>{activityStatusLabels[row.status]}</Badge>
        ),
    },
    {
      key: 'actions',
      header: '',
      align: 'right',
      render: (row) => (
        <div className="cell-actions">
          <Button variant="ghost" size="sm" icon="users" onClick={() => navigate(`/app/actividades/${row.id}`)}>
            Inscripciones
          </Button>
          {canManage && (
            <Button variant="ghost" size="sm" icon="edit" onClick={() => setEditing(row)}>
              Editar
            </Button>
          )}
        </div>
      ),
    },
  ]

  const handleExport = async () => {
    setExporting(true)
    try {
      const blob = await api.analytics.exportCsv('activities', organizationId ?? undefined)
      downloadBlob(blob, `abudemy-actividades-${new Date().toISOString().slice(0, 10)}.csv`)
      toast.success('Exportación lista', 'El CSV incluye todas las actividades del ámbito seleccionado.')
    } catch (error) {
      toast.error('No se pudo exportar', humanizeError(error))
    } finally {
      setExporting(false)
    }
  }

  return (
    <>
      <PageHeader
        title="Actividades, talleres y cursos"
        description="Administra la programación de tu organización, controla cupos y revisa las inscripciones."
        actions={
          <>
            <Button variant="secondary" icon="download" onClick={handleExport} loading={exporting}>
              Exportar CSV
            </Button>
            {canManage && (
              <Button icon="plus" onClick={() => setCreateOpen(true)}>
                Nueva actividad
              </Button>
            )}
          </>
        }
      />

      <Card>
        <div className="filters-bar">
          <div className="field toolbar__search" style={{ flex: '2 1 220px' }}>
            <label className="field__label" htmlFor="activity-search">
              Buscar
            </label>
            <TextInput
              id="activity-search"
              leadingIcon="search"
              placeholder="Título de la actividad…"
              value={search}
              onChange={(event) => setSearch(event.target.value)}
            />
          </div>
          <SelectInput label="Tipo" value={type} onChange={(event) => setType(event.target.value)}>
            <option value="">Todos</option>
            <option value="workshop">Taller</option>
            <option value="course">Curso</option>
            <option value="event">Evento</option>
            <option value="support_group">Grupo de apoyo</option>
          </SelectInput>
          <SelectInput label="Estado" value={status} onChange={(event) => setStatus(event.target.value)}>
            <option value="">Todos</option>
            <option value="draft">Borrador / en revisión</option>
            <option value="published">Publicada</option>
            <option value="cancelled">Cancelada</option>
            <option value="finished">Finalizada</option>
          </SelectInput>
          <SelectInput label="Modalidad" value={modality} onChange={(event) => setModality(event.target.value)}>
            <option value="">Todas</option>
            <option value="in_person">Presencial</option>
            <option value="online">En línea</option>
            <option value="hybrid">Híbrida</option>
          </SelectInput>
          <TextInput label="Comuna" value={comuna} onChange={(event) => setComuna(event.target.value)} />
          <TextInput label="Desde" type="date" value={from} onChange={(event) => setFrom(event.target.value)} />
          <TextInput label="Hasta" type="date" value={to} onChange={(event) => setTo(event.target.value)} />
          <div className="toolbar__actions">
            <Button
              variant="ghost"
              size="sm"
              icon="close"
              onClick={() => {
                setSearch('')
                setType('')
                setStatus('')
                setModality('')
                setComuna('')
                setFrom('')
                setTo('')
              }}
            >
              Limpiar
            </Button>
          </div>
        </div>

        <CardBody>
          {activities.isError ? (
            <ErrorState description={humanizeError(activities.error)} onRetry={() => void activities.refetch()} />
          ) : (
            <>
              <DataTable
                columns={columns}
                rows={activities.data?.items ?? []}
                rowKey={(row) => row.id}
                loading={activities.isLoading}
                empty={
                  <EmptyState
                    title="No hay actividades"
                    description="Crea un taller, curso o evento para tu comunidad."
                    icon="activities"
                    action={
                      canManage ? (
                        <Button icon="plus" size="sm" onClick={() => setCreateOpen(true)}>
                          Nueva actividad
                        </Button>
                      ) : undefined
                    }
                  />
                }
              />
              {activities.data && activities.data.pagination.total > 0 && (
                <Pagination
                  page={activities.data.pagination.page}
                  perPage={activities.data.pagination.per_page}
                  total={activities.data.pagination.total}
                  totalPages={activities.data.pagination.total_pages}
                  onPageChange={setPage}
                />
              )}
              {activities.data && (
                <p className="split-note" style={{ marginTop: 8 }}>
                  {formatNumber(activities.data.pagination.total)} actividades · actualizado{' '}
                  {relativeTime(activities.dataUpdatedAt)}. Las categorías disponibles son{' '}
                  {(categories.data ?? []).length}.
                </p>
              )}
            </>
          )}
        </CardBody>
      </Card>

      <Modal
        open={createOpen || editing !== null}
        onClose={closeForm}
        title={editing ? 'Editar actividad' : 'Nueva actividad'}
        description="Los cambios quedarán visibles para los usuarios de la aplicación móvil."
        size="lg"
      >
        <ActivityForm
          initial={editing}
          submitting={save.isPending}
          error={save.error}
          onCancel={closeForm}
          onSubmit={(input: ActivityInput) => {
            save.mutate(
              { id: editing?.id, input },
              {
                onSuccess: () => {
                  toast.success(editing ? 'Actividad actualizada' : 'Actividad creada', input.title)
                  closeForm()
                },
                onError: (error) => {
                  if (!isApiError(error) || !error.isValidation) {
                    toast.error('No se pudo guardar', humanizeError(error))
                  }
                },
              },
            )
          }}
        />
      </Modal>

      <p className="split-note">
        ¿Necesitas revisar quién está inscrito? Abre <Link to="/app/actividades">una actividad</Link>{' '}
        y entra a su detalle de inscripciones.
      </p>
    </>
  )
}
