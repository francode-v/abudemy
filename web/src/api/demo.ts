import type { RequestAdapter } from './client'
import { ApiError } from './errors'
import type {
  Activity,
  ActivityInput,
  AnalyticsOverview,
  AssistanceRequest,
  Enrollment,
  Organization,
  OrganizationMember,
  Opportunity,
  Page,
  Service,
  ServiceCategory,
  ServiceInput,
  Session,
  TokenPair,
  User,
} from './types'

/**
 * Adaptador de DEMOSTRACIÓN.
 *
 * Solo se activa con `VITE_DEMO_MODE=true`. Devuelve datos **ficticios** que
 * se identifican con el prefijo `[DEMO]` y que la interfaz anuncia mediante un
 * aviso permanente. Nunca se mezcla con los datos reales del backend.
 *
 * La sesión demo es de **administrador** para que todas las secciones del
 * portal sean navegables (incluidas Personas registradas, Usuarios y Cuentas
 * institucionales). La cuenta institucional pendiente permite probar el flujo
 * de aprobación.
 */

const DEMO_TAG = '[DEMO]'
const now = Date.now()
const iso = (offsetDays = 0) => new Date(now + offsetDays * 86_400_000).toISOString()
const DEMO_USER_ID = '00000000-0000-4000-8000-000000000001'
const DEMO_ORG_A = '00000000-0000-4000-8000-0000000000a1'
const DEMO_ORG_B = '00000000-0000-4000-8000-0000000000a2'

const categories: ServiceCategory[] = [
  { id: 'cat-salud', name: `${DEMO_TAG} Salud y bienestar`, slug: 'salud', description: 'Consultorios, CESFAM y programas de salud.' },
  { id: 'cat-social', name: `${DEMO_TAG} Apoyo social`, slug: 'apoyo-social', description: 'Acompañamiento y ayudas sociales.' },
  { id: 'cat-cultura', name: `${DEMO_TAG} Cultura y recreación`, slug: 'cultura', description: 'Talleres culturales y recreativos.' },
  { id: 'cat-deporte', name: `${DEMO_TAG} Deporte y movimiento`, slug: 'deporte', description: 'Gimnasia, caminatas y deporte adaptado.' },
  { id: 'cat-digital', name: `${DEMO_TAG} Inclusión digital`, slug: 'digital', description: 'Alfabetización digital y trámites en línea.' },
]

const DEMO_ORG_NAMES: Record<string, string> = {
  [DEMO_ORG_A]: `${DEMO_TAG} Municipalidad de Santiago`,
  [DEMO_ORG_B]: `${DEMO_TAG} Fundación Manos Amigas`,
}

function makeService(
  id: string,
  name: string,
  categoryId: string,
  comuna: string,
  region: string,
  lat: number,
  lng: number,
  organizationId: string | null,
  status: Service['status'] = 'active',
): Service {
  const category = categories.find((c) => c.id === categoryId) ?? null
  return {
    id,
    organization_id: organizationId,
    organization_name: organizationId ? DEMO_ORG_NAMES[organizationId] ?? '' : '',
    category_id: categoryId,
    category,
    name: `${DEMO_TAG} ${name}`,
    description: 'Registro ficticio utilizado únicamente en modo demostración.',
    type: 'establishment',
    status,
    address: `Calle Demo 123, ${comuna}`,
    comuna,
    region,
    latitude: lat,
    longitude: lng,
    phone: '+56 2 2000 0000',
    email: 'demo@abudemy.example',
    website: '',
    schedule: [
      { day_of_week: 1, opens_at: '09:00', closes_at: '13:00' },
      { day_of_week: 3, opens_at: '09:00', closes_at: '13:00' },
    ],
    created_by: DEMO_USER_ID,
    created_at: iso(-40),
    updated_at: iso(-3),
  }
}

