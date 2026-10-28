#!/usr/bin/env node
/**
 * Semilla de datos de desarrollo para el PORTAL WEB de Abudemy.
 *
 * Crea —a través de la API REST real, nunca por SQL directo— un conjunto de
 * datos representativo de las secciones del portal institucional: cuentas,
 * organizaciones, integrantes, servicios con horarios, actividades con
 * inscripciones, oportunidades con postulaciones, avisos y solicitudes de
 * asistencia.
 *
 * Uso:
 *   SEED_API_URL=http://localhost:8080 npm run seed:web
 *
 * Variables:
 *   SEED_API_URL / VITE_API_BASE_URL  URL base de la API (por defecto http://localhost:8080)
 *   SEED_ADMIN_EMAIL / SEED_ADMIN_PASSWORD  cuenta administradora (por defecto admin@abudemy.cl / Admin1234)
 *
 * Nota: cada ejecución crea recursos nuevos. Está pensada para desarrollo.
 * Si detecta que la organización ancla ya existe, se detiene para no duplicar.
 * El administrador es obligatorio: la cuenta institucional debe aprobarse (0009)
 * antes de crear organizaciones.
 */

const API = (process.env.SEED_API_URL || process.env.VITE_API_BASE_URL || 'http://localhost:8080').replace(/\/+$/, '')
const ADMIN_EMAIL = process.env.SEED_ADMIN_EMAIL || 'admin@abudemy.cl'
const ADMIN_PASSWORD = process.env.SEED_ADMIN_PASSWORD || 'Admin1234'

const INSTITUTIONAL = {
  email: 'web.institucional@abudemy.cl',
  password: 'WebDemo1234',
  fullName: 'Portal Institucional Abudemy',
}

// Personas mayores (rol mobile_user por defecto) que participan del contenido.
// El perfil sociodemográfico es el que pide la app móvil al registrarse; así
// "Personas registradas" y sus filtros muestran datos coherentes.
const MOBILE_USERS = [
  { key: 'carmen', email: 'carmen.rojas@ejemplo.cl', password: 'Demo1234', fullName: 'Carmen Rojas Vega', phone: '+56940112233', birthDate: '1947-03-18', comuna: 'Talcahuano', healthSystem: 'fonasa', sex: 'female', livingSituation: 'with_children' },
  { key: 'luis', email: 'luis.vera@ejemplo.cl', password: 'Demo1234', fullName: 'Luis Vera Contreras', phone: '+56941223344', birthDate: '1950-08-02', comuna: 'Concepción', healthSystem: 'fonasa', sex: 'male', livingSituation: 'with_partner' },
  { key: 'elena', email: 'elena.torres@ejemplo.cl', password: 'Demo1234', fullName: 'Elena Torres Molina', phone: '+56942334455', birthDate: '1944-12-24', comuna: 'Concepción', healthSystem: 'isapre', sex: 'female', livingSituation: 'with_partner_and_children' },
  { key: 'patricia', email: 'patricia.nunez@ejemplo.cl', password: 'Demo1234', fullName: 'Patricia Núñez Salazar', phone: '+56945667788', birthDate: '1955-06-09', comuna: 'Concepción', healthSystem: 'fonasa', sex: 'female', livingSituation: 'alone' },
  { key: 'hugo', email: 'hugo.fuentes@ejemplo.cl', password: 'Demo1234', fullName: 'Hugo Fuentes Rivas', phone: '+56946778899', birthDate: '1949-01-30', comuna: 'Hualpén', healthSystem: 'isapre', sex: 'male', livingSituation: 'with_partner' },
  { key: 'teresa', email: 'teresa.aguilera@ejemplo.cl', password: 'Demo1234', fullName: 'Teresa Aguilera Pinto', phone: '+56947889900', birthDate: '1960-09-14', comuna: 'Chiguayante', healthSystem: 'capredena', sex: 'female', livingSituation: 'with_children' },
  { key: 'raul', email: 'raul.paredes@ejemplo.cl', password: 'Demo1234', fullName: 'Raúl Paredes Moreno', phone: '+56948990011', birthDate: '1958-04-22', comuna: 'San Pedro de la Paz', healthSystem: 'fonasa', sex: 'male', livingSituation: 'with_family' },
  { key: 'silvia', email: 'silvia.castro@ejemplo.cl', password: 'Demo1234', fullName: 'Silvia Castro Vidal', phone: '+56949001122', birthDate: '1953-11-05', comuna: 'Talcahuano', healthSystem: 'dipreca', sex: 'female', livingSituation: 'with_caregiver' },
]

