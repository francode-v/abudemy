# Abudemy

Plataforma de **Abudemy**: conecta a personas mayores con servicios de salud,
asistencia, actividades comunitarias y cursos, y oportunidades para compartir
conocimiento y generar ingresos.

El repositorio está organizado como **monorepo** con tres partes independientes:

| Carpeta    | Descripción                                                        | Stack                              |
| ---------- | ------------------------------------------------------------------ | ---------------------------------- |
| `backend/` | API REST única, compartida por la app móvil y la web institucional | Go 1.27 + PostgreSQL 16            |
| `web/`     | Plataforma web institucional                                       | React 19 + TypeScript + Vite       |
| `mobile/`  | Aplicación móvil para personas mayores                             | Expo SDK 57 + React Native + TS    |

Los clientes consumen la misma API JSON (`/api/v1/...`) y la misma base de
datos PostgreSQL. **Ningún cliente se conecta directamente a la base de datos:**
todo pasa por la API.

```
Session/
├── backend/            # API REST en Go (docs, migraciones, Dockerfile, Makefile)
│   ├── cmd/api/
│   ├── internal/
│   ├── migrations/
│   ├── docs/openapi.yaml
│   └── README.md
├── web/                # Plataforma web institucional (React + TypeScript)
│   └── README.md
├── mobile/             # App móvil (Expo + React Native + TypeScript)
│   └── README.md
├── docker-compose.yml  # Orquestación local: PostgreSQL + API
└── .env.example        # Variables del stack de Docker Compose
```

## Puesta en marcha rápida

### 1. Backend (API + base de datos)

Desde la **raíz del repositorio**:

```bash
cp .env.example .env          # opcional; Docker Compose trae valores por defecto
docker compose up --build     # PostgreSQL + API en http://localhost:8080
```

¿Prefieres ejecutar la API sin Docker contra tu propio PostgreSQL? Sigue las
instrucciones de [`backend/README.md`](backend/README.md).

### 2. Web institucional

```bash
cd web
npm install
cp .env.example .env
npm run dev                   # http://localhost:5173 (proxea /api al backend)
```

### 3. Aplicación móvil

```bash
cd mobile
npm install
cp .env.example .env          # ajusta EXPO_PUBLIC_API_BASE_URL si usas un teléfono físico
npm start                     # abre Expo; escanea el QR con Expo Go
```

¿Pruebas en un teléfono físico? Usa la IP de tu equipo en
`EXPO_PUBLIC_API_BASE_URL` (por ejemplo `http://192.168.0.10:8080`) en lugar de
`localhost`.

## Documentación

- **Backend**: [`backend/README.md`](backend/README.md) — arquitectura, configuración, endpoints y pruebas.
- **Contrato de la API**: [`backend/docs/openapi.yaml`](backend/docs/openapi.yaml).
- **Web**: [`web/README.md`](web/README.md) — puesta en marcha, arquitectura y dependencias pendientes del backend.
- **Móvil**: [`mobile/README.md`](mobile/README.md) — puesta en marcha, accesibilidad y dependencias pendientes del backend.

## Pruebas

```bash
# Backend (Go)
cd backend
make test                 # unitarias (sin base de datos)
make test-integration     # integración (requiere TEST_DATABASE_URL)

# Web (Vitest)
cd web
npm test

# Móvil (Jest)
cd mobile
npm test
```
