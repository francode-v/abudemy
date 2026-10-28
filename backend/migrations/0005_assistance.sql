-- 0005_assistance.sql
-- Trusted contacts and assistance requests with honest notification tracking.

CREATE TABLE IF NOT EXISTS trusted_contacts (
    id           uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id      uuid NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    name         text NOT NULL,
    relationship text NOT NULL DEFAULT '',
    phone        text NOT NULL DEFAULT '',
    email        text NOT NULL DEFAULT '',
    priority     integer NOT NULL DEFAULT 1 CHECK (priority BETWEEN 1 AND 10),
    created_at   timestamptz NOT NULL DEFAULT now(),
    updated_at   timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT trusted_contacts_reachable CHECK (phone <> '' OR email <> '')
);

CREATE INDEX IF NOT EXISTS trusted_contacts_user_idx ON trusted_contacts (user_id);

CREATE TABLE IF NOT EXISTS assistance_requests (
    id          uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id     uuid NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    type        text NOT NULL DEFAULT 'general',
    description text NOT NULL DEFAULT '',
    latitude    double precision,
    longitude   double precision,
    status      text NOT NULL DEFAULT 'pending'
                CHECK (status IN ('pending','notified','resolved','cancelled')),
    resolved_at timestamptz,
    created_at  timestamptz NOT NULL DEFAULT now(),
    updated_at  timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT assistance_latitude_range  CHECK (latitude IS NULL OR latitude BETWEEN -90 AND 90),
    CONSTRAINT assistance_longitude_range CHECK (longitude IS NULL OR longitude BETWEEN -180 AND 180)
);

CREATE INDEX IF NOT EXISTS assistance_requests_user_idx   ON assistance_requests (user_id);
CREATE INDEX IF NOT EXISTS assistance_requests_status_idx ON assistance_requests (status);
CREATE INDEX IF NOT EXISTS assistance_requests_created_idx ON assistance_requests (created_at);

CREATE TABLE IF NOT EXISTS assistance_notifications (
    id                    uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    assistance_request_id uuid NOT NULL REFERENCES assistance_requests (id) ON DELETE CASCADE,
    contact_id            uuid REFERENCES trusted_contacts (id) ON DELETE SET NULL,
    channel               text NOT NULL DEFAULT 'sms',
    recipient             text NOT NULL DEFAULT '',
    provider              text NOT NULL DEFAULT 'none',
    status                text NOT NULL DEFAULT 'pending'
                          CHECK (status IN ('pending','sent','failed','skipped')),
    error                 text NOT NULL DEFAULT '',
    sent_at               timestamptz,
    created_at            timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS assistance_notifications_request_idx ON assistance_notifications (assistance_request_id);
CREATE INDEX IF NOT EXISTS assistance_notifications_status_idx  ON assistance_notifications (status);
