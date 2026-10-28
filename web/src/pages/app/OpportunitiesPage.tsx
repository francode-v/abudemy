import { useMemo, useState, type FormEvent } from 'react'
import { humanizeError, type Opportunity, type OpportunityInput, type OpportunityStatus, type OpportunityType } from '@/api'
import { useOpportunities, useSaveOpportunity, useUpdateOpportunityStatus } from '@/hooks/queries'
import { useAuth } from '@/auth/AuthProvider'
import { useOrgScope } from '@/app/orgScope'
import { useToast } from '@/hooks/useToast'
import { Badge, Button, Card, CardBody, EmptyState, ErrorState, PageHeader, SelectInput, TextAreaInput, TextInput } from '@/components/ui/primitives'
import { DataTable, type Column } from '@/components/ui/data'
import { Modal } from '@/components/ui/overlays'
import { opportunityStatusLabels, opportunityTypeLabels } from '@/lib/dictionaries'
import { opportunityStatusTone } from '@/lib/status'
import { formatCurrency, formatDate, relativeTime } from '@/lib/format'

function emptyForm() {
  return {
    organization_id: '',
    title: '',
    description: '',
    type: 'job' as OpportunityType,
    status: 'draft' as OpportunityStatus,
    comuna: '',
    region: '',
    salary_min: '',
    salary_max: '',
    currency: 'CLP',
    contact_email: '',
    expires_at: '',
  }
}

