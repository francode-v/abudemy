-- 0009_institutional_approval.sql
-- Institutional client accounts must be approved by an administrator before
-- they can use the institutional portal. Existing accounts keep their access.

ALTER TABLE users
    ADD COLUMN IF NOT EXISTS institutional_approved boolean NOT NULL DEFAULT true;

-- Only institutional clients are subject to the approval flow. Every other
-- role is either administrative or unrelated to the institutional portal.
UPDATE users
SET institutional_approved = true
WHERE role <> 'institutional_client';

CREATE INDEX IF NOT EXISTS users_institutional_approved_idx
    ON users (institutional_approved)
    WHERE role = 'institutional_client';
