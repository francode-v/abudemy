import { useEffect, useMemo, useState } from 'react'
import { humanizeError, type User, type UserStatus } from '@/api'
import { useUserAdminMutations, useUsers } from '@/hooks/queries'
import { useToast } from '@/hooks/useToast'
import { useDebouncedValue } from '@/hooks/useDebouncedValue'
import {
  Badge,
  Button,
  Callout,
  Card,
  CardBody,
  EmptyState,
  ErrorState,
  PageHeader,
  SelectInput,
  TextInput,
} from '@/components/ui/primitives'
import { DataTable, Pagination, StatCard, type Column } from '@/components/ui/data'
import { userStatusLabels } from '@/lib/dictionaries'
import { userStatusTone } from '@/lib/status'
import { formatDate, formatNumber, relativeTime } from '@/lib/format'

const PER_PAGE = 20
const STATUSES: UserStatus[] = ['active', 'inactive', 'suspended']

/** Panel de administración para aprobar cuentas institucionales. */
export function InstitutionalAccountsPage() {
  const toast = useToast()
  const [search, setSearch] = useState('')
  const [status, setStatus] = useState('')
  const [page, setPage] = useState(1)

  const debouncedSearch = useDebouncedValue(search)

  useEffect(() => {
    setPage(1)
  }, [debouncedSearch, status])

  const query = useMemo(
    () => ({
      role: 'institutional_client',
      search: debouncedSearch || undefined,
      status: status || undefined,
      page,
      per_page: PER_PAGE,
    }),
    [debouncedSearch, status, page],
  )

  const accounts = useUsers(query)
  const { setInstitutionalApproval } = useUserAdminMutations()

  const items = accounts.data?.items ?? []
  const pending = items.filter((item) => !item.institutional_approved).length
  const approved = items.filter((item) => item.institutional_approved).length
  const active = items.filter((item) => item.status === 'active').length

  const columns: Column<User>[] = [
    {
      key: 'person',
      header: 'Persona',
      render: (row) => (
        <div>
          <strong>{row.full_name}</strong>
          <div className="muted text-sm">{row.email}</div>
        </div>
      ),
    },
    {
      key: 'status',
      header: 'Cuenta',
      render: (row) => (
        <Badge tone={userStatusTone(row.status)}>{userStatusLabels[row.status]}</Badge>
      ),
      hideOnMobile: true,
    },
    {
      key: 'approval',
      header: 'Acceso institucional',
      render: (row) =>
        row.institutional_approved ? (
          <Badge tone="success">Activo</Badge>
        ) : (
          <Badge tone="warning">Pendiente</Badge>
        ),
    },
    {
      key: 'created',
      header: 'Registro',
      render: (row) => <span className="text-sm muted">{formatDate(row.created_at)}</span>,
      hideOnMobile: true,
    },
    {
      key: 'login',
      header: 'Último acceso',
      render: (row) => (
        <span className="text-sm muted">
          {row.last_login_at ? relativeTime(row.last_login_at) : 'Sin registro'}
        </span>
      ),
      hideOnMobile: true,
    },
    {
      key: 'actions',
      header: '',
      align: 'right',
      render: (row) => (
        <Button
          variant={row.institutional_approved ? 'ghost' : 'primary'}
          size="sm"
          icon={row.institutional_approved ? 'close' : 'check'}
          loading={setInstitutionalApproval.isPending && setInstitutionalApproval.variables?.id === row.id}
          onClick={() =>
            setInstitutionalApproval.mutate(
              { id: row.id, approved: !row.institutional_approved },
              {
                onSuccess: () =>
                  toast.success(
                    row.institutional_approved ? 'Acceso desactivado' : 'Cuenta aprobada',
                    row.full_name,
                  ),
                onError: (error) =>
                  toast.error('No se pudo actualizar', humanizeError(error)),
              },
            )
          }
        >
          {row.institutional_approved ? 'Desactivar acceso' : 'Activar acceso'}
        </Button>
      ),
    },
  ]

  return (
    <>
      <PageHeader
        title="Cuentas institucionales"
        description="Aprueba o revoca el acceso de los clientes institucionales al portal. Solo los administradores pueden entrar a esta vista."
        actions={
          <Button
            variant="secondary"
            icon="refresh"
            onClick={() => void accounts.refetch()}
            loading={accounts.isFetching}
          >
            Actualizar
          </Button>
        }
      />

      <Callout tone="warning" title="Aprobación de acceso">
        Activar una cuenta habilita de inmediato el portal institucional; desactivarla bloquea sus
        secciones sin borrar sus datos. Los cambios se aplican en la siguiente petición de la
        persona, aunque ya tenga una sesión abierta.
      </Callout>

      <section className="stat-grid">
        <StatCard
          label="Pendientes (página actual)"
          value={formatNumber(pending)}
          tone="warning"
          icon="clock"
          description="Cuentas institucionales que aún no pueden usar el portal."
          updatedAt={accounts.dataUpdatedAt ? new Date(accounts.dataUpdatedAt) : undefined}
        />
        <StatCard
          label="Aprobadas (página actual)"
          value={formatNumber(approved)}
          tone="success"
          icon="check"
          description="Cuentas institucionales con acceso habilitado."
          updatedAt={accounts.dataUpdatedAt ? new Date(accounts.dataUpdatedAt) : undefined}
        />
        <StatCard
          label="Activas (página actual)"
          value={formatNumber(active)}
          icon="users"
          description="Cuentas cuyo estado de cuenta es «activo»."
          updatedAt={accounts.dataUpdatedAt ? new Date(accounts.dataUpdatedAt) : undefined}
        />
        <StatCard
          label="Total institucional"
          value={formatNumber(accounts.data?.pagination.total)}
          tone="primary"
          icon="organizations"
          description="Cuentas institucionales que coinciden con los filtros."
          updatedAt={accounts.dataUpdatedAt ? new Date(accounts.dataUpdatedAt) : undefined}
        />
      </section>

      <Card>
        <div className="filters-bar">
          <div className="field toolbar__search" style={{ flex: '2 1 240px' }}>
            <label className="field__label" htmlFor="institutional-search">
              Buscar
            </label>
            <TextInput
              id="institutional-search"
              leadingIcon="search"
              placeholder="Nombre o correo…"
              value={search}
              onChange={(event) => setSearch(event.target.value)}
            />
          </div>

          <SelectInput label="Estado de la cuenta" value={status} onChange={(event) => setStatus(event.target.value)}>
            <option value="">Todos</option>
            {STATUSES.map((value) => (
              <option key={value} value={value}>
                {userStatusLabels[value]}
              </option>
            ))}
          </SelectInput>

          <div className="toolbar__actions">
            <Button
              variant="ghost"
              size="sm"
              icon="close"
              onClick={() => {
                setSearch('')
                setStatus('')
              }}
            >
              Limpiar filtros
            </Button>
          </div>
        </div>

        <CardBody>
          {accounts.isError ? (
            <ErrorState
              description={humanizeError(accounts.error)}
              onRetry={() => void accounts.refetch()}
            />
          ) : (
            <>
              <DataTable
                columns={columns}
                rows={items}
                rowKey={(row) => row.id}
                loading={accounts.isLoading}
                empty={
                  <EmptyState
                    title="No encontramos cuentas institucionales"
                    description="Ajusta los filtros de búsqueda o estado."
                    icon="users"
                  />
                }
              />
              {accounts.data && accounts.data.pagination.total > PER_PAGE && (
                <Pagination
                  page={accounts.data.pagination.page}
                  perPage={accounts.data.pagination.per_page}
                  total={accounts.data.pagination.total}
                  totalPages={accounts.data.pagination.total_pages}
                  onPageChange={setPage}
                />
              )}
            </>
          )}
        </CardBody>
      </Card>
    </>
  )
}
