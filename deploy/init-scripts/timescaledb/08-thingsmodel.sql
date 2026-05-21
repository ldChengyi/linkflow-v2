-- ============================================================
-- LinkFlow v2 - thingsmodel
-- ============================================================
-- Stores versioned thing-model definitions owned by a product.
--
-- A thing model describes the product-level capabilities that device
-- property reports, events, and service calls should conform to.
--
-- Shape:
--   - one product can have many model versions
--   - one product can have at most one current published model
--   - properties/events/services are kept as JSONB contract objects
--     so the backend can validate and evolve their detailed schema
--     without changing table shape for every new capability type.
--
-- RLS model:
--   Backend code sets app.current_user_id inside each request
--   transaction. Thing-model access is allowed when the actor owns
--   the parent tenant.
-- ============================================================

CREATE TABLE IF NOT EXISTS thingsmodel (
    id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),

    tenant_id uuid NOT NULL,
    product_id uuid NOT NULL,

    model_version integer NOT NULL DEFAULT 1,
    model_name text NOT NULL DEFAULT '',
    description text NOT NULL DEFAULT '',

    status text NOT NULL DEFAULT 'draft',
    is_current boolean NOT NULL DEFAULT false,

    properties jsonb NOT NULL DEFAULT '{}'::jsonb,
    events jsonb NOT NULL DEFAULT '{}'::jsonb,
    services jsonb NOT NULL DEFAULT '{}'::jsonb,

    published_at timestamptz,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),

    CONSTRAINT thingsmodel_version_positive
        CHECK (model_version > 0),

    CONSTRAINT thingsmodel_status_valid
        CHECK (status IN ('draft', 'published', 'deprecated')),

    CONSTRAINT thingsmodel_current_requires_published
        CHECK (is_current = false OR status = 'published'),

    CONSTRAINT thingsmodel_published_at_matches_status
        CHECK (
            (status = 'published' AND published_at IS NOT NULL)
            OR (status <> 'published')
        ),

    CONSTRAINT thingsmodel_properties_object
        CHECK (jsonb_typeof(properties) = 'object'),

    CONSTRAINT thingsmodel_events_object
        CHECK (jsonb_typeof(events) = 'object'),

    CONSTRAINT thingsmodel_services_object
        CHECK (jsonb_typeof(services) = 'object')
);

ALTER TABLE thingsmodel
    DROP CONSTRAINT IF EXISTS thingsmodel_tenant_id_fkey,
    DROP CONSTRAINT IF EXISTS thingsmodel_product_tenant_fk;

CREATE UNIQUE INDEX IF NOT EXISTS idx_thingsmodel_product_version_unique
    ON thingsmodel (tenant_id, product_id, model_version);

CREATE UNIQUE INDEX IF NOT EXISTS idx_thingsmodel_product_current_unique
    ON thingsmodel (tenant_id, product_id)
    WHERE is_current;

CREATE INDEX IF NOT EXISTS idx_thingsmodel_tenant_status
    ON thingsmodel (tenant_id, status);

CREATE INDEX IF NOT EXISTS idx_thingsmodel_product_status
    ON thingsmodel (tenant_id, product_id, status);

CREATE INDEX IF NOT EXISTS idx_thingsmodel_properties_gin
    ON thingsmodel
    USING gin (properties);

CREATE INDEX IF NOT EXISTS idx_thingsmodel_events_gin
    ON thingsmodel
    USING gin (events);

CREATE INDEX IF NOT EXISTS idx_thingsmodel_services_gin
    ON thingsmodel
    USING gin (services);

ALTER TABLE thingsmodel ENABLE ROW LEVEL SECURITY;
ALTER TABLE thingsmodel FORCE ROW LEVEL SECURITY;

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM pg_policies
        WHERE schemaname = current_schema()
          AND tablename = 'thingsmodel'
          AND policyname = 'thingsmodel_select_for_tenant_owner'
    ) THEN
        CREATE POLICY thingsmodel_select_for_tenant_owner
            ON thingsmodel
            FOR SELECT
            USING (
                EXISTS (
                    SELECT 1
                    FROM tenants t
                    WHERE t.id = thingsmodel.tenant_id
                      AND t.owner_user_id = current_app_user_id()
                )
            );
    END IF;

    IF NOT EXISTS (
        SELECT 1
        FROM pg_policies
        WHERE schemaname = current_schema()
          AND tablename = 'thingsmodel'
          AND policyname = 'thingsmodel_select_for_admin'
    ) THEN
        CREATE POLICY thingsmodel_select_for_admin
            ON thingsmodel
            FOR SELECT
            USING (current_app_admin_service() IS NOT NULL);
    END IF;

    IF NOT EXISTS (
        SELECT 1
        FROM pg_policies
        WHERE schemaname = current_schema()
          AND tablename = 'thingsmodel'
          AND policyname = 'thingsmodel_insert_for_tenant_owner'
    ) THEN
        CREATE POLICY thingsmodel_insert_for_tenant_owner
            ON thingsmodel
            FOR INSERT
            WITH CHECK (
                EXISTS (
                    SELECT 1
                    FROM tenants t
                    WHERE t.id = thingsmodel.tenant_id
                      AND t.owner_user_id = current_app_user_id()
                )
                AND EXISTS (
                    SELECT 1
                    FROM products p
                    WHERE p.id = thingsmodel.product_id
                      AND p.tenant_id = thingsmodel.tenant_id
                )
            );
    END IF;

    IF NOT EXISTS (
        SELECT 1
        FROM pg_policies
        WHERE schemaname = current_schema()
          AND tablename = 'thingsmodel'
          AND policyname = 'thingsmodel_update_for_tenant_owner'
    ) THEN
        CREATE POLICY thingsmodel_update_for_tenant_owner
            ON thingsmodel
            FOR UPDATE
            USING (
                EXISTS (
                    SELECT 1
                    FROM tenants t
                    WHERE t.id = thingsmodel.tenant_id
                      AND t.owner_user_id = current_app_user_id()
                )
            )
            WITH CHECK (
                EXISTS (
                    SELECT 1
                    FROM tenants t
                    WHERE t.id = thingsmodel.tenant_id
                      AND t.owner_user_id = current_app_user_id()
                )
                AND EXISTS (
                    SELECT 1
                    FROM products p
                    WHERE p.id = thingsmodel.product_id
                      AND p.tenant_id = thingsmodel.tenant_id
                )
            );
    END IF;

    IF NOT EXISTS (
        SELECT 1
        FROM pg_policies
        WHERE schemaname = current_schema()
          AND tablename = 'thingsmodel'
          AND policyname = 'thingsmodel_delete_for_tenant_owner'
    ) THEN
        CREATE POLICY thingsmodel_delete_for_tenant_owner
            ON thingsmodel
            FOR DELETE
            USING (
                EXISTS (
                    SELECT 1
                    FROM tenants t
                    WHERE t.id = thingsmodel.tenant_id
                      AND t.owner_user_id = current_app_user_id()
                )
            );
    END IF;
END $$;

DO $$
BEGIN
    RAISE NOTICE 'Initialized table: thingsmodel';
END $$;
