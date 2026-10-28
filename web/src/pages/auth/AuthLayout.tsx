import { Link } from 'react-router-dom'
import type { ReactNode } from 'react'
import { AbudemyLogo, Icon } from '@/components/ui/Icon'

export function AuthLayout({
  title,
  subtitle,
  children,
  footer,
}: {
  title: string
  subtitle: string
  children: ReactNode
  footer?: ReactNode
}) {
  return (
    <div className="auth">
      <aside className="auth__aside">
        <Link to="/" aria-label="Abudemy">
          <AbudemyLogo />
        </Link>
        <div>
          <h2>Portal institucional Abudemy</h2>
          <p>
            Consulta indicadores territoriales, administra servicios y actividades, y descarga
            informes para tu organización.
          </p>
          <ul className="auth__list">
            <li>
              <Icon name="check" size={17} /> Datos compartidos con la aplicación móvil de personas mayores.
            </li>
            <li>
              <Icon name="shield" size={17} /> Acceso protegido por roles y organizaciones.
            </li>
            <li>
              <Icon name="map" size={17} /> Mapas y análisis por comuna.
            </li>
          </ul>
        </div>
        <p style={{ fontSize: 12.5 }}>
          © {new Date().getFullYear()} Abudemy. Plataforma para organizaciones que trabajan con
          personas mayores.
        </p>
      </aside>

      <section className="auth__main">
        <div className="auth__card">
          <h1>{title}</h1>
          <p>{subtitle}</p>
          {children}
          {footer && <div className="auth__footer">{footer}</div>}
        </div>
      </section>
    </div>
  )
}