const ANCHOR_ORG = 'Corporación Cultural Biobío'

async function request(method, path, { token, body } = {}) {
  const headers = { Accept: 'application/json' }
  if (body !== undefined) headers['Content-Type'] = 'application/json'
  if (token) headers.Authorization = `Bearer ${token}`

  const response = await fetch(`${API}${path}`, {
    method,
    headers,
    body: body === undefined ? undefined : JSON.stringify(body),
  })

  const text = await response.text()
  let data
  try {
    data = text ? JSON.parse(text) : undefined
  } catch {
    data = text
  }

  if (!response.ok) {
    const message = data?.error?.message || data?.message || `HTTP ${response.status}`
    const error = new Error(`${method} ${path} → ${response.status}: ${message}`)
    error.status = response.status
    error.data = data
    throw error
  }
  return data
}

async function session({
  email,
  password,
  fullName,
  role,
  phone,
  birthDate,
  comuna,
  healthSystem,
  sex,
  livingSituation,
}) {
  try {
    const pair = await request('POST', '/api/v1/auth/login', { body: { email, password } })
    console.log(`  · sesión iniciada como ${email}`)
    return pair.access_token
  } catch (error) {
    if (error.status !== 400 && error.status !== 401) throw error
  }
  const body = { email, password, full_name: fullName }
  if (role) body.role = role
  if (phone) body.phone = phone
  if (birthDate) body.birth_date = birthDate
  if (comuna) body.comuna = comuna
  if (healthSystem) body.health_system = healthSystem
  if (sex) body.sex = sex
  if (livingSituation) body.living_situation = livingSituation
  await request('POST', '/api/v1/auth/register', { body })
  const pair = await request('POST', '/api/v1/auth/login', { body: { email, password } })
  console.log(`  · cuenta registrada e iniciada: ${email}`)
  return pair.access_token
}

const daysFromNow = (days, hour, minute = 0) => {
  const d = new Date(Date.now() + days * 86_400_000)
  d.setUTCHours(hour, minute, 0, 0)
  return d.toISOString()
}

