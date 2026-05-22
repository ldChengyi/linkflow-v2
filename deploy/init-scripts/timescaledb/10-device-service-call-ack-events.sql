-- ============================================================
-- LinkFlow v2 - device.service.call.acknowledged storage
-- ============================================================
-- Stores device service-call acknowledgements consumed from Kafka topic:
--   lf.v1.device.events
--
-- Source event:
--   event_type:    device.service.call.acknowledged
--   event_version: 1
--
-- The event producer owns event_id. This table keeps the producer's
-- event_id for idempotency and traceability.
-- ============================================================

CREATE TABLE IF NOT EXISTS device_service_call_ack_events (
    event_id uuid NOT NULL,
    command_id uuid NOT NULL,
    tenant_id text NOT NULL,
    product_key text NOT NULL,
    device_slug text NOT NULL,
    service_name text NOT NULL,
    protocol text NOT NULL,
    success boolean NOT NULL,
    code text,
    message text,
    occurred_at timestamptz NOT NULL,
    received_at timestamptz NOT NULL DEFAULT now(),
    producer text NOT NULL,
    trace_id text,
    correlation_id text,
    causation_id uuid,
    output jsonb NOT NULL,
    raw jsonb,

    PRIMARY KEY (event_id, occurred_at),
    CONSTRAINT device_service_call_ack_events_output_object
        CHECK (jsonb_typeof(output) = 'object'),
    CONSTRAINT device_service_call_ack_events_raw_object
        CHECK (raw IS NULL OR jsonb_typeof(raw) = 'object')
);

SELECT create_hypertable(
    'device_service_call_ack_events',
    'occurred_at',
    if_not_exists => TRUE
);

CREATE INDEX IF NOT EXISTS idx_device_service_call_ack_events_device_time
    ON device_service_call_ack_events (
        tenant_id,
        product_key,
        device_slug,
        service_name,
        occurred_at DESC
    );

CREATE INDEX IF NOT EXISTS idx_device_service_call_ack_events_command_id_time
    ON device_service_call_ack_events (
        command_id,
        occurred_at DESC
    );

CREATE INDEX IF NOT EXISTS idx_device_service_call_ack_events_output_gin
    ON device_service_call_ack_events
    USING gin (output);

DO $$
BEGIN
    RAISE NOTICE 'Initialized table: device_service_call_ack_events';
END $$;
