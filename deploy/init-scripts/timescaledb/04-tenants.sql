-- ============================================================
-- LinkFlow v2 - tenants
-- ============================================================
-- Stores first-stage tenant workspaces.
--
-- Simple tenancy model:
--   - users are login identities
--   - a user can create one or more tenants
--   - each tenant has one owner user
--
-- This version intentionally does not include tenant_members,
-- invitations, or shared workspaces.
--
-- RLS model:
--   Backend code should set this PostgreSQL session variable inside
--   each request transaction before querying tenant-scoped tables:
--
--     SET LOCAL app.current_user_id = '<user uuid>';
-- ============================================================

CREATE TABLE IF NOT EXISTS tenants (
    id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),

    owner_user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,

    tenant_slug text NOT NULL,
    tenant_name text NOT NULL,

    status text NOT NULL DEFAULT 'active',

    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),

    CONSTRAINT tenants_slug_not_blank
        CHECK (length(trim(tenant_slug)) > 0),

    CONSTRAINT tenants_slug_format
        CHECK (tenant_slug ~ '^[a-z0-9][a-z0-9-]{1,61}[a-z0-9]$'),

    CONSTRAINT tenants_name_not_blank
        CHECK (length(trim(tenant_name)) > 0),

    CONSTRAINT tenants_status_valid
        CHECK (status IN ('active', 'disabled'))
);

ALTER TABLE tenants
    DROP COLUMN IF EXISTS description;

CREATE UNIQUE INDEX IF NOT EXISTS idx_tenants_slug_unique
    ON tenants (lower(tenant_slug));

CREATE INDEX IF NOT EXISTS idx_tenants_owner_user
    ON tenants (owner_user_id);

CREATE INDEX IF NOT EXISTS idx_tenants_status
    ON tenants (status);

CREATE OR REPLACE FUNCTION current_app_user_id()
RETURNS uuid
LANGUAGE sql
STABLE
AS $$
    SELECT NULLIF(current_setting('app.current_user_id', true), '')::uuid
$$;

CREATE OR REPLACE FUNCTION current_app_internal_service()
RETURNS text
LANGUAGE sql
STABLE
AS $$
    SELECT NULLIF(current_setting('app.internal_service', true), '')
$$;

ALTER TABLE tenants ENABLE ROW LEVEL SECURITY;
ALTER TABLE tenants FORCE ROW LEVEL SECURITY;

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM pg_policies
        WHERE schemaname = current_schema()
          AND tablename = 'tenants'
          AND policyname = 'tenants_select_for_owner'
    ) THEN
        CREATE POLICY tenants_select_for_owner
            ON tenants
            FOR SELECT
            USING (owner_user_id = current_app_user_id());
    END IF;

    IF NOT EXISTS (
        SELECT 1
        FROM pg_policies
        WHERE schemaname = current_schema()
          AND tablename = 'tenants'
          AND policyname = 'tenants_select_for_internal_service'
    ) THEN
        CREATE POLICY tenants_select_for_internal_service
            ON tenants
            FOR SELECT
            USING (current_app_internal_service() = 'backend');
    END IF;

    IF NOT EXISTS (
        SELECT 1
        FROM pg_policies
        WHERE schemaname = current_schema()
          AND tablename = 'tenants'
          AND policyname = 'tenants_insert_for_owner'
    ) THEN
        CREATE POLICY tenants_insert_for_owner
            ON tenants
            FOR INSERT
            WITH CHECK (owner_user_id = current_app_user_id());
    END IF;

    IF NOT EXISTS (
        SELECT 1
        FROM pg_policies
        WHERE schemaname = current_schema()
          AND tablename = 'tenants'
          AND policyname = 'tenants_update_for_owner'
    ) THEN
        CREATE POLICY tenants_update_for_owner
            ON tenants
            FOR UPDATE
            USING (owner_user_id = current_app_user_id())
            WITH CHECK (owner_user_id = current_app_user_id());
    END IF;

    IF NOT EXISTS (
        SELECT 1
        FROM pg_policies
        WHERE schemaname = current_schema()
          AND tablename = 'tenants'
          AND policyname = 'tenants_delete_for_owner'
    ) THEN
        CREATE POLICY tenants_delete_for_owner
            ON tenants
            FOR DELETE
            USING (owner_user_id = current_app_user_id());
    END IF;
END $$;

DO $$
BEGIN
    RAISE NOTICE 'Initialized table: tenants';
END $$;
