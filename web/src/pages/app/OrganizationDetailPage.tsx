import { useState, type FormEvent } from 'react'
import { Link, useParams } from 'react-router-dom'
import {
  humanizeError,
  type OrganizationInput,
  type OrganizationMember,
  type OrgMemberRole,
} from '@/api'
import {
  useOrganization,
  useOrganizationMemberMutations,
  useOrganizationMembers,
  useSaveOrganization,
} from '@/hooks/queries'
import { useAuth } from '@/auth/AuthProvider'
import { useToast } from '@/hooks/useToast'
import {
  Badge,
  Button,
  Callout,
  Card,
  CardBody,
  CardHeader,
  EmptyState,
  ErrorState,
  PageHeader,
  SelectInput,
  Spinner,
  TextInput,
} from '@/components/ui/primitives'
import { DataTable, KeyValue, type Column } from '@/components/ui/data'
import { ConfirmDialog, Modal } from '@/components/ui/overlays'
import { OrganizationForm } from '@/components/forms/OrganizationForm'
import { Icon } from '@/components/ui/Icon'
import { organizationTypeLabels, orgMemberRoleLabels } from '@/lib/dictionaries'
import { formatDate, formatDateTime, relativeTime } from '@/lib/format'

const MEMBER_ROLES: OrgMemberRole[] = ['owner', 'admin', 'editor', 'viewer']

