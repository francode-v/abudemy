import { Link } from 'react-router-dom'
import { Button, Card, CardBody } from '@/components/ui/primitives'
import { Icon } from '@/components/ui/Icon'
import { useAuth } from '@/auth/AuthProvider'

export function NotFoundPage() {
  const { status } = useAuth()
  const authenticated = status === 'authenticated'

  return (
    <div style={{ maxWidth: 640, margin: '56px auto', padding: '0 16px' }}>
      <Card>
        <CardBody>
          <div className="state">
            <span className="state__icon" aria-hidden="true">
              <Icon name="search" size={22} />
            </span>
            <h1 className="state__title">No encontramos esta página</h1>
            <p className="state__description">
              La dirección que abriste no existe o cambió de ubicación. Revisa el enlace o vuelve a una
              sección conocida.
            </p>
            <div className="state__action row">
              <Link className="btn btn--secondary" to="/">
                Ir al inicio
              </Link>
              {authenticated ? (
                <Link className="btn btn--primary" to="/app">
                  Ir al panel
                </Link>
              ) : (
                <Link className="btn btn--primary" to="/login">
                  Iniciar sesión
                </Link>
              )}
            </div>
          </div>
        </CardBody>
      </Card>
      <div className="row" style={{ justifyContent: 'center', marginTop: 16 }}>
        <Button variant="ghost" size="sm" icon="list" onClick={() => window.history.back()}>
          Volver atrás
        </Button>
      </div>
    </div>
  )
}
