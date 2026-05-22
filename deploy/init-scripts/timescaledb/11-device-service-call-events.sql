-- ============================================================
-- LinkFlow v2 - cloud-to-device service call storage
-- ============================================================
-- Stores service calls dispatched by the cloud platform to devices.
--
-- MQTT topic:
--   lf/v1/{tenant_slug}/{product_key}/{device_slug}/service/down/{service_name}
--
-- The cloud producer owns command_id. The device acknowledgement is stored
-- separately in device_service_call_ack_events and can correlate through the
-- command_id when devices include it in the reply payload.
-- ============================================================

CREATE TABLE IF NOT EXISTS device_service_call_events (
    command_id uuid NOT NULL,
    tenant_id text NOT NULL,
    product_key text NOT NULL,
    device_slug text NOT NULL,
    service_name text NOT NULL,
    protocol text NOT NULL,
    topic text NOT NULL,
    occurred_at timestamptz NOT NULL,
    received_at timestamptz NOT NULL DEFAULT now(),
    producer text NOT NULL,
    requested_by uuid NOT NULL,
    input jsonb NOT NULL,

    PRIMARY KEY (command_id, occurred_at),
    CONSTRAINT device_service_call_events_input_object
        CHECK (jsonb_typeof(input) = 'object')
);

SELECT create_hypertable(
    'device_service_call_events',
    'occurred_at',
    if_not_exists => TRUE
);

CREATE INDEX IF NOT EXISTS idx_device_service_call_events_device_time
    ON device_service_call_events (
        tenant_id,
        product_key,
        device_slug,
        service_name,
        occurred_at DESC
    );

CREATE INDEX IF NOT EXISTS idx_device_service_call_events_requested_by_time
    ON device_service_call_events (
        requested_by,
        occurred_at DESC
    );

CREATE INDEX IF NOT EXISTS idx_device_service_call_events_input_gin
    ON device_service_call_events
    USING gin (input);

DO $$
BEGIN
    RAISE NOTICE 'Initialized table: device_service_call_events';
END $$;
