import { useMemo, useState, type FormEvent } from 'react'
import type { Activity, ActivityInput, ActivityStatus, ActivityType, Modality } from '@/api'
import { ApiError, humanizeError } from '@/api'
import { Button, Callout, Field, SelectInput, TextAreaInput, TextInput } from '@/components/ui/primitives'
import { useAuth } from '@/auth/AuthProvider'
import { canWriteOrgRole } from '@/auth/permissions'
import { toDateTimeInputValue } from '@/lib/format'

export interface ActivityFormProps {
  initial?: Activity | null
  submitting?: boolean
  error?: unknown
  onSubmit: (input: ActivityInput) => void
  onCancel: () => void
}

function toFormState(activity?: Activity | null) {
  return {
    organization_id: activity?.organization_id ?? '',
    title: activity?.title ?? '',
    description: activity?.description ?? '',
    type: (activity?.type ?? 'workshop') as ActivityType,
    modality: (activity?.modality ?? 'in_person') as Modality,
    status: (activity?.status ?? 'draft') as ActivityStatus,
    location: activity?.location ?? '',
    comuna: activity?.comuna ?? '',
    region: activity?.region ?? '',
    latitude: activity?.latitude != null ? String(activity.latitude) : '',
    longitude: activity?.longitude != null ? String(activity.longitude) : '',
    start_at: toDateTimeInputValue(activity?.start_at),
    end_at: toDateTimeInputValue(activity?.end_at),
    capacity: activity?.capacity != null ? String(activity.capacity) : '0',
    price: activity?.price != null ? String(activity.price) : '0',
    currency: activity?.currency ?? 'CLP',
  }
}

