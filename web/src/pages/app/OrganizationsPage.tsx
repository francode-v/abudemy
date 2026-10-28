import { useEffect, useMemo, useState } from 'react'
import { Link } from 'react-router-dom'
import {
  humanizeError,
  isApiError,
  type Organization,
  type OrganizationInput,
  type OrganizationType,
} from '@/api'
import { useOrganizations, useSaveOrganization } from '@/hooks/queries'
import { useAuth } from '@/auth/AuthProvider'
import { useToast } from '@/hooks/useToast'
import { useDebouncedValue } from '@/hooks/useDebouncedValue'
import { Badge, Button, Callout, Card, CardBody, EmptyState, ErrorState, PageHeader, SelectInput, TextInput } from '@/components/ui/primitives'
import { DataTable, Pagination, type Column } from '@/components/ui/data'
import { Modal } from '@/components/ui/overlays'
import { OrganizationForm } from '@/components/forms/OrganizationForm'
import { organizationTypeLabels } from '@/lib/dictionaries'
import { formatNumber, relativeTime } from '@/lib/format'

const PER_PAGE = 20

export function OrganizationsPage() {
  const { permissions, refreshMemberships } = useAuth()
  const toast = useToast()

  const [search, setSearch] = useState('')
  const [type, setType] = useState('')
  const [active, setActive] = useState('')
  const [page, setPage] = useState(1)

  const [formOpen, setFormOpen] = useState(false)
  const [editing, setEditing] = useState<Organization | null>(null)

  const debouncedSearch = useDebouncedValue(search)

  useEffect(() => {
    setPage(1)
  }, [debouncedSearch, type, active])

  const query = useMemo(
    () => ({
      search: debouncedSearch || undefined,
      type: type || undefined,
      active: active ? active === 'true' : undefined,
      page,
      per_page: PER_PAGE,
    }),
    [debouncedSearch, type, active, page],
  )

  const organizations = useOrganizations(query)
  const save = useSaveOrganization()

  const isAdmin = permissions.isAdmin
  const canCreate = permissions.canCreateOrganization

  // El backend pagina y filtra solo para administradores; el resto recibe todas
  // sus organizaciones, por lo que se filtran localmente para mantener la UX.
  const rows = useMemo(() => {
    const items = organizations.data?.items ?? []
    if (isAdmin) return items
    const term = debouncedSearch.trim().toLowerCase()
    return items.filter((org) => {
      if (term && !org.name.toLowerCase().includes(term)) return false
      if (type && org.type !== type) return false
      if (active && String(org.active) !== active) return false
      return true
    })
  }, [organizations.data, isAdmin, debouncedSearch, type, active])

  const openCreate = () => {
    setEditing(null)
    setFormOpen(true)
    save.reset()
  }

  const openEdit = (org: Organization) => {
    setEditing(org)
    setFormOpen(true)
    save.reset()
  }

  const columns: Column<Organization>[] = [
    {
      key: 'name',
      header: 'Organización',
      render: (row) => (
        <div>
          <Link to={`/app/organizaciones/${row.id}`}>
            <strong>{row.name}</strong>
          </Link>
          <div className="muted text-sm">{organizationTypeLabels[row.type] ?? row.type}</div>
        </div>
      ),
    },
    {
      key: 'territory',
      header: 'Territorio',
      render: (row) =>
        row.comuna ? (
          <span>
            {row.comuna}
            <div className="muted text-sm">{row.region}</div>
          </span>
        ) : (
          <span className="muted">Sin comuna</span>
        ),
      hideOnMobile: true,
    },
    {
      key: 'contact',
      header: 'Contacto',
      render: (row) => (
        <span className="text-sm">
          {row.email || row.phone || <span className="muted">Sin datos de contacto</span>}
        </span>
      ),
      hideOnMobile: true,
    },
    {
      key: 'active',
      header: 'Estado',
      render: (row) => (
        <Badge tone={row.active ? 'success' : 'neutral'}>{row.active ? 'Activa' : 'Inactiva'}</Badge>
      ),
    },
    {
      key: 'updated',
      header: 'Actualizada',
      render: (row) => <span className="muted text-sm">{relativeTime(row.updated_at)}</span>,
      hideOnMobile: true,
    },
    {
      key: 'actions',
      header: '',
      align: 'right',
      render: (row) => (
        <div className="cell-actions">
          <Link className="btn btn--ghost btn--sm" to={`/app/organizaciones/${row.id}`}>
            Ver ficha
          </Link>
          {permissions.canWriteOrganization(row.id) && (
            <Button variant="ghost" size="sm" icon="edit" onClick={() => openEdit(row)}>
              Editar
            </Button>
          )}
        </div>
      ),
    },
  ]

  return (
    <>
      <PageHeader
        title="Organizaciones"
        description="Instituciones que publican servicios, actividades y oportunidades en Abudemy, con sus integrantes y permisos."
        actions={
          canCreate ? (
            <Button icon="plus" onClick={openCreate}>
              Nueva organización
            </Button>
          ) : undefined
        }
      />

      <Callout tone="info" title="Alcance según tu rol">
        {isAdmin
          ? 'Como administrador ves todas las organizaciones de la plataforma y puedes crearlas o editarlas.'
          : 'Ves únicamente las organizaciones a las que perteneces. Puedes editar la información si tu rol dentro de la organización es propietario, administrador o editor.'}
      </Callout>

      <Card>
        <div className="filters-bar">
          <div className="field toolbar__search" style={{ flex: '2 1 240px' }}>
            <label className="field__label" htmlFor="org-search">
              Buscar
            </label>
            <TextInput
              id="org-search"
              leadingIcon="search"
              placeholder="Nombre de la organización…"
              value={search}
              onChange={(event) => setSearch(event.target.value)}
            />
          </div>

          <SelectInput label="Tipo" value={type} onChange={(event) => setType(event.target.value)}>
            <option value="">Todos</option>
            {(Object.keys(organizationTypeLabels) as OrganizationType[]).map((orgType) => (
              <option key={orgType} value={orgType}>
                {organizationTypeLabels[orgType]}
              </option>
            ))}
          </SelectInput>

          <SelectInput label="Estado" value={active} onChange={(event) => setActive(event.target.value)}>
            <option value="">Todos</option>
            <option value="true">Activas</option>
            <option value="false">Inactivas</option>
          </SelectInput>

          <div className="toolbar__actions">
            <Button
              variant="ghost"
              size="sm"
              icon="close"
              onClick={() => {
                setSearch('')
                setType('')
                setActive('')
              }}
            >
              Limpiar filtros
            </Button>
          </div>
        </div>

        <CardBody>
          {organizations.isError ? (
            <ErrorState
              description={humanizeError(organizations.error)}
              onRetry={() => void organizations.refetch()}
            />
          ) : (
            <>
              <DataTable
                columns={columns}
                rows={rows}
                rowKey={(row) => row.id}
                loading={organizations.isLoading}
                empty={
                  <EmptyState
                    title="No encontramos organizaciones"
                    description={
                      canCreate
                        ? 'Ajusta los filtros o crea la primera organización.'
                        : 'Ajusta los filtros o solicita a tu administrador que te asigne una organización.'
                    }
                    icon="organizations"
                    action={
                      canCreate ? (
                        <Button icon="plus" size="sm" onClick={openCreate}>
                          Nueva organización
                        </Button>
                      ) : undefined
                    }
                  />
                }
              />
              {isAdmin && organizations.data && organizations.data.pagination.total > PER_PAGE && (
                <Pagination
                  page={organizations.data.pagination.page}
                  perPage={organizations.data.pagination.per_page}
                  total={organizations.data.pagination.total}
                  totalPages={organizations.data.pagination.total_pages}
                  onPageChange={setPage}
                />
              )}
              <p className="split-note" style={{ marginTop: 8 }}>
                {isAdmin
                  ? `${formatNumber(organizations.data?.pagination.total ?? 0)} organizaciones registradas.`
                  : `${formatNumber(rows.length)} organizaciones a las que perteneces.`}{' '}
                Actualizado {relativeTime(organizations.dataUpdatedAt)}.
              </p>
            </>
          )}
        </CardBody>
      </Card>

      <Modal
        open={formOpen}
        onClose={() => setFormOpen(false)}
        title={editing ? 'Editar organización' : 'Nueva organización'}
        description={
          editing
            ? 'Actualiza los datos de la organización.'
            : 'Registra una institución. Quien la crea queda como propietario.'
        }
        size="lg"
      >
        <OrganizationForm
          initial={editing}
          submitting={save.isPending}
          error={save.error}
          onCancel={() => setFormOpen(false)}
          onSubmit={(input: OrganizationInput) => {
            save.mutate(
              { id: editing?.id, input },
              {
                onSuccess: (created) => {
                  toast.success(
                    editing ? 'Organización actualizada' : 'Organización creada',
                    created.name,
                  )
                  setFormOpen(false)
                  if (!editing) void refreshMemberships()
                },
                onError: (err) => {
                  if (!isApiError(err) || !err.isValidation) {
                    toast.error('No se pudo guardar', humanizeError(err))
                  }
                },
              },
            )
          }}
        />
      </Modal>
    </>
  )
}
