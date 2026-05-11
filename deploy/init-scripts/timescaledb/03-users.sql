-- ============================================================
-- LinkFlow v2 - users
-- ============================================================
-- Stores platform users for backend-api authentication.
--
-- This table is the source of truth for:
--   - login identities: email / phone
--   - password hash
--   - platform-level role
--   - account status
--   - token invalidation through password_changed_at
--
-- Tenant membership is intentionally not defined here. A user can
-- later belong to multiple tenants through a separate membership table.
-- ============================================================

CREATE TABLE IF NOT EXISTS users (
    id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),

    email text,
    phone text,

    password_hash text NOT NULL,

    status text NOT NULL DEFAULT 'active',
    role text NOT NULL DEFAULT 'user',

    email_verified_at timestamptz,
    phone_verified_at timestamptz,
    password_changed_at timestamptz NOT NULL DEFAULT now(),

    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    last_login_at timestamptz,

    CONSTRAINT users_identity_required
        CHECK (email IS NOT NULL OR phone IS NOT NULL),

    CONSTRAINT users_email_not_blank
        CHECK (email IS NULL OR length(trim(email)) > 0),

    CONSTRAINT users_phone_not_blank
        CHECK (phone IS NULL OR length(trim(phone)) > 0),

    CONSTRAINT users_password_hash_not_blank
        CHECK (length(trim(password_hash)) > 0),

    CONSTRAINT users_status_valid
        CHECK (status IN ('active', 'disabled')),

    CONSTRAINT users_role_valid
        CHECK (role IN ('platform_admin', 'user'))
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_users_email_unique
    ON users (lower(email))
    WHERE email IS NOT NULL;

CREATE UNIQUE INDEX IF NOT EXISTS idx_users_phone_unique
    ON users (phone)
    WHERE phone IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_users_status
    ON users (status);

CREATE INDEX IF NOT EXISTS idx_users_role
    ON users (role);

DO $$
BEGIN
    RAISE NOTICE 'Initialized table: users';
END $$;
