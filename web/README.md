# Abudemy — Plataforma web institucional

Aplicación web (React + TypeScript) que consume la API REST del backend de
Abudemy sin modificarlo. Está pensada para **organizaciones, municipalidades,
centros de salud y administradores**: consulta de indicadores, gestión de
servicios y actividades, mapa territorial, oportunidades, informes y
administración de cuentas.

> La web **nunca** se conecta directamente a PostgreSQL. Toda lectura y
> escritura pasa por la API REST versionada (`/api/v1/...`) usando el mismo
> contrato que la aplicación móvil. La autorización real la aplica siempre el
> backend; la interfaz solo oculta acciones según el rol del usuario.

---

## Stack

| Concern            | Elección                                             |
| ------------------ | ---------------------------------------------------- |
| Framework          | React 19 + TypeScript 5.7                            |
| Build              | Vite 7                                               |
| Enrutado           | React Router 7                                       |
| Estado de datos    | TanStack Query 5                                     |
| Gráficos           | Recharts                                             |
| Mapas              | Leaflet + React Leaflet (+ leaflet.heat)             |
| PDF (cliente)      | jsPDF + jspdf-autotable                              |
| Pruebas            | Vitest                                               |
| Estilos            | Sistema de diseño propio en `src/styles.css`         |

### Requisitos

- **Node.js 20+** (probado con Node 24) y npm.
- El **backend de Abudemy** en marcha (por defecto `http://localhost:8080`).
  Consulta [`backend/README.md`](../backend/README.md) y el
  [`README.md`](../README.md) de la raíz para levantarlo con Docker Compose.

---

## Puesta en marcha

```bash
cd web
npm install
cp .env.example .env      # ajusta los valores si es necesario
npm run dev               # http://localhost:5173
```

El servidor de desarrollo **proxea** `/api`, `/healthz` y `/readyz` hacia
`VITE_PROXY_TARGET` (por defecto `http://localhost:8080`), de modo que el
navegador habla con el mismo origen y no hay problemas de CORS en desarrollo.

### Scripts disponibles

| Comando              | Descripción                                                        |
| -------------------- | ------------------------------------------------------------------ |
| `npm run dev`        | Servidor de desarrollo con recarga en caliente.                    |
| `npm run build`      | Verifica tipos (`tsc --noEmit`) y genera `dist/`.                  |
| `npm run preview`    | Sirve el build de producción localmente.                           |
| `npm run typecheck`  | Solo verificación de tipos.                                        |
| `npm test`           | Ejecuta las pruebas unitarias (Vitest).                            |
| `npm run seed:demo`  | Crea datos de ejemplo vía la API real (ver más abajo).             |
| `npm run seed:web`   | Crea el dataset del portal vía la API real (ver más abajo).        |

---

## Configuración

Variables en `web/.env` (ver `.env.example`). Solo las que empiezan con `VITE_`
llegan al navegador.

| Variable             | Por defecto             | Descripción                                                             |
| -------------------- | ----------------------- | ---------------------------------------------------------------------- |
| `VITE_API_BASE_URL`  | *(vacío)*               | URL base de la API. Vacío = mismo origen (proxy en desarrollo).         |
| `VITE_DEMO_MODE`     | `false`                 | `true` activa datos ficticios `[DEMO]` con aviso permanente.            |
| `VITE_PROXY_TARGET`  | `http://localhost:8080` | Destino del proxy de desarrollo para `/api`, `/healthz` y `/readyz`.    |

---

## Modo demostración

Con `VITE_DEMO_MODE=true` todas las respuestas provienen de un adaptador en
memoria (`src/api/demo.ts`). Cada registro lleva el prefijo `[DEMO]` y la
interfaz muestra **un aviso permanente** indicando que los datos son ficticios.
El modo demostración **nunca** se mezcla con datos reales: es una capa de
transporte alternativa, no un "mock" dentro de los componentes.

El adaptador incluye personas con perfil sociodemográfico completo (comuna,
previsión de salud, sexo y convivencia), una cuenta institucional pendiente y
un administrador, de modo que **Personas registradas**, **Usuarios** y
**Cuentas institucionales** funcionan con filtros, indicadores y acciones de
aprobación sin backend. La sesión demo entra como **administrador**, así que
todas las secciones del portal son navegables.

Por defecto (`false`) la aplicación usa el backend real.

---

## Arquitectura

