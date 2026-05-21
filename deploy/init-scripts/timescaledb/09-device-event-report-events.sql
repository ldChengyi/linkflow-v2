-- ============================================================
-- LinkFlow v2 - device.event.reported storage
-- ============================================================
-- Stores device business events consumed from Kafka topic:
--   lf.v1.device.events
--
-- Source event:
--   event_type:    device.event.reported
--   event_version: 1
--
-- The event producer owns event_id. This table keeps the producer's
-- event_id for idempotency and traceability.
-- ============================================================

CREATE TABLE IF NOT EXISTS device_event_report_events (
    event_id uuid NOT NULL,
    tenant_id text NOT NULL,
    product_key text NOT NULL,
    device_slug text NOT NULL,
    event_name text NOT NULL,
    protocol text NOT NULL,
    occurred_at timestamptz NOT NULL,
    received_at timestamptz NOT NULL DEFAULT now(),
    producer text NOT NULL,
    trace_id text,
    correlation_id text,
    causation_id uuid,
    params jsonb NOT NULL,
    raw jsonb,

    PRIMARY KEY (event_id, occurred_at),
    CONSTRAINT device_event_report_events_params_object
        CHECK (jsonb_typeof(params) = 'object'),
    CONSTRAINT device_event_report_events_raw_object
        CHECK (raw IS NULL OR jsonb_typeof(raw) = 'object')
);

SELECT create_hypertable(
    'device_event_report_events',
    'occurred_at',
    if_not_exists => TRUE
);

CREATE INDEX IF NOT EXISTS idx_device_event_report_events_device_time
    ON device_event_report_events (
        tenant_id,
        product_key,
        device_slug,
        event_name,
        occurred_at DESC
    );

CREATE INDEX IF NOT EXISTS idx_device_event_report_events_params_gin
    ON device_event_report_events
    USING gin (params);

DO $$
BEGIN
    RAISE NOTICE 'Initialized table: device_event_report_events';
END $$;
