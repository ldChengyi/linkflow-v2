# rule for: device.connection.connected
# source: $events/client_connected

RULE_ID="lf-device-connected"
RULE_NAME="device.connection.connected"

RULE_SQL=$(cat <<'SQL'
SELECT
  uuid_v4() as event_id,
  'device.connection.connected' as event_type,
  1 as event_version,
  format_date('millisecond', '+00:00', '%Y-%m-%dT%H:%M:%S.%3NZ', now_timestamp('millisecond')) as occurred_at,
  'emqx-rule-engine' as producer,
  client_attrs.tenant_id as tenant_id,
  client_attrs.tenant_id as payload.tenant_id,
  client_attrs.product_id as payload.product_id,
  client_attrs.device_id as payload.device_id,
  client_attrs.tenant_slug as payload.tenant_slug,
  client_attrs.product_key as payload.product_key,
  client_attrs.device_slug as payload.device_slug,
  'mqtt' as payload.protocol,
  keepalive as payload.keepalive
FROM "$events/client_connected"
WHERE client_attrs.tenant_id != '' AND client_attrs.device_id != ''
SQL
)

register_rule "${RULE_ID}" "${RULE_NAME}" "${RULE_SQL}" device_events