export function OpportunitiesPage() {
  const { permissions, memberships } = useAuth()
  const { organizationId } = useOrgScope()
  const toast = useToast()

  const [statusFilter, setStatusFilter] = useState('')
  const [typeFilter, setTypeFilter] = useState('')
  const [formOpen, setFormOpen] = useState(false)
  const [editing, setEditing] = useState<Opportunity | null>(null)
  const [form, setForm] = useState(emptyForm)

  const query = useMemo(
    () => ({
      status: statusFilter || undefined,
      type: typeFilter || undefined,
      per_page: 50,
    }),
    [statusFilter, typeFilter],
  )

  const opportunities = useOpportunities(query)
  const save = useSaveOpportunity()
  const statusMutation = useUpdateOpportunityStatus()

  const canManage = permissions.canManageAnyContent
  const canManageRow = (organizationId?: string | null) =>
    permissions.canWriteOrganization(organizationId)

  const openCreate = () => {
    setEditing(null)
    setForm({ ...emptyForm(), organization_id: organizationId ?? '' })
    setFormOpen(true)
    save.reset()
  }

  const openEdit = (opportunity: Opportunity) => {
    setEditing(opportunity)
    setForm({
      organization_id: opportunity.organization_id ?? '',
      title: opportunity.title,
      description: opportunity.description ?? '',
      type: opportunity.type,
      status: opportunity.status,
      comuna: opportunity.comuna ?? '',
      region: opportunity.region ?? '',
      salary_min: opportunity.salary_min != null ? String(opportunity.salary_min) : '',
      salary_max: opportunity.salary_max != null ? String(opportunity.salary_max) : '',
      currency: opportunity.currency ?? 'CLP',
      contact_email: opportunity.contact_email ?? '',
      expires_at: opportunity.expires_at ? opportunity.expires_at.slice(0, 10) : '',
    })
    setFormOpen(true)
    save.reset()
  }

  const submit = (event: FormEvent) => {
    event.preventDefault()
    if (form.title.trim().length < 3) {
      toast.warning('Título requerido', 'Ingresa un título de al menos 3 caracteres.')
      return
    }
    const input: OpportunityInput = {
      organization_id: form.organization_id || null,
      title: form.title.trim(),
      description: form.description.trim(),
      type: form.type,
      status: form.status,
      comuna: form.comuna.trim(),
      region: form.region.trim(),
      salary_min: form.salary_min ? Number(form.salary_min) : null,
      salary_max: form.salary_max ? Number(form.salary_max) : null,
      currency: form.currency.trim() || 'CLP',
      contact_email: form.contact_email.trim(),
      expires_at: form.expires_at ? new Date(`${form.expires_at}T23:59:59`).toISOString() : null,
    }
    save.mutate(
      { id: editing?.id, input },
      {
        onSuccess: () => {
          toast.success(editing ? 'Oportunidad actualizada' : 'Oportunidad creada', input.title)
          setFormOpen(false)
        },
        onError: (error) => toast.error('No se pudo guardar', humanizeError(error)),
      },
    )
  }

  const columns: Column<Opportunity>[] = [
    {
      key: 'title',
      header: 'Oportunidad',
      render: (row) => (
        <div>
          <strong>{row.title}</strong>
          <div className="muted text-sm">{opportunityTypeLabels[row.type]}</div>
        </div>
      ),
    },
    {
      key: 'location',
      header: 'Ubicación',
      render: (row) => row.comuna || '—',
      hideOnMobile: true,
    },
    {
      key: 'salary',
      header: 'Rango',
      render: (row) =>
        row.salary_min || row.salary_max
          ? `${row.salary_min ? formatCurrency(row.salary_min) : '—'} – ${row.salary_max ? formatCurrency(row.salary_max) : '—'}`
          : 'No informado',
      hideOnMobile: true,
    },
    {
      key: 'expires',
      header: 'Vence',
      render: (row) => (row.expires_at ? formatDate(row.expires_at) : 'Sin fecha'),
      hideOnMobile: true,
    },
    {
      key: 'status',
      header: 'Estado',
      render: (row) =>
        canManageRow(row.organization_id) ? (
          <select
            className="input input--select"
            style={{ minHeight: 32, paddingTop: 2, paddingBottom: 2, width: 130 }}
            value={row.status}
            aria-label={`Estado de ${row.title}`}
            disabled={statusMutation.isPending}
            onChange={(event) =>
              statusMutation.mutate(
                { id: row.id, status: event.target.value as Opportunity['status'] },
                {
                  onSuccess: () => toast.success('Estado actualizado'),
                  onError: (error) => toast.error('No se pudo actualizar', humanizeError(error)),
                },
              )
            }
          >
            <option value="draft">Borrador</option>
            <option value="open">Abierta</option>
            <option value="closed">Cerrada</option>
          </select>
        ) : (
          <Badge tone={opportunityStatusTone(row.status)}>{opportunityStatusLabels[row.status]}</Badge>
        ),
    },
    {
      key: 'actions',
      header: '',
      align: 'right',
      render: (row) =>
        canManageRow(row.organization_id) ? (
          <Button variant="ghost" size="sm" icon="edit" onClick={() => openEdit(row)}>
            Editar
          </Button>
        ) : null,
    },
  ]

  return (
    <>
      <PageHeader
        title="Oportunidades"
        description="Empleos, trabajos puntuales, voluntariados y oportunidades económicas para personas mayores."
        actions={
          canManage ? (
            <Button icon="plus" onClick={openCreate}>
              Nueva oportunidad
            </Button>
          ) : undefined
        }
      />

      <Card>
        <div className="filters-bar">
          <SelectInput label="Tipo" value={typeFilter} onChange={(event) => setTypeFilter(event.target.value)}>
            <option value="">Todos</option>
            <option value="job">Empleo</option>
            <option value="gig">Trabajo puntual</option>
            <option value="volunteer">Voluntariado</option>
            <option value="economic">Oportunidad económica</option>
          </SelectInput>
          <SelectInput label="Estado" value={statusFilter} onChange={(event) => setStatusFilter(event.target.value)}>
            <option value="">Todos</option>
            <option value="draft">Borrador</option>
            <option value="open">Abierta</option>
            <option value="closed">Cerrada</option>
          </SelectInput>
        </div>

        <CardBody>
          {opportunities.isError ? (
            <ErrorState description={humanizeError(opportunities.error)} onRetry={() => void opportunities.refetch()} />
          ) : (
            <>
              <DataTable
                columns={columns}
                rows={opportunities.data?.items ?? []}
                rowKey={(row) => row.id}
                loading={opportunities.isLoading}
                empty={
                  <EmptyState
                    title="No hay oportunidades"
                    description="Crea una oportunidad para que las personas mayores puedan postular."
                    icon="trend"
                  />
                }
              />
              <p className="split-note" style={{ marginTop: 8 }}>
                Actualizado {relativeTime(opportunities.dataUpdatedAt)}.
              </p>
            </>
          )}
        </CardBody>
      </Card>

      <Modal
        open={formOpen}
        onClose={() => setFormOpen(false)}
        title={editing ? 'Editar oportunidad' : 'Nueva oportunidad'}
        description="La información quedará publicada en la plataforma."
        size="lg"
        footer={
          <>
            <Button variant="secondary" onClick={() => setFormOpen(false)} disabled={save.isPending}>
              Cancelar
            </Button>
            <Button form="opportunity-form" type="submit" loading={save.isPending}>
              {editing ? 'Guardar cambios' : 'Crear oportunidad'}
            </Button>
          </>
        }
      >
        <form id="opportunity-form" onSubmit={submit} noValidate>
          <div className="form-grid">
            <div className="form-grid--full">
              <TextInput
                label="Título"
                required
                value={form.title}
                onChange={(event) => setForm({ ...form, title: event.target.value })}
              />
            </div>
            <SelectInput
              label="Organización"
              value={form.organization_id}
              onChange={(event) => setForm({ ...form, organization_id: event.target.value })}
            >
              <option value="">Sin organización</option>
              {memberships.map((membership) => (
                <option key={membership.organization.id} value={membership.organization.id}>
                  {membership.organization.name}
                </option>
              ))}
            </SelectInput>
            <SelectInput
              label="Tipo"
              value={form.type}
              onChange={(event) => setForm({ ...form, type: event.target.value as OpportunityType })}
            >
              <option value="job">Empleo</option>
              <option value="gig">Trabajo puntual</option>
              <option value="volunteer">Voluntariado</option>
              <option value="economic">Oportunidad económica</option>
            </SelectInput>
            <SelectInput
              label="Estado"
              value={form.status}
              onChange={(event) => setForm({ ...form, status: event.target.value as OpportunityStatus })}
            >
              <option value="draft">Borrador</option>
              <option value="open">Abierta</option>
              <option value="closed">Cerrada</option>
            </SelectInput>
            <TextInput
              label="Comuna"
              value={form.comuna}
              onChange={(event) => setForm({ ...form, comuna: event.target.value })}
            />
            <TextInput
              label="Región"
              value={form.region}
              onChange={(event) => setForm({ ...form, region: event.target.value })}
            />
            <TextInput
              label="Remuneración mínima"
              type="number"
              value={form.salary_min}
              onChange={(event) => setForm({ ...form, salary_min: event.target.value })}
            />
            <TextInput
              label="Remuneración máxima"
              type="number"
              value={form.salary_max}
              onChange={(event) => setForm({ ...form, salary_max: event.target.value })}
            />
            <TextInput
              label="Correo de contacto"
              type="email"
              value={form.contact_email}
              onChange={(event) => setForm({ ...form, contact_email: event.target.value })}
            />
            <TextInput
              label="Vence el"
              type="date"
              value={form.expires_at}
              onChange={(event) => setForm({ ...form, expires_at: event.target.value })}
            />
            <div className="form-grid--full">
              <TextAreaInput
                label="Descripción"
                rows={4}
                value={form.description}
                onChange={(event) => setForm({ ...form, description: event.target.value })}
              />
            </div>
          </div>
        </form>
      </Modal>
    </>
  )
}
