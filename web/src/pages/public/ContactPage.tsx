import { useState, type FormEvent } from 'react'
import { Icon } from '@/components/ui/Icon'
import { Button, Field, TextAreaInput, TextInput } from '@/components/ui/primitives'
import { PublicSection } from '@/components/layout/PublicLayout'
import { useToast } from '@/hooks/useToast'

const CONTACT_EMAIL = 'contacto@abudemy.cl'

interface FormState {
  name: string
  organization: string
  email: string
  phone: string
  message: string
}

const initialState: FormState = { name: '', organization: '', email: '', phone: '', message: '' }

export function ContactPage() {
  const [form, setForm] = useState<FormState>(initialState)
  const [errors, setErrors] = useState<Partial<Record<keyof FormState, string>>>({})
  const toast = useToast()

  const update = (key: keyof FormState, value: string) => {
    setForm((prev) => ({ ...prev, [key]: value }))
    setErrors((prev) => ({ ...prev, [key]: undefined }))
  }

  const validate = (): boolean => {
    const next: Partial<Record<keyof FormState, string>> = {}
    if (form.name.trim().length < 2) next.name = 'Ingresa tu nombre.'
    if (form.organization.trim().length < 2) next.organization = 'Indica tu organización.'
    if (!/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(form.email.trim())) next.email = 'Ingresa un correo válido.'
    if (form.message.trim().length < 20) next.message = 'Cuéntanos brevemente qué necesitas (mínimo 20 caracteres).'
    setErrors(next)
    return Object.keys(next).length === 0
  }

  const onSubmit = (event: FormEvent) => {
    event.preventDefault()
    if (!validate()) return

    const subject = `Solicitud de acceso a Abudemy — ${form.organization}`
    const body = [
      `Nombre: ${form.name}`,
      `Organización: ${form.organization}`,
      `Correo: ${form.email}`,
      form.phone ? `Teléfono: ${form.phone}` : '',
      '',
      form.message,
    ]
      .filter(Boolean)
      .join('\n')

    window.location.href = `mailto:${CONTACT_EMAIL}?subject=${encodeURIComponent(subject)}&body=${encodeURIComponent(body)}`
    toast.success(
      'Preparamos tu mensaje',
      'Se abrirá tu cliente de correo con los datos completados para que lo envíes.',
    )
  }

  return (
    <PublicSection>
      <div className="section__head">
        <h2>Contacto</h2>
        <p>
          Escríbenos para solicitar acceso al portal institucional, coordinar una demostración o
          pedir información sobre la integración con tus sistemas.
        </p>
      </div>

      <div className="contact-grid">
        <aside className="contact-aside">
          <div className="contact-aside__item">
            <Icon name="mail" size={20} />
            <div>
              <strong>Correo</strong>
              <p className="muted text-sm">{CONTACT_EMAIL}</p>
            </div>
          </div>
          <div className="contact-aside__item">
            <Icon name="building" size={20} />
            <div>
              <strong>¿Quiénes pueden acceder?</strong>
              <p className="muted text-sm">
                Municipalidades, centros de salud, fundaciones, organizaciones comunitarias y
                equipos que trabajan con personas mayores.
              </p>
            </div>
          </div>
          <div className="contact-aside__item">
            <Icon name="clock" size={20} />
            <div>
              <strong>Tiempo de respuesta</strong>
              <p className="muted text-sm">
                Respondemos en horario hábil. Si ya tienes cuenta, puedes ingresar directamente al
                portal.
              </p>
            </div>
          </div>

        </aside>

        <form className="contact-form" onSubmit={onSubmit} noValidate>
          <div className="form-grid">
            <TextInput
              label="Nombre y apellido"
              required
              value={form.name}
              error={errors.name}
              autoComplete="name"
              onChange={(event) => update('name', event.target.value)}
            />
            <TextInput
              label="Organización"
              required
              value={form.organization}
              error={errors.organization}
              autoComplete="organization"
              onChange={(event) => update('organization', event.target.value)}
            />
            <TextInput
              label="Correo electrónico"
              type="email"
              required
              value={form.email}
              error={errors.email}
              autoComplete="email"
              onChange={(event) => update('email', event.target.value)}
            />
            <TextInput
              label="Teléfono"
              type="tel"
              value={form.phone}
              autoComplete="tel"
              onChange={(event) => update('phone', event.target.value)}
            />
          </div>

          <div style={{ marginTop: 16 }}>
            <Field label="Mensaje" required error={errors.message}>
              <TextAreaInput
                rows={6}
                placeholder="Cuéntanos qué necesitas consultar y qué tipo de organización representas."
                value={form.message}
                onChange={(event) => update('message', event.target.value)}
              />
            </Field>
          </div>

          <div className="row row--between" style={{ marginTop: 20 }}>
            <p className="split-note">
              Al enviar, se abrirá tu cliente de correo. No almacenamos estos datos en la plataforma.
            </p>
            <Button type="submit" icon="mail">
              Redactar mensaje
            </Button>
          </div>
        </form>
      </div>
    </PublicSection>
  )
}