const services: Service[] = [
  makeService('svc-1', 'CESFAM Centro', 'cat-salud', 'Santiago', 'Metropolitana', -33.4489, -70.6693, DEMO_ORG_A),
  makeService('svc-2', 'Club de adulto mayor Los Aromos', 'cat-social', 'Providencia', 'Metropolitana', -33.4315, -70.6093, DEMO_ORG_A),
  makeService('svc-3', 'Taller de memoria', 'cat-salud', 'Maipú', 'Metropolitana', -33.5111, -70.7577, DEMO_ORG_B),
  makeService('svc-4', 'Gimnasia suave municipal', 'cat-deporte', 'La Florida', 'Metropolitana', -33.5225, -70.5983, DEMO_ORG_B),
  makeService('svc-5', 'Punto de inclusión digital', 'cat-digital', 'Puente Alto', 'Metropolitana', -33.6116, -70.5755, DEMO_ORG_A),
  makeService('svc-6', 'Casa de la cultura', 'cat-cultura', 'Ñuñoa', 'Metropolitana', -33.4569, -70.5974, DEMO_ORG_B, 'draft'),
  makeService('svc-7', 'Centro diurno El Roble', 'cat-social', 'Valparaíso', 'Valparaíso', -33.0472, -71.6127, DEMO_ORG_A),
  makeService('svc-8', 'Consultorio Playa Ancha', 'cat-salud', 'Valparaíso', 'Valparaíso', -33.0311, -71.6299, DEMO_ORG_A, 'inactive'),
  makeService('svc-9', 'Taller de tejido', 'cat-cultura', 'Viña del Mar', 'Valparaíso', -33.0245, -71.5518, DEMO_ORG_B),
  makeService('svc-10', 'Caminatas guiadas', 'cat-deporte', 'Concepción', 'Biobío', -36.8269, -73.0498, null),
  makeService('svc-11', 'Alfabetización digital', 'cat-digital', 'Concepción', 'Biobío', -36.8270, -73.0503, null),
  makeService('svc-12', 'Programa de acompañamiento', 'cat-social', 'Temuco', 'Araucanía', -38.7359, -72.5904, null),
]

function makeActivity(
  id: string,
  title: string,
  comuna: string,
  region: string,
  lat: number,
  lng: number,
  organizationId: string | null,
  capacity: number,
  enrolled: number,
  startOffset: number,
  type: Activity['type'],
  status: Activity['status'],
): Activity {
  return {
    id,
    organization_id: organizationId,
    service_id: null,
    category_id: null,
    title: `${DEMO_TAG} ${title}`,
    description: 'Actividad ficticia del modo demostración.',
    type,
    modality: 'in_person',
    status,
    location: `Sede ${comuna}`,
    comuna,
    region,
    latitude: lat,
    longitude: lng,
    start_at: iso(startOffset),
    end_at: iso(startOffset + 0.08),
    capacity,
    enrolled_count: enrolled,
    price: 0,
    currency: 'CLP',
    created_at: iso(-20),
    updated_at: iso(-2),
  }
}

const activities: Activity[] = [
  makeActivity('act-1', 'Taller de memoria activa', 'Santiago', 'Metropolitana', -33.4489, -70.6693, DEMO_ORG_A, 30, 27, 3, 'workshop', 'published'),
  makeActivity('act-2', 'Curso de celulares', 'Providencia', 'Metropolitana', -33.4315, -70.6093, DEMO_ORG_A, 20, 20, 5, 'course', 'published'),
  makeActivity('act-3', 'Encuentro de clubes', 'Maipú', 'Metropolitana', -33.5111, -70.7577, DEMO_ORG_B, 100, 64, 10, 'event', 'published'),
  makeActivity('act-4', 'Grupo de apoyo mutuo', 'La Florida', 'Metropolitana', -33.5225, -70.5983, DEMO_ORG_B, 15, 9, 2, 'support_group', 'published'),
  makeActivity('act-5', 'Yoga para mayores', 'Viña del Mar', 'Valparaíso', -33.0245, -71.5518, DEMO_ORG_B, 25, 0, 14, 'workshop', 'draft'),
  makeActivity('act-6', 'Caminata patrimonial', 'Valparaíso', 'Valparaíso', -33.0472, -71.6127, DEMO_ORG_A, 40, 12, -6, 'event', 'finished'),
]

const enrollments: Enrollment[] = Array.from({ length: 24 }).map((_, i) => ({
  id: `enr-${i}`,
  activity_id: activities[i % activities.length].id,
  user_id: `user-${i}`,
  status: i % 7 === 0 ? 'attended' : 'enrolled',
  created_at: iso(-i),
  updated_at: iso(-i),
  activity_title: activities[i % activities.length].title,
  user_email: `persona${i}@demo.example`,
  user_full_name: `Persona Demo ${i}`,
}))

