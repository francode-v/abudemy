#!/usr/bin/env node
/**
 * Semilla de datos de desarrollo para la plataforma web de Abudemy.
 *
 * Crea —a través de la API REST real— una cuenta institucional, una
 * organización, un servicio y una actividad. No escribe directamente en
 * PostgreSQL: usa exactamente los mismos endpoints que la interfaz.
 *
 * Uso:
 *   SEED_EMAIL=tu@correo.cl SEED_PASSWORD=secreto npm run seed:demo
 *
 * Variables:
 *   VITE_API_BASE_URL / SEED_API_URL  URL base de la API (por defecto http://localhost:8080)
 *   SEED_EMAIL                        correo de la cuenta (se registra si no existe)
 *   SEED_PASSWORD                     contraseña (mínimo 8 caracteres)
 *   SEED_NAME                         nombre completo de la cuenta
 *   SEED_ADMIN_EMAIL / SEED_ADMIN_PASSWORD  administrador que aprueba la cuenta
 *                                     institucional (por defecto admin@abudemy.cl /
 *                                     Admin1234, creado por la migración 0011)
 *
 * Desde 0009 las cuentas institucionales necesitan aprobación de un
 * administrador antes de poder crear organizaciones. El script la solicita
 * automáticamente si detecta que la cuenta está pendiente.
 *
 * Nota: cada ejecución crea recursos nuevos. Está pensado para entornos de
 * desarrollo, nunca para producción.
 */

const API = (process.env.VITE_API_BASE_URL || process.env.SEED_API_URL || 'http://localhost:8080').replace(/\/+$/, '')
const EMAIL = process.env.SEED_EMAIL || 'institucional@abudemy.cl'
const PASSWORD = process.env.SEED_PASSWORD || 'Demo1234'
const FULL_NAME = process.env.SEED_NAME || 'Cuenta institucional de ejemplo'
const ADMIN_EMAIL = process.env.SEED_ADMIN_EMAIL || 'admin@abudemy.cl'
const ADMIN_PASSWORD = process.env.SEED_ADMIN_PASSWORD || 'Admin1234'

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

async function ensureSession() {
  try {
    const pair = await request('POST', '/api/v1/auth/login', { body: { email: EMAIL, password: PASSWORD } })
    console.log(`✓ Sesión iniciada como ${EMAIL}`)
    return pair.access_token
  } catch (error) {
    if (error.status !== 401 && error.status !== 400) throw error
    console.log(`… La cuenta ${EMAIL} no existe o la contraseña no coincide. Registrando…`)
    await request('POST', '/api/v1/auth/register', {
      body: { email: EMAIL, password: PASSWORD, full_name: FULL_NAME, role: 'institutional_client' },
    })
    const pair = await request('POST', '/api/v1/auth/login', { body: { email: EMAIL, password: PASSWORD } })
    console.log(`✓ Cuenta registrada e iniciada como ${EMAIL}`)
    return pair.access_token
  }
}

/** Inicia sesión como administrador para aprobar la cuenta institucional (0009). */
async function signInAdmin() {
  try {
    const pair = await request('POST', '/api/v1/auth/login', {
      body: { email: ADMIN_EMAIL, password: ADMIN_PASSWORD },
    })
    return pair.access_token
  } catch (error) {
    throw new Error(
      `La cuenta institucional está pendiente de aprobación y no se pudo entrar como administrador (${ADMIN_EMAIL}).\n` +
        '  La cuenta la crea la migración 0011_demo_data_refresh.sql al arrancar la API.\n' +
        '  Si usas otro administrador, define SEED_ADMIN_EMAIL y SEED_ADMIN_PASSWORD.\n' +
        `  Detalle: ${error.message}`,
    )
  }
}

/**
 * Aprueba la cuenta si está pendiente. Sin la aprobación la API devuelve
 * organizaciones vacías y rechaza crear organizaciones.
 */
async function ensureInstitutionalApproval(token) {
  const me = await request('GET', '/api/v1/users/me', { token })
  if (me.role !== 'institutional_client' || me.institutional_approved) return
  const adminToken = await signInAdmin()
  const updated = await request('PATCH', `/api/v1/users/${me.id}/institutional-approval`, {
    token: adminToken,
    body: { approved: true },
  })
  if (updated.institutional_approved !== true) {
    throw new Error(`La aprobación de ${EMAIL} no quedó registrada en la API.`)
  }
  console.log(`✓ Cuenta institucional aprobada por ${ADMIN_EMAIL}`)
}

async function main() {
  console.log(`Sembrando datos de desarrollo en ${API}\n`)

  const token = await ensureSession()
  await ensureInstitutionalApproval(token)

  const organization = await request('POST', '/api/v1/organizations', {
    token,
    body: {
      name: 'Organización de ejemplo (semilla)',
      type: 'community',
      comuna: 'Santiago',
      region: 'Metropolitana',
      latitude: -33.4489,
      longitude: -70.6693,
      email: EMAIL,
      description: 'Organización creada por el script de semilla para desarrollo.',
    },
  })
  console.log(`✓ Organización creada: ${organization.name} (${organization.id})`)

  const service = await request('POST', '/api/v1/services', {
    token,
    body: {
      organization_id: organization.id,
      name: 'Servicio de ejemplo (semilla)',
      type: 'establishment',
      status: 'active',
      comuna: 'Santiago',
      region: 'Metropolitana',
      address: 'Calle de ejemplo 123',
      latitude: -33.4489,
      longitude: -70.6693,
      schedule: [{ day_of_week: 1, opens_at: '09:00', closes_at: '13:00' }],
    },
  })
  console.log(`✓ Servicio creado: ${service.name} (${service.id})`)

  const start = new Date(Date.now() + 7 * 86_400_000)
  const activity = await request('POST', '/api/v1/activities', {
    token,
    body: {
      organization_id: organization.id,
      service_id: service.id,
      title: 'Actividad de ejemplo (semilla)',
      type: 'workshop',
      modality: 'in_person',
      status: 'published',
      comuna: 'Santiago',
      region: 'Metropolitana',
      capacity: 20,
      start_at: start.toISOString(),
      end_at: new Date(start.getTime() + 2 * 3_600_000).toISOString(),
    },
  })
  console.log(`✓ Actividad creada: ${activity.title} (${activity.id})`)

  console.log('\nListo. Inicia la web con `npm run dev` y entra con la cuenta indicada.')
}

main().catch((error) => {
  console.error('\n✗ La semilla falló:', error.message)
  console.error('  Verifica que el backend esté en marcha (por defecto http://localhost:8080).')
  process.exitCode = 1
})
