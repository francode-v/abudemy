import { useEffect, useState, Suspense } from 'react'
import { NavLink, Outlet, Link, useLocation, useNavigate } from 'react-router-dom'
import { AbudemyLogo, Icon, type IconName } from '@/components/ui/Icon'
import { Button } from '@/components/ui/primitives'
import { useAuth } from '@/auth/AuthProvider'
import { useOrgScope } from '@/app/orgScope'
import { initials } from '@/lib/format'
import { DEMO_MODE } from '@/api'
import { roleLabels } from '@/lib/dictionaries'
import { FullPageLoader, Toaster } from '@/components/ui/overlays'

interface NavItem {
  to: string
  label: string
  icon: IconName
  end?: boolean
  adminOnly?: boolean
}

const mainNav: NavItem[] = [
  { to: '/app', label: 'Panel', icon: 'dashboard', end: true },
  { to: '/app/mapa', label: 'Mapa territorial', icon: 'map' },
  { to: '/app/servicios', label: 'Servicios', icon: 'services' },
  { to: '/app/actividades', label: 'Actividades', icon: 'activities' },
  { to: '/app/asistencia', label: 'Solicitudes de asistencia', icon: 'heart' },
  { to: '/app/oportunidades', label: 'Oportunidades', icon: 'trend' },
  { to: '/app/informes', label: 'Informes', icon: 'reports' },
  { to: '/app/organizaciones', label: 'Organizaciones', icon: 'organizations' },
]

const accountNav: NavItem[] = [
  { to: '/app/cuenta', label: 'Mi cuenta', icon: 'user' },
  { to: '/app/cuenta/sesiones', label: 'Sesiones y seguridad', icon: 'shield' },
]

const adminNav: NavItem[] = [
  { to: '/app/admin/usuarios', label: 'Usuarios', icon: 'users', adminOnly: true },
  {
    to: '/app/admin/institucionales',
    label: 'Cuentas institucionales',
    icon: 'shield',
    adminOnly: true,
  },
]

const communityNav: NavItem[] = [
  { to: '/app/personas', label: 'Personas registradas', icon: 'user' },
]

export function AppLayout() {
  const { user, logout, permissions } = useAuth()
  const { organizationId, setOrganizationId, options, activeName } = useOrgScope()
  const [sidebarOpen, setSidebarOpen] = useState(false)
  const location = useLocation()
  const navigate = useNavigate()

  useEffect(() => {
    setSidebarOpen(false)
  }, [location.pathname])

  if (!user) return null

  return (
    <div className="app-shell">
      <aside className={`sidebar ${sidebarOpen ? 'is-open' : ''}`} aria-label="Navegación institucional">
        <div className="sidebar__brand">
          <Link to="/" aria-label="Abudemy">
            <AbudemyLogo />
          </Link>
        </div>

        {permissions.isInstitutional && (
          <nav className="sidebar__section" aria-label="Secciones principales">
            {mainNav.map((item) => (
              <SidebarLink key={item.to} item={item} />
            ))}
          </nav>
        )}

        {permissions.canViewRegisteredPeople && (
          <div className="sidebar__section">
            <p className="sidebar__section-label">Comunidad</p>
            {communityNav.map((item) => (
              <SidebarLink key={item.to} item={item} />
            ))}
          </div>
        )}

        <div className="sidebar__section">
          <p className="sidebar__section-label">Mi cuenta</p>
          {accountNav.map((item) => (
            <SidebarLink key={item.to} item={item} />
          ))}
        </div>

        {permissions.isAdmin && (
          <div className="sidebar__section">
            <p className="sidebar__section-label">Administración</p>
            {adminNav.map((item) => (
              <SidebarLink key={item.to} item={item} />
            ))}
          </div>
        )}

        <div className="sidebar__footer">
          <div className="sidebar__user">
            <span className="avatar">{initials(user.full_name)}</span>
            <div style={{ minWidth: 0 }}>
              <p className="sidebar__user-name" title={user.full_name}>
                {user.full_name}
              </p>
              <p className="sidebar__user-role">{roleLabels[user.role]}</p>
            </div>
          </div>
          <Button
            variant="ghost"
            size="sm"
            icon="logout"
            block
            onClick={() => void logout().then(() => navigate('/'))}
          >
            Cerrar sesión
          </Button>
        </div>
      </aside>

      <div
        className={`sidebar-backdrop ${sidebarOpen ? 'is-open' : ''}`}
        onClick={() => setSidebarOpen(false)}
        role="presentation"
      />

      <div className="main">
        <header className="topbar">
          <button
            className="btn btn--ghost btn--icon topbar__toggle"
            onClick={() => setSidebarOpen((prev) => !prev)}
            aria-label="Abrir menú de navegación"
            aria-expanded={sidebarOpen}
          >
            <Icon name="menu" size={18} />
          </button>

          {options.length > 0 && (
            <div className="topbar__org">
              <Icon name="organizations" size={16} />
              <label htmlFor="org-scope" className="sr-only">
                Organización activa
              </label>
              <select
                id="org-scope"
                className="input input--select"
                style={{ minHeight: 36, paddingTop: 4, paddingBottom: 4, maxWidth: 280 }}
                value={organizationId ?? ''}
                onChange={(event) => setOrganizationId(event.target.value || null)}
              >
                <option value="">Todas mis organizaciones</option>
                {options.map((option) => (
                  <option key={option.id} value={option.id}>
                    {option.name}
                  </option>
                ))}
              </select>
              <span className="badge badge--primary">{activeName}</span>
            </div>
          )}

          <div className="topbar__spacer" />

          {DEMO_MODE && (
            <span className="topbar__badge" title="Los datos mostrados son ficticios">
              <Icon name="alert" size={13} /> Datos de demostración
            </span>
          )}
        </header>

        {DEMO_MODE && (
          <div className="demo-banner" role="status">
            <Icon name="info" size={15} />
            <span>
              <strong>Modo demostración activo.</strong> Todos los registros son ficticios y están
              marcados con «[DEMO]». Desactiva <code>VITE_DEMO_MODE</code> para usar el backend real.
            </span>
          </div>
        )}

        <main className="content">
          <Suspense fallback={<FullPageLoader message="Cargando sección…" />}>
            <Outlet />
          </Suspense>
        </main>
      </div>

      <Toaster />
    </div>
  )
}

function SidebarLink({ item }: { item: NavItem }) {
  return (
    <NavLink
      to={item.to}
      end={item.end}
      className={({ isActive }) => `sidebar__link ${isActive ? 'is-active' : ''}`}
    >
      <Icon name={item.icon} size={17} />
      <span>{item.label}</span>
    </NavLink>
  )
}