const organizations: Organization[] = [
  {
    id: DEMO_ORG_A,
    name: DEMO_ORG_NAMES[DEMO_ORG_A],
    type: 'government',
    legal_id: 'DEMO-76.111.111-1',
    description: 'Organización ficticia del modo demostración.',
    email: 'muni@demo.example',
    phone: '+56 2 2000 0001',
    address: 'Av. Demo 100',
    comuna: 'Santiago',
    region: 'Metropolitana',
    latitude: -33.4489,
    longitude: -70.6693,
    active: true,
    created_at: iso(-90),
    updated_at: iso(-10),
  },
  {
    id: DEMO_ORG_B,
    name: DEMO_ORG_NAMES[DEMO_ORG_B],
    type: 'ngo',
    legal_id: 'DEMO-65.222.222-2',
    description: 'Organización ficticia del modo demostración.',
    email: 'fundacion@demo.example',
    phone: '+56 2 2000 0002',
    address: 'Av. Demo 200',
    comuna: 'Maipú',
    region: 'Metropolitana',
    latitude: -33.5111,
    longitude: -70.7577,
    active: true,
    created_at: iso(-80),
    updated_at: iso(-9),
  },
]

const members: OrganizationMember[] = [
  {
    id: 'mem-1',
    organization_id: DEMO_ORG_A,
    user_id: DEMO_USER_ID,
    role: 'owner',
    user_email: 'admin@demo.example',
    user_full_name: 'Administración de la plataforma (demo)',
    created_at: iso(-80),
  },
]

const opportunities: Opportunity[] = [
  {
    id: 'opp-1',
    organization_id: DEMO_ORG_A,
    title: `${DEMO_TAG} Monitor de talleres`,
    description: 'Oportunidad ficticia del modo demostración.',
    type: 'gig',
    status: 'open',
    comuna: 'Santiago',
    region: 'Metropolitana',
    salary_min: 250000,
    salary_max: 350000,
    currency: 'CLP',
    contact_email: 'demo@abudemy.example',
    expires_at: iso(20),
    created_at: iso(-5),
    updated_at: iso(-1),
  },
]

const assistance: AssistanceRequest[] = Array.from({ length: 9 }).map((_, i) => ({
  id: `asi-${i}`,
  user_id: `user-${i}`,
  type: ['medical', 'social', 'domestic'][i % 3],
  description: 'Solicitud ficticia del modo demostración.',
  latitude: null,
  longitude: null,
  status: (['pending', 'notified', 'resolved'][i % 3] as AssistanceRequest['status']),
  created_at: iso(-i * 2),
  updated_at: iso(-i),
  resolved_at: i % 3 === 2 ? iso(-i) : null,
  notifications: [],
}))

const user: User = {
  id: DEMO_USER_ID,
  email: 'admin@demo.example',
  full_name: 'Administración de la plataforma (demo)',
  phone: '+56 9 0000 0000',
  role: 'admin',
  status: 'active',
  institutional_approved: true,
  created_at: iso(-120),
  updated_at: iso(-1),
  last_login_at: iso(0),
}

/**
 * Cuentas de la comunidad para las páginas "Personas registradas", "Usuarios"
 * y "Cuentas institucionales". Reflejan el perfil sociodemográfico que la app
 * móvil pide al registrarse.
 */
function makeDemoPerson(
  id: string,
  email: string,
  fullName: string,
  birthDate: string,
  comuna: string,
  healthSystem: NonNullable<User['health_system']>,
  sex: NonNullable<User['sex']>,
  livingSituation: NonNullable<User['living_situation']>,
  status: User['status'] = 'active',
): User {
  return {
    id,
    email,
    full_name: `${DEMO_TAG} ${fullName}`,
    phone: '+56 9 0000 0000',
    birth_date: birthDate,
    comuna,
    health_system: healthSystem,
    sex,
    living_situation: livingSituation,
    role: 'mobile_user',
    status,
    institutional_approved: true,
    created_at: iso(-45),
    updated_at: iso(-2),
    last_login_at: iso(-1),
  }
}

