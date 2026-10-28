-- 0003_activities.sql
-- Workshops, events, courses and their enrollments.

CREATE TABLE IF NOT EXISTS activities (
    id              uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id uuid REFERENCES organizations (id) ON DELETE SET NULL,
    service_id      uuid REFERENCES services (id) ON DELETE SET NULL,
    category_id     uuid REFERENCES service_categories (id) ON DELETE SET NULL,
    title           text NOT NULL,
    description     text NOT NULL DEFAULT '',
    type            text NOT NULL DEFAULT 'workshop'
                    CHECK (type IN ('workshop','event','course','support_group')),
    modality        text NOT NULL DEFAULT 'in_person'
                    CHECK (modality IN ('in_person','online','hybrid')),
    status          text NOT NULL DEFAULT 'draft'
                    CHECK (status IN ('draft','published','cancelled','finished')),
    location        text NOT NULL DEFAULT '',
    comuna          text NOT NULL DEFAULT '',
    region          text NOT NULL DEFAULT '',
    latitude        double precision,
    longitude       double precision,
    start_at        timestamptz NOT NULL,
    end_at          timestamptz,
    capacity        integer NOT NULL DEFAULT 0 CHECK (capacity >= 0),
    price           numeric(12,2) NOT NULL DEFAULT 0 CHECK (price >= 0),
    currency        text NOT NULL DEFAULT 'CLP',
    created_by      uuid REFERENCES users (id) ON DELETE SET NULL,
    created_at      timestamptz NOT NULL DEFAULT now(),
    updated_at      timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT activities_period CHECK (end_at IS NULL OR end_at >= start_at),
    CONSTRAINT activities_latitude_range  CHECK (latitude IS NULL OR latitude BETWEEN -90 AND 90),
    CONSTRAINT activities_longitude_range CHECK (longitude IS NULL OR longitude BETWEEN -180 AND 180)
);

CREATE INDEX IF NOT EXISTS activities_organization_idx ON activities (organization_id);
CREATE INDEX IF NOT EXISTS activities_status_idx       ON activities (status);
CREATE INDEX IF NOT EXISTS activities_type_idx         ON activities (type);
CREATE INDEX IF NOT EXISTS activities_start_idx        ON activities (start_at);
CREATE INDEX IF NOT EXISTS activities_comuna_idx       ON activities (comuna);

CREATE TABLE IF NOT EXISTS activity_enrollments (
    id          uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    activity_id uuid NOT NULL REFERENCES activities (id) ON DELETE CASCADE,
    user_id     uuid NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    status      text NOT NULL DEFAULT 'enrolled'
                CHECK (status IN ('enrolled','waitlist','cancelled','attended')),
    created_at  timestamptz NOT NULL DEFAULT now(),
    updated_at  timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT activity_enrollments_unique UNIQUE (activity_id, user_id)
);

CREATE INDEX IF NOT EXISTS activity_enrollments_activity_idx ON activity_enrollments (activity_id);
CREATE INDEX IF NOT EXISTS activity_enrollments_user_idx     ON activity_enrollments (user_id);
CREATE INDEX IF NOT EXISTS activity_enrollments_status_idx   ON activity_enrollments (status);
