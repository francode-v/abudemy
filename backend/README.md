# Abudemy — Backend

Backend of **Abudemy**, a platform that connects older adults with health
services, assistance, community activities and courses, and opportunities to
share knowledge and generate income.

The single REST API in this repository is shared by two clients:

- the **mobile application** for older adults, and
- the **institutional web platform** for organizations and administrators.

Both consume the same JSON API and the same PostgreSQL datastore.

---

## Stack

| Concern            | Choice                                             |
| ------------------ | -------------------------------------------------- |
| Language           | Go 1.27                                            |
| Database           | PostgreSQL 16                                       |
| HTTP               | `net/http` (Go 1.22+ routing), JSON                 |
| Auth               | JWT access tokens + rotating refresh tokens         |
| Migrations         | Versioned SQL, embedded and applied at startup      |
| Local development  | Docker Compose                                      |
| API documentation  | OpenAPI 3.0 (`docs/openapi.yaml`)                   |

### Dependencies

Kept intentionally small: `jackc/pgx` (PostgreSQL), `golang-jwt/jwt`,
`google/uuid`, `golang.org/x/crypto` (bcrypt) and `joho/godotenv`.

---

## Architecture

A **modular monolith** with a strict one-way dependency flow:

```
cmd/api            entrypoint: config, DB, migration, HTTP server, shutdown
internal/api       HTTP transport: routes, middleware, decoding, responses
internal/service   business logic, authorization, transactions, validation
internal/store     repository layer: every SQL query lives here
internal/models    domain types shared across layers
internal/auth      password hashing + JWT
internal/notify    notification providers (none | log | webhook)
internal/database  connection pool + migration runner
internal/config    environment configuration
internal/validate  field validation helpers
migrations/        versioned SQL (embedded in the binary)
docs/openapi.yaml  OpenAPI contract
```

- Handlers only decode, delegate and encode. **No business logic in handlers.**
- Services own authorization and orchestration; repositories own SQL.
- All queries are **parameterized**; list queries are paginated.
- Multi-step operations (capacity control, schedule replacement) run in
  **transactions** driven by `context.Context`.

```
cmd/api ─▶ internal/api ─▶ internal/service ─▶ internal/store ─▶ PostgreSQL
                                   │
                                   ├─▶ internal/auth
                                   └─▶ internal/notify
```

---

## Repository layout

This repository is a monorepo with two independent parts:

- **`backend/`** — the Go REST API described in this document.
- **`web/`** — the institutional web platform (React + TypeScript).

The `docker-compose.yml` that starts PostgreSQL and the API lives at the
repository root and builds the API image from `backend/`.

---

## Getting started

### 1. With Docker Compose (recommended)

From the **repository root**:

```bash
cp .env.example .env          # optional; compose ships sensible defaults
docker compose up --build
```

Compose reads the root `.env` for variable overrides. This starts PostgreSQL and
the API. Migrations run automatically on boot (idempotently). The API is
available at `http://localhost:8080`.

### 2. Running locally against your own PostgreSQL

From this **`backend/`** directory:

```bash
cp .env.example .env          # set DATABASE_URL and JWT_SECRET
go run ./cmd/api
```

The server applies pending migrations at startup and then listens on
`HTTP_ADDR` (default `:8080`).

### 3. Useful endpoints

```bash
curl localhost:8080/healthz
curl localhost:8080/readyz
```

---

## Configuration

All configuration is read from the environment (optionally from a local `.env`).
See [`.env.example`](.env.example). No secrets are ever committed.

