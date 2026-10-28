import { NavLink, Outlet, Link } from 'react-router-dom'
import { AbudemyLogo, Icon } from '@/components/ui/Icon'
import { Button } from '@/components/ui/primitives'
import { Toaster } from '@/components/ui/overlays'
import { useAuth } from '@/auth/AuthProvider'

export function PublicLayout() {
  const { status } = useAuth()
  const authenticated = status === 'authenticated'

  return (
    <div className="public">
      <header className="public-header">
        <Link to="/" className="brand-link" aria-label="Abudemy, inicio">
          <AbudemyLogo />
        </Link>
        <nav className="public-nav" aria-label="Navegación principal">
          <NavLink
            to="/"
            end
            className={({ isActive }) => `public-nav__link ${isActive ? 'is-active' : ''}`}
          >
            Inicio
          </NavLink>
          <NavLink
            to="/funcionalidades"
            className={({ isActive }) => `public-nav__link ${isActive ? 'is-active' : ''}`}
          >
            Funcionalidades
          </NavLink>
          <NavLink
            to="/contacto"
            className={({ isActive }) => `public-nav__link ${isActive ? 'is-active' : ''}`}
          >
            Contacto
          </NavLink>
          <Link to={authenticated ? '/app' : '/login'} className="btn btn--primary btn--sm">
            <Icon name={authenticated ? 'dashboard' : 'lock'} size={15} />
            {authenticated ? 'Ir al panel' : 'Portal institucional'}
          </Link>
        </nav>
      </header>

      <main>
        <Outlet />
      </main>

      <footer className="public-footer">
        <div className="public-footer__grid">
          <div>
            <AbudemyLogo />
            <p style={{ marginTop: 14, maxWidth: 320, fontSize: 14 }}>
              Abudemy conecta a personas mayores con servicios de salud, apoyo, actividades
              comunitarias y oportunidades para compartir conocimiento.
            </p>
          </div>
          <div>
            <h4>Plataforma</h4>
            <Link to="/funcionalidades">Funcionalidades</Link>
            <Link to="/contacto">Solicitar acceso</Link>
            <Link to="/login">Portal institucional</Link>
          </div>
          <div>
            <h4>Recursos</h4>
            <a href="https://www.who.int/es/health-topics/ageing" target="_blank" rel="noreferrer">
              Envejecimiento saludable (OMS)
            </a>
            <a href="/api/v1/analytics/reports" target="_blank" rel="noreferrer">
              API de informes
            </a>
          </div>
        </div>
        <div className="public-footer__legal">
          © {new Date().getFullYear()} Abudemy. Plataforma institucional para organizaciones que
          trabajan con personas mayores.
        </div>
      </footer>

      <Toaster />
    </div>
  )
}

export function PublicSection({
  children,
  tint,
  id,
}: {
  children: React.ReactNode
  tint?: boolean
  id?: string
}) {
  return (
    <section id={id} className={`section ${tint ? 'section--tint' : ''}`}>
      <div className="container">{children}</div>
    </section>
  )
}

export { Button }
