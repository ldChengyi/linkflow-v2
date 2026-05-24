-- ============================================================
-- LinkFlow v2 - cloud-to-device property set storage
-- ============================================================
-- Stores property set commands dispatched by the cloud platform to devices.
--
-- MQTT topic:
--   lf/v1/{tenant_slug}/{product_key}/{device_slug}/property/down/set
--
-- The cloud producer owns command_id. The device acknowledgement is stored
-- separately in device_property_set_ack_events and can correlate through the
-- command_id when devices include it in the reply payload.
-- ============================================================

CREATE TABLE IF NOT EXISTS device_property_set_events (
    command_id uuid NOT NULL,
    tenant_id text NOT NULL,
    product_key text NOT NULL,
    device_slug text NOT NULL,
    protocol text NOT NULL,
    topic text NOT NULL,
    occurred_at timestamptz NOT NULL,
    received_at timestamptz NOT NULL DEFAULT now(),
    producer text NOT NULL,
    requested_by uuid NOT NULL,
    properties jsonb NOT NULL,

    PRIMARY KEY (command_id, occurred_at),
    CONSTRAINT device_property_set_events_properties_object
        CHECK (jsonb_typeof(properties) = 'object')
);

SELECT create_hypertable(
    'device_property_set_events',
    'occurred_at',
    if_not_exists => TRUE
);

CREATE INDEX IF NOT EXISTS idx_device_property_set_events_device_time
    ON device_property_set_events (
        tenant_id,
        product_key,
        device_slug,
        occurred_at DESC
    );

CREATE INDEX IF NOT EXISTS idx_device_property_set_events_requested_by_time
    ON device_property_set_events (
        requested_by,
        occurred_at DESC
    );

CREATE INDEX IF NOT EXISTS idx_device_property_set_events_properties_gin
    ON device_property_set_events
    USING gin (properties);

DO $$
BEGIN
    RAISE NOTICE 'Initialized table: device_property_set_events';
END $$;
