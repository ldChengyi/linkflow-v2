# rule for: device.property.reported
# topic: lf/v1/{tenant_slug}/{product_key}/{device_slug}/property/up/post

RULE_ID="lf-device-property-reported"
RULE_NAME="device.property.reported"

RULE_SQL=$(cat <<'SQL'
SELECT
  json_decode(payload) as payload.properties,
  uuid_v4() as event_id,
  'device.property.reported' as event_type,
  1 as event_version,
  format_date('millisecond', '+00:00', '%Y-%m-%dT%H:%M:%S.%3NZ', now_timestamp('millisecond')) as occurred_at,
  'emqx-rule-engine' as producer,
  client_attrs.tenant_id as tenant_id,
  client_attrs.tenant_id as payload.tenant_id,
  client_attrs.product_id as payload.product_id,
  client_attrs.device_id as payload.device_id,
  nth(3, tokens(topic, '/')) as payload.tenant_slug,
  nth(4, tokens(topic, '/')) as payload.product_key,
  nth(5, tokens(topic, '/')) as payload.device_slug,
  'mqtt' as payload.protocol
FROM "lf/v1/+/+/+/property/up/post"
SQL
)

register_rule "${RULE_ID}" "${RULE_NAME}" "${RULE_SQL}" device_events