```
src/
├── api/           Capa de servicio HTTP (única responsable del transporte)
│   ├── client.ts     ApiClient: fetch, timeout, refresh single-flight, errores
│   ├── endpoints.ts  Api: métodos tipados por recurso
│   ├── types.ts      Espejo TypeScript del contrato OpenAPI
│   ├── errors.ts     ApiError + humanizeError (envelope del backend)
│   ├── tokenStore.ts Persistencia de tokens + id de sesión
│   └── demo.ts       Adaptador del modo demostración
├── auth/          Sesión (AuthProvider) y modelo de permisos
├── app/           Ámbito de organización (OrgScope) y guards de ruta
├── hooks/         Hooks de TanStack Query (lecturas y mutaciones)
├── components/    UI reutilizable (primitivas, tablas, overlays, mapas, forms)
├── pages/         Vistas: public, auth y app
├── lib/           Utilidades: formatos, diccionarios, CSV, PDF, geo, estados
└── styles.css     Sistema de diseño (tokens, componentes, responsive)
```

**Regla de oro:** los componentes nunca construyen URLs ni concatenan SQL. Piden
datos a los hooks, que llaman a `Api` (`src/api/endpoints.ts`), que usa el
`ApiClient`. Esto centraliza autenticación, paginación, manejo de errores y
refresco de tokens en un único lugar.

### Cliente HTTP

- Inyecta el `Bearer` desde `tokenStore`.
- **Refresco single-flight**: ante un `401`, rota el refresh token una sola vez
  (aunque haya varias peticiones en paralelo) y repite la solicitud.
- Normaliza el envelope de error `{ error: { code, message, fields } }`.
- Timeout de 20 s y cancelación por `AbortSignal`.
- Si el refresh falla, notifica la expiración y la sesión local se cierra.

---

## Autenticación y permisos

- **Roles de plataforma:** `mobile_user`, `institutional_client`,
  `org_collaborator`, `admin`. Los últimos tres acceden al portal institucional,
  pero una cuenta `institutional_client` recién registrada queda **pendiente de
  aprobación** hasta que un administrador la active.
- **Roles de organización:** `owner`, `admin`, `editor` (escritura) y `viewer`
  (solo lectura). `owner`/`admin` además gestionan integrantes.
- Los recursos **sin organización** son de plataforma y quedan **solo para
  administradores**.
- `src/app/guards.tsx` protege las rutas: `RequireAuth`, `RequireInstitutional`
  y `RequireAdmin`. `src/auth/permissions.ts` decide qué acciones mostrar.
- El **ámbito de organización** activo (`OrgScope`) se elige en la barra
  superior y se recuerda en `localStorage`; alimenta los filtros de analítica y
  exportaciones.

---

## Secciones del portal

- **Panel**: indicadores del endpoint de analítica, gráficos y resumen
  territorial exportable a PDF.
- **Mapa territorial**: servicios, actividades y organizaciones con coordenadas
  (no incluye solicitudes de asistencia ni datos personales).
- **Servicios** y **Actividades**: listados con filtros, paginación, alta/edición
  y control de estado. Los servicios muestran la organización que los provee; las
  actividades incluyen la gestión de inscripciones.
- **Solicitudes de asistencia**: lectura y gestión de estado (los
  administradores ven todas; el resto, solo las propias).
- **Oportunidades**: empleos, trabajos puntuales, voluntariado y oportunidades
  económicas.
- **Informes**: la vista previa cambia según el **tipo de informe**; indicadores
  por ámbito/período, exportación **CSV** (backend) y **PDF** (cliente).
- **Organizaciones**: listado, ficha e integrantes con sus roles. Solo los
  administradores y los clientes institucionales aprobados pueden crear una.
- **Mi cuenta**: datos personales, cambio de contraseña, sesiones y seguridad.
- **Usuarios** (solo administradores): roles y estados.
- **Cuentas institucionales** (solo administradores): aprobar o revocar el
  acceso de las cuentas `institutional_client`.

---

## Pruebas

```bash
npm test
```

Pruebas unitarias de lógica pura (sin navegador): modelo de permisos, formatos y
exportación CSV.

---

## Semilla de datos de desarrollo

`npm run seed:demo` crea —usando la API real— una cuenta institucional, una
organización, un servicio y una actividad:

```bash
SEED_EMAIL=tu@correo.cl SEED_PASSWORD=tu-clave npm run seed:demo
```

La cuenta institucional se registra sin aprobar (flujo de 0009), así que el
script pide a un administrador que la apruebe antes de crear recursos. Por
defecto usa `admin@abudemy.cl` / `Admin1234`; puedes cambiarlo con
`SEED_ADMIN_EMAIL` y `SEED_ADMIN_PASSWORD`.

