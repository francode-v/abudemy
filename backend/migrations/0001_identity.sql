-- 0001_identity.sql
-- Identity, authentication and organization membership.

CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE IF NOT EXISTS users (
    id            uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    email         text NOT NULL,
    password_hash text NOT NULL,
    full_name     text NOT NULL,
    phone         text NOT NULL DEFAULT '',
    birth_date    date,
    role          text NOT NULL DEFAULT 'mobile_user'
                  CHECK (role IN ('mobile_user','institutional_client','org_collaborator','admin')),
    status        text NOT NULL DEFAULT 'active'
                  CHECK (status IN ('active','inactive','suspended')),
    last_login_at timestamptz,
    created_at    timestamptz NOT NULL DEFAULT now(),
    updated_at    timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT users_email_unique UNIQUE (email),
    CONSTRAINT users_email_lowercase CHECK (email = lower(email))
);

CREATE INDEX IF NOT EXISTS users_role_idx   ON users (role);
CREATE INDEX IF NOT EXISTS users_status_idx ON users (status);

CREATE TABLE IF NOT EXISTS sessions (
    id                 uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id            uuid NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    refresh_token_hash text NOT NULL,
    user_agent         text NOT NULL DEFAULT '',
    ip_address         text NOT NULL DEFAULT '',
    expires_at         timestamptz NOT NULL,
    revoked_at         timestamptz,
    last_used_at       timestamptz,
    created_at         timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT sessions_refresh_token_unique UNIQUE (refresh_token_hash)
);

CREATE INDEX IF NOT EXISTS sessions_user_idx    ON sessions (user_id);
CREATE INDEX IF NOT EXISTS sessions_expires_idx ON sessions (expires_at);

CREATE TABLE IF NOT EXISTS password_reset_tokens (
    id         uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id    uuid NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    token_hash text NOT NULL,
    expires_at timestamptz NOT NULL,
    used_at    timestamptz,
    created_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT password_reset_token_unique UNIQUE (token_hash)
);

CREATE INDEX IF NOT EXISTS password_reset_user_idx ON password_reset_tokens (user_id);

CREATE TABLE IF NOT EXISTS organizations (
    id          uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    name        text NOT NULL,
    type        text NOT NULL DEFAULT 'other'
                CHECK (type IN ('health','community','ngo','employer','government','educational','other')),
    legal_id    text NOT NULL DEFAULT '',
    description text NOT NULL DEFAULT '',
    email       text NOT NULL DEFAULT '',
    phone       text NOT NULL DEFAULT '',
    address     text NOT NULL DEFAULT '',
    comuna      text NOT NULL DEFAULT '',
    region      text NOT NULL DEFAULT '',
    latitude    double precision,
    longitude   double precision,
    active      boolean NOT NULL DEFAULT true,
    created_at  timestamptz NOT NULL DEFAULT now(),
    updated_at  timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT organizations_latitude_range  CHECK (latitude IS NULL OR latitude BETWEEN -90 AND 90),
    CONSTRAINT organizations_longitude_range CHECK (longitude IS NULL OR longitude BETWEEN -180 AND 180)
);

CREATE INDEX IF NOT EXISTS organizations_comuna_idx ON organizations (comuna);
CREATE INDEX IF NOT EXISTS organizations_type_idx   ON organizations (type);

CREATE TABLE IF NOT EXISTS organization_members (
    id              uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id uuid NOT NULL REFERENCES organizations (id) ON DELETE CASCADE,
    user_id         uuid NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    role            text NOT NULL DEFAULT 'viewer'
                    CHECK (role IN ('owner','admin','editor','viewer')),
    created_at      timestamptz NOT NULL DEFAULT now(),
    updated_at      timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT organization_members_unique UNIQUE (organization_id, user_id)
);

CREATE INDEX IF NOT EXISTS organization_members_user_idx ON organization_members (user_id);
