-- 0006_engagement.sql
-- Engagement features requested by the mobile application: user preferences,
-- push device registry, in-app notification feed, activity reminders and
-- suspicious-content reports.

-- Preferences are stored per user as JSON documents so the mobile client can
-- evolve its accessibility/notification toggles without a schema change.
CREATE TABLE IF NOT EXISTS user_preferences (
    user_id       uuid PRIMARY KEY REFERENCES users (id) ON DELETE CASCADE,
    interests     jsonb NOT NULL DEFAULT '[]'::jsonb,
    accessibility jsonb NOT NULL DEFAULT '{}'::jsonb,
    notifications jsonb NOT NULL DEFAULT '{}'::jsonb,
    created_at    timestamptz NOT NULL DEFAULT now(),
    updated_at    timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT user_preferences_interests_array CHECK (jsonb_typeof(interests) = 'array'),
    CONSTRAINT user_preferences_accessibility_object CHECK (jsonb_typeof(accessibility) = 'object'),
    CONSTRAINT user_preferences_notifications_object CHECK (jsonb_typeof(notifications) = 'object')
);

-- Registered push devices. A device is soft-revoked (revoked_at) so the
-- delivery history stays auditable.
CREATE TABLE IF NOT EXISTS devices (
    id           uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id      uuid NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    platform     text NOT NULL CHECK (platform IN ('ios', 'android', 'web')),
    token        text NOT NULL,
    device_name  text NOT NULL DEFAULT '',
    app_version  text NOT NULL DEFAULT '',
    last_seen_at timestamptz NOT NULL DEFAULT now(),
    revoked_at   timestamptz,
    created_at   timestamptz NOT NULL DEFAULT now(),
    updated_at   timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT devices_token_unique UNIQUE (token)
);

CREATE INDEX IF NOT EXISTS devices_user_idx     ON devices (user_id);
CREATE INDEX IF NOT EXISTS devices_active_idx   ON devices (user_id) WHERE revoked_at IS NULL;

-- In-app notification feed. `push_status` records the honest delivery attempt
-- against a push provider: when no provider is configured nothing is sent and
-- the attempt is stored as 'skipped'.
CREATE TABLE IF NOT EXISTS notifications (
    id              uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id         uuid NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    kind            text NOT NULL DEFAULT 'general',
    title           text NOT NULL,
    body            text NOT NULL DEFAULT '',
    data            jsonb NOT NULL DEFAULT '{}'::jsonb,
    read_at         timestamptz,
    push_status     text NOT NULL DEFAULT 'skipped'
                    CHECK (push_status IN ('pending', 'sent', 'failed', 'skipped')),
    push_provider   text NOT NULL DEFAULT 'none',
    push_error      text NOT NULL DEFAULT '',
    created_at      timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT notifications_data_object CHECK (jsonb_typeof(data) = 'object')
);

CREATE INDEX IF NOT EXISTS notifications_user_idx        ON notifications (user_id, created_at DESC);
CREATE INDEX IF NOT EXISTS notifications_unread_idx      ON notifications (user_id) WHERE read_at IS NULL;

-- Reminders a user sets for an activity. Delivery depends on a scheduler and a
-- notification provider; without them the reminder is simply stored.
CREATE TABLE IF NOT EXISTS activity_reminders (
    id          uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id     uuid NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    activity_id uuid NOT NULL REFERENCES activities (id) ON DELETE CASCADE,
    remind_at   timestamptz NOT NULL,
    status      text NOT NULL DEFAULT 'pending'
                CHECK (status IN ('pending', 'cancelled', 'sent', 'skipped')),
    created_at  timestamptz NOT NULL DEFAULT now(),
    updated_at  timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT activity_reminders_unique UNIQUE (activity_id, user_id)
);

CREATE INDEX IF NOT EXISTS activity_reminders_user_idx ON activity_reminders (user_id);

-- Reports of suspicious or abusive opportunity posts.
CREATE TABLE IF NOT EXISTS opportunity_reports (
    id             uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    opportunity_id uuid NOT NULL REFERENCES opportunities (id) ON DELETE CASCADE,
    user_id        uuid NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    reason         text NOT NULL,
    details        text NOT NULL DEFAULT '',
    status         text NOT NULL DEFAULT 'open'
                   CHECK (status IN ('open', 'reviewing', 'dismissed', 'actioned')),
    created_at     timestamptz NOT NULL DEFAULT now(),
    updated_at     timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT opportunity_reports_unique UNIQUE (opportunity_id, user_id)
);

CREATE INDEX IF NOT EXISTS opportunity_reports_opportunity_idx ON opportunity_reports (opportunity_id);
CREATE INDEX IF NOT EXISTS opportunity_reports_status_idx      ON opportunity_reports (status);
