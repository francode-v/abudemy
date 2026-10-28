# Despliegue: Neon + Render + Cloudflare Pages

```
Navegador ──► Cloudflare Pages (web/, React estático)
                  │  fetch con Bearer token (VITE_API_BASE_URL)
                  ▼
              Render (backend/, Go en Docker)  ──►  Neon (PostgreSQL)
```

Orden: **1) Neon → 2) Render → 3) Cloudflare Pages → 4) volver a Render para el CORS.**
Antes, sube el proyecto a GitHub/GitLab (los `.env` están en `.gitignore`; no los subas).

## 1. Neon (base de datos)

1. Crea un proyecto en https://neon.tech. Elige una región AWS cercana a la de Render
   (Render *Virginia* ↔ Neon `us-east-1`; Render *Ohio* ↔ Neon `us-east-2`).
2. En **Connect** copia la cadena de conexión. Desmarca *Connection pooling* para obtener la
   conexión directa (el backend ya tiene su propio pool). Debe terminar en `?sslmode=require`:
   `postgres://USUARIO:CLAVE@ep-xxxx.us-east-1.aws.neon.tech/neondb?sslmode=require`
3. No hay que crear tablas: el backend aplica las migraciones solo al arrancar.

## 2. Render (backend)

1. **New → Blueprint** y elige el repositorio. Render lee `render.yaml`.
   (Alternativa manual: *New → Web Service*, Runtime **Docker**, Root Directory `backend`,
   Health Check Path `/healthz`.)
2. Completa las variables que `render.yaml` deja en blanco:

| Variable | Valor |
|---|---|
| `DATABASE_URL` | la cadena de Neon del paso 1 |
| `CORS_ALLOWED_ORIGINS` | por ahora `*`; se ajusta en el paso 4 |
| `SEED_ADMIN_EMAIL` | tu correo de administrador |
| `SEED_ADMIN_PASSWORD` | una clave nueva y fuerte (mín. según la política del backend) |

   `JWT_SECRET` lo genera Render solo. `APP_ENV=production` ya viene puesto.
3. Cuando termine el deploy abre `https://TU-SERVICIO.onrender.com/healthz` → debe responder OK.
   `/readyz` además comprueba la base de datos.

## 3. Cloudflare Pages (web)

1. **Workers & Pages → Create → Pages → Connect to Git**, elige el repositorio.
2. Configuración de build:
   - **Root directory:** `web`
   - **Build command:** `npm run build`
   - **Build output directory:** `dist`
   - **Variables de entorno (Production y Preview):**
     - `VITE_API_BASE_URL` = `https://TU-SERVICIO.onrender.com` (sin `/` final)
     - `VITE_DEMO_MODE` = `false`
   - El Node se fija con `web/.node-version` (22).
3. Deploy. Tu sitio queda en `https://TU-PROYECTO.pages.dev`.

> `VITE_*` se incrusta **al compilar**: si cambias la URL del backend, hay que volver a desplegar la web.

## 4. Cerrar el CORS

En Render cambia `CORS_ALLOWED_ORIGINS` a tu dominio de Pages (incluye las previews con comodín):

```
https://TU-PROYECTO.pages.dev,https://*.TU-PROYECTO.pages.dev
```

Si más adelante usas dominio propio, agrégalo a la lista (`https://www.tudominio.cl`).

## Cambios que se hicieron al proyecto

- **Backend lee `PORT`** (Render lo inyecta). `HTTP_ADDR` sigue mandando si está definido.
- **Datos demo desactivados fuera de desarrollo.** Las migraciones `0007`, `0008` y `0011`
  crean cuentas con claves públicas (incluido `admin@abudemy.cl / Admin1234`). Ahora solo corren si
  `SEED_DEMO_DATA=true` (por defecto `true` solo con `APP_ENV=development`). Las categorías reales
  (`0006_seed_categories`) siguen aplicándose siempre. Tu administrador se crea con
  `SEED_ADMIN_EMAIL`/`SEED_ADMIN_PASSWORD`.
- **Pool compatible con Neon:** `MinConns=0` (la base puede dormirse), reintentos al conectar
  (cold start de Neon) y `DB_MAX_CONNS` configurable.
- **CORS** acepta comodín de subdominio (`https://*.proyecto.pages.dev`) y expone `X-Request-Id`.
- **Web:** `public/_redirects` (rutas SPA de React Router), `public/_headers` (cache y cabeceras),
  `.node-version` y `engines` en `package.json`.
- Nuevos: `render.yaml`, este `DEPLOY.md`, tests de CORS, PORT y migraciones demo.

## Cosas a tener en cuenta

- **Render gratis se duerme a los 15 min** sin tráfico; la primera petición tarda ~30–60 s.
  Neon gratis también se suspende y despierta en ~1 s. Para uso real, plan *Starter* en Render.
- `BCRYPT_COST=11` en `render.yaml` por la CPU limitada del plan gratis; súbelo a 12 en un plan pago.
- Los `.env` del zip traen valores de desarrollo (`change-me…`, clave de admin de ejemplo).
  No los reutilices en producción.
- Si quieres una demo con datos de ejemplo, pon `SEED_DEMO_DATA=true` y **reinicia**; desactívalo
  después y borra esas cuentas si la URL es pública.