| Variable               | Required | Default       | Notes                                             |
| ---------------------- | -------- | ------------- | ------------------------------------------------- |
| `APP_ENV`              | no       | `development` | `development`/`staging`/`production`              |
| `HTTP_ADDR`            | no       | `:8080`       | Listen address                                    |
| `DATABASE_URL`         | **yes**  | —             | PostgreSQL DSN                                    |
| `JWT_SECRET`           | **yes**  | —             | ≥ 32 bytes outside development                    |
| `JWT_ISSUER`           | no       | `abudemy`     | Token issuer                                      |
| `JWT_ACCESS_TTL`       | no       | `15m`         | Access token lifetime                             |
| `JWT_REFRESH_TTL`      | no       | `720h`        | Refresh token lifetime                            |
| `PASSWORD_RESET_TTL`   | no       | `1h`          | Reset token lifetime                              |
| `BCRYPT_COST`          | no       | `12`          | 10–16                                             |
| `LOG_LEVEL`            | no       | `info`        | `debug`/`info`/`warn`/`error`                     |
| `NOTIFY_PROVIDER`      | no       | `none`        | `none`/`log`/`webhook`                            |
| `NOTIFY_WEBHOOK_URL`   | cond.    | —             | Required when `NOTIFY_PROVIDER=webhook`           |
| `NOTIFY_WEBHOOK_TOKEN` | no       | —             | Sent as `Authorization: Bearer <token>`           |
| `ASSISTANT_PROVIDER`   | no       | `none`        | `none`/`webhook` (curated KB by default)          |
| `ASSISTANT_WEBHOOK_URL`| cond.    | —             | Required when `ASSISTANT_PROVIDER=webhook`        |
| `ASSISTANT_WEBHOOK_TOKEN` | no    | —             | Sent as `Authorization: Bearer <token>`           |
| `GEOCODE_PROVIDER`     | no       | `none`        | `none`/`nominatim` (501 when `none`)              |
| `GEOCODE_BASE_URL`     | cond.    | Nominatim     | Overridable when `GEOCODE_PROVIDER=nominatim`     |
| `GEOCODE_USER_AGENT`   | cond.    | —             | Required when `GEOCODE_PROVIDER=nominatim`        |
| `PAYMENTS_PROVIDER`    | no       | `none`        | `none` (capability reporting only)                |
| `CORS_ALLOWED_ORIGINS` | no       | `*`           | Comma-separated origins                           |
| `SEED_ADMIN_EMAIL`     | no       | —             | Optional bootstrap admin                          |
| `SEED_ADMIN_PASSWORD`  | no       | —             | Optional bootstrap admin                          |

---

## Authentication & authorization

- **Registration / login** return an access token (JWT, short lived) and a
  refresh token. Refresh tokens are stored **hashed** (SHA-256) and **rotated**
  on every use; reuse of a rotated token is rejected.
- Every authenticated request validates the token signature, expiry and the
  backing **session**. Revoking a session (logout, password change/reset,
  status change) immediately invalidates its access token.
- **Roles:** `mobile_user`, `institutional_client`, `org_collaborator`, `admin`.
  New `institutional_client` accounts start **unapproved**: they can sign in but
  cannot use institutional features until an administrator enables their access.
  The approval flag is re-read from the database on every request, so revoking it
  takes effect immediately.
- **Organization scoping:** write access to a resource requires ownership of the
  target organization (`owner`/`admin`/`editor`). Resources without an
  organization are platform-level and admin-only. Administrators can access
  everything. Only administrators and approved institutional clients may create
  organizations.
- Cross-user and cross-organization access is rejected with `403`, and the
  checks live in the service layer — never in the client.

Passwords are hashed with bcrypt and never returned or logged. Personal data and
coordinates are only exposed to their owner (or administrators).

---

## Feature map (API v1)

All routes are versioned under `/api/v1/`. Full contract in
[`docs/openapi.yaml`](docs/openapi.yaml).

| Area                  | Routes                                                                        |
| --------------------- | ----------------------------------------------------------------------------- |
| Auth                  | `POST /auth/register`, `/login`, `/refresh`, `/logout`, `/logout-all`; `GET /auth/sessions`, `DELETE /auth/sessions/{id}`; `POST /auth/password/forgot`, `/reset` |
| Users                 | `GET/PATCH /users/me`, `POST /users/me/password`, `POST /users/me/email`, `GET/PUT /users/me/preferences`, `GET /users/me/enrollments`, `GET /users/me/applications`, `GET /users/me/reminders`, `GET/POST /users/me/skills`; admin: `GET /users`, `PATCH /users/{id}/status`, `/role`, `/institutional-approval` |
| Services              | `GET/POST /services`, `GET/PUT/DELETE /services/{id}`, `PATCH /services/{id}/status`, `GET /service-categories` |
| Activities            | `GET/POST /activities`, `GET/PUT /activities/{id}`, `PATCH .../status`, `POST .../enrollments`, `GET .../enrollments`, `DELETE .../enrollments/me`, `PATCH .../enrollments/{userId}`, `POST .../reminders`, `DELETE .../reminders/me` |
| Opportunities         | `GET/POST /offerings`, `GET/PUT/PATCH /offerings/{id}`; `GET/POST /opportunities`, `GET/PUT/PATCH /opportunities/{id}`, `GET/POST .../applications`, `PATCH .../applications/{appId}`, `POST .../reports`; `PUT/DELETE /skills/{id}` |
| Assistance            | `GET/POST /assistance/contacts`, `PUT/DELETE /assistance/contacts/{id}`; `GET/POST /assistance/requests`, `GET /assistance/requests/{id}`, `PATCH .../status` |
| Devices & notifications | `GET/POST /devices`, `DELETE /devices/{id}`; `GET /notifications`, `GET /notifications/unread-count`, `POST /notifications/{id}/read`, `POST /notifications/read-all` |
| Assistant             | `GET /assistant/topics`, `GET /assistant/answer` |
| Geo                   | `GET /geo/search`, `GET /geo/reverse` (501 unless a provider is configured) |
| Payments              | `GET /payments/config` (reports whether a gateway is configured) |
| Organizations         | `GET/POST /organizations`, `GET /organizations/mine`, `GET/PUT /organizations/{id}`, `GET/POST .../members`, `PATCH/DELETE .../members/{userId}` |
| Analytics & reports   | `GET /analytics/overview`, `GET /analytics/reports`, `GET /analytics/reports/{type}` (CSV, or bounded JSON preview with `format=json`) |