`npm run seed:web` crea un dataset más completo, pensado para recorrer el portal
de punta a punta: administrador, cuenta institucional, personas mayores con
perfil sociodemográfico, tres organizaciones con integrantes, servicios con
horarios, actividades con inscripciones (incluida lista de espera),
oportunidades con postulaciones, avisos y solicitudes de asistencia.

```bash
SEED_API_URL=http://localhost:8080 npm run seed:web
```

Cuentas creadas (desarrollo):

| Cuenta                            | Contraseña     | Rol                                  |
| --------------------------------- | -------------- | ------------------------------------ |
| `admin@abudemy.cl`                | `Admin1234`    | `admin` (aprueba instituciones)      |
| `web.institucional@abudemy.cl`    | `WebDemo1234`  | `institutional_client` (owner)       |
| `pendiente.institucional@abudemy.cl` | `Demo1234`  | `institutional_client` (pendiente)   |
| `carmen.rojas@ejemplo.cl`         | `Demo1234`     | `mobile_user`                        |
| `luis.vera@ejemplo.cl`            | `Demo1234`     | `org_collaborator` (editor)          |
| `elena.torres@ejemplo.cl`         | `Demo1234`     | `mobile_user`                        |
| `patricia.nunez@ejemplo.cl`       | `Demo1234`     | `mobile_user`                        |
| `hugo.fuentes@ejemplo.cl`         | `Demo1234`     | `mobile_user`                        |
| `teresa.aguilera@ejemplo.cl`      | `Demo1234`     | `mobile_user`                        |
| `raul.paredes@ejemplo.cl`         | `Demo1234`     | `mobile_user`                        |
| `silvia.castro@ejemplo.cl`        | `Demo1234`     | `mobile_user`                        |

Ambas semillas son solo para desarrollo: cada ejecución crea recursos nuevos.
El mismo dataset del portal se carga también con la migración
[`0008_seed_web_demo_data.sql`](../backend/migrations/0008_seed_web_demo_data.sql)
y la migración [`0011_demo_data_refresh.sql`](../backend/migrations/0011_demo_data_refresh.sql)
lo pone al día (perfiles, aprobación institucional y fechas futuras).

---

## Dependencias pendientes del backend

La web se construyó sin inventar endpoints ni simular resultados. Estas
capacidades **no existen todavía** en la API y se documentan aquí como
dependencias abiertas:

| Necesidad                              | Estado en el backend                         | Cómo lo aborda la web                                                        |
| -------------------------------------- | -------------------------------------------- | ---------------------------------------------------------------------------- |
| Exportación de informes en PDF         | `format=pdf` responde `501 Not Implemented`  | El PDF se compone en el navegador con los datos consultados (se indica).     |
| Series temporales / evolución          | No hay endpoint de series                    | La serie mensual se calcula en el cliente desde `start_at` (hasta 5 páginas).|
| Filtros de fecha/comuna en analítica   | `/analytics/overview` solo acepta `organization_id` | Los filtros viven en Servicios/Actividades/Informes, donde sí se soportan.   |
| Formulario público de contacto         | No hay endpoint público                      | Se abre el cliente de correo (`mailto:`) y se explica en la interfaz.        |
| Cambiar el correo electrónico          | No hay endpoint                              | El campo se muestra de solo lectura.                                         |
| Revisión editorial de oportunidades    | No hay endpoint de revisión                  | Se usa el estado `draft` / `open` / `closed`.                                |
| Geocodificación de direcciones         | No disponible                                | Las coordenadas se ingresan manualmente.                                     |
| Exportar sin organización (>1 org)     | El backend lo rechaza para no-admin          | La interfaz pide elegir una organización antes de exportar.                  |
| Contactos y solicitudes de asistencia  | Sin endpoints de escritura para la web       | Solo lectura y cambio de estado (administradores).                           |

---

## Despliegue

- Genera el build con `npm run build` y sirve `dist/` como sitio estático.
- Configura **fallback de SPA** hacia `index.html` (React Router).
- Define `VITE_API_BASE_URL` con la URL absoluta de la API y añade el origen de
  la web a `CORS_ALLOWED_ORIGINS` del backend.
- No expongas credenciales en el frontend: no hay secretos en el cliente.

---

## Seguridad

- El transporte añade el `Bearer` y rota tokens; la autorización es del backend.
- Los tokens se guardan en `localStorage` (access + refresh rotativo). En
  despliegues con requisitos estrictos puede moverse el access token a memoria.
- Nunca se muestran datos personales en los indicadores agregados ni en el mapa.