export function ActivityForm({ initial, submitting, error, onSubmit, onCancel }: ActivityFormProps) {
  const { memberships, permissions } = useAuth()
  const [form, setForm] = useState(() => toFormState(initial))
  const [localErrors, setLocalErrors] = useState<Record<string, string>>({})

  const apiError = error instanceof ApiError ? error : undefined
  const fieldError = (field: string) => localErrors[field] ?? apiError?.fieldError(field)

  const organizationOptions = useMemo(
    () =>
      memberships
        .filter((membership) => permissions.isAdmin || canWriteOrgRole(membership.role))
        .map((membership) => ({ id: membership.organization.id, name: membership.organization.name })),
    [memberships, permissions.isAdmin],
  )

  const update = <K extends keyof typeof form>(key: K, value: (typeof form)[K]) => {
    setForm((prev) => ({ ...prev, [key]: value }))
    setLocalErrors((prev) => ({ ...prev, [key]: '' }))
  }

  const validate = (): boolean => {
    const errors: Record<string, string> = {}
    if (form.title.trim().length < 3) errors.title = 'Ingresa un título de al menos 3 caracteres.'
    if (!form.start_at) errors.start_at = 'Indica la fecha y hora de inicio.'
    if (form.end_at && form.start_at && new Date(form.end_at) < new Date(form.start_at)) {
      errors.end_at = 'La fecha de término debe ser posterior al inicio.'
    }
    if (Number(form.capacity) < 0) errors.capacity = 'El cupo no puede ser negativo.'
    if (Number(form.price) < 0) errors.price = 'El precio no puede ser negativo.'
    if (Boolean(form.latitude) !== Boolean(form.longitude)) {
      errors.latitude = 'Completa latitud y longitud juntas si deseas ubicar la actividad.'
    }
    setLocalErrors(errors)
    return Object.keys(errors).length === 0
  }

  const submit = (event: FormEvent) => {
    event.preventDefault()
    if (!validate()) return
    const hasLocation = form.latitude !== '' && form.longitude !== ''
    const input: ActivityInput = {
      organization_id: form.organization_id || null,
      title: form.title.trim(),
      description: form.description.trim(),
      type: form.type,
      modality: form.modality,
      status: form.status,
      location: form.location.trim(),
      comuna: form.comuna.trim(),
      region: form.region.trim(),
      latitude: hasLocation ? Number(form.latitude) : null,
      longitude: hasLocation ? Number(form.longitude) : null,
      start_at: new Date(form.start_at).toISOString(),
      end_at: form.end_at ? new Date(form.end_at).toISOString() : null,
      capacity: Number(form.capacity) || 0,
      price: Number(form.price) || 0,
      currency: form.currency.trim() || 'CLP',
    }
    onSubmit(input)
  }

  return (
    <form onSubmit={submit} noValidate>
      {apiError && (
        <div style={{ marginBottom: 16 }}>
          <Callout tone="danger" title="No se pudo guardar">
            {humanizeError(apiError)}
            {apiError.isForbidden && (
              <p style={{ marginTop: 6 }}>
                Para crear actividades necesitas permisos de edición en la organización
                seleccionada, o ser administrador de la plataforma.
              </p>
            )}
          </Callout>
        </div>
      )}

      <div className="form-grid">
        <div className="form-grid--full">
          <SelectInput
            label="Organización responsable"
            value={form.organization_id}
            error={fieldError('organization_id')}
            hint="Las actividades sin organización son plataforma y requieren rol de administrador."
            onChange={(event) => update('organization_id', event.target.value)}
          >
            <option value="">Sin organización (plataforma)</option>
            {organizationOptions.map((option) => (
              <option key={option.id} value={option.id}>
                {option.name}
              </option>
            ))}
          </SelectInput>
        </div>

        <div className="form-grid--full">
          <TextInput
            label="Título de la actividad"
            required
            value={form.title}
            error={fieldError('title')}
            onChange={(event) => update('title', event.target.value)}
          />
        </div>

        <div className="form-grid--full">
          <Field label="Descripción" error={fieldError('description')}>
            <TextAreaInput
              rows={3}
              value={form.description}
              onChange={(event) => update('description', event.target.value)}
            />
          </Field>
        </div>

        <SelectInput
          label="Tipo"
          value={form.type}
          error={fieldError('type')}
          onChange={(event) => update('type', event.target.value as ActivityType)}
        >
          <option value="workshop">Taller</option>
          <option value="course">Curso</option>
          <option value="event">Evento</option>
          <option value="support_group">Grupo de apoyo</option>
        </SelectInput>

        <SelectInput
          label="Modalidad"
          value={form.modality}
          error={fieldError('modality')}
          onChange={(event) => update('modality', event.target.value as Modality)}
        >
          <option value="in_person">Presencial</option>
          <option value="online">En línea</option>
          <option value="hybrid">Híbrida</option>
        </SelectInput>

        <SelectInput
          label="Estado de publicación"
          value={form.status}
          error={fieldError('status')}
          hint="«Borrador / en revisión» permite prepararla antes de publicarla."
          onChange={(event) => update('status', event.target.value as ActivityStatus)}
        >
          <option value="draft">Borrador / en revisión</option>
          <option value="published">Publicada</option>
          <option value="cancelled">Cancelada</option>
          <option value="finished">Finalizada</option>
        </SelectInput>
      </div>

      <div className="form-section">
        <p className="form-section__title">Programación</p>
        <div className="form-grid">
          <TextInput
            label="Inicio"
            type="datetime-local"
            required
            value={form.start_at}
            error={fieldError('start_at')}
            onChange={(event) => update('start_at', event.target.value)}
          />
          <TextInput
            label="Término"
            type="datetime-local"
            value={form.end_at}
            error={fieldError('end_at')}
            onChange={(event) => update('end_at', event.target.value)}
          />
          <TextInput
            label="Cupo"
            type="number"
            min={0}
            value={form.capacity}
            error={fieldError('capacity')}
            hint="0 significa sin cupo máximo."
            onChange={(event) => update('capacity', event.target.value)}
          />
          <TextInput
            label="Precio"
            type="number"
            min={0}
            step="0.01"
            value={form.price}
            error={fieldError('price')}
            onChange={(event) => update('price', event.target.value)}
          />
          <TextInput
            label="Moneda"
            value={form.currency}
            error={fieldError('currency')}
            onChange={(event) => update('currency', event.target.value)}
          />
        </div>
      </div>

      <div className="form-section">
        <p className="form-section__title">Lugar</p>
        <div className="form-grid">
          <div className="form-grid--full">
            <TextInput
              label="Lugar / sede"
              value={form.location}
              error={fieldError('location')}
              onChange={(event) => update('location', event.target.value)}
            />
          </div>
          <TextInput
            label="Comuna"
            value={form.comuna}
            error={fieldError('comuna')}
            onChange={(event) => update('comuna', event.target.value)}
          />
          <TextInput
            label="Región"
            value={form.region}
            error={fieldError('region')}
            onChange={(event) => update('region', event.target.value)}
          />
          <TextInput
            label="Latitud"
            inputMode="decimal"
            value={form.latitude}
            error={fieldError('latitude')}
            onChange={(event) => update('latitude', event.target.value)}
          />
          <TextInput
            label="Longitud"
            inputMode="decimal"
            value={form.longitude}
            error={fieldError('longitude')}
            onChange={(event) => update('longitude', event.target.value)}
          />
        </div>
      </div>

      <div className="modal__footer" style={{ marginTop: 20, border: 'none', background: 'none', padding: 0 }}>
        <Button type="button" variant="secondary" onClick={onCancel} disabled={submitting}>
          Cancelar
        </Button>
        <Button type="submit" loading={submitting}>
          {initial ? 'Guardar cambios' : 'Crear actividad'}
        </Button>
      </div>
    </form>
  )
}
