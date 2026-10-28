import { useEffect, useState, type FormEvent } from 'react'
import { Link } from 'react-router-dom'
import { humanizeError } from '@/api'
import { useChangePassword, useUpdateProfile } from '@/hooks/queries'
import { useAuth } from '@/auth/AuthProvider'
import { useToast } from '@/hooks/useToast'
import {
  Badge,
  Button,
  Callout,
  Card,
  CardBody,
  CardHeader,
  PageHeader,
  TextInput,
} from '@/components/ui/primitives'
import { KeyValue, StatCard } from '@/components/ui/data'
import { Icon } from '@/components/ui/Icon'
import { orgMemberRoleLabels, roleLabels, userStatusLabels } from '@/lib/dictionaries'
import { userStatusTone } from '@/lib/status'
import { formatDateTime, relativeTime, toDateInputValue } from '@/lib/format'
import { PASSWORD_REQUIREMENTS, passwordError } from '@/lib/password'

export function ProfilePage() {
  const { user, memberships, reload } = useAuth()
  const toast = useToast()
  const updateProfile = useUpdateProfile()
  const changePassword = useChangePassword()

  const [form, setForm] = useState({ full_name: '', phone: '', birth_date: '' })
  const [passwordForm, setPasswordForm] = useState({ current: '', next: '', confirm: '' })

  useEffect(() => {
    if (!user) return
    setForm({
      full_name: user.full_name ?? '',
      phone: user.phone ?? '',
      birth_date: toDateInputValue(user.birth_date),
    })
  }, [user])

  if (!user) {
    return (
      <div className="content">
        <Callout tone="warning" title="Sesión no disponible">
          Vuelve a iniciar sesión para ver tu cuenta.
        </Callout>
      </div>
    )
  }

  const submitProfile = (event: FormEvent) => {
    event.preventDefault()
    if (form.full_name.trim().length < 2) {
      toast.warning('Nombre requerido', 'Ingresa tu nombre completo.')
      return
    }
    updateProfile.mutate(
      {
        full_name: form.full_name.trim(),
        phone: form.phone.trim() || undefined,
        birth_date: form.birth_date || undefined,
      },
      {
        onSuccess: async () => {
          toast.success('Perfil actualizado', 'Tus datos se guardaron correctamente.')
          await reload()
        },
        onError: (error) => toast.error('No se pudo actualizar', humanizeError(error)),
      },
    )
  }

  const submitPassword = (event: FormEvent) => {
    event.preventDefault()
    const passwordIssue = passwordError(passwordForm.next)
    if (passwordIssue) {
      toast.warning('Contraseña débil', passwordIssue)
      return
    }
    if (passwordForm.next !== passwordForm.confirm) {
      toast.warning('Las contraseñas no coinciden')
      return
    }
    changePassword.mutate(
      { currentPassword: passwordForm.current, newPassword: passwordForm.next },
      {
        onSuccess: () => {
          toast.success('Contraseña actualizada', 'Usa tu nueva contraseña en el próximo inicio de sesión.')
          setPasswordForm({ current: '', next: '', confirm: '' })
        },
        onError: (error) => toast.error('No se pudo cambiar', humanizeError(error)),
      },
    )
  }

  return (
    <>
      <PageHeader
        title="Mi cuenta"
        description="Administra tus datos personales, tu contraseña y revisa las organizaciones a las que perteneces."
        actions={
          <Link className="btn btn--secondary" to="/app/cuenta/sesiones">
            <Icon name="shield" size={15} /> Sesiones y seguridad
          </Link>
        }
      />

      {user.role === 'institutional_client' && !user.institutional_approved && (
        <Callout tone="warning" title="Cuenta institucional pendiente">
          Un administrador debe aprobar tu cuenta antes de que puedas usar servicios,
          actividades, informes y organizaciones del portal institucional.
        </Callout>
      )}

      <section className="stat-grid">
        <StatCard
          label="Rol en la plataforma"
          value={roleLabels[user.role]}
          icon="user"
          description="El rol determina qué secciones del portal institucional puedes usar."
        />
        <StatCard
          label="Estado de la cuenta"
          value={<Badge tone={userStatusTone(user.status)}>{userStatusLabels[user.status]}</Badge>}
          icon="shield"
          description="Una cuenta suspendida o inactiva pierde el acceso a la plataforma."
        />
        {user.role === 'institutional_client' && (
          <StatCard
            label="Aprobación institucional"
            value={
              <Badge tone={user.institutional_approved ? 'success' : 'warning'}>
                {user.institutional_approved ? 'Aprobada' : 'Pendiente'}
              </Badge>
            }
            icon="check"
            description="Solo las cuentas aprobadas por un administrador acceden al portal institucional."
          />
        )}
        <StatCard
          label="Organizaciones"
          value={String(memberships.length)}
          icon="organizations"
          description="Organizaciones a las que perteneces con algún rol de integrante."
        />
        <StatCard
          label="Cuenta creada"
          value={formatDateTime(user.created_at)}
          hint={`Último acceso: ${user.last_login_at ? relativeTime(user.last_login_at) : 'sin registro'}`}
          icon="clock"
          description="Fecha de creación de la cuenta y último inicio de sesión registrado."
        />
      </section>

      <div className="grid-2">
        <Card>
          <CardHeader title="Datos personales" icon="user" />
          <CardBody>
            <form onSubmit={submitProfile} noValidate>
              <div className="form-grid">
                <div className="form-grid--full">
                  <TextInput
                    label="Nombre completo"
                    required
                    value={form.full_name}
                    onChange={(event) => setForm((prev) => ({ ...prev, full_name: event.target.value }))}
                  />
                </div>
                <TextInput
                  label="Correo electrónico"
                  value={user.email}
                  readOnly
                  disabled
                  hint="El correo electrónico no se puede modificar desde el portal."
                />
                <TextInput
                  label="Teléfono"
                  type="tel"
                  value={form.phone}
                  onChange={(event) => setForm((prev) => ({ ...prev, phone: event.target.value }))}
                />
                <TextInput
                  label="Fecha de nacimiento"
                  type="date"
                  value={form.birth_date}
                  onChange={(event) => setForm((prev) => ({ ...prev, birth_date: event.target.value }))}
                />
              </div>
              <div className="detail-actions" style={{ marginTop: 16 }}>
                <Button type="submit" loading={updateProfile.isPending}>
                  Guardar cambios
                </Button>
              </div>
            </form>
          </CardBody>
        </Card>

        <Card>
          <CardHeader
            title="Cambiar contraseña"
            description="Al cambiarla, se cierran tus otras sesiones por seguridad."
            icon="lock"
          />
          <CardBody>
            <form onSubmit={submitPassword} noValidate>
              <div className="form-grid">
                <div className="form-grid--full">
                  <TextInput
                    label="Contraseña actual"
                    type="password"
                    autoComplete="current-password"
                    leadingIcon="lock"
                    value={passwordForm.current}
                    onChange={(event) =>
                      setPasswordForm((prev) => ({ ...prev, current: event.target.value }))
                    }
                  />
                </div>
                <div className="form-grid--full">
                  <TextInput
                    label="Nueva contraseña"
                    type="password"
                    autoComplete="new-password"
                    leadingIcon="lock"
                    hint={PASSWORD_REQUIREMENTS}
                    value={passwordForm.next}
                    onChange={(event) => setPasswordForm((prev) => ({ ...prev, next: event.target.value }))}
                  />
                </div>
                <div className="form-grid--full">
                  <TextInput
                    label="Repetir nueva contraseña"
                    type="password"
                    autoComplete="new-password"
                    leadingIcon="lock"
                    value={passwordForm.confirm}
                    onChange={(event) =>
                      setPasswordForm((prev) => ({ ...prev, confirm: event.target.value }))
                    }
                  />
                </div>
              </div>
              <div className="detail-actions" style={{ marginTop: 16 }}>
                <Button type="submit" variant="secondary" loading={changePassword.isPending}>
                  Cambiar contraseña
                </Button>
              </div>
            </form>
          </CardBody>
        </Card>
      </div>

      <Card>
        <CardHeader
          title="Mis organizaciones"
          description="Roles que tienes en cada organización. Los administradores de la organización gestionan estos roles."
          icon="organizations"
        />
        <CardBody>
          {memberships.length === 0 ? (
            <Callout tone="info" title="Sin organizaciones">
              Todavía no perteneces a ninguna organización. Si representas a una institución, solicita
              acceso desde el formulario de contacto o pide a un administrador que te agregue.
            </Callout>
          ) : (
            <KeyValue
              items={memberships.map((membership) => ({
                label: membership.organization.name,
                value: (
                  <Badge tone={membership.role === 'viewer' ? 'neutral' : 'info'}>
                    {orgMemberRoleLabels[membership.role]}
                  </Badge>
                ),
              }))}
            />
          )}
        </CardBody>
      </Card>
    </>
  )
}
