-- 0010_user_profile.sql
-- Socio-demographic profile collected when a person creates a mobile account:
-- comuna, health system, sex and living situation. Existing accounts keep
-- empty values (the mobile form requires them, but the API stays backwards
-- compatible for institutional accounts and integrations).

ALTER TABLE users
    ADD COLUMN IF NOT EXISTS comuna text NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS health_system text NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS sex text NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS living_situation text NOT NULL DEFAULT '';

ALTER TABLE users
    ADD CONSTRAINT users_health_system_check
        CHECK (health_system IN (
            '', 'fonasa', 'isapre', 'capredena', 'dipreca', 'private', 'other'
        )),
    ADD CONSTRAINT users_sex_check
        CHECK (sex IN (
            '', 'female', 'male', 'other', 'prefer_not_to_say'
        )),
    ADD CONSTRAINT users_living_situation_check
        CHECK (living_situation IN (
            '', 'alone', 'with_partner', 'with_children',
            'with_partner_and_children', 'with_family', 'with_caregiver', 'other'
        ));
