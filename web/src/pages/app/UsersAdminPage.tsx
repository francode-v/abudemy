import { useEffect, useMemo, useState } from 'react'
import { humanizeError, type Role, type User, type UserStatus } from '@/api'
import { useUserAdminMutations, useUsers } from '@/hooks/queries'
import { useAuth } from '@/auth/AuthProvider'
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
import { roleLabels, userStatusLabels } from '@/lib/dictionaries'
import { userStatusTone } from '@/lib/status'
import { formatDate, formatNumber, relativeTime } from '@/lib/format'

const PER_PAGE = 20
const ROLES: Role[] = ['mobile_user', 'institutional_client', 'org_collaborator', 'admin']
const STATUSES: UserStatus[] = ['active', 'inactive', 'suspended']

export function UsersAdminPage() {
  const { user: currentUser } = useAuth()
  const toast = useToast()

  const [search, setSearch] = useState('')
  const [role, setRole] = useState('')
  const [status, setStatus] = useState('')
  const [page, setPage] = useState(1)

  const debouncedSearch = useDebouncedValue(search)

  useEffect(() => {
    setPage(1)
  }, [debouncedSearch, role, status])

  const query = useMemo(
    () => ({
      search: debouncedSearch || undefined,
      role: role || undefined,
      status: status || undefined,
      page,
      per_page: PER_PAGE,
    }),
    [debouncedSearch, role, status, page],
  )

  const users = useUsers(query)
  const { updateRole, updateStatus } = useUserAdminMutations()

  const items = users.data?.items ?? []
  const active = items.filter((item) => item.status === 'active').length
  const admins = items.filter((item) => item.role === 'admin').length

  const columns: Column<User>[] = [
    {
      key: 'user',
      header: 'Persona',
      render: (row) => (
        <div>
          <strong>{row.full_name}</strong>
          <div className="muted text-sm">{row.email}</div>
        </div>
      ),
    },
    {
      key: 'role',
      header: 'Rol',
      render: (row) =>
        row.id === currentUser?.id ? (
          <Badge tone="primary">{roleLabels[row.role]}</Badge>
        ) : (
          <select
            className="input input--select"
            style={{ minHeight: 32, paddingTop: 2, paddingBottom: 2, width: 200 }}
            value={row.role}
            aria-label={`Rol de ${row.full_name}`}
            disabled={updateRole.isPending}
            onChange={(event) =>
              updateRole.mutate(
                { id: row.id, role: event.target.value as Role },
                {
                  onSuccess: () => toast.success('Rol actualizado', row.full_name),
                  onError: (error) => toast.error('No se pudo actualizar', humanizeError(error)),
                },
              )
            }
          >
            {ROLES.map((value) => (
              <option key={value} value={value}>
                {roleLabels[value]}
              </option>
            ))}
          </select>
        ),
    },
    {
      key: 'status',
      header: 'Estado',
      render: (row) =>
        row.id === currentUser?.id ? (
          <Badge tone={userStatusTone(row.status)}>{userStatusLabels[row.status]}</Badge>
        ) : (
          <select
            className="input input--select"
            style={{ minHeight: 32, paddingTop: 2, paddingBottom: 2, width: 140 }}
            value={row.status}
            aria-label={`Estado de ${row.full_name}`}
            disabled={updateStatus.isPending}
            onChange={(event) => {
              const nextStatus = event.target.value as UserStatus
              updateStatus.mutate(
                { id: row.id, status: nextStatus },
                {
                  onSuccess: () =>
                    toast.success('Estado actualizado', `${row.full_name}: ${userStatusLabels[nextStatus]}`),
                  onError: (error) => toast.error('No se pudo actualizar', humanizeError(error)),
                },
              )
            }}
          >
            {STATUSES.map((value) => (
              <option key={value} value={value}>
                {userStatusLabels[value]}
              </option>
            ))}
          </select>
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
  ]

  return (
    <>
      <PageHeader
        title="Usuarios de la plataforma"
        description="Administración de cuentas, roles y estados. Visible solo para administradores de Abudemy."
        actions={
          <Button variant="secondary" icon="refresh" onClick={() => void users.refetch()} loading={users.isFetching}>
            Actualizar
          </Button>
        }
      />

      <Callout tone="warning" title="Acciones sensibles">
        Cambiar el rol o suspender una cuenta afecta de inmediato su acceso. Al suspender o
        desactivar a una persona, se cierran todas sus sesiones. No puedes modificar tu propia
        cuenta desde esta pantalla para evitar un bloqueo accidental.
      </Callout>

      <section className="stat-grid">
        <StatCard
          label="Usuarios (página actual)"
          value={formatNumber(items.length)}
          icon="users"
          description="Cuentas listadas en esta página según los filtros aplicados."
          updatedAt={users.dataUpdatedAt ? new Date(users.dataUpdatedAt) : undefined}
        />
        <StatCard
          label="Activos"
          value={formatNumber(active)}
          tone="success"
          icon="check"
          description="Cuentas con estado «activo» dentro de la página actual."
          updatedAt={users.dataUpdatedAt ? new Date(users.dataUpdatedAt) : undefined}
        />
        <StatCard
          label="Administradores"
          value={formatNumber(admins)}
          tone="warning"
          icon="shield"
          description="Cuentas con rol de administrador de plataforma dentro de la página actual."
          updatedAt={users.dataUpdatedAt ? new Date(users.dataUpdatedAt) : undefined}
        />
        <StatCard
          label="Total en la plataforma"
          value={formatNumber(users.data?.pagination.total)}
          tone="primary"
          icon="users"
          description="Total de cuentas que coinciden con los filtros, en todas las páginas."
          updatedAt={users.dataUpdatedAt ? new Date(users.dataUpdatedAt) : undefined}
        />
      </section>

      <Card>
        <div className="filters-bar">
          <div className="field toolbar__search" style={{ flex: '2 1 240px' }}>
            <label className="field__label" htmlFor="user-search">
              Buscar
            </label>
            <TextInput
              id="user-search"
              leadingIcon="search"
              placeholder="Nombre o correo…"
              value={search}
              onChange={(event) => setSearch(event.target.value)}
            />
          </div>

          <SelectInput label="Rol" value={role} onChange={(event) => setRole(event.target.value)}>
            <option value="">Todos</option>
            {ROLES.map((value) => (
              <option key={value} value={value}>
                {roleLabels[value]}
              </option>
            ))}
          </SelectInput>

          <SelectInput label="Estado" value={status} onChange={(event) => setStatus(event.target.value)}>
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
                setRole('')
                setStatus('')
              }}
            >
              Limpiar filtros
            </Button>
          </div>
        </div>

        <CardBody>
          {users.isError ? (
            <ErrorState description={humanizeError(users.error)} onRetry={() => void users.refetch()} />
          ) : (
            <>
              <DataTable
                columns={columns}
                rows={items}
                rowKey={(row) => row.id}
                loading={users.isLoading}
                empty={
                  <EmptyState
                    title="No encontramos usuarios"
                    description="Ajusta los filtros de búsqueda, rol o estado."
                    icon="users"
                  />
                }
              />
              {users.data && users.data.pagination.total > PER_PAGE && (
                <Pagination
                  page={users.data.pagination.page}
                  perPage={users.data.pagination.per_page}
                  total={users.data.pagination.total}
                  totalPages={users.data.pagination.total_pages}
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
