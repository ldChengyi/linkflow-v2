-- ============================================================
-- LinkFlow v2 - products
-- ============================================================
-- Stores device product definitions.
--
-- A product describes a class of devices. Devices will bind to a
-- product later through product_id, while thing-model definitions
-- should live in versioned product-owned tables instead of this table.
--
-- RLS model:
--   Backend code sets app.current_user_id inside each request
--   transaction. Product access is allowed when the actor owns the
--   parent tenant.
-- ============================================================

CREATE TABLE IF NOT EXISTS products (
    id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),

    tenant_id uuid NOT NULL,

    product_key text NOT NULL,
    product_name text NOT NULL,
    description text NOT NULL DEFAULT '',

    node_type text NOT NULL DEFAULT 'direct',
    auth_type text NOT NULL DEFAULT 'secret',
    protocol_type text NOT NULL DEFAULT 'mqtt',

    status text NOT NULL DEFAULT 'active',

    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),

    CONSTRAINT products_key_not_blank
        CHECK (length(trim(product_key)) > 0),

    CONSTRAINT products_key_format
        CHECK (product_key ~ '^[a-z0-9][a-z0-9-]{1,61}[a-z0-9]$'),

    CONSTRAINT products_name_not_blank
        CHECK (length(trim(product_name)) > 0),

    CONSTRAINT products_node_type_valid
        CHECK (node_type IN ('direct', 'gateway', 'sub_device')),

    CONSTRAINT products_auth_type_valid
        CHECK (auth_type IN ('secret', 'certificate', 'anonymous')),

    CONSTRAINT products_protocol_type_valid
        CHECK (protocol_type IN ('mqtt', 'http', 'coap', 'modbus', 'opcua', 'lora')),

    CONSTRAINT products_status_valid
        CHECK (status IN ('active', 'disabled'))
);

ALTER TABLE products
    DROP CONSTRAINT IF EXISTS products_tenant_id_fkey;

CREATE UNIQUE INDEX IF NOT EXISTS idx_products_tenant_key_unique
    ON products (tenant_id, lower(product_key));

CREATE UNIQUE INDEX IF NOT EXISTS idx_products_id_tenant_unique
    ON products (id, tenant_id);

CREATE INDEX IF NOT EXISTS idx_products_tenant_status
    ON products (tenant_id, status);

ALTER TABLE products ENABLE ROW LEVEL SECURITY;
ALTER TABLE products FORCE ROW LEVEL SECURITY;

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM pg_policies
        WHERE schemaname = current_schema()
          AND tablename = 'products'
          AND policyname = 'products_select_for_tenant_owner'
    ) THEN
        CREATE POLICY products_select_for_tenant_owner
            ON products
            FOR SELECT
            USING (
                EXISTS (
                    SELECT 1
                    FROM tenants t
                    WHERE t.id = products.tenant_id
                      AND t.owner_user_id = current_app_user_id()
                )
            );
    END IF;

    IF NOT EXISTS (
        SELECT 1
        FROM pg_policies
        WHERE schemaname = current_schema()
          AND tablename = 'products'
          AND policyname = 'products_insert_for_tenant_owner'
    ) THEN
        CREATE POLICY products_insert_for_tenant_owner
            ON products
            FOR INSERT
            WITH CHECK (
                EXISTS (
                    SELECT 1
                    FROM tenants t
                    WHERE t.id = products.tenant_id
                      AND t.owner_user_id = current_app_user_id()
                )
            );
    END IF;

    IF NOT EXISTS (
        SELECT 1
        FROM pg_policies
        WHERE schemaname = current_schema()
          AND tablename = 'products'
          AND policyname = 'products_update_for_tenant_owner'
    ) THEN
        CREATE POLICY products_update_for_tenant_owner
            ON products
            FOR UPDATE
            USING (
                EXISTS (
                    SELECT 1
                    FROM tenants t
                    WHERE t.id = products.tenant_id
                      AND t.owner_user_id = current_app_user_id()
                )
            )
            WITH CHECK (
                EXISTS (
                    SELECT 1
                    FROM tenants t
                    WHERE t.id = products.tenant_id
                      AND t.owner_user_id = current_app_user_id()
                )
            );
    END IF;

    IF NOT EXISTS (
        SELECT 1
        FROM pg_policies
        WHERE schemaname = current_schema()
          AND tablename = 'products'
          AND policyname = 'products_delete_for_tenant_owner'
    ) THEN
        CREATE POLICY products_delete_for_tenant_owner
            ON products
            FOR DELETE
            USING (
                EXISTS (
                    SELECT 1
                    FROM tenants t
                    WHERE t.id = products.tenant_id
                      AND t.owner_user_id = current_app_user_id()
                )
            );
    END IF;
END $$;

DO $$
BEGIN
    RAISE NOTICE 'Initialized table: products';
END $$;
