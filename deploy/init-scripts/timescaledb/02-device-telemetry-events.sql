-- ============================================================
-- LinkFlow v2 - device.telemetry.received storage
-- ============================================================
-- Stores telemetry facts consumed from Kafka topic:
--   lf.v1.device.events
--
-- Source event:
--   event_type:    device.telemetry.received
--   event_version: 1
--
-- The event producer owns event_id. This table keeps the producer's
-- event_id for idempotency and traceability.
-- ============================================================

CREATE TABLE IF NOT EXISTS device_telemetry_events (
    event_id uuid NOT NULL,
    tenant_id text NOT NULL,
    product_key text NOT NULL,
    device_id text NOT NULL,
    protocol text NOT NULL,
    occurred_at timestamptz NOT NULL,
    received_at timestamptz NOT NULL DEFAULT now(),
    producer text NOT NULL,
    trace_id text,
    correlation_id text,
    causation_id uuid,
    metrics jsonb NOT NULL,
    raw jsonb,

    PRIMARY KEY (event_id, occurred_at),
    CONSTRAINT device_telemetry_events_metrics_object
        CHECK (jsonb_typeof(metrics) = 'object'),
    CONSTRAINT device_telemetry_events_raw_object
        CHECK (raw IS NULL OR jsonb_typeof(raw) = 'object')
);

SELECT create_hypertable(
    'device_telemetry_events',
    'occurred_at',
    if_not_exists => TRUE
);

CREATE INDEX IF NOT EXISTS idx_device_telemetry_events_device_time
    ON device_telemetry_events (
        tenant_id,
        product_key,
        device_id,
        occurred_at DESC
    );

CREATE INDEX IF NOT EXISTS idx_device_telemetry_events_metrics_gin
    ON device_telemetry_events
    USING gin (metrics);

DO $$
BEGIN
    RAISE NOTICE 'Initialized table: device_telemetry_events';
END $$;
