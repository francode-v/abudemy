import { useEffect, useMemo, useState } from 'react'
import { useSearchParams } from 'react-router-dom'
import { api, humanizeError, isApiError, type Service, type ServiceInput } from '@/api'
import { useCategories, useDeleteService, useSaveService, useServices, useServiceStatus } from '@/hooks/queries'
import { useOrgScope } from '@/app/orgScope'
import { useAuth } from '@/auth/AuthProvider'
import { useDebouncedValue } from '@/hooks/useDebouncedValue'
import { useToast } from '@/hooks/useToast'
import { Badge, Button, Callout, Card, CardBody, ErrorState, EmptyState, PageHeader, SelectInput, TextInput } from '@/components/ui/primitives'
import { DataTable, Pagination, type Column } from '@/components/ui/data'
import { Modal, ConfirmDialog } from '@/components/ui/overlays'
import { ServiceForm } from '@/components/forms/ServiceForm'
import { serviceStatusLabels, serviceTypeLabels } from '@/lib/dictionaries'
import { serviceStatusTone } from '@/lib/status'
import { formatNumber, relativeTime } from '@/lib/format'
import { downloadBlob } from '@/lib/csv'

const PER_PAGE = 20

export function ServicesPage() {
  const [params, setParams] = useSearchParams()
  const { organizationId } = useOrgScope()
  const { permissions } = useAuth()
  const toast = useToast()
  const categories = useCategories()

  const [search, setSearch] = useState('')
  const [categoryId, setCategoryId] = useState('')
  const [type, setType] = useState('')
  const [status, setStatus] = useState('')
  const [comuna, setComuna] = useState(params.get('comuna') ?? '')
  const [page, setPage] = useState(1)

  const [createOpen, setCreateOpen] = useState(false)
  const [editing, setEditing] = useState<Service | null>(null)
  const [deleting, setDeleting] = useState<Service | null>(null)
  const [exporting, setExporting] = useState(false)

  const debouncedSearch = useDebouncedValue(search)

  useEffect(() => {
    setPage(1)
  }, [debouncedSearch, categoryId, type, status, comuna, organizationId])

  const query = useMemo(
    () => ({
      search: debouncedSearch || undefined,
      category_id: categoryId || undefined,
      type: type || undefined,
      status: status || undefined,
      comuna: comuna || undefined,
      organization_id: organizationId ?? undefined,
      page,
      per_page: PER_PAGE,
    }),
    [debouncedSearch, categoryId, type, status, comuna, organizationId, page],
  )

  const services = useServices(query)
  const save = useSaveService()
  const statusMutation = useServiceStatus()
  const remove = useDeleteService()

  const canManage = permissions.canManageAnyContent

  const closeForm = () => {
    setCreateOpen(false)
    setEditing(null)
    save.reset()
  }

  const columns: Column<Service>[] = [
    {
      key: 'name',
      header: 'Servicio',
      render: (row) => (
        <div>
          <strong>{row.name}</strong>
          <div className="muted text-sm">{row.address || 'Sin dirección registrada'}</div>
        </div>
      ),
    },
    {
      key: 'organization',
      header: 'Organización',
      render: (row) =>
        row.organization_name ? (
          <span>{row.organization_name}</span>
        ) : (
          <span className="muted">Sin organización</span>
        ),
      hideOnMobile: true,
    },
    {
      key: 'category',
      header: 'Categoría',
      render: (row) => row.category?.name ?? <span className="muted">Sin categoría</span>,
      hideOnMobile: true,
    },
    {
      key: 'type',
      header: 'Tipo',
      render: (row) => serviceTypeLabels[row.type],
      hideOnMobile: true,
    },
    {
      key: 'comuna',
      header: 'Comuna',
      render: (row) =>
        row.comuna ? (
          <span>
            {row.comuna}
            <div className="muted text-sm">{row.region}</div>
          </span>
        ) : (
          <span className="muted">Sin comuna</span>
        ),
    },
    {
      key: 'updated',
      header: 'Actualizado',
      render: (row) => <span className="text-sm muted">{relativeTime(row.updated_at)}</span>,
      hideOnMobile: true,
    },
    {
      key: 'status',
      header: 'Estado',
      render: (row) =>
        canManage ? (
          <select
            className="input input--select"
            style={{ minHeight: 32, paddingTop: 2, paddingBottom: 2, width: 168 }}
            value={row.status}
            aria-label={`Estado de ${row.name}`}
            disabled={statusMutation.isPending}
            onChange={(event) =>
              statusMutation.mutate(
                { id: row.id, status: event.target.value as Service['status'] },
                {
                  onSuccess: () => toast.success('Estado actualizado', row.name),
                  onError: (error) => toast.error('No se pudo cambiar el estado', humanizeError(error)),
                },
              )
            }
          >
            <option value="active">Activo</option>
            <option value="inactive">Inactivo</option>
            <option value="draft">Borrador / en revisión</option>
          </select>
        ) : (
          <Badge tone={serviceStatusTone(row.status)}>{serviceStatusLabels[row.status]}</Badge>
        ),
    },
    {
      key: 'actions',
      header: '',
      align: 'right',
      render: (row) =>
        canManage ? (
          <div className="cell-actions">
            <Button variant="ghost" size="sm" icon="edit" onClick={() => setEditing(row)}>
              Editar
            </Button>
            <Button variant="ghost" size="sm" icon="trash" onClick={() => setDeleting(row)}>
              Eliminar
            </Button>
          </div>
        ) : (
          <span className="muted text-sm">Solo lectura</span>
        ),
    },
  ]

  const handleExport = async () => {
    setExporting(true)
    try {
      const blob = await api.analytics.exportCsv('services', organizationId ?? undefined)
      downloadBlob(blob, `abudemy-servicios-${new Date().toISOString().slice(0, 10)}.csv`)
      toast.success('Exportación lista', 'Se descargó el CSV de servicios.')
    } catch (error) {
      toast.error('No se pudo exportar', humanizeError(error))
    } finally {
      setExporting(false)
    }
  }

  return (
    <>
      <PageHeader
        title="Servicios y establecimientos"
        description="Catálogo de servicios de tu organización. Los cambios que guardes quedan disponibles de inmediato para la aplicación móvil."
        actions={
          <>
            <Button variant="secondary" icon="download" onClick={handleExport} loading={exporting}>
              Exportar CSV
            </Button>
            {canManage && (
              <Button icon="plus" onClick={() => setCreateOpen(true)}>
                Nuevo servicio
              </Button>
            )}
          </>
        }
      />

      {!canManage && (
        <Callout tone="info" title="Modo consulta">
          Tu rol permite ver el catálogo pero no modificarlo. Si necesitas editar servicios,
          solicita un rol de edición dentro de tu organización.
        </Callout>
      )}

      <Card>
        <div className="filters-bar">
          <div className="field toolbar__search" style={{ flex: '2 1 240px' }}>
            <label className="field__label" htmlFor="service-search">
              Buscar
            </label>
            <TextInput
              id="service-search"
              leadingIcon="search"
              placeholder="Nombre del servicio…"
              value={search}
              onChange={(event) => setSearch(event.target.value)}
            />
          </div>

          <SelectInput label="Categoría" value={categoryId} onChange={(event) => setCategoryId(event.target.value)}>
            <option value="">Todas</option>
            {(categories.data ?? []).map((category) => (
              <option key={category.id} value={category.id}>
                {category.name}
              </option>
            ))}
          </SelectInput>

          <SelectInput label="Tipo" value={type} onChange={(event) => setType(event.target.value)}>
            <option value="">Todos</option>
            <option value="establishment">Establecimiento</option>
            <option value="resource">Recurso de apoyo</option>
          </SelectInput>

          <SelectInput label="Estado" value={status} onChange={(event) => setStatus(event.target.value)}>
            <option value="">Todos</option>
            <option value="active">Activo</option>
            <option value="inactive">Inactivo</option>
            <option value="draft">Borrador / en revisión</option>
          </SelectInput>

          <TextInput
            label="Comuna"
            value={comuna}
            placeholder="Ej: Maipú"
            onChange={(event) => {
              setComuna(event.target.value)
              const next = new URLSearchParams(params)
              if (event.target.value) next.set('comuna', event.target.value)
              else next.delete('comuna')
              setParams(next, { replace: true })
            }}
          />

          <div className="toolbar__actions">
            <Button
              variant="ghost"
              size="sm"
              icon="close"
              onClick={() => {
                setSearch('')
                setCategoryId('')
                setType('')
                setStatus('')
                setComuna('')
                setParams({}, { replace: true })
              }}
            >
              Limpiar filtros
            </Button>
          </div>
        </div>

        <CardBody>
          {services.isError ? (
            <ErrorState description={humanizeError(services.error)} onRetry={() => void services.refetch()} />
          ) : (
            <>
              <DataTable
                columns={columns}
                rows={services.data?.items ?? []}
                rowKey={(row) => row.id}
                loading={services.isLoading}
                empty={
                  <EmptyState
                    title="No encontramos servicios"
                    description="Ajusta los filtros o crea el primer servicio de tu organización."
                    icon="services"
                    action={
                      canManage ? (
                        <Button icon="plus" size="sm" onClick={() => setCreateOpen(true)}>
                          Nuevo servicio
                        </Button>
                      ) : undefined
                    }
                  />
                }
              />
              {services.data && services.data.pagination.total > 0 && (
                <Pagination
                  page={services.data.pagination.page}
                  perPage={services.data.pagination.per_page}
                  total={services.data.pagination.total}
                  totalPages={services.data.pagination.total_pages}
                  onPageChange={setPage}
                />
              )}
              {services.data && (
                <p className="split-note" style={{ marginTop: 8 }}>
                  {formatNumber(services.data.pagination.total)} servicios encontrados. Los datos se
                  actualizaron {relativeTime(services.dataUpdatedAt)}.
                </p>
              )}
            </>
          )}
        </CardBody>
      </Card>

      <Modal
        open={createOpen || editing !== null}
        onClose={closeForm}
        title={editing ? 'Editar servicio' : 'Nuevo servicio'}
        description={
          editing
            ? 'Actualiza la información del servicio.'
            : 'Registra un establecimiento o recurso de apoyo.'
        }
        size="lg"
      >
        <ServiceForm
          initial={editing}
          submitting={save.isPending}
          error={save.error}
          onCancel={closeForm}
          onSubmit={(input: ServiceInput) => {
            save.mutate(
              { id: editing?.id, input },
              {
                onSuccess: () => {
                  toast.success(editing ? 'Servicio actualizado' : 'Servicio creado', input.name)
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

      <ConfirmDialog
        open={deleting !== null}
        title="Eliminar servicio"
        message={
          <>
            ¿Confirmas eliminar <strong>{deleting?.name}</strong>? Esta acción no se puede deshacer y
            el servicio dejará de estar disponible para la comunidad.
          </>
        }
        confirmLabel="Eliminar"
        loading={remove.isPending}
        onCancel={() => setDeleting(null)}
        onConfirm={() => {
          if (!deleting) return
          remove.mutate(deleting.id, {
            onSuccess: () => {
              toast.success('Servicio eliminado', deleting.name)
              setDeleting(null)
            },
            onError: (error) => {
              toast.error('No se pudo eliminar', humanizeError(error))
              setDeleting(null)
            },
          })
        }}
      />
    </>
  )
}
