-- ============================================================
-- LinkFlow v2 - audit logs
-- ============================================================
-- Stores business audit events for backend API operations.
--
-- This is not a request log. It records who performed which business
-- action on which resource and whether the operation succeeded.
-- ============================================================

CREATE TABLE IF NOT EXISTS audit_logs (
    id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),

    actor_user_id uuid REFERENCES users(id) ON DELETE SET NULL,
    actor_role text,

    action text NOT NULL,
    resource_type text NOT NULL,
    resource_id text,

    result text NOT NULL,
    error_code text,

    method text NOT NULL,
    path text NOT NULL,
    status_code integer NOT NULL,
    duration_ms integer NOT NULL,

    ip text,
    user_agent text,
    request_id text,
    trace_id text,
    operation_id text,

    metadata jsonb NOT NULL DEFAULT '{}'::jsonb,

    created_at timestamptz NOT NULL DEFAULT now(),

    CONSTRAINT audit_logs_action_not_blank
        CHECK (length(trim(action)) > 0),

    CONSTRAINT audit_logs_resource_type_not_blank
        CHECK (length(trim(resource_type)) > 0),

    CONSTRAINT audit_logs_result_valid
        CHECK (result IN ('success', 'failure')),

    CONSTRAINT audit_logs_metadata_object
        CHECK (jsonb_typeof(metadata) = 'object')
);

ALTER TABLE audit_logs
    ADD COLUMN IF NOT EXISTS method text NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS path text NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS status_code integer NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS duration_ms integer NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS operation_id text;

CREATE INDEX IF NOT EXISTS idx_audit_logs_actor_created
    ON audit_logs (actor_user_id, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_audit_logs_action_created
    ON audit_logs (action, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_audit_logs_resource_created
    ON audit_logs (resource_type, resource_id, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_audit_logs_operation_created
    ON audit_logs (operation_id, created_at DESC)
    WHERE operation_id IS NOT NULL;

ALTER TABLE audit_logs ENABLE ROW LEVEL SECURITY;
ALTER TABLE audit_logs FORCE ROW LEVEL SECURITY;

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM pg_policies
        WHERE schemaname = current_schema()
          AND tablename = 'audit_logs'
          AND policyname = 'audit_logs_select_for_actor'
    ) THEN
        CREATE POLICY audit_logs_select_for_actor
            ON audit_logs
            FOR SELECT
            USING (actor_user_id = current_app_user_id());
    END IF;

    IF NOT EXISTS (
        SELECT 1
        FROM pg_policies
        WHERE schemaname = current_schema()
          AND tablename = 'audit_logs'
          AND policyname = 'audit_logs_insert_for_actor'
    ) THEN
        CREATE POLICY audit_logs_insert_for_actor
            ON audit_logs
            FOR INSERT
            WITH CHECK (actor_user_id = current_app_user_id());
    END IF;
END $$;

DO $$
BEGIN
    RAISE NOTICE 'Initialized table: audit_logs';
END $$;
