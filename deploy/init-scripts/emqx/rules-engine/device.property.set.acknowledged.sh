# rule for: device.property.set.acknowledged
# topic: lf/v1/{tenant_slug}/{product_key}/{device_slug}/property/up/set_reply
#
# Device payload shape:
#   {"command_id":"018f56d3-7cb7-7f1a-9b41-3f3a63fd3db8","success":true,"code":"ok","message":"applied","properties":{"led_on":true}}

RULE_ID="lf-device-property-set-ack"
RULE_NAME="device.property.set.acknowledged"

RULE_SQL=$(cat <<'SQL'
SELECT
  json_decode(payload) as payload.raw,
  payload.command_id as payload.command_id,
  payload.success as payload.success,
  payload.code as payload.code,
  payload.message as payload.message,
  payload.properties as payload.properties,
  uuid_v4() as event_id,
  'device.property.set.acknowledged' as event_type,
  1 as event_version,
  format_date('millisecond', '+00:00', '%Y-%m-%dT%H:%M:%S.%3NZ', now_timestamp('millisecond')) as occurred_at,
  'emqx-rule-engine' as producer,
  payload.command_id as causation_id,
  client_attrs.tenant_id as tenant_id,
  client_attrs.tenant_id as payload.tenant_id,
  client_attrs.product_id as payload.product_id,
  client_attrs.device_id as payload.device_id,
  nth(3, tokens(topic, '/')) as payload.tenant_slug,
  nth(4, tokens(topic, '/')) as payload.product_key,
  nth(5, tokens(topic, '/')) as payload.device_slug,
  'mqtt' as payload.protocol
FROM "lf/v1/+/+/+/property/up/set_reply"
SQL
)

register_rule "${RULE_ID}" "${RULE_NAME}" "${RULE_SQL}" device_events