### Capacity control

Enrolling is transactional: the activity row is locked (`SELECT … FOR UPDATE`),
the current enrollment count is read, and the enrollment is written with status
`enrolled` or `waitlist`. Concurrent requests can never exceed `capacity`.
Duplicate enrollments return `409`.

### Assistance notifications (honest by design)

When a user creates an assistance request, the trusted contacts are notified
through the configured provider. If **no provider is configured**, each attempt
is recorded with status `skipped` and the request is **not** marked as notified.
The API never reports a delivery that did not happen. With `NOTIFY_PROVIDER=webhook`
the attempt is recorded as `sent` only when the provider returns 2xx.

### In-app notifications, devices and push

Domain events (an application accepted/rejected, an assistance request resolved)
write an **in-app notification** for the affected user. The same event attempts a
push to the user's registered device tokens. Each notification row carries the
real delivery state in `push_status`:

- `skipped` — no push provider configured **or** the user has no registered
  device (this takes precedence);
- `sent` — the provider accepted the push (2xx);
- `failed` — the provider rejected it or the request errored.

The feed (`GET /notifications`) and `unread-count` are always truthful; a client
must render `push_status` as-is and never claim a push was delivered when it was
not. Device tokens are stored per user and are never exposed to other users.

### Assistant, geocoding and payments (honest by design)

These areas depend on third-party services and degrade honestly:

- **Assistant** (`ASSISTANT_PROVIDER=none|webhook`) answers from a curated,
  versioned knowledge base and reports its source (`knowledge_base`) and
  `ai: false`. With a webhook provider configured it forwards the question and
  reports `source: provider` / `ai: true`. It never fabricates an AI answer.
- **Geocoding** (`GEOCODE_PROVIDER=none|nominatim`) returns `501 Not Implemented`
  when no provider is configured, instead of inventing coordinates.
- **Payments** (`PAYMENTS_PROVIDER=none`) reports capability only
  (`available: false` with an explanatory message); there is no gateway and no
  charge is ever simulated.

---

## Database

- Normalized schema with foreign keys, `CHECK` constraints, unique constraints
  and indexes (category, comuna, region, status, coordinates, dates).
- UUID primary keys (`gen_random_uuid()`), `timestamptz` timestamps.
- Enum-like columns are `TEXT` + `CHECK` for painless evolution.
- Migrations live in [`migrations/`](migrations) and are embedded into the
  binary (`migrations.FS`). Applied in filename order inside a transaction and
  recorded in `schema_migrations`.

Adding a migration: create `migrations/000N_description.sql`. It is picked up
automatically (a restart applies it).

Initial data: `0006_seed_categories.sql` seeds the service catalogue,
`0007_seed_demo_data.sql` loads the mobile/backend demo dataset and
`0008_seed_web_demo_data.sql` loads the institutional web demo dataset (all
idempotent via fixed ids and `ON CONFLICT ... DO NOTHING`).
`0010_user_profile.sql` adds the socio-demographic profile columns and
`0011_demo_data_refresh.sql` brings the demo dataset up to date (profiles,
admin + pending institutional account, future-dated activities/opportunities);
it only runs when the demo dataset is present. The mobile dataset
can be recreated against a running API with `npm run seed:demo` in `mobile/`;
the web dataset with `npm run seed:web` in `web/`.