const pendingInstitutional: User = {
  id: 'demo-institutional-pending',
  email: 'pendiente.institucional@demo.example',
  full_name: `${DEMO_TAG} Clínica Los Cerros (solicitud pendiente)`,
  phone: '+56 41 222 9988',
  comuna: 'Talcahuano',
  role: 'institutional_client',
  status: 'active',
  institutional_approved: false,
  created_at: iso(-10),
  updated_at: iso(-10),
}

const people: User[] = [
  user,
  pendingInstitutional,
  makeDemoPerson('person-1', 'maria.gonzalez@demo.example', 'María González Pérez', '1948-05-12', 'Concepción', 'fonasa', 'female', 'with_family'),
  makeDemoPerson('person-2', 'juan.soto@demo.example', 'Juan Soto Muñoz', '1951-11-03', 'Talcahuano', 'fonasa', 'male', 'with_partner'),
  makeDemoPerson('person-3', 'rosa.munoz@demo.example', 'Rosa Muñoz Riquelme', '1945-02-27', 'Talcahuano', 'capredena', 'female', 'alone'),
  makeDemoPerson('person-4', 'carmen.rojas@demo.example', 'Carmen Rojas Vega', '1947-03-18', 'Talcahuano', 'fonasa', 'female', 'with_children'),
  makeDemoPerson('person-5', 'elena.torres@demo.example', 'Elena Torres Molina', '1944-12-24', 'Concepción', 'isapre', 'female', 'with_partner_and_children'),
  makeDemoPerson('person-6', 'patricia.nunez@demo.example', 'Patricia Núñez Salazar', '1955-06-09', 'Concepción', 'fonasa', 'female', 'alone'),
  makeDemoPerson('person-7', 'hugo.fuentes@demo.example', 'Hugo Fuentes Rivas', '1949-01-30', 'Hualpén', 'isapre', 'male', 'with_partner'),
  makeDemoPerson('person-8', 'teresa.aguilera@demo.example', 'Teresa Aguilera Pinto', '1960-09-14', 'Chiguayante', 'capredena', 'female', 'with_children'),
  makeDemoPerson('person-9', 'raul.paredes@demo.example', 'Raúl Paredes Moreno', '1958-04-22', 'San Pedro de la Paz', 'fonasa', 'male', 'with_family'),
  makeDemoPerson('person-10', 'silvia.castro@demo.example', 'Silvia Castro Vidal', '1953-11-05', 'Talcahuano', 'dipreca', 'female', 'with_caregiver'),
  makeDemoPerson('person-11', 'andres.molina@demo.example', 'Andrés Molina Fuentes', '1970-02-17', 'Concepción', 'fonasa', 'male', 'with_partner', 'inactive'),
]

/** Edad cumplida a partir de la fecha de nacimiento. */
function ageOf(birthDate?: string): number | undefined {
  if (!birthDate) return undefined
  const birth = new Date(birthDate)
  if (Number.isNaN(birth.getTime())) return undefined
  const today = new Date()
  let age = today.getFullYear() - birth.getFullYear()
  const month = today.getMonth() - birth.getMonth()
  if (month < 0 || (month === 0 && today.getDate() < birth.getDate())) age -= 1
  return age
}

/** Filtros de `/api/v1/users` que usa el portal (role, estado, búsqueda y perfil). */
function filterPeople(query: Record<string, unknown> | undefined): User[] {
  let items = [...people]
  if (query?.role) items = items.filter((p) => p.role === query.role)
  if (query?.status) items = items.filter((p) => p.status === query.status)
  if (query?.comuna) {
    const term = String(query.comuna).toLowerCase()
    items = items.filter((p) => (p.comuna ?? '').toLowerCase().includes(term))
  }
  if (query?.health_system) items = items.filter((p) => p.health_system === query.health_system)
  if (query?.sex) items = items.filter((p) => p.sex === query.sex)
  if (query?.living_situation) {
    items = items.filter((p) => p.living_situation === query.living_situation)
  }
  if (query?.search) {
    const term = String(query.search).toLowerCase()
    items = items.filter(
      (p) => p.full_name.toLowerCase().includes(term) || p.email.toLowerCase().includes(term),
    )
  }
  const ageMin = query?.age_min === undefined ? undefined : Number(query.age_min)
  const ageMax = query?.age_max === undefined ? undefined : Number(query.age_max)
  if (ageMin !== undefined && Number.isFinite(ageMin)) {
    items = items.filter((p) => {
      const age = ageOf(p.birth_date)
      return age !== undefined && age >= ageMin
    })
  }
  if (ageMax !== undefined && Number.isFinite(ageMax)) {
    items = items.filter((p) => {
      const age = ageOf(p.birth_date)
      return age !== undefined && age <= ageMax
    })
  }
  return items
}

