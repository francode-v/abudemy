import { Navigate, Outlet, useLocation } from 'react-router-dom'
import { useAuth } from '@/auth/AuthProvider'
import { FullPageLoader } from '@/components/ui/overlays'
import { Card, CardBody } from '@/components/ui/primitives'
import { Link } from 'react-router-dom'

/** Exige una sesión activa; en caso contrario redirige al inicio de sesión. */
export function RequireAuth() {
  const { status } = useAuth()
  const location = useLocation()

  if (status === 'loading') {
    return <FullPageLoader message="Verificando tu sesión…" />
  }
  if (status !== 'authenticated') {
    return <Navigate to="/login" replace state={{ from: location.pathname + location.search }} />
  }
  return <Outlet />
}

/** Exige un rol con acceso al portal institucional. */
export function RequireInstitutional() {
  const { permissions, status, user } = useAuth()

  if (status === 'loading') {
    return <FullPageLoader message="Cargando tu perfil…" />
  }
  if (!permissions.isInstitutional) {
    const pendingApproval =
      user?.role === 'institutional_client' && !user.institutional_approved
    return <NotAuthorized pendingApproval={pendingApproval} />
  }
  return <Outlet />
}

/** Exige rol de administrador de plataforma. */
export function RequireAdmin() {
  const { permissions } = useAuth()
  if (!permissions.isAdmin) return <NotAuthorized adminOnly />
  return <Outlet />
}

/** Exige permiso para ver el listado de personas registradas. */
export function RequirePeopleViewer() {
  const { permissions, status } = useAuth()
  if (status === 'loading') {
    return <FullPageLoader message="Cargando tu perfil…" />
  }
  if (!permissions.canViewRegisteredPeople) return <NotAuthorized peopleOnly />
  return <Outlet />
}

export function NotAuthorized({
  adminOnly = false,
  peopleOnly = false,
  pendingApproval = false,
}: {
  adminOnly?: boolean
  peopleOnly?: boolean
  pendingApproval?: boolean
}) {
  return (
    <div className="content">
      <Card>
        <CardBody>
          <div className="state">
            <h2 className="state__title">
              {pendingApproval ? 'Cuenta pendiente de aprobación' : 'Acceso restringido'}
            </h2>
            <p className="state__description">
              {adminOnly
                ? 'Esta sección está disponible únicamente para administradores de la plataforma Abudemy.'
                : peopleOnly
                  ? 'Esta sección está disponible para administradores y clientes institucionales aprobados.'
                  : pendingApproval
                    ? 'Tu cuenta institucional fue creada correctamente, pero un administrador todavía debe activarla. Mientras esperas, puedes revisar los datos de tu cuenta.'
                    : 'Tu cuenta no tiene un rol institucional habilitado. El portal está destinado a clientes institucionales, municipalidades, centros de salud, fundaciones y organizaciones.'}
            </p>
            {!pendingApproval && (
              <p className="state__description">
                Si necesitas acceso, solicítalo a través del formulario de contacto indicando tu
                organización.
              </p>
            )}
            <div className="state__action row">
              {!pendingApproval && (
                <Link className="btn btn--secondary" to="/contacto">
                  Solicitar acceso
                </Link>
              )}
              {!adminOnly && (
                <Link className="btn btn--ghost" to="/app/cuenta">
                  Ir a mi cuenta
                </Link>
              )}
            </div>
          </div>
        </CardBody>
      </Card>
    </div>
  )
}
