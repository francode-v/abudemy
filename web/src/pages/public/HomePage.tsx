import { Link } from 'react-router-dom'
import { Icon, type IconName } from '@/components/ui/Icon'
import { PublicSection } from '@/components/layout/PublicLayout'

const features: { icon: IconName; title: string; text: string }[] = [
  {
    icon: 'dashboard',
    title: 'Panel de indicadores',
    text: 'Tarjetas, gráficos y tablas con la situación de servicios, actividades, inscripciones y solicitudes de asistencia.',
  },
  {
    icon: 'map',
    title: 'Mapa territorial',
    text: 'Ubica establecimientos y servicios, agrega datos por comuna y detecta zonas con menor cobertura.',
  },
  {
    icon: 'services',
    title: 'Gestión de servicios',
    text: 'Mantén al día direcciones, contactos, horarios y categorías de cada establecimiento o recurso de apoyo.',
  },
  {
    icon: 'activities',
    title: 'Actividades e inscripciones',
    text: 'Publica talleres, cursos y eventos, controla cupos y revisa quién está inscrito en cada instancia.',
  },
  {
    icon: 'reports',
    title: 'Informes y exportaciones',
    text: 'Resume indicadores por período y territorio, y exporta los datos para tus informes de gestión.',
  },
  {
    icon: 'shield',
    title: 'Seguridad por roles',
    text: 'Cada persona ve y edita únicamente lo que le corresponde según su rol y su organización.',
  },
]

const benefits = [
  { title: 'Información validada', text: 'Los datos provienen de la misma base que alimenta la aplicación móvil para personas mayores.' },
  { title: 'Decisiones territoriales', text: 'Compara comunas y prioriza dónde faltan servicios o actividades.' },
  { title: 'Menos trabajo manual', text: 'Actualiza una vez y la información queda disponible para toda la comunidad.' },
  { title: 'Trazabilidad', text: 'Cada servicio registra quién lo creó y cuándo se actualizó por última vez.' },
]

export function HomePage() {
  return (
    <>
      <section className="hero">
        <div className="container hero__inner">
          <div>
            <span className="eyebrow">
              <Icon name="sparkles" size={14} /> Plataforma institucional
            </span>
            <h1>
              Información clara sobre <span>servicios y actividades</span> para personas mayores
            </h1>
            <p className="hero__lead">
              Abudemy reúne en un solo lugar los establecimientos, programas y talleres disponibles
              en cada territorio. Municipalidades, centros de salud, fundaciones y organizaciones
              pueden consultar indicadores, analizar necesidades y administrar su oferta.
            </p>
            <div className="hero__actions">
              <Link to="/login" className="btn btn--primary btn--lg">
                <Icon name="lock" size={16} /> Acceder al portal
              </Link>
              <Link to="/funcionalidades" className="btn btn--secondary btn--lg">
                Ver funcionalidades
              </Link>
            </div>
          </div>

          <div className="hero__panel" aria-label="Resumen de la plataforma">
            <p className="hero__panel-title">Lo que encontrarás</p>
            <div className="hero__metric">
              <span>Servicios y establecimientos</span>
              <strong>Catálogo</strong>
            </div>
            <div className="hero__metric">
              <span>Agregación territorial</span>
              <strong>Por comuna</strong>
            </div>
            <div className="hero__metric">
              <span>Actividades e inscripciones</span>
              <strong>Cupos</strong>
            </div>
            <div className="hero__metric">
              <span>Exportaciones</span>
              <strong>CSV y PDF</strong>
            </div>
          </div>
        </div>
      </section>

      <PublicSection>
        <div className="section__head">
          <h2>¿Qué es Abudemy?</h2>
          <p>
            Abudemy conecta a las personas mayores con servicios de salud, apoyo y actividades
            comunitarias, y les abre oportunidades para compartir sus conocimientos. Esta
            plataforma es el espacio de trabajo de las organizaciones que hacen posible esa red.
          </p>
        </div>
        <div className="feature-grid">
          {features.map((feature) => (
            <article className="feature" key={feature.title}>
              <span className="feature__icon" aria-hidden="true">
                <Icon name={feature.icon} size={20} />
              </span>
              <h3>{feature.title}</h3>
              <p>{feature.text}</p>
            </article>
          ))}
        </div>
      </PublicSection>

      <PublicSection tint>
        <div className="section__head">
          <h2>Beneficios para tu organización</h2>
          <p>
            Una herramienta pensada para equipos municipales, de salud y de la sociedad civil que
            necesitan evidencia para planificar y gestionar.
          </p>
        </div>
        <div className="benefit-list">
          {benefits.map((benefit) => (
            <article className="benefit" key={benefit.title}>
              <Icon name="check" size={18} className="benefit__icon" />
              <div>
                <strong>{benefit.title}</strong>
                <p>{benefit.text}</p>
              </div>
            </article>
          ))}
        </div>
      </PublicSection>

      <PublicSection>
        <div className="section__head">
          <h2>Cómo empezar</h2>
          <p>Tres pasos para incorporar a tu organización a la plataforma.</p>
        </div>
        <div className="steps">
          <article className="step">
            <h3>Solicita tu acceso</h3>
            <p>Cuéntanos qué organización representas y qué necesitas consultar.</p>
          </article>
          <article className="step">
            <h3>Configura tu organización</h3>
            <p>Registra tus servicios, categorías y datos de contacto.</p>
          </article>
          <article className="step">
            <h3>Analiza y publica</h3>
            <p>Consulta indicadores territoriales, publica actividades y descarga informes.</p>
          </article>
        </div>
      </PublicSection>

      <PublicSection tint>
        <div className="section__head">
          <h2>Compromiso con la privacidad</h2>
          <p>
            La plataforma muestra información de establecimientos y servicios de acceso público.
            Nunca expone domicilios particulares, ubicaciones personales ni datos identificables de
            las personas mayores.
          </p>
        </div>
        <div className="hero__actions">
          <Link to="/contacto" className="btn btn--primary">
            Solicitar una demostración
          </Link>
          <Link to="/login" className="btn btn--secondary">
            Entrar al portal institucional
          </Link>
        </div>
      </PublicSection>
    </>
  )
}
