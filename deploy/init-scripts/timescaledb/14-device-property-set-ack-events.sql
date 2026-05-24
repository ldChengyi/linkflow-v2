-- ============================================================
-- LinkFlow v2 - device.property.set.acknowledged storage
-- ============================================================
-- Stores device property-set acknowledgements consumed from Kafka topic:
--   lf.v1.device.events
--
-- Source event:
--   event_type:    device.property.set.acknowledged
--   event_version: 1
-- ============================================================

CREATE TABLE IF NOT EXISTS device_property_set_ack_events (
    event_id uuid NOT NULL,
    command_id uuid,
    tenant_id text NOT NULL,
    product_key text NOT NULL,
    device_slug text NOT NULL,
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
    properties jsonb NOT NULL,
    raw jsonb,

    PRIMARY KEY (event_id, occurred_at),
    CONSTRAINT device_property_set_ack_events_properties_object
        CHECK (jsonb_typeof(properties) = 'object'),
    CONSTRAINT device_property_set_ack_events_raw_object
        CHECK (raw IS NULL OR jsonb_typeof(raw) = 'object')
);

SELECT create_hypertable(
    'device_property_set_ack_events',
    'occurred_at',
    if_not_exists => TRUE
);

CREATE INDEX IF NOT EXISTS idx_device_property_set_ack_events_device_time
    ON device_property_set_ack_events (
        tenant_id,
        product_key,
        device_slug,
        occurred_at DESC
    );

CREATE INDEX IF NOT EXISTS idx_device_property_set_ack_events_command_id_time
    ON device_property_set_ack_events (
        command_id,
        occurred_at DESC
    );

CREATE INDEX IF NOT EXISTS idx_device_property_set_ack_events_properties_gin
    ON device_property_set_ack_events
    USING gin (properties);

DO $$
BEGIN
    RAISE NOTICE 'Initialized table: device_property_set_ack_events';
END $$;