const sessions: Session[] = [
  {
    id: 'ses-demo',
    user_id: DEMO_USER_ID,
    user_agent: 'Navegador de demostración',
    ip_address: '127.0.0.1',
    created_at: iso(0),
    expires_at: iso(30),
    current: true,
  },
]

const demoPair: TokenPair = {
  access_token: 'demo-access-token',
  refresh_token: 'demo-refresh-token',
  token_type: 'Bearer',
  expires_in: 3600,
  session_id: 'ses-demo',
  user,
}

function paginate<T>(items: T[], query: Record<string, unknown> | undefined): Page<T> {
  const page = Math.max(1, Number(query?.page ?? 1) || 1)
  const perPage = Math.min(100, Math.max(1, Number(query?.per_page ?? 20) || 20))
  const start = (page - 1) * perPage
  return {
    items: items.slice(start, start + perPage),
    pagination: {
      page,
      per_page: perPage,
      total: items.length,
      total_pages: Math.ceil(items.length / perPage),
    },
  }
}

function buildOverview(orgId?: string): AnalyticsOverview {
  const svc = orgId ? services.filter((s) => s.organization_id === orgId) : services
  const act = orgId ? activities.filter((a) => a.organization_id === orgId) : activities
  const actIds = new Set(act.map((a) => a.id))
  const enr = enrollments.filter((e) => actIds.has(e.activity_id))
  const byCategory = new Map<string, number>()
  for (const s of svc) {
    const label = s.category?.name ?? 'Sin categoría'
    byCategory.set(label, (byCategory.get(label) ?? 0) + 1)
  }
  const byType = new Map<string, number>()
  for (const a of act) byType.set(a.type, (byType.get(a.type) ?? 0) + 1)
  const byComuna = new Map<string, { region: string; count: number }>()
  for (const s of svc) {
    if (!s.comuna) continue
    const entry = byComuna.get(s.comuna) ?? { region: s.region ?? '', count: 0 }
    entry.count += 1
    byComuna.set(s.comuna, entry)
  }
  const occ = act.filter((a) => a.capacity > 0)
  return {
    services_total: svc.length,
    services_active: svc.filter((s) => s.status === 'active').length,
    activities_total: act.length,
    activities_published: act.filter((a) => a.status === 'published').length,
    activities_upcoming: act.filter((a) => a.status === 'published' && new Date(a.start_at) >= new Date()).length,
    enrollments_total: enr.length,
    enrollments_active: enr.filter((e) => e.status === 'enrolled' || e.status === 'attended').length,
    users_total: people.length,
    users_mobile: people.filter((p) => p.role === 'mobile_user').length,
    organizations_total: organizations.length,
    opportunities_open: opportunities.filter((o) => o.status === 'open').length,
    assistance_total: assistance.length,
    assistance_pending: assistance.filter((a) => a.status === 'pending').length,
    average_occupancy: occ.length
      ? occ.reduce((acc, a) => acc + Math.min(1, a.enrolled_count / a.capacity), 0) / occ.length
      : 0,
    services_by_category: [...byCategory.entries()].map(([label, count]) => ({ label, count })),
    activities_by_type: [...byType.entries()].map(([label, count]) => ({ label, count })),
    territorial_distribution: [...byComuna.entries()].map(([comuna, v]) => ({
      comuna,
      region: v.region,
      count: v.count,
    })),
  }
}

function csvBlob(rows: (string | number)[][]): Blob {
  const content = rows.map((r) => r.map((c) => `"${String(c).replace(/"/g, '""')}"`).join(',')).join('\n')
  return new Blob([content], { type: 'text/csv;charset=utf-8' })
}

