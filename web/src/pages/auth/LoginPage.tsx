import { useState, type FormEvent } from 'react'
import { Link, Navigate, useLocation, useNavigate } from 'react-router-dom'
import { AuthLayout } from './AuthLayout'
import { Button, Callout, TextInput } from '@/components/ui/primitives'
import { useAuth } from '@/auth/AuthProvider'
import { humanizeError } from '@/api'
import { useToast } from '@/hooks/useToast'

export function LoginPage() {
  const { login, status, user } = useAuth()
  const navigate = useNavigate()
  const location = useLocation()
  const toast = useToast()

  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [showPassword, setShowPassword] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [submitting, setSubmitting] = useState(false)

  const from = (location.state as { from?: string } | null)?.from ?? '/app'

  // Un cliente institucional pendiente de aprobación no puede entrar al
  // portal todavía: lo llevamos a su cuenta, donde ve el estado del trámite.
  const landingFor = (candidate?: typeof user | null) =>
    candidate?.role === 'institutional_client' && !candidate.institutional_approved
      ? '/app/cuenta'
      : from

  if (status === 'authenticated') {
    return <Navigate to={landingFor(user)} replace />
  }

  const onSubmit = async (event: FormEvent) => {
    event.preventDefault()
    setError(null)

    if (!/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(email.trim())) {
      setError('Ingresa un correo electrónico válido.')
      return
    }
    if (password.length < 1) {
      setError('Ingresa tu contraseña.')
      return
    }

    setSubmitting(true)
    try {
      const authenticated = await login(email.trim(), password)
      toast.success('Sesión iniciada', `Bienvenido/a, ${authenticated.full_name}.`)
      navigate(landingFor(authenticated), { replace: true })
    } catch (err) {
      setError(humanizeError(err))
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <AuthLayout
      title="Inicia sesión"
      subtitle="Accede con las credenciales que te entregó tu organización."
      footer={
        <>
          ¿No tienes cuenta? <Link to="/registro">Crea tu cuenta institucional</Link>
        </>
      }
    >
      <form className="auth__form" onSubmit={onSubmit} noValidate>
        {error && (
          <Callout tone="danger" title="No pudimos iniciar tu sesión">
            {error}
          </Callout>
        )}

        <TextInput
          label="Correo electrónico"
          type="email"
          autoComplete="email"
          required
          leadingIcon="mail"
          value={email}
          onChange={(event) => setEmail(event.target.value)}
        />

        <div>
          <TextInput
            label="Contraseña"
            type={showPassword ? 'text' : 'password'}
            autoComplete="current-password"
            required
            leadingIcon="lock"
            value={password}
            onChange={(event) => setPassword(event.target.value)}
          />
          <div className="auth__meta">
            <label className="checkbox">
              <input
                type="checkbox"
                checked={showPassword}
                onChange={(event) => setShowPassword(event.target.checked)}
              />
              <span>Mostrar contraseña</span>
            </label>
            <Link to="/recuperar">¿Olvidaste tu contraseña?</Link>
          </div>
        </div>

        <Button type="submit" block size="lg" loading={submitting}>
          Ingresar
        </Button>
      </form>
    </AuthLayout>
  )
}
