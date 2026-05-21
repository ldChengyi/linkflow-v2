-- ============================================================
-- LinkFlow v2 - devices
-- ============================================================
-- Stores device identities and credentials.
--
-- devices.device_slug is the platform-facing device identifier used
-- by APIs, MQTT topics, and event payloads at protocol boundaries.
-- devices.id remains the internal stable UUID used by foreign keys.
--
-- RLS model:
--   Backend code sets app.current_user_id inside each request
--   transaction. Device access is allowed when the actor owns the
--   parent tenant.
-- ============================================================

CREATE TABLE IF NOT EXISTS devices (
    id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),

    tenant_id uuid NOT NULL,
    product_id uuid NOT NULL,

    device_slug text NOT NULL,
    device_name text NOT NULL,
    description text NOT NULL DEFAULT '',

    status text NOT NULL DEFAULT 'active',
    connection_status text NOT NULL DEFAULT 'offline',

    gateway_device_id uuid,

    firmware_version text,
    ip_address text,
    last_seen_at timestamptz,

    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),

    CONSTRAINT devices_id_tenant_unique
        UNIQUE (id, tenant_id),

    CONSTRAINT devices_slug_not_blank
        CHECK (length(trim(device_slug)) > 0),

    CONSTRAINT devices_slug_format
        CHECK (device_slug ~ '^[a-zA-Z0-9][a-zA-Z0-9._:-]{0,127}$'),

    CONSTRAINT devices_name_not_blank
        CHECK (length(trim(device_name)) > 0),

    CONSTRAINT devices_status_valid
        CHECK (status IN ('active', 'disabled')),

    CONSTRAINT devices_connection_status_valid
        CHECK (connection_status IN ('online', 'offline'))
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_devices_tenant_slug_unique
    ON devices (tenant_id, lower(device_slug));

CREATE INDEX IF NOT EXISTS idx_devices_tenant_product
    ON devices (tenant_id, product_id);

CREATE INDEX IF NOT EXISTS idx_devices_tenant_status
    ON devices (tenant_id, status);

CREATE INDEX IF NOT EXISTS idx_devices_tenant_connection
    ON devices (tenant_id, connection_status);

CREATE INDEX IF NOT EXISTS idx_devices_gateway
    ON devices (gateway_device_id)
    WHERE gateway_device_id IS NOT NULL;

DROP INDEX IF EXISTS idx_devices_attributes_gin;

ALTER TABLE devices
    DROP CONSTRAINT IF EXISTS devices_tenant_id_fkey,
    DROP CONSTRAINT IF EXISTS devices_product_tenant_fk,
    DROP CONSTRAINT IF EXISTS devices_gateway_tenant_fk,
    DROP CONSTRAINT IF EXISTS devices_attributes_object;

ALTER TABLE devices
    DROP COLUMN IF EXISTS attributes;

CREATE TABLE IF NOT EXISTS device_credentials (
    id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),

    tenant_id uuid NOT NULL,
    device_id uuid NOT NULL,

    secret_hash text,
    certificate_fingerprint text,
    certificate_pem text,

    status text NOT NULL DEFAULT 'active',

    issued_at timestamptz NOT NULL DEFAULT now(),
    expires_at timestamptz,
    last_used_at timestamptz,

    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),

    CONSTRAINT device_credentials_status_valid
        CHECK (status IN ('active', 'rotating', 'revoked'))
);

ALTER TABLE device_credentials
    DROP CONSTRAINT IF EXISTS device_credentials_tenant_id_fkey,
    DROP CONSTRAINT IF EXISTS device_credentials_device_id_fkey,
    DROP CONSTRAINT IF EXISTS device_credentials_device_tenant_fk,
    DROP CONSTRAINT IF EXISTS device_credentials_type_valid,
    DROP CONSTRAINT IF EXISTS device_credentials_status_valid,
    DROP CONSTRAINT IF EXISTS device_credentials_secret_required,
    DROP CONSTRAINT IF EXISTS device_credentials_certificate_required;

DROP INDEX IF EXISTS idx_device_credentials_device_type_unique;