export function OrganizationDetailPage() {
  const { id } = useParams<{ id: string }>()
  const { permissions, user } = useAuth()
  const toast = useToast()

  const organization = useOrganization(id)
  const members = useOrganizationMembers(id)
  const save = useSaveOrganization()
  const memberMutations = useOrganizationMemberMutations(id ?? '')

  const [editOpen, setEditOpen] = useState(false)
  const [addOpen, setAddOpen] = useState(false)
  const [removing, setRemoving] = useState<OrganizationMember | null>(null)
  const [memberForm, setMemberForm] = useState<{ email: string; role: OrgMemberRole }>({
    email: '',
    role: 'viewer',
  })

  const canWrite = permissions.canWriteOrganization(id)
  const canManage = permissions.canManageOrganization(id)

  const openAdd = () => {
    setMemberForm({ email: '', role: 'viewer' })
    setAddOpen(true)
    memberMutations.add.reset()
  }

  const submitMember = (event: FormEvent) => {
    event.preventDefault()
    if (!/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(memberForm.email.trim())) {
      toast.warning('Correo inválido', 'Ingresa el correo de una persona ya registrada en Abudemy.')
      return
    }
    memberMutations.add.mutate(
      { email: memberForm.email.trim(), role: memberForm.role },
      {
        onSuccess: () => {
          toast.success('Integrante agregado', memberForm.email.trim())
          setAddOpen(false)
        },
        onError: (error) => toast.error('No se pudo agregar', humanizeError(error)),
      },
    )
  }

  if (organization.isLoading) {
    return (
      <div className="content">
        <Spinner size="lg" />
      </div>
    )
  }

  if (organization.isError || !organization.data) {
    return (
      <div className="content">
        <ErrorState
          title="No pudimos cargar la organización"
          description={humanizeError(organization.error)}
          onRetry={() => void organization.refetch()}
        />
        <p className="split-note">
          <Link to="/app/organizaciones">
            <Icon name="chevron-left" size={12} /> Volver a organizaciones
          </Link>
        </p>
      </div>
    )
  }

  const item = organization.data

  const columns: Column<OrganizationMember>[] = [
    {
      key: 'member',
      header: 'Integrante',
      render: (row) => (
        <div>
          <strong>{row.user_full_name || 'Sin nombre'}</strong>
          <div className="muted text-sm">{row.user_email || row.user_id}</div>
          {row.user_id === user?.id && <Badge tone="info">Tú</Badge>}
        </div>
      ),
    },
    {
      key: 'role',
      header: 'Rol',
      render: (row) =>
        canManage && row.user_id !== user?.id ? (
          <select
            className="input input--select"
            style={{ minHeight: 32, paddingTop: 2, paddingBottom: 2, width: 160 }}
            value={row.role}
            aria-label={`Rol de ${row.user_full_name ?? row.user_id}`}
            disabled={memberMutations.updateRole.isPending}
            onChange={(event) =>
              memberMutations.updateRole.mutate(
                { userId: row.user_id, role: event.target.value as OrgMemberRole },
                {
                  onSuccess: () => toast.success('Rol actualizado'),
                  onError: (error) => toast.error('No se pudo actualizar', humanizeError(error)),
                },
              )
            }
          >
            {MEMBER_ROLES.map((role) => (
              <option key={role} value={role}>
                {orgMemberRoleLabels[role]}
              </option>
            ))}
          </select>
        ) : (
          <Badge tone={row.role === 'owner' ? 'primary' : row.role === 'viewer' ? 'neutral' : 'info'}>
            {orgMemberRoleLabels[row.role]}
          </Badge>
        ),
    },
    {
      key: 'since',
      header: 'Incorporación',
      render: (row) => <span className="text-sm muted">{formatDate(row.created_at)}</span>,
      hideOnMobile: true,
    },
    {
      key: 'actions',
      header: '',
      align: 'right',
      render: (row) =>
        canManage && row.user_id !== user?.id ? (
          <Button variant="ghost" size="sm" icon="trash" onClick={() => setRemoving(row)}>
            Quitar
          </Button>
        ) : row.user_id === user?.id ? (
          <span className="muted text-sm">Tu propia cuenta</span>
        ) : null,
    },
  ]

  return (
    <>
      <PageHeader
        title={item.name}
        breadcrumb={
          <Link to="/app/organizaciones">
            <Icon name="chevron-left" size={12} /> Volver a organizaciones
          </Link>
        }
        description={item.description || `${organizationTypeLabels[item.type] ?? 'Organización'} en Abudemy.`}
        actions={
          canWrite ? (
            <Button variant="secondary" icon="edit" onClick={() => setEditOpen(true)}>
              Editar organización
            </Button>
          ) : undefined
        }
      />

      <div className="row">
        <Badge tone={item.active ? 'success' : 'neutral'}>{item.active ? 'Activa' : 'Inactiva'}</Badge>
        <Badge tone="primary">{organizationTypeLabels[item.type] ?? item.type}</Badge>
        {item.legal_id && <Badge tone="neutral">{item.legal_id}</Badge>}
      </div>

      {!canManage && (
        <Callout tone="info" title="Permisos de integrantes">
          Puedes consultar la organización, pero solo los roles <strong>propietario</strong> y{' '}
          <strong>administrador</strong> pueden agregar integrantes o cambiar sus roles.
        </Callout>
      )}

      <div className="grid-2">
        <Card>
          <CardHeader title="Ficha de la organización" icon="organizations" />
          <CardBody>
            <KeyValue
              items={[
                { label: 'Tipo', value: organizationTypeLabels[item.type] ?? item.type },
                { label: 'RUT / Identificación', value: item.legal_id || '—' },
                { label: 'Correo', value: item.email || '—' },
                { label: 'Teléfono', value: item.phone || '—' },
                { label: 'Dirección', value: item.address || '—' },
                { label: 'Comuna', value: item.comuna || '—' },
                { label: 'Región', value: item.region || '—' },
                {
                  label: 'Coordenadas',
                  value:
                    item.latitude != null && item.longitude != null
                      ? `${item.latitude.toFixed(5)}, ${item.longitude.toFixed(5)}`
                      : 'Sin coordenadas',
                },
                { label: 'Creada', value: formatDateTime(item.created_at) },
                {
                  label: 'Última actualización',
                  value: `${formatDateTime(item.updated_at)} (${relativeTime(item.updated_at)})`,
                },
              ]}
            />
          </CardBody>
        </Card>

        <Card>
          <CardHeader
            title="Integrantes"
            description="Personas con acceso a los recursos de esta organización."
            icon="users"
            actions={
              canManage ? (
                <Button size="sm" icon="plus" onClick={openAdd}>
                  Agregar
                </Button>
              ) : undefined
            }
          />
          <CardBody>
            {members.isError ? (
              <ErrorState description={humanizeError(members.error)} onRetry={() => void members.refetch()} />
            ) : (
              <DataTable
                columns={columns}
                rows={members.data ?? []}
                rowKey={(row) => row.id}
                loading={members.isLoading}
                empty={
                  <EmptyState
                    title="Sin integrantes"
                    description="Agrega integrantes para que puedan administrar los recursos de la organización."
                    icon="users"
                  />
                }
              />
            )}
          </CardBody>
        </Card>
      </div>

      <Modal
        open={editOpen}
        onClose={() => setEditOpen(false)}
        title="Editar organización"
        description="Actualiza los datos de la organización."
        size="lg"
      >
        <OrganizationForm
          initial={item}
          submitting={save.isPending}
          error={save.error}
          onCancel={() => setEditOpen(false)}
          onSubmit={(input: OrganizationInput) => {
            save.mutate(
              { id: item.id, input },
              {
                onSuccess: () => {
                  toast.success('Organización actualizada', item.name)
                  setEditOpen(false)
                },
                onError: (error) => toast.error('No se pudo guardar', humanizeError(error)),
              },
            )
          }}
        />
      </Modal>

      <Modal
        open={addOpen}
        onClose={() => setAddOpen(false)}
        title="Agregar integrante"
        description="La persona debe tener una cuenta ya registrada en Abudemy."
        size="sm"
        footer={
          <>
            <Button variant="secondary" onClick={() => setAddOpen(false)} disabled={memberMutations.add.isPending}>
              Cancelar
            </Button>
            <Button form="member-form" type="submit" loading={memberMutations.add.isPending}>
              Agregar integrante
            </Button>
          </>
        }
      >
        <form id="member-form" onSubmit={submitMember} noValidate>
          <div className="stack">
            <TextInput
              label="Correo de la persona"
              type="email"
              required
              leadingIcon="mail"
              placeholder="persona@organizacion.cl"
              value={memberForm.email}
              onChange={(event) => setMemberForm((prev) => ({ ...prev, email: event.target.value }))}
            />
            <SelectInput
              label="Rol dentro de la organización"
              value={memberForm.role}
              onChange={(event) =>
                setMemberForm((prev) => ({ ...prev, role: event.target.value as OrgMemberRole }))
              }
            >
              {MEMBER_ROLES.map((role) => (
                <option key={role} value={role}>
                  {orgMemberRoleLabels[role]}
                </option>
              ))}
            </SelectInput>
            <Callout tone="info">
              Los roles <strong>editor</strong> permiten modificar servicios y actividades; los roles{' '}
              <strong>propietario</strong> y <strong>administrador</strong> también gestionan
              integrantes.
            </Callout>
          </div>
        </form>
      </Modal>

      <ConfirmDialog
        open={removing !== null}
        title="Quitar integrante"
        message={
          <>
            ¿Confirmas quitar a <strong>{removing?.user_full_name || removing?.user_email}</strong> de
            esta organización? Perderá el acceso a sus recursos de inmediato.
          </>
        }
        confirmLabel="Quitar"
        loading={memberMutations.remove.isPending}
        onCancel={() => setRemoving(null)}
        onConfirm={() => {
          if (!removing) return
          memberMutations.remove.mutate(removing.user_id, {
            onSuccess: () => {
              toast.success('Integrante eliminado')
              setRemoving(null)
            },
            onError: (error) => {
              toast.error('No se pudo quitar', humanizeError(error))
              setRemoving(null)
            },
          })
        }}
      />
    </>
  )
}