function reportRows(type: string): (string | number)[][] {
  switch (type) {
    case 'services':
      return [['id', 'name', 'organization', 'category', 'type', 'status', 'comuna', 'region', 'phone', 'email'],
        ...services.map((s) => [s.id, s.name, s.organization_name ?? '', s.category?.name ?? '', s.type, s.status, s.comuna ?? '', s.region ?? '', s.phone ?? '', s.email ?? ''])]
    case 'activities':
      return [['id', 'title', 'type', 'modality', 'status', 'comuna', 'region', 'start_at', 'capacity', 'enrolled'],
        ...activities.map((a) => [a.id, a.title, a.type, a.modality, a.status, a.comuna ?? '', a.region ?? '', a.start_at, a.capacity, a.enrolled_count])]
    case 'enrollments':
      return [['id', 'activity', 'user_email', 'user_name', 'status'],
        ...enrollments.map((e) => [e.id, e.activity_title ?? '', e.user_email ?? '', e.user_full_name ?? '', e.status])]
    case 'assistance':
      return [['id', 'type', 'status', 'created_at'],
        ...assistance.map((a) => [a.id, a.type, a.status, a.created_at])]
    case 'opportunities':
      return [['id', 'title', 'type', 'status', 'comuna', 'region'],
        ...opportunities.map((o) => [o.id, o.title, o.type, o.status, o.comuna ?? '', o.region ?? ''])]
    case 'territorial':
      return [['comuna', 'region', 'services'],
        ...buildOverview().territorial_distribution.map((t) => [t.comuna, t.region, t.count])]
    default:
      throw new ApiError(404, 'not_found', 'El informe solicitado no existe.')
  }
}