ALTER TABLE device_credentials
    DROP COLUMN IF EXISTS credential_type;

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM pg_constraint
        WHERE conrelid = 'device_credentials'::regclass
          AND conname = 'device_credentials_status_valid'
    ) THEN
        ALTER TABLE device_credentials
            ADD CONSTRAINT device_credentials_status_valid
            CHECK (status IN ('active', 'rotating', 'revoked'));
    END IF;
END $$;

CREATE UNIQUE INDEX IF NOT EXISTS idx_device_credentials_one_active
    ON device_credentials (device_id)
    WHERE status = 'active';

CREATE INDEX IF NOT EXISTS idx_device_credentials_tenant_status
    ON device_credentials (tenant_id, status);

CREATE INDEX IF NOT EXISTS idx_device_credentials_cert_fingerprint
    ON device_credentials (certificate_fingerprint)
    WHERE certificate_fingerprint IS NOT NULL;

ALTER TABLE devices ENABLE ROW LEVEL SECURITY;
ALTER TABLE devices FORCE ROW LEVEL SECURITY;

ALTER TABLE device_credentials ENABLE ROW LEVEL SECURITY;
ALTER TABLE device_credentials FORCE ROW LEVEL SECURITY;

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM pg_policies
        WHERE schemaname = current_schema()
          AND tablename = 'devices'
          AND policyname = 'devices_select_for_tenant_owner'
    ) THEN
        CREATE POLICY devices_select_for_tenant_owner
            ON devices
            FOR SELECT
            USING (
                EXISTS (
                    SELECT 1
                    FROM tenants t
                    WHERE t.id = devices.tenant_id
                      AND t.owner_user_id = current_app_user_id()
                )
            );
    END IF;

    IF NOT EXISTS (
        SELECT 1
        FROM pg_policies
        WHERE schemaname = current_schema()
          AND tablename = 'devices'
          AND policyname = 'devices_select_for_internal_service'
    ) THEN
        CREATE POLICY devices_select_for_internal_service
            ON devices
            FOR SELECT
            USING (current_app_internal_service() = 'backend');
    END IF;

    IF NOT EXISTS (
        SELECT 1
        FROM pg_policies
        WHERE schemaname = current_schema()
          AND tablename = 'devices'
          AND policyname = 'devices_select_for_admin'
    ) THEN
        CREATE POLICY devices_select_for_admin
            ON devices
            FOR SELECT
            USING (current_app_admin_service() IS NOT NULL);
    END IF;

    -- Admin UPDATE on devices is intended for connection-state writes from
    -- backend services (e.g. device-event-processor handling client_connected /
    -- client_disconnected). Column scoping (connection_status, last_seen_at)
    -- is enforced by the calling code, not by RLS.
    IF NOT EXISTS (
        SELECT 1
        FROM pg_policies
        WHERE schemaname = current_schema()
          AND tablename = 'devices'
          AND policyname = 'devices_update_for_admin'
    ) THEN
        CREATE POLICY devices_update_for_admin
            ON devices
            FOR UPDATE
            USING (current_app_admin_service() IS NOT NULL)
            WITH CHECK (current_app_admin_service() IS NOT NULL);
    END IF;

    IF NOT EXISTS (
        SELECT 1
        FROM pg_policies
        WHERE schemaname = current_schema()
          AND tablename = 'devices'
          AND policyname = 'devices_insert_for_tenant_owner'
    ) THEN
        CREATE POLICY devices_insert_for_tenant_owner
            ON devices
            FOR INSERT
            WITH CHECK (
                EXISTS (
                    SELECT 1
                    FROM tenants t
                    WHERE t.id = devices.tenant_id
                      AND t.owner_user_id = current_app_user_id()
                )
            );
    END IF;

    IF NOT EXISTS (
        SELECT 1
        FROM pg_policies
        WHERE schemaname = current_schema()
          AND tablename = 'devices'
          AND policyname = 'devices_update_for_tenant_owner'
    ) THEN
        CREATE POLICY devices_update_for_tenant_owner
            ON devices
            FOR UPDATE
            USING (
                EXISTS (
                    SELECT 1
                    FROM tenants t
                    WHERE t.id = devices.tenant_id
                      AND t.owner_user_id = current_app_user_id()
                )
            )
            WITH CHECK (
                EXISTS (
                    SELECT 1
                    FROM tenants t
                    WHERE t.id = devices.tenant_id
                      AND t.owner_user_id = current_app_user_id()
                )
            );
    END IF;

    IF NOT EXISTS (
        SELECT 1
        FROM pg_policies
        WHERE schemaname = current_schema()
          AND tablename = 'devices'
          AND policyname = 'devices_delete_for_tenant_owner'
    ) THEN
        CREATE POLICY devices_delete_for_tenant_owner
            ON devices
            FOR DELETE
            USING (
                EXISTS (
                    SELECT 1
                    FROM tenants t
                    WHERE t.id = devices.tenant_id
                      AND t.owner_user_id = current_app_user_id()
                )
            );
    END IF;

    IF NOT EXISTS (
        SELECT 1
        FROM pg_policies
        WHERE schemaname = current_schema()
          AND tablename = 'device_credentials'
          AND policyname = 'device_credentials_select_for_tenant_owner'
    ) THEN
        CREATE POLICY device_credentials_select_for_tenant_owner
            ON device_credentials
            FOR SELECT
            USING (
                EXISTS (
                    SELECT 1
                    FROM tenants t
                    WHERE t.id = device_credentials.tenant_id
                      AND t.owner_user_id = current_app_user_id()
                )
            );
    END IF;

    IF NOT EXISTS (
        SELECT 1
        FROM pg_policies
        WHERE schemaname = current_schema()
          AND tablename = 'device_credentials'
          AND policyname = 'device_credentials_select_for_internal_service'
    ) THEN
        CREATE POLICY device_credentials_select_for_internal_service
            ON device_credentials
            FOR SELECT
            USING (current_app_internal_service() = 'backend');
    END IF;

    IF NOT EXISTS (
        SELECT 1
        FROM pg_policies
        WHERE schemaname = current_schema()
          AND tablename = 'device_credentials'
          AND policyname = 'device_credentials_select_for_admin'
    ) THEN
        CREATE POLICY device_credentials_select_for_admin
            ON device_credentials
            FOR SELECT
            USING (current_app_admin_service() IS NOT NULL);
    END IF;

    IF NOT EXISTS (
        SELECT 1
        FROM pg_policies
        WHERE schemaname = current_schema()
          AND tablename = 'device_credentials'
          AND policyname = 'device_credentials_insert_for_tenant_owner'
    ) THEN
        CREATE POLICY device_credentials_insert_for_tenant_owner
            ON device_credentials
            FOR INSERT
            WITH CHECK (
                EXISTS (
                    SELECT 1
                    FROM tenants t
                    WHERE t.id = device_credentials.tenant_id
                      AND t.owner_user_id = current_app_user_id()
                )
            );
    END IF;

    IF NOT EXISTS (
        SELECT 1
        FROM pg_policies
        WHERE schemaname = current_schema()
          AND tablename = 'device_credentials'
          AND policyname = 'device_credentials_update_for_tenant_owner'
    ) THEN
        CREATE POLICY device_credentials_update_for_tenant_owner
            ON device_credentials
            FOR UPDATE
            USING (
                EXISTS (
                    SELECT 1
                    FROM tenants t
                    WHERE t.id = device_credentials.tenant_id
                      AND t.owner_user_id = current_app_user_id()
                )
            )
            WITH CHECK (
                EXISTS (
                    SELECT 1
                    FROM tenants t
                    WHERE t.id = device_credentials.tenant_id
                      AND t.owner_user_id = current_app_user_id()
                )
            );
    END IF;

    IF NOT EXISTS (
        SELECT 1
        FROM pg_policies
        WHERE schemaname = current_schema()
          AND tablename = 'device_credentials'
          AND policyname = 'device_credentials_delete_for_tenant_owner'
    ) THEN
        CREATE POLICY device_credentials_delete_for_tenant_owner
            ON device_credentials
            FOR DELETE
            USING (
                EXISTS (
                    SELECT 1
                    FROM tenants t
                    WHERE t.id = device_credentials.tenant_id
                      AND t.owner_user_id = current_app_user_id()
                )
            );
    END IF;
END $$;

DO $$
BEGIN
    RAISE NOTICE 'Initialized tables: devices, device_credentials';
END $$;
