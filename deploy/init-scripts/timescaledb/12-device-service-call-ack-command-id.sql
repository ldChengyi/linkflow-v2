-- ============================================================
-- LinkFlow v2 - device service-call ACK command correlation
-- ============================================================
-- Adds command_id to existing ACK tables created before the service-call
-- correlation contract required devices to echo command_id.
-- ============================================================

ALTER TABLE device_service_call_ack_events
    ADD COLUMN IF NOT EXISTS command_id uuid;

CREATE INDEX IF NOT EXISTS idx_device_service_call_ack_events_command_id_time
    ON device_service_call_ack_events (
        command_id,
        occurred_at DESC
    );

DO $$
BEGIN
    RAISE NOTICE 'Ensured column: device_service_call_ack_events.command_id';
END $$;
