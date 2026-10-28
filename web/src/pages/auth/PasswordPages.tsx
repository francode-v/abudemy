import { useState, type FormEvent } from 'react'
import { Link, useNavigate, useSearchParams } from 'react-router-dom'
import { AuthLayout } from './AuthLayout'
import { Button, Callout, TextInput } from '@/components/ui/primitives'
import { api, humanizeError } from '@/api'
import { useToast } from '@/hooks/useToast'
import { PASSWORD_REQUIREMENTS, passwordError } from '@/lib/password'

export function ForgotPasswordPage() {
  const [email, setEmail] = useState('')
  const [sent, setSent] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [loading, setLoading] = useState(false)

  const onSubmit = async (event: FormEvent) => {
    event.preventDefault()
    setError(null)
    if (!/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(email.trim())) {
      setError('Ingresa un correo electrónico válido.')
      return
    }
    setLoading(true)
    try {
      await api.auth.forgotPassword(email.trim())
      setSent(true)
    } catch (err) {
      setError(humanizeError(err))
    } finally {
      setLoading(false)
    }
  }

  return (
    <AuthLayout
      title="Recuperar contraseña"
      subtitle="Te enviaremos instrucciones si el correo está registrado."
      footer={
        <>
          <Link to="/login">Volver a iniciar sesión</Link>
        </>
      }
    >
      {sent ? (
        <Callout tone="success" title="Solicitud recibida">
          Si el correo está registrado, enviaremos las instrucciones de restablecimiento. Por
          seguridad, la respuesta es la misma exista o no la cuenta.
          <p style={{ marginTop: 10 }}>
            <Link to="/restablecer">Ya tengo un token de restablecimiento</Link>
          </p>
        </Callout>
      ) : (
        <form className="auth__form" onSubmit={onSubmit} noValidate>
          {error && <Callout tone="danger">{error}</Callout>}
          <TextInput
            label="Correo electrónico"
            type="email"
            required
            leadingIcon="mail"
            value={email}
            onChange={(event) => setEmail(event.target.value)}
          />
          <Button type="submit" block size="lg" loading={loading}>
            Enviar instrucciones
          </Button>
        </form>
      )}
    </AuthLayout>
  )
}

export function ResetPasswordPage() {
  const [params] = useSearchParams()
  const navigate = useNavigate()
  const toast = useToast()
  const [token, setToken] = useState(params.get('token') ?? '')
  const [password, setPassword] = useState('')
  const [confirm, setConfirm] = useState('')
  const [error, setError] = useState<string | null>(null)
  const [loading, setLoading] = useState(false)

  const onSubmit = async (event: FormEvent) => {
    event.preventDefault()
    setError(null)
    if (token.trim().length < 8) {
      setError('Ingresa el token que recibiste.')
      return
    }
    const passwordIssue = passwordError(password)
    if (passwordIssue) {
      setError(passwordIssue)
      return
    }
    if (password !== confirm) {
      setError('Las contraseñas no coinciden.')
      return
    }
    setLoading(true)
    try {
      await api.auth.resetPassword(token.trim(), password)
      toast.success('Contraseña actualizada', 'Ya puedes iniciar sesión con tu nueva contraseña.')
      navigate('/login', { replace: true })
    } catch (err) {
      setError(humanizeError(err))
    } finally {
      setLoading(false)
    }
  }

  return (
    <AuthLayout
      title="Restablecer contraseña"
      subtitle="Ingresa el token recibido y define una nueva contraseña."
      footer={<Link to="/login">Volver a iniciar sesión</Link>}
    >
      <form className="auth__form" onSubmit={onSubmit} noValidate>
        {error && <Callout tone="danger">{error}</Callout>}
        <TextInput
          label="Token de restablecimiento"
          required
          value={token}
          onChange={(event) => setToken(event.target.value)}
          hint="Lo encuentras en el correo enviado por Abudemy."
        />
        <TextInput
          label="Nueva contraseña"
          type="password"
          required
          leadingIcon="lock"
          hint={PASSWORD_REQUIREMENTS}
          value={password}
          onChange={(event) => setPassword(event.target.value)}
        />
        <TextInput
          label="Repetir contraseña"
          type="password"
          required
          leadingIcon="lock"
          value={confirm}
          onChange={(event) => setConfirm(event.target.value)}
        />
        <Button type="submit" block size="lg" loading={loading}>
          Restablecer contraseña
        </Button>
      </form>
    </AuthLayout>
  )
}
