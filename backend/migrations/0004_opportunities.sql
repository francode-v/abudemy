-- 0004_opportunities.sql
-- Skills, user offerings (classes/services) and economic opportunities.

CREATE TABLE IF NOT EXISTS skills (
    id          uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id     uuid NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    name        text NOT NULL,
    description text NOT NULL DEFAULT '',
    level       text NOT NULL DEFAULT 'beginner'
                CHECK (level IN ('beginner','intermediate','advanced','expert')),
    created_at  timestamptz NOT NULL DEFAULT now(),
    updated_at  timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT skills_user_name_unique UNIQUE (user_id, name)
);

CREATE INDEX IF NOT EXISTS skills_user_idx ON skills (user_id);

CREATE TABLE IF NOT EXISTS offerings (
    id          uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id     uuid NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    skill_id    uuid REFERENCES skills (id) ON DELETE SET NULL,
    title       text NOT NULL,
    description text NOT NULL DEFAULT '',
    type        text NOT NULL DEFAULT 'class' CHECK (type IN ('class','service')),
    status      text NOT NULL DEFAULT 'active' CHECK (status IN ('draft','active','inactive')),
    modality    text NOT NULL DEFAULT 'in_person'
                CHECK (modality IN ('in_person','online','hybrid')),
    price       numeric(12,2) NOT NULL DEFAULT 0 CHECK (price >= 0),
    currency    text NOT NULL DEFAULT 'CLP',
    comuna      text NOT NULL DEFAULT '',
    region      text NOT NULL DEFAULT '',
    created_at  timestamptz NOT NULL DEFAULT now(),
    updated_at  timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS offerings_user_idx   ON offerings (user_id);
CREATE INDEX IF NOT EXISTS offerings_status_idx ON offerings (status);

CREATE TABLE IF NOT EXISTS opportunities (
    id              uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id uuid REFERENCES organizations (id) ON DELETE SET NULL,
    title           text NOT NULL,
    description     text NOT NULL DEFAULT '',
    type            text NOT NULL DEFAULT 'job'
                    CHECK (type IN ('job','gig','volunteer','economic')),
    status          text NOT NULL DEFAULT 'open' CHECK (status IN ('draft','open','closed')),
    comuna          text NOT NULL DEFAULT '',
    region          text NOT NULL DEFAULT '',
    salary_min      numeric(12,2),
    salary_max      numeric(12,2),
    currency        text NOT NULL DEFAULT 'CLP',
    contact_email   text NOT NULL DEFAULT '',
    created_by      uuid REFERENCES users (id) ON DELETE SET NULL,
    expires_at      timestamptz,
    created_at      timestamptz NOT NULL DEFAULT now(),
    updated_at      timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT opportunities_salary_range CHECK (salary_max IS NULL OR salary_min IS NULL OR salary_max >= salary_min)
);

CREATE INDEX IF NOT EXISTS opportunities_status_idx ON opportunities (status);
CREATE INDEX IF NOT EXISTS opportunities_type_idx   ON opportunities (type);
CREATE INDEX IF NOT EXISTS opportunities_comuna_idx ON opportunities (comuna);

CREATE TABLE IF NOT EXISTS opportunity_applications (
    id             uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    opportunity_id uuid NOT NULL REFERENCES opportunities (id) ON DELETE CASCADE,
    user_id        uuid NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    message        text NOT NULL DEFAULT '',
    status         text NOT NULL DEFAULT 'applied' CHECK (status IN ('applied','accepted','rejected')),
    created_at     timestamptz NOT NULL DEFAULT now(),
    updated_at     timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT opportunity_applications_unique UNIQUE (opportunity_id, user_id)
);

CREATE INDEX IF NOT EXISTS opportunity_applications_user_idx ON opportunity_applications (user_id);
