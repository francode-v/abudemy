import { useMemo, useState, type FormEvent } from 'react'
import type { ScheduleEntry, Service, ServiceInput, ServiceStatus, ServiceType } from '@/api'
import { ApiError, humanizeError } from '@/api'
import { Button, Callout, Field, IconButton, SelectInput, TextAreaInput, TextInput } from '@/components/ui/primitives'
import { useAuth } from '@/auth/AuthProvider'
import { useCategories } from '@/hooks/queries'
import { canWriteOrgRole } from '@/auth/permissions'
import { dayOfWeekLabels } from '@/lib/dictionaries'

export interface ServiceFormProps {
  initial?: Service | null
  submitting?: boolean
  error?: unknown
  onSubmit: (input: ServiceInput) => void
  onCancel: () => void
}

const emptyScheduleRow: ScheduleEntry = { day_of_week: 1, opens_at: '09:00', closes_at: '13:00' }

function toFormState(service?: Service | null) {
  return {
    organization_id: service?.organization_id ?? '',
    category_id: service?.category_id ?? '',
    name: service?.name ?? '',
    description: service?.description ?? '',
    type: (service?.type ?? 'establishment') as ServiceType,
    status: (service?.status ?? 'active') as ServiceStatus,
    address: service?.address ?? '',
    comuna: service?.comuna ?? '',
    region: service?.region ?? '',
    phone: service?.phone ?? '',
    email: service?.email ?? '',
    website: service?.website ?? '',
    latitude: service?.latitude != null ? String(service.latitude) : '',
    longitude: service?.longitude != null ? String(service.longitude) : '',
    schedule: service?.schedule?.length ? service.schedule : [],
  }
}

