-- 0002_services.sql
-- Establishments, support resources, categories and schedules.

CREATE TABLE IF NOT EXISTS service_categories (
    id          uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    name        text NOT NULL,
    slug        text NOT NULL,
    description text NOT NULL DEFAULT '',
    created_at  timestamptz NOT NULL DEFAULT now(),
    updated_at  timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT service_categories_slug_unique UNIQUE (slug)
);

CREATE TABLE IF NOT EXISTS services (
    id              uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id uuid REFERENCES organizations (id) ON DELETE SET NULL,
    category_id     uuid REFERENCES service_categories (id) ON DELETE SET NULL,
    name            text NOT NULL,
    description     text NOT NULL DEFAULT '',
    type            text NOT NULL DEFAULT 'establishment'
                    CHECK (type IN ('establishment','resource')),
    status          text NOT NULL DEFAULT 'active'
                    CHECK (status IN ('active','inactive','draft')),
    address         text NOT NULL DEFAULT '',
    comuna          text NOT NULL DEFAULT '',
    region          text NOT NULL DEFAULT '',
    latitude        double precision,
    longitude       double precision,
    phone           text NOT NULL DEFAULT '',
    email           text NOT NULL DEFAULT '',
    website         text NOT NULL DEFAULT '',
    created_by      uuid REFERENCES users (id) ON DELETE SET NULL,
    created_at      timestamptz NOT NULL DEFAULT now(),
    updated_at      timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT services_latitude_range  CHECK (latitude IS NULL OR latitude BETWEEN -90 AND 90),
    CONSTRAINT services_longitude_range CHECK (longitude IS NULL OR longitude BETWEEN -180 AND 180)
);

CREATE INDEX IF NOT EXISTS services_category_idx     ON services (category_id);
CREATE INDEX IF NOT EXISTS services_organization_idx ON services (organization_id);
CREATE INDEX IF NOT EXISTS services_comuna_idx        ON services (comuna);
CREATE INDEX IF NOT EXISTS services_region_idx        ON services (region);
CREATE INDEX IF NOT EXISTS services_status_idx        ON services (status);
CREATE INDEX IF NOT EXISTS services_type_idx          ON services (type);
CREATE INDEX IF NOT EXISTS services_name_trgm_idx     ON services (lower(name));

CREATE TABLE IF NOT EXISTS service_schedules (
    id          uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    service_id  uuid NOT NULL REFERENCES services (id) ON DELETE CASCADE,
    day_of_week smallint NOT NULL CHECK (day_of_week BETWEEN 0 AND 6),
    opens_at    time NOT NULL,
    closes_at   time NOT NULL,
    CONSTRAINT service_schedules_order CHECK (closes_at > opens_at)
);

CREATE INDEX IF NOT EXISTS service_schedules_service_idx ON service_schedules (service_id);
