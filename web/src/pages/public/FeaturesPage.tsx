import { Link } from 'react-router-dom'
import { Icon, type IconName } from '@/components/ui/Icon'
import { PublicSection } from '@/components/layout/PublicLayout'

const groups: { title: string; icon: IconName; description: string; items: string[] }[] = [
  {
    title: 'Panel institucional',
    icon: 'dashboard',
    description: 'Todos los indicadores disponibles en la API de Abudemy, con su significado y frescura.',
    items: [
      'Tarjetas de indicadores: servicios, actividades, inscripciones, usuarios y asistencia.',
      'Gráficos de servicios por categoría y actividades por tipo.',
      'Distribución territorial de servicios por comuna y región.',
      'Filtro por organización cuando tu cuenta administra más de una.',
      'Explicación de cada indicador y la fecha en que se consultaron los datos.',
    ],
  },
  {
    title: 'Mapas e información territorial',
    icon: 'map',
    description: 'Explora dónde están los servicios y qué zonas requieren atención.',
    items: [
      'Mapa interactivo con establecimientos y actividades georreferenciadas.',
      'Filtros por comuna, categoría, tipo y período.',
      'Capa de calor cuando existen suficientes datos geográficos.',
      'Resumen de indicadores de la zona seleccionada.',
      'Solo se publican coordenadas de establecimientos, nunca domicilios particulares.',
    ],
  },
  {
    title: 'Servicios y actividades',
    icon: 'services',
    description: 'Mantén la oferta de tu organización actualizada para la comunidad.',
    items: [
      'Crear, consultar y editar servicios y establecimientos.',
      'Actualizar direcciones, contactos, horarios y categorías.',
      'Crear talleres, cursos y eventos con control de cupos.',
      'Publicar, dejar en borrador o desactivar publicaciones según tus permisos.',
    ],
  },
  {
    title: 'Informes',
    icon: 'reports',
    description: 'Prepara la información que necesitas para tus reportes de gestión.',
    items: [
      'Consulta de indicadores por período y territorio.',
      'Resúmenes con gráficos y tablas listos para presentar.',
      'Exportación de datos tabulares en CSV.',
      'Generación de documentos PDF listos para presentar.',
    ],
  },
  {
    title: 'Administración',
    icon: 'shield',
    description: 'Gestiona usuarios, organizaciones y permisos cuando tu rol lo permite.',
    items: [
      'Administración de usuarios: estado de la cuenta y rol.',
      'Gestión de organizaciones y de sus integrantes.',
      'Control de acceso por rol y por organización.',
      'Categorías de servicios disponibles para clasificar la oferta.',
    ],
  },
]

export function FeaturesPage() {
  return (
    <>
      <PublicSection>
        <div className="section__head">
          <h2>Funcionalidades de la plataforma</h2>
          <p>
            Estas son las herramientas que encontrarás en el portal institucional para gestionar tu
            organización y su oferta de servicios.
          </p>
        </div>
        <div className="stack" style={{ gap: 22 }}>
          {groups.map((group) => (
            <article className="card" key={group.title}>
              <div className="card__header">
                <span className="card__icon" aria-hidden="true">
                  <Icon name={group.icon} size={18} />
                </span>
                <div className="card__heading">
                  <h3 className="card__title">{group.title}</h3>
                  <p className="card__description">{group.description}</p>
                </div>
              </div>
              <div className="card__body">
                <div className="benefit-list">
                  {group.items.map((item) => (
                    <div className="benefit" key={item}>
                      <Icon name="check" size={17} className="benefit__icon" />
                      <p>{item}</p>
                    </div>
                  ))}
                </div>
              </div>
            </article>
          ))}
        </div>
      </PublicSection>

      <PublicSection tint>
        <div className="section__head">
          <h2>Roles y permisos</h2>
          <p>
            Cada acción respeta los permisos de tu rol y de la organización en la que trabajas.
          </p>
        </div>
        <div className="feature-grid">
          <article className="feature">
            <h3>Administrador</h3>
            <p>Acceso completo a indicadores, organizaciones, usuarios y publicaciones de toda la plataforma.</p>
          </article>
          <article className="feature">
            <h3>Cliente institucional</h3>
            <p>Una vez aprobada su cuenta, crea su organización, administra sus integrantes y gestiona los servicios y actividades de la institución.</p>
          </article>
          <article className="feature">
            <h3>Colaborador de organización</h3>
            <p>Trabaja sobre las organizaciones a las que fue invitado, con permisos de lectura, edición o administración.</p>
          </article>
        </div>
        <div className="hero__actions">
          <Link to="/login" className="btn btn--primary btn--lg">
            Acceder al portal
          </Link>
          <Link to="/contacto" className="btn btn--secondary btn--lg">
            Solicitar acceso
          </Link>
        </div>
      </PublicSection>
    </>
  )
}
