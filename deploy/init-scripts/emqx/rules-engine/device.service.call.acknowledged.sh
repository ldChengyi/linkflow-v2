# rule for: device.service.call.acknowledged
# topic: lf/v1/{tenant_slug}/{product_key}/{device_slug}/service/up/{service_name}_reply
#
# Device payload shape:
#   {"success":true,"code":"ok","message":"done","output":{}}

RULE_ID="lf-device-service-call-ack"
RULE_NAME="device.service.call.acknowledged"

RULE_SQL=$(cat <<'SQL'
SELECT
  json_decode(payload) as payload.raw,
  regex_replace(nth(8, tokens(topic, '/')), '_reply$', '') as payload.service_name,
  payload.success as payload.success,
  payload.code as payload.code,
  payload.message as payload.message,
  payload.output as payload.output,
  uuid_v4() as event_id,
  'device.service.call.acknowledged' as event_type,
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
FROM "lf/v1/+/+/+/service/up/+"
WHERE regex_match(nth(8, tokens(topic, '/')), '^[a-z][a-z0-9_]*_reply$')
SQL
)

register_rule "${RULE_ID}" "${RULE_NAME}" "${RULE_SQL}" device_events
