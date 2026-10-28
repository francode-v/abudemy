import { useState, type FormEvent } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import { AuthLayout } from './AuthLayout'
import { Button, Callout, SelectInput, TextInput } from '@/components/ui/primitives'
import { humanizeError, isApiError, type Role } from '@/api'
import { useAuth } from '@/auth/AuthProvider'
import { useRegister } from '@/hooks/queries'
import { useToast } from '@/hooks/useToast'
import { PASSWORD_REQUIREMENTS, passwordError } from '@/lib/password'

const FIELD_MESSAGES: Record<string, string> = {
  email: 'Ingresa un correo electrónico válido.',
  password: 'Revisa la contraseña: debe cumplir todos los requisitos.',
  full_name: 'Ingresa tu nombre completo.',
  phone: 'El teléfono es demasiado largo.',
  role: 'Selecciona un tipo de cuenta válido.',
}

export function RegisterPage() {
  const register = useRegister()
  const { login } = useAuth()
  const navigate = useNavigate()
  const toast = useToast()

  const [form, setForm] = useState({
    full_name: '',
    email: '',
    phone: '',
    password: '',
    confirm: '',
    role: 'institutional_client' as Role,
  })
  const [errors, setErrors] = useState<Record<string, string>>({})
  const [formError, setFormError] = useState<string | null>(null)

  const update = (key: keyof typeof form, value: string) => {
    setForm((prev) => ({ ...prev, [key]: value }))
    setErrors((prev) => ({ ...prev, [key]: '' }))
  }

  const validate = (): boolean => {
    const next: Record<string, string> = {}
    const fullName = form.full_name.trim()
    const email = form.email.trim()
    const phone = form.phone.trim()

    if (fullName.length < 2) next.full_name = 'Ingresa tu nombre completo.'
    if (fullName.length > 120) next.full_name = 'El nombre no puede superar los 120 caracteres.'
    if (!/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(email)) next.email = 'Ingresa un correo válido.'
    if (email.length > 254) next.email = 'El correo no puede superar los 254 caracteres.'
    if (phone.length > 40) next.phone = 'El teléfono no puede superar los 40 caracteres.'

    const passwordIssue = passwordError(form.password)
    if (passwordIssue) next.password = passwordIssue
    if (!form.confirm) next.confirm = 'Repite la contraseña.'
    else if (form.password !== form.confirm) next.confirm = 'Las contraseñas no coinciden.'

    setErrors(next)
    return Object.keys(next).length === 0
  }

  /** Muestra los errores de campo que devuelve el backend junto al input. */
  const applyServerFieldErrors = (err: unknown): boolean => {
    if (!isApiError(err)) return false
    if (err.isConflict) {
      setFormError('Ya existe una cuenta con ese correo electrónico.')
      return true
    }
    if (err.fields.length === 0) return false
    const next: Record<string, string> = {}
    for (const field of err.fields) {
      next[field.field] = FIELD_MESSAGES[field.field] ?? field.message
    }
    setErrors((prev) => ({ ...prev, ...next }))
    return Object.keys(next).length > 0
  }

  const onSubmit = async (event: FormEvent) => {
    event.preventDefault()
    setFormError(null)
    if (!validate()) return

    try {
      await register.mutateAsync({
        email: form.email.trim(),
        password: form.password,
        full_name: form.full_name.trim(),
        phone: form.phone.trim() || undefined,
        role: form.role,
      })
    } catch (err) {
      if (!applyServerFieldErrors(err)) setFormError(humanizeError(err))
      return
    }

    try {
      await login(form.email.trim(), form.password)
    } catch {
      toast.success('Cuenta creada', 'Inicia sesión con tu nueva contraseña.')
      navigate('/login', { replace: true })
      return
    }

    if (form.role === 'institutional_client') {
      toast.success(
        'Cuenta creada',
        'Tu cuenta institucional queda pendiente de aprobación por un administrador.',
      )
      navigate('/app/cuenta', { replace: true })
    } else {
      toast.success('Cuenta creada', 'Ya puedes usar la aplicación móvil de Abudemy.')
      navigate('/', { replace: true })
    }
  }

  return (
    <AuthLayout
      title="Crear cuenta"
      subtitle="Regístrate para usar la aplicación móvil o solicitar acceso institucional."
      footer={
        <>
          ¿Ya tienes cuenta? <Link to="/login">Inicia sesión</Link>
        </>
      }
    >
      <form className="auth__form" onSubmit={onSubmit} noValidate>
        {formError && (
          <Callout tone="danger" title="No pudimos crear la cuenta">
            {formError}
          </Callout>
        )}

        <TextInput
          label="Nombre completo"
          required
          autoComplete="name"
          leadingIcon="user"
          value={form.full_name}
          error={errors.full_name}
          onChange={(event) => update('full_name', event.target.value)}
        />
        <TextInput
          label="Correo electrónico"
          type="email"
          required
          autoComplete="email"
          leadingIcon="mail"
          value={form.email}
          error={errors.email}
          onChange={(event) => update('email', event.target.value)}
        />
        <TextInput
          label="Teléfono"
          type="tel"
          autoComplete="tel"
          leadingIcon="phone"
          value={form.phone}
          error={errors.phone}
          onChange={(event) => update('phone', event.target.value)}
        />
        <SelectInput
          label="Tipo de cuenta"
          value={form.role}
          onChange={(event) => update('role', event.target.value)}
        >
          <option value="institutional_client">Cliente institucional</option>
          <option value="mobile_user">Persona mayor (aplicación móvil)</option>
        </SelectInput>

        {form.role === 'institutional_client' && (
          <Callout tone="info" title="Aprobación institucional">
            Las cuentas institucionales quedan pendientes hasta que un administrador las active.
            Podrás iniciar sesión y revisar tu cuenta mientras esperas la aprobación.
          </Callout>
        )}

        <TextInput
          label="Contraseña"
          type="password"
          required
          autoComplete="new-password"
          leadingIcon="lock"
          hint={PASSWORD_REQUIREMENTS}
          value={form.password}
          error={errors.password}
          onChange={(event) => update('password', event.target.value)}
        />
        <TextInput
          label="Repetir contraseña"
          type="password"
          required
          autoComplete="new-password"
          leadingIcon="lock"
          hint="Debe coincidir con la contraseña anterior."
          value={form.confirm}
          error={errors.confirm}
          onChange={(event) => update('confirm', event.target.value)}
        />

        <Button type="submit" block size="lg" loading={register.isPending}>
          Crear cuenta
        </Button>
      </form>
    </AuthLayout>
  )
}