export function ServiceForm({ initial, submitting, error, onSubmit, onCancel }: ServiceFormProps) {
  const { memberships, permissions } = useAuth()
  const categories = useCategories()
  const [form, setForm] = useState(() => toFormState(initial))
  const [localErrors, setLocalErrors] = useState<Record<string, string>>({})

  const apiError = error instanceof ApiError ? error : undefined
  const fieldError = (field: string) => localErrors[field] ?? apiError?.fieldError(field)

  const organizationOptions = useMemo(() => {
    if (permissions.isAdmin) {
      return memberships.map((membership) => ({
        id: membership.organization.id,
        name: membership.organization.name,
      }))
    }
    return memberships
      .filter((membership) => canWriteOrgRole(membership.role))
      .map((membership) => ({ id: membership.organization.id, name: membership.organization.name }))
  }, [memberships, permissions.isAdmin])

  const update = <K extends keyof typeof form>(key: K, value: (typeof form)[K]) => {
    setForm((prev) => ({ ...prev, [key]: value }))
    setLocalErrors((prev) => ({ ...prev, [key]: '' }))
  }

  const updateScheduleRow = (index: number, patch: Partial<ScheduleEntry>) => {
    setForm((prev) => ({
      ...prev,
      schedule: prev.schedule.map((row, i) => (i === index ? { ...row, ...patch } : row)),
    }))
  }

  const validate = (): boolean => {
    const errors: Record<string, string> = {}
    if (form.name.trim().length < 3) errors.name = 'Ingresa un nombre de al menos 3 caracteres.'
    if (!permissions.isAdmin && !initial && !form.organization_id) {
      errors.organization_id = 'Selecciona la organización responsable del servicio.'
    }
    if (form.latitude && (Number.isNaN(Number(form.latitude)) || Math.abs(Number(form.latitude)) > 90)) {
      errors.latitude = 'La latitud debe estar entre -90 y 90.'
    }
    if (form.longitude && (Number.isNaN(Number(form.longitude)) || Math.abs(Number(form.longitude)) > 180)) {
      errors.longitude = 'La longitud debe estar entre -180 y 180.'
    }
    if (form.email && !/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(form.email)) {
      errors.email = 'Ingresa un correo válido.'
    }
    if (Boolean(form.latitude) !== Boolean(form.longitude)) {
      errors.latitude = 'Completa latitud y longitud juntas para ubicar el servicio en el mapa.'
    }
    setLocalErrors(errors)
    return Object.keys(errors).length === 0
  }

  const submit = (event: FormEvent) => {
    event.preventDefault()
    if (!validate()) return
    const hasLocation = form.latitude !== '' && form.longitude !== ''
    const input: ServiceInput = {
      organization_id: form.organization_id || null,
      category_id: form.category_id || null,
      name: form.name.trim(),
      description: form.description.trim(),
      type: form.type,
      status: form.status,
      address: form.address.trim(),
      comuna: form.comuna.trim(),
      region: form.region.trim(),
      phone: form.phone.trim(),
      email: form.email.trim(),
      website: form.website.trim(),
      latitude: hasLocation ? Number(form.latitude) : null,
      longitude: hasLocation ? Number(form.longitude) : null,
      schedule: form.schedule,
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
                Solo los administradores y los integrantes con rol de edición pueden modificar
                servicios de una organización.
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
            hint={
              permissions.isAdmin
                ? 'Como administrador puedes dejar el servicio a nivel de plataforma sin organización.'
                : 'Solo puedes elegir organizaciones donde tienes permisos de edición.'
            }
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
            label="Nombre del servicio"
            required
            value={form.name}
            error={fieldError('name')}
            onChange={(event) => update('name', event.target.value)}
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
          label="Categoría"
          value={form.category_id}
          error={fieldError('category_id')}
          onChange={(event) => update('category_id', event.target.value)}
        >
          <option value="">Sin categoría</option>
          {(categories.data ?? []).map((category) => (
            <option key={category.id} value={category.id}>
              {category.name}
            </option>
          ))}
        </SelectInput>

        <SelectInput
          label="Tipo"
          value={form.type}
          error={fieldError('type')}
          onChange={(event) => update('type', event.target.value as ServiceType)}
        >
          <option value="establishment">Establecimiento</option>
          <option value="resource">Recurso de apoyo</option>
        </SelectInput>

        <SelectInput
          label="Estado"
          value={form.status}
          error={fieldError('status')}
          hint="«Borrador / en revisión» oculta el servicio de la consulta pública activa."
          onChange={(event) => update('status', event.target.value as ServiceStatus)}
        >
          <option value="active">Activo</option>
          <option value="inactive">Inactivo</option>
          <option value="draft">Borrador / en revisión</option>
        </SelectInput>
      </div>

      <div className="form-section">
        <p className="form-section__title">Ubicación</p>
        <div className="form-grid">
          <div className="form-grid--full">
            <TextInput
              label="Dirección"
              value={form.address}
              error={fieldError('address')}
              onChange={(event) => update('address', event.target.value)}
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
            placeholder="-33.4489"
            value={form.latitude}
            error={fieldError('latitude')}
            onChange={(event) => update('latitude', event.target.value)}
          />
          <TextInput
            label="Longitud"
            inputMode="decimal"
            placeholder="-70.6693"
            value={form.longitude}
            error={fieldError('longitude')}
            hint="Se usan para ubicar el servicio en el mapa territorial."
            onChange={(event) => update('longitude', event.target.value)}
          />
        </div>
      </div>

      <div className="form-section">
        <p className="form-section__title">Contacto</p>
        <div className="form-grid">
          <TextInput
            label="Teléfono"
            value={form.phone}
            error={fieldError('phone')}
            onChange={(event) => update('phone', event.target.value)}
          />
          <TextInput
            label="Correo electrónico"
            type="email"
            value={form.email}
            error={fieldError('email')}
            onChange={(event) => update('email', event.target.value)}
          />
          <div className="form-grid--full">
            <TextInput
              label="Sitio web"
              placeholder="https://"
              value={form.website}
              error={fieldError('website')}
              onChange={(event) => update('website', event.target.value)}
            />
          </div>
        </div>
      </div>

      <div className="form-section">
        <div className="row row--between" style={{ marginBottom: 10 }}>
          <p className="form-section__title" style={{ marginBottom: 0 }}>
            Horarios de atención
          </p>
          <Button
            type="button"
            variant="secondary"
            size="sm"
            icon="plus"
            onClick={() => update('schedule', [...form.schedule, { ...emptyScheduleRow }])}
          >
            Agregar horario
          </Button>
        </div>

        {form.schedule.length === 0 ? (
          <p className="muted text-sm">Sin horarios declarados. Puedes agregar los tramos de atención.</p>
        ) : (
          <div className="schedule-list">
            {form.schedule.map((row, index) => (
              <div className="schedule-row" key={index}>
                <SelectInput
                  label="Día"
                  value={String(row.day_of_week)}
                  onChange={(event) => updateScheduleRow(index, { day_of_week: Number(event.target.value) })}
                >
                  {dayOfWeekLabels.map((label, day) => (
                    <option key={label} value={day}>
                      {label}
                    </option>
                  ))}
                </SelectInput>
                <TextInput
                  label="Apertura"
                  type="time"
                  value={row.opens_at}
                  onChange={(event) => updateScheduleRow(index, { opens_at: event.target.value })}
                />
                <TextInput
                  label="Cierre"
                  type="time"
                  value={row.closes_at}
                  onChange={(event) => updateScheduleRow(index, { closes_at: event.target.value })}
                />
                <div className="schedule-row__remove">
                  <IconButton
                    icon="trash"
                    label="Quitar horario"
                    variant="ghost"
                    onClick={() =>
                      update(
                        'schedule',
                        form.schedule.filter((_, i) => i !== index),
                      )
                    }
                  />
                </div>
              </div>
            ))}
          </div>
        )}
      </div>

      <div className="modal__footer" style={{ marginTop: 20, border: 'none', background: 'none', padding: 0 }}>
        <Button type="button" variant="secondary" onClick={onCancel} disabled={submitting}>
          Cancelar
        </Button>
        <Button type="submit" loading={submitting}>
          {initial ? 'Guardar cambios' : 'Crear servicio'}
        </Button>
      </div>
    </form>
  )
}