---

## Testing

```bash
make test               # unit tests (no database required)
make test-integration   # integration tests (requires TEST_DATABASE_URL)
make cover              # coverage report
```

- **Unit tests** cover password hashing/policy, JWT issuance/verification,
  validation, pagination, roles, configuration and notification providers.
- **Integration tests** (`internal/api/integration_test.go`,
  `internal/api/engagement_integration_test.go`) boot the full HTTP server against
  a real PostgreSQL and exercise registration, login, token rotation,
  organizations/permissions, services, activities + capacity, the assistance flow,
  the engagement endpoints (preferences, email change, my applications,
  reminders, opportunity reports, devices, in-app notifications, assistant, geo
  and payments), analytics, CSV export and schema constraints.

Integration tests require a **dedicated database whose name contains `test`**
(the suite truncates tables) and skip otherwise:

```bash
docker run -d --name abudemy-pg-test -e POSTGRES_PASSWORD=postgres \
  -e POSTGRES_USER=postgres -e POSTGRES_DB=abudemy_test -p 5432:5432 postgres:16-alpine
make test-integration
# or: TEST_DATABASE_URL="postgres://postgres:postgres@localhost:5432/abudemy_test?sslmode=disable" \
#     go test ./internal/api/... -run TestIntegration -v
```

---

## Security notes

- Parameterized queries everywhere; identifiers are never built from user input.
- Passwords (bcrypt) and tokens (SHA-256) are stored hashed; secrets come only
  from the environment and are never logged.
- Consistent error envelope; internal errors never leak details to clients.
- Request body size cap, strict JSON decoding, field validation, pagination caps.
- Structured logging with request IDs; panic recovery; CORS allow-list.
- Access tokens are bound to a revocable session.

---

## Dependencies on external services

These are intentionally **not simulated**:

- **Outbound notifications** (SMS/email) require a provider. Set
  `NOTIFY_PROVIDER=webhook` (and `NOTIFY_WEBHOOK_URL`) or `log` for development.
  Until then, assistance notification attempts are recorded as `skipped`.
- **Password reset delivery** requires the same notification provider. The reset
  token is always stored; delivery depends on the provider. In development,
  `NOTIFY_PROVIDER=log` prints the dispatch to the log.
- **PDF report export** is documented as an optional feature and currently
  returns `501 Not Implemented`; only CSV is implemented.
- **Push notifications** require a push provider. Until one is configured, each
  notification is stored with `push_status = skipped`; the in-app feed still
  works.
- **Assistant** uses a curated knowledge base by default. A real AI provider can
  be plugged in via `ASSISTANT_PROVIDER=webhook` (`ASSISTANT_WEBHOOK_URL`,
  `ASSISTANT_WEBHOOK_TOKEN`); without it the API answers from the knowledge base
  and says so.
- **Geocoding** requires `GEOCODE_PROVIDER=nominatim` (`GEOCODE_BASE_URL`,
  `GEOCODE_USER_AGENT`). Without a provider the endpoints return `501`.
- **Payments** have no gateway (`PAYMENTS_PROVIDER=none`); the capability
  endpoint reports `available: false` rather than simulating a charge.
- **Geospatial search** uses a portable haversine query (no PostGIS). For very
  large datasets a `PostGIS`/`earthdistance` index would be the next step.

---

## Project layout

```
backend/
├── cmd/api/                 # entrypoint
├── docs/openapi.yaml        # OpenAPI contract
├── internal/
│   ├── api/                 # HTTP handlers, router, middleware, integration tests
│   ├── apperr/              # error model
│   ├── auth/                # bcrypt + JWT (+ unit tests)
│   ├── config/              # env config (+ unit tests)
│   ├── database/            # pool + migration runner
│   ├── logging/             # structured logging
│   ├── models/              # domain types (+ unit tests)
│   ├── notify/              # notification providers (+ unit tests)
│   ├── service/             # business logic & authorization
│   ├── store/               # repositories (all SQL)
│   └── validate/            # validation helpers (+ unit tests)
├── migrations/              # versioned SQL (embedded)
├── Dockerfile
├── Makefile
├── .env.example
└── README.md
```

At the repository root: `docker-compose.yml`, `.env.example` and the `web/`
platform.

## Common tasks

```bash
make help            # list targets
make fmt vet build   # format, vet, build
make run             # run locally
make docker-up       # local stack
make docker-down     # stop stack
```