async function main() {
  console.log(`Sembrando datos del portal web en ${API}\n`)

  // 1. Sesiones
  console.log('Cuentas:')
  const inst = await session({ ...INSTITUTIONAL, role: 'institutional_client' })
  const users = {}
  for (const user of MOBILE_USERS) {
    users[user.key] = { token: await session({ ...user, role: 'mobile_user' }), ...user }
  }
  let admin
  try {
    admin = await session({ email: ADMIN_EMAIL, password: ADMIN_PASSWORD })
  } catch (error) {
    throw new Error(
      `No se pudo iniciar sesión como administrador (${ADMIN_EMAIL}).\n` +
        '  La cuenta la crea la migración 0011_demo_data_refresh.sql al arrancar la API.\n' +
        '  Si usas otro administrador, define SEED_ADMIN_EMAIL y SEED_ADMIN_PASSWORD.\n' +
        `  Detalle: ${error.message}`,
    )
  }

  // Aprueba la cuenta institucional (0009). Sin la aprobación, la API devuelve
  // organizaciones vacías y rechaza crear organizaciones.
  const instUser = await request('GET', '/api/v1/users/me', { token: inst })
  const approved = await request('PATCH', `/api/v1/users/${instUser.id}/institutional-approval`, {
    token: admin,
    body: { approved: true },
  })
  if (approved.institutional_approved !== true) {
    throw new Error(`La aprobación de ${INSTITUTIONAL.email} no quedó registrada en la API.`)
  }
  console.log(`  ✓ ${INSTITUTIONAL.email} aprobada para el portal institucional`)

  // Guarda contra duplicados: la organización ancla ya existe.
  const mine = await request('GET', '/api/v1/organizations/mine', { token: inst })
  if ((mine.items || []).some((m) => (m.organization?.name ?? m.name) === ANCHOR_ORG)) {
    console.log(`\n✗ La organización "${ANCHOR_ORG}" ya existe para ${INSTITUTIONAL.email}.`)
    console.log('  La semilla web ya fue aplicada; no se crean duplicados.')
    process.exitCode = 2
    return
  }

  // 2. Categorías de servicio (para enlazar servicios y actividades)
  const categories = (await request('GET', '/api/v1/service-categories', { token: inst })).items
  const cat = (slug) => categories.find((c) => c.slug === slug)?.id

  // 3. Organizaciones
  console.log('\nOrganizaciones:')
  const orgs = {}
  const orgDefs = [
    { key: 'cultural', name: ANCHOR_ORG, type: 'educational', comuna: 'Concepción', region: 'Biobío', latitude: -36.8270, longitude: -73.0503, address: 'Av. Chacabuco 1120', phone: '+56412554400', email: 'hola@culturalbiobio.cl', legalId: '65.221.480-3', description: 'Corporación dedicada a talleres artísticos, culturales y de memoria para personas mayores.' },
    { key: 'salud', name: 'Cesfam Lorenzo Arenas', type: 'health', comuna: 'Concepción', region: 'Biobío', latitude: -36.8185, longitude: -73.0620, address: 'Los Carrera 2950', phone: '+56412554411', email: 'contacto@cesfamlorenzo.cl', legalId: '', description: 'Centro de salud familiar con programas preventivos para personas mayores.' },
    { key: 'vecinos', name: 'Junta de Vecinos Barrio Norte', type: 'community', comuna: 'Talcahuano', region: 'Biobío', latitude: -36.7220, longitude: -73.1210, address: 'Pasaje Los Aromos 45', phone: '+56414115522', email: 'directiva@jjvvbarriotorto.cl', legalId: '', description: 'Organización vecinal de Talcahuano: olla común, mingas y transporte solidario.' },
  ]
  for (const def of orgDefs) {
    const org = await request('POST', '/api/v1/organizations', {
      token: inst,
      body: {
        name: def.name,
        type: def.type,
        legal_id: def.legalId,
        description: def.description,
        email: def.email,
        phone: def.phone,
        address: def.address,
        comuna: def.comuna,
        region: def.region,
        latitude: def.latitude,
        longitude: def.longitude,
      },
    })
    orgs[def.key] = org
    console.log(`  ✓ ${org.name} (${org.id})`)
  }

  // 4. Integrantes de organización
  console.log('\nIntegrantes:')
  const memberDefs = [
    { org: 'cultural', email: users.luis.email, role: 'editor' },
    { org: 'salud', email: users.elena.email, role: 'viewer' },
    { org: 'vecinos', email: users.carmen.email, role: 'viewer' },
  ]
  for (const def of memberDefs) {
    try {
      await request('POST', `/api/v1/organizations/${orgs[def.org].id}/members`, {
        token: inst,
        body: { email: def.email, role: def.role },
      })
      console.log(`  ✓ ${def.email} → ${orgs[def.org].name} (${def.role})`)
    } catch (error) {
      console.log(`  ! ${def.email} → ${orgs[def.org].name}: ${error.message}`)
    }
  }

  // Promover a un integrante a colaborador institucional (rol de plataforma).
  try {
    const me = await request('GET', '/api/v1/users/me', { token: users.luis.token })
    await request('PATCH', `/api/v1/users/${me.id}/role`, { token: admin, body: { role: 'org_collaborator' } })
    console.log(`  ✓ ${users.luis.email} promovido a org_collaborator`)
  } catch (error) {
    console.log(`  ! no se pudo promover a ${users.luis.email}: ${error.message}`)
  }

  // 5. Servicios con horarios
  console.log('\nServicios:')
  const serviceDefs = [
    { key: 'teatro', org: 'cultural', category: 'actividades', name: 'Escuela de teatro para mayores', type: 'resource', address: 'Av. Chacabuco 1120, Concepción', schedule: [{ day_of_week: 2, opens_at: '16:00', closes_at: '18:00' }, { day_of_week: 4, opens_at: '16:00', closes_at: '18:00' }], description: 'Montaje teatral con vecinos: expresión corporal, memoria y trabajo en grupo.' },
    { key: 'biblioteca', org: 'cultural', category: 'educacion', name: 'Biblioteca y club de lectura', type: 'establishment', address: 'Av. Chacabuco 1120, Concepción', schedule: [{ day_of_week: 1, opens_at: '10:00', closes_at: '17:00' }, { day_of_week: 3, opens_at: '10:00', closes_at: '17:00' }, { day_of_week: 5, opens_at: '10:00', closes_at: '14:00' }], description: 'Préstamo de libros, audiolibros con letra grande y encuentros de lectura.' },
    { key: 'nutricion', org: 'salud', category: 'salud-bienestar', name: 'Consulta de nutrición para personas mayores', type: 'establishment', address: 'Los Carrera 2950, Concepción', schedule: [{ day_of_week: 1, opens_at: '08:30', closes_at: '12:30' }, { day_of_week: 3, opens_at: '08:30', closes_at: '12:30' }], description: 'Evaluación nutricional y plan de alimentación adaptado a patologías crónicas.' },
    { key: 'caidas', org: 'salud', category: 'salud-bienestar', name: 'Programa de prevención de caídas', type: 'resource', address: 'Los Carrera 2950, Concepción', schedule: [{ day_of_week: 4, opens_at: '11:00', closes_at: '12:30' }], description: 'Ejercicios de equilibrio y revisión del hogar para reducir el riesgo de caídas.' },
    { key: 'bodega', org: 'vecinos', category: 'asistencia-social', name: 'Bodega solidaria de alimentos', type: 'establishment', address: 'Pasaje Los Aromos 45, Talcahuano', schedule: [{ day_of_week: 6, opens_at: '10:00', closes_at: '13:00' }], description: 'Entrega de cajas de alimentos a vecinos mayores de la junta de vecinos.' },
    { key: 'transporte', org: 'vecinos', category: 'transporte', name: 'Transporte asistido a controles médicos', type: 'resource', address: 'Pasaje Los Aromos 45, Talcahuano', schedule: [{ day_of_week: 2, opens_at: '08:00', closes_at: '12:00' }], description: 'Traslado acompañado desde el domicilio al centro de salud y de regreso.' },
  ]
  const services = {}
  for (const def of serviceDefs) {
    const org = orgs[def.org]
    const service = await request('POST', '/api/v1/services', {
      token: inst,
      body: {
        organization_id: org.id,
        category_id: cat(def.category),
        name: def.name,
        description: def.description,
        type: def.type,
        status: 'active',
        address: def.address,
        comuna: org.comuna,
        region: org.region,
        latitude: org.latitude,
        longitude: org.longitude,
        phone: org.phone,
        schedule: def.schedule,
      },
    })
    services[def.key] = service
    console.log(`  ✓ ${service.name}`)
  }

  // 6. Actividades
  console.log('\nActividades:')
  const activityDefs = [
    { key: 'obra', org: 'cultural', service: 'teatro', category: 'actividades', title: 'Obra de teatro: Voces del barrio', type: 'event', modality: 'in_person', status: 'published', capacity: 60, price: 0, start: daysFromNow(10, 18, 30), end: daysFromNow(10, 20, 0), location: 'Teatro Municipal de Concepción', description: 'Estreno del montaje del taller de teatro. Entrada liberada con inscripción previa.' },
    { key: 'lectura', org: 'cultural', service: 'biblioteca', category: 'educacion', title: 'Club de lectura: poesía chilena', type: 'workshop', modality: 'hybrid', status: 'published', capacity: 25, price: 0, start: daysFromNow(6, 11, 0), end: daysFromNow(6, 12, 30), location: 'Biblioteca Vecinal, Av. Chacabuco 1120', description: 'Comentario de poemas de Gabriela Mistral y Pablo Neruda, con lectura en voz alta.' },
    { key: 'nutricion_charla', org: 'salud', service: 'nutricion', category: 'salud-bienestar', title: 'Charla: alimentación y diabetes', type: 'event', modality: 'online', status: 'published', capacity: 100, price: 0, start: daysFromNow(4, 11, 0), end: daysFromNow(4, 12, 0), location: '', description: 'Recomendaciones prácticas de alimentación para personas con diabetes tipo 2.' },
    { key: 'equilibrio', org: 'salud', service: 'caidas', category: 'salud-bienestar', title: 'Taller de equilibrio y fuerza', type: 'workshop', modality: 'in_person', status: 'published', capacity: 2, price: 0, start: daysFromNow(8, 11, 0), end: daysFromNow(8, 12, 30), location: 'Cesfam Lorenzo Arenas, Los Carrera 2950', description: 'Rutina guiada de equilibrio y fuerza para prevenir caídas. Cupos muy limitados.' },
    { key: 'minga', org: 'vecinos', service: 'bodega', category: 'asistencia-social', title: 'Minga de invierno del barrio', type: 'event', modality: 'in_person', status: 'published', capacity: 40, price: 0, start: daysFromNow(12, 10, 0), end: daysFromNow(12, 14, 0), location: 'Sede vecinal, Pasaje Los Aromos 45', description: 'Jornada comunitaria de limpieza, huerto y once compartida.' },
    { key: 'huerto', org: 'vecinos', service: null, category: 'educacion', title: 'Curso de huerto urbano en macetas', type: 'course', modality: 'in_person', status: 'draft', capacity: 20, price: 0, start: daysFromNow(20, 10, 0), end: daysFromNow(20, 12, 0), location: 'Sede vecinal, Pasaje Los Aromos 45', description: 'Borrador: aprender a cultivar hortalizas en espacios pequeños.' },
  ]
  const activities = {}
  for (const def of activityDefs) {
    const org = orgs[def.org]
    const activity = await request('POST', '/api/v1/activities', {
      token: inst,
      body: {
        organization_id: org.id,
        service_id: def.service ? services[def.service].id : undefined,
        category_id: cat(def.category),
        title: def.title,
        description: def.description,
        type: def.type,
        modality: def.modality,
        status: def.status,
        location: def.location,
        comuna: org.comuna,
        region: org.region,
        latitude: def.modality === 'online' ? undefined : org.latitude,
        longitude: def.modality === 'online' ? undefined : org.longitude,
        start_at: def.start,
        end_at: def.end,
        capacity: def.capacity,
        price: def.price,
        currency: 'CLP',
      },
    })
    activities[def.key] = activity
    console.log(`  ✓ ${activity.title} [${activity.status}]`)
  }

  // 7. Inscripciones (cupos: el taller de equilibrio tiene capacidad 2 → lista de espera)
  console.log('\nInscripciones:')
  const enrollmentDefs = [
    { activity: 'obra', user: 'carmen' },
    { activity: 'obra', user: 'luis' },
    { activity: 'obra', user: 'elena' },
    { activity: 'lectura', user: 'carmen' },
    { activity: 'lectura', user: 'elena' },
    { activity: 'nutricion_charla', user: 'luis' },
    { activity: 'equilibrio', user: 'carmen' },
    { activity: 'equilibrio', user: 'luis' },
    { activity: 'equilibrio', user: 'elena' }, // → lista de espera
    { activity: 'minga', user: 'luis' },
  ]
  for (const def of enrollmentDefs) {
    try {
      const res = await request('POST', `/api/v1/activities/${activities[def.activity].id}/enrollments`, { token: users[def.user].token })
      console.log(`  ✓ ${users[def.user].fullName} → ${activities[def.activity].title}${res.waitlisted ? ' (lista de espera)' : ''}`)
    } catch (error) {
      console.log(`  ! ${users[def.user].email} → ${activities[def.activity].title}: ${error.message}`)
    }
  }

  // 8. Oportunidades y postulaciones
  console.log('\nOportunidades:')
  const opportunityDefs = [
    { key: 'monitor', org: 'cultural', title: 'Monitor/a de teatro para personas mayores', type: 'job', status: 'open', salaryMin: 400000, salaryMax: 550000, description: 'Dirigir el taller de teatro dos tardes por semana y preparar el montaje anual.', contactEmail: 'rrhh@culturalbiobio.cl', expires: daysFromNow(25, 23, 59) },
    { key: 'asistente', org: 'salud', title: 'Asistente administrativo/a part-time', type: 'job', status: 'open', salaryMin: 320000, salaryMax: 380000, description: 'Agenda de horas, orientación a usuarios y apoyo en el programa de personas mayores.', contactEmail: 'contacto@cesfamlorenzo.cl', expires: daysFromNow(18, 23, 59) },
    { key: 'olla', org: 'vecinos', title: 'Voluntariado en la olla común', type: 'volunteer', status: 'open', salaryMin: null, salaryMax: null, description: 'Apoyo en cocina y reparto de almuerzos los sábados. No requiere experiencia.', contactEmail: 'directiva@jjvvbarriotorto.cl', expires: daysFromNow(30, 23, 59) },
    { key: 'manualidades', org: 'cultural', title: 'Tallerista de manualidades (honorarios)', type: 'economic', status: 'open', salaryMin: 12000, salaryMax: 18000, description: 'Dictar un taller semanal remunerado de manualidades para personas mayores.', contactEmail: 'hola@culturalbiobio.cl', expires: daysFromNow(22, 23, 59) },
  ]
  const opportunities = {}
  for (const def of opportunityDefs) {
    const org = orgs[def.org]
    const opportunity = await request('POST', '/api/v1/opportunities', {
      token: inst,
      body: {
        organization_id: org.id,
        title: def.title,
        description: def.description,
        type: def.type,
        status: def.status,
        comuna: org.comuna,
        region: org.region,
        salary_min: def.salaryMin,
        salary_max: def.salaryMax,
        currency: 'CLP',
        contact_email: def.contactEmail,
        expires_at: def.expires,
      },
    })
    opportunities[def.key] = opportunity
    console.log(`  ✓ ${opportunity.title}`)
  }

  console.log('\nPostulaciones:')
  const applicationDefs = [
    { opportunity: 'monitor', user: 'carmen', message: 'Tengo 12 años de experiencia en teatro comunitario y disponibilidad en las tardes.' },
    { opportunity: 'asistente', user: 'luis', message: 'Trabajé en atención de público en una notaría y manejo agenda digital.' },
    { opportunity: 'olla', user: 'elena', message: 'Vivo a dos cuadras de la sede y puedo ayudar los sábados por la mañana.' },
    { opportunity: 'monitor', user: 'elena', message: 'Puedo apoyar con vestuario y utilería.' },
  ]
  const applications = []
  for (const def of applicationDefs) {
    const app = await request('POST', `/api/v1/opportunities/${opportunities[def.opportunity].id}/applications`, {
      token: users[def.user].token,
      body: { message: def.message },
    })
    applications.push({ ...def, id: app.id })
    console.log(`  ✓ ${users[def.user].fullName} → ${opportunities[def.opportunity].title}`)
  }

  // Resolver algunas postulaciones: genera avisos en la app (notifications).
  console.log('\nResolución de postulaciones:')
  const decisions = [
    { opportunity: 'monitor', user: 'carmen', status: 'accepted' },
    { opportunity: 'monitor', user: 'elena', status: 'rejected' },
  ]
  for (const def of decisions) {
    const app = applications.find((a) => a.opportunity === def.opportunity && a.user === def.user)
    if (!app) continue
    await request('PATCH', `/api/v1/opportunities/${opportunities[def.opportunity].id}/applications/${app.id}`, {
      token: inst,
      body: { status: def.status },
    })
    console.log(`  ✓ ${users[def.user].fullName} → ${def.status}`)
  }

  // 9. Preferencias, dispositivos y asistencia (datos de personas mayores)
  console.log('\nPreferencias, dispositivos y asistencia:')
  await request('PUT', '/api/v1/users/me/preferences', {
    token: users.carmen.token,
    body: {
      interests: ['cultura', 'actividades', 'salud'],
      accessibility: { large_text: true, high_contrast: true, reduce_motion: false, screen_reader: false },
      notifications: { courses: true, assistance: true, opportunities: true, activity_reminders: true },
    },
  })
  console.log('  ✓ preferencias de Carmen Rojas')

  await request('POST', '/api/v1/devices', {
    token: users.carmen.token,
    body: { platform: 'android', token: 'demo-android-token-carmen-rojas', device_name: 'Celular de Carmen', app_version: '1.1.0' },
  })
  console.log('  ✓ dispositivo de Carmen Rojas')

  const contact = await request('POST', '/api/v1/assistance/contacts', {
    token: users.carmen.token,
    body: { name: 'Daniela Rojas', relationship: 'hija', phone: '+56955556677', email: 'daniela.rojas@ejemplo.cl', priority: 1 },
  })
  console.log(`  ✓ contacto de confianza: ${contact.name}`)

  const helpRequest = await request('POST', '/api/v1/assistance/requests', {
    token: users.carmen.token,
    body: {
      type: 'medical',
      description: 'Necesito acompañamiento para ir a un control al hospital el viernes a las 9:00.',
      latitude: orgs.cultural.latitude,
      longitude: orgs.cultural.longitude,
      contact_id: contact.id,
    },
  })
  console.log(`  ✓ solicitud de asistencia (${helpRequest.status})`)

  const resolved = await request('POST', '/api/v1/assistance/requests', {
    token: users.luis.token,
    body: { type: 'home', description: 'Se me trabó la puerta de entrada y no puedo abrirla desde adentro.' },
  })
  await request('PATCH', `/api/v1/assistance/requests/${resolved.id}/status`, {
    token: users.luis.token,
    body: { status: 'resolved' },
  })
  console.log('  ✓ solicitud de asistencia resuelta (Luis Vera)')

  console.log('\nListo. Datos del portal web creados.')
  console.log(`  Administrador: ${ADMIN_EMAIL} / ${ADMIN_PASSWORD}`)
  console.log(`  Cuenta institucional: ${INSTITUTIONAL.email} / ${INSTITUTIONAL.password}`)
  console.log(`  Personas mayores: ${MOBILE_USERS.map((u) => u.email).join(', ')} / Demo1234`)
  console.log('  Cuenta pendiente de aprobación: pendiente.institucional@abudemy.cl / Demo1234')
}

main().catch((error) => {
  console.error('\n✗ La semilla web falló:', error.message)
  console.error(`  Verifica que el backend esté en marcha (por defecto ${API}).`)
  process.exitCode = 1
})
