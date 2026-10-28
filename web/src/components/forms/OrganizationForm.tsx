import { useState, type FormEvent } from 'react'
import type { Organization, OrganizationInput, OrganizationType } from '@/api'
import { ApiError, humanizeError } from '@/api'
import { Button, Callout, CheckboxInput, Field, SelectInput, TextAreaInput, TextInput } from '@/components/ui/primitives'
import { organizationTypeLabels } from '@/lib/dictionaries'

export interface OrganizationFormProps {
  initial?: Organization | null
  submitting?: boolean
  error?: unknown
  onSubmit: (input: OrganizationInput) => void
  onCancel: () => void
}

function toFormState(org?: Organization | null) {
  return {
    name: org?.name ?? '',
    type: (org?.type ?? 'community') as OrganizationType,
    legal_id: org?.legal_id ?? '',
    description: org?.description ?? '',
    email: org?.email ?? '',
    phone: org?.phone ?? '',
    address: org?.address ?? '',
    comuna: org?.comuna ?? '',
    region: org?.region ?? '',
    latitude: org?.latitude != null ? String(org.latitude) : '',
    longitude: org?.longitude != null ? String(org.longitude) : '',
    active: org?.active ?? true,
  }
}

/**
 * Formulario de organización. Reutilizado por el listado y por el detalle para
 * mantener una única fuente de validación en el cliente (el backend revalida y
 * sigue siendo la autoridad final).
 */
export function OrganizationForm({ initial, submitting, error, onSubmit, onCancel }: OrganizationFormProps) {
  const [form, setForm] = useState(() => toFormState(initial))
  const [localErrors, setLocalErrors] = useState<Record<string, string>>({})

  const apiError = error instanceof ApiError ? error : undefined
  const fieldError = (field: string) => localErrors[field] ?? apiError?.fieldError(field)

  const update = <K extends keyof typeof form>(key: K, value: (typeof form)[K]) => {
    setForm((prev) => ({ ...prev, [key]: value }))
    setLocalErrors((prev) => ({ ...prev, [key]: '' }))
  }

  const validate = (): boolean => {
    const errors: Record<string, string> = {}
    if (form.name.trim().length < 3) errors.name = 'Ingresa un nombre de al menos 3 caracteres.'
    if (form.email && !/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(form.email)) {
      errors.email = 'Ingresa un correo válido.'
    }
    if (form.latitude && (Number.isNaN(Number(form.latitude)) || Math.abs(Number(form.latitude)) > 90)) {
      errors.latitude = 'La latitud debe estar entre -90 y 90.'
    }
    if (form.longitude && (Number.isNaN(Number(form.longitude)) || Math.abs(Number(form.longitude)) > 180)) {
      errors.longitude = 'La longitud debe estar entre -180 y 180.'
    }
    if (Boolean(form.latitude) !== Boolean(form.longitude)) {
      errors.latitude = 'Completa latitud y longitud juntas para ubicar la organización en el mapa.'
    }
    setLocalErrors(errors)
    return Object.keys(errors).length === 0
  }

  const submit = (event: FormEvent) => {
    event.preventDefault()
    if (!validate()) return
    const hasLocation = form.latitude !== '' && form.longitude !== ''
    const input: OrganizationInput = {
      name: form.name.trim(),
      type: form.type,
      legal_id: form.legal_id.trim(),
      description: form.description.trim(),
      email: form.email.trim(),
      phone: form.phone.trim(),
      address: form.address.trim(),
      comuna: form.comuna.trim(),
      region: form.region.trim(),
      latitude: hasLocation ? Number(form.latitude) : null,
      longitude: hasLocation ? Number(form.longitude) : null,
      active: form.active,
    }
    onSubmit(input)
  }

  return (
    <form onSubmit={submit} noValidate>
      {apiError && (
        <div style={{ marginBottom: 16 }}>
          <Callout tone="danger" title="No se pudo guardar la organización">
            {humanizeError(apiError)}
            {apiError.isForbidden && (
              <p style={{ marginTop: 6 }}>
                Solo administradores y clientes institucionales pueden crear organizaciones; la
                edición requiere rol propietario, administrador o editor.
              </p>
            )}
          </Callout>
        </div>
      )}

      <div className="form-grid">
        <div className="form-grid--full">
          <TextInput
            label="Nombre de la organización"
            required
            value={form.name}
            error={fieldError('name')}
            onChange={(event) => update('name', event.target.value)}
          />
        </div>

        <SelectInput
          label="Tipo"
          value={form.type}
          error={fieldError('type')}
          onChange={(event) => update('type', event.target.value as OrganizationType)}
        >
          {(Object.keys(organizationTypeLabels) as OrganizationType[]).map((type) => (
            <option key={type} value={type}>
              {organizationTypeLabels[type]}
            </option>
          ))}
        </SelectInput>

        <TextInput
          label="RUT / Identificación legal"
          value={form.legal_id}
          error={fieldError('legal_id')}
          onChange={(event) => update('legal_id', event.target.value)}
        />

        <div className="form-grid--full">
          <Field label="Descripción" error={fieldError('description')}>
            <TextAreaInput
              rows={3}
              value={form.description}
              onChange={(event) => update('description', event.target.value)}
            />
          </Field>
        </div>

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
            hint="Se usan para mostrar la organización en el mapa territorial."
            onChange={(event) => update('longitude', event.target.value)}
          />
        </div>
      </div>

      <div className="form-section">
        <CheckboxInput
          label="Organización activa"
          checked={form.active}
          onChange={(event) => update('active', event.target.checked)}
        />
        <p className="field__hint" style={{ marginTop: 4 }}>
          Desmarca para marcar la organización como inactiva. Las organizaciones inactivas siguen
          siendo visibles para sus integrantes, pero se señalan como fuera de operación.
        </p>
      </div>

      <div className="modal__footer" style={{ marginTop: 20, border: 'none', background: 'none', padding: 0 }}>
        <Button type="button" variant="secondary" onClick={onCancel} disabled={submitting}>
          Cancelar
        </Button>
        <Button type="submit" loading={submitting}>
          {initial ? 'Guardar cambios' : 'Crear organización'}
        </Button>
      </div>
    </form>
  )
}