/** Punto de entrada del adaptador. */
export const demoAdapter: RequestAdapter = async (method, path, options) => {
  const query = options.query as Record<string, unknown> | undefined
  await new Promise((resolve) => setTimeout(resolve, 120))

  // ---- Auth ----
  if (path === '/api/v1/auth/login' && method === 'POST') {
    const body = options.body as { email?: string; password?: string }
    if (!body?.email || !body?.password) {
      throw new ApiError(400, 'validation_error', 'Email y contraseña son obligatorios.')
    }
    return demoPair
  }
  if (path === '/api/v1/auth/refresh') return demoPair
  if (path === '/api/v1/auth/logout' || path === '/api/v1/auth/logout-all') return undefined
  if (path === '/api/v1/auth/sessions' && method === 'GET') return { items: sessions }
  if (path.startsWith('/api/v1/auth/sessions/') && method === 'DELETE') return undefined
  if (path === '/api/v1/auth/password/forgot') {
    return { message: 'Modo demostración: no se envían correos reales.' }
  }

  // ---- Users ----
  if (path === '/api/v1/users/me' && method === 'GET') return user
  if (path === '/api/v1/users/me' && method === 'PATCH') {
    const body = options.body as { full_name?: string; phone?: string; birth_date?: string }
    if (body.full_name) user.full_name = body.full_name
    if (body.phone !== undefined) user.phone = body.phone
    if (body.birth_date !== undefined) user.birth_date = body.birth_date
    user.updated_at = iso(0)
    return user
  }
  if (path === '/api/v1/users/me/enrollments') return paginate(enrollments, query)
  if (path === '/api/v1/users/me/password' && method === 'POST') return undefined
  if (path === '/api/v1/users' && method === 'GET') return paginate(filterPeople(query), query)
  if (path.startsWith('/api/v1/users/') && path.endsWith('/status') && method === 'PATCH') {
    const body = options.body as { status: User['status'] }
    const target = people.find((p) => p.id === path.split('/')[4]) ?? user
    target.status = body.status
    target.updated_at = iso(0)
    return target
  }
  if (path.startsWith('/api/v1/users/') && path.endsWith('/role') && method === 'PATCH') {
    const body = options.body as { role: User['role'] }
    const target = people.find((p) => p.id === path.split('/')[4]) ?? user
    target.role = body.role
    target.updated_at = iso(0)
    return target
  }
  if (path.startsWith('/api/v1/users/') && path.endsWith('/institutional-approval') && method === 'PATCH') {
    const body = options.body as { approved: boolean }
    const target = people.find((p) => p.id === path.split('/')[4]) ?? user
    target.institutional_approved = body.approved
    target.updated_at = iso(0)
    return target
  }
  if (path.startsWith('/api/v1/users/') && method === 'GET') {
    return people.find((p) => p.id === path.split('/')[4]) ?? user
  }
  if (path.includes('/api/v1/users/') && (method === 'PATCH' || method === 'POST')) {
    return people.find((p) => p.id === path.split('/')[4]) ?? user
  }

  // ---- Catálogo ----
  if (path === '/api/v1/service-categories') return { items: categories }
  if (path === '/api/v1/services' && method === 'GET') {
    let items = [...services]
    if (query?.comuna) items = items.filter((s) => s.comuna === query.comuna)
    if (query?.category_id) items = items.filter((s) => s.category_id === query.category_id)
    if (query?.status) items = items.filter((s) => s.status === query.status)
    if (query?.type) items = items.filter((s) => s.type === query.type)
    if (query?.organization_id) items = items.filter((s) => s.organization_id === query.organization_id)
    if (query?.search) {
      const term = String(query.search).toLowerCase()
      items = items.filter((s) => s.name.toLowerCase().includes(term))
    }
    return paginate(items, query)
  }
  if (path === '/api/v1/services' && method === 'POST') {
    const input = options.body as ServiceInput
    const created: Service = {
      ...makeService(`svc-${Date.now()}`, input.name, input.category_id ?? 'cat-social', input.comuna ?? '', input.region ?? '', input.latitude ?? 0, input.longitude ?? 0, input.organization_id ?? null),
      ...input,
      id: `svc-${Date.now()}`,
      schedule: input.schedule ?? [],
      created_at: iso(0),
      updated_at: iso(0),
    }
    services.push(created)
    return created
  }
  if (path.startsWith('/api/v1/services/')) {
    const id = path.split('/')[4]
    const index = services.findIndex((s) => s.id === id)
    if (method === 'GET') {
      if (index < 0) throw new ApiError(404, 'not_found', 'Servicio no encontrado.')
      return services[index]
    }
    if (method === 'DELETE') {
      if (index >= 0) services.splice(index, 1)
      return undefined
    }
    if (method === 'PUT') {
      const input = options.body as ServiceInput
      const updated = { ...services[index], ...input, id, updated_at: iso(0) } as Service
      services[index] = updated
      return updated
    }
    if (method === 'PATCH') {
      const body = options.body as { status: Service['status'] }
      const updated = { ...services[index], status: body.status as Service['status'], updated_at: iso(0) }
      services[index] = updated
      return updated
    }
  }

  // ---- Actividades ----
  if (path === '/api/v1/activities' && method === 'GET') {
    let items = [...activities]
    if (query?.comuna) items = items.filter((a) => a.comuna === query.comuna)
    if (query?.status) items = items.filter((a) => a.status === query.status)
    if (query?.type) items = items.filter((a) => a.type === query.type)
    if (query?.modality) items = items.filter((a) => a.modality === query.modality)
    if (query?.organization_id) items = items.filter((a) => a.organization_id === query.organization_id)
    if (query?.search) {
      const term = String(query.search).toLowerCase()
      items = items.filter((a) => a.title.toLowerCase().includes(term))
    }
    return paginate(items, query)
  }
  if (path === '/api/v1/activities' && method === 'POST') {
    const input = options.body as ActivityInput
    const created = makeActivity(`act-${Date.now()}`, input.title, input.comuna ?? '', input.region ?? '', input.latitude ?? 0, input.longitude ?? 0, input.organization_id ?? null, input.capacity ?? 0, 0, 7, input.type ?? 'workshop', input.status ?? 'draft')
    activities.push(created)
    return created
  }
  if (path.startsWith('/api/v1/activities/')) {
    const parts = path.split('/')
    const id = parts[4]
    const index = activities.findIndex((a) => a.id === id)
    if (parts[5] === 'enrollments' && parts[6] && method === 'PATCH') return undefined
    if (parts[5] === 'enrollments' && method === 'GET') {
      return paginate(enrollments.filter((e) => e.activity_id === id), query)
    }
    if (method === 'GET') {
      if (index < 0) throw new ApiError(404, 'not_found', 'Actividad no encontrada.')
      return activities[index]
    }
    if (method === 'PUT') {
      const input = options.body as ActivityInput
      const updated = { ...activities[index], ...input, id, updated_at: iso(0) } as Activity
      activities[index] = updated
      return updated
    }
    if (method === 'PATCH') {
      const body = options.body as { status: Activity['status'] }
      const updated = { ...activities[index], status: body.status, updated_at: iso(0) }
      activities[index] = updated
      return updated
    }
  }

  // ---- Organizaciones ----
  if (path === '/api/v1/organizations/mine') {
    return {
      items: [
        { organization: organizations[0], role: 'owner' },
        { organization: organizations[1], role: 'editor' },
      ],
    }
  }
  if (path === '/api/v1/organizations' && method === 'GET') return paginate(organizations, query)
  if (path === '/api/v1/organizations' && method === 'POST') {
    const input = options.body as Partial<Organization>
    const created: Organization = {
      id: `org-${Date.now()}`,
      name: input.name ?? `${DEMO_TAG} Organización`,
      type: input.type ?? 'other',
      active: true,
      created_at: iso(0),
      updated_at: iso(0),
    }
    organizations.push(created)
    return created
  }
  if (path.startsWith('/api/v1/organizations/')) {
    const parts = path.split('/')
    const id = parts[4]
    if (parts[5] === 'members') {
      if (method === 'GET') return { items: members.filter((m) => m.organization_id === id) }
      if (method === 'POST') {
        const body = options.body as { role: OrganizationMember['role']; email?: string }
        const created: OrganizationMember = {
          id: `mem-${Date.now()}`,
          organization_id: id,
          user_id: `user-${Date.now()}`,
          role: body.role ?? 'viewer',
          user_email: body.email,
          user_full_name: body.email ?? 'Nuevo integrante (demo)',
          created_at: iso(0),
        }
        members.push(created)
        return created
      }
      const memberUserId = parts[6]
      const memberIndex = members.findIndex(
        (m) => m.organization_id === id && m.user_id === memberUserId,
      )
      if (method === 'PATCH' && memberIndex >= 0) {
        const body = options.body as { role: OrganizationMember['role'] }
        members[memberIndex] = { ...members[memberIndex], role: body.role }
        return undefined
      }
      if (method === 'DELETE' && memberIndex >= 0) {
        members.splice(memberIndex, 1)
        return undefined
      }
      return undefined
    }
    const index = organizations.findIndex((o) => o.id === id)
    if (method === 'GET') {
      if (index < 0) throw new ApiError(404, 'not_found', 'Organización no encontrada.')
      return organizations[index]
    }
    if (method === 'PUT') {
      const input = options.body as Partial<Organization>
      const updated = { ...organizations[index], ...input, id, updated_at: iso(0) }
      organizations[index] = updated
      return updated
    }
  }

  // ---- Oportunidades ----
  if (path === '/api/v1/opportunities' && method === 'GET') return paginate(opportunities, query)

  // ---- Asistencia ----
  if (path === '/api/v1/assistance/requests' && method === 'GET') {
    let items = [...assistance]
    if (query?.status) items = items.filter((a) => a.status === query.status)
    return paginate(items, query)
  }

  // ---- Analítica ----
  if (path === '/api/v1/analytics/overview') {
    return buildOverview(query?.organization_id ? String(query.organization_id) : undefined)
  }
  if (path === '/api/v1/analytics/reports') {
    return { items: ['services', 'activities', 'enrollments', 'assistance', 'opportunities', 'territorial'], formats: ['csv', 'json'] }
  }
  if (path.startsWith('/api/v1/analytics/reports/')) {
    const type = path.split('/')[5]
    const rows = reportRows(type)
    if (query?.format === 'json') {
      const body = rows.slice(1)
      return { type, headers: rows[0], rows: body, total: body.length, truncated: false }
    }
    return csvBlob(rows)
  }

  throw new ApiError(404, 'not_found', `Modo demostración: endpoint no implementado (${method} ${path}).`)
}
