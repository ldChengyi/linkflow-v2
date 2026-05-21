#!/usr/bin/env sh
set -eu

# ---------- env ----------
EMQX_API_URL="${EMQX_API_URL:-http://emqx:18083}"
EMQX_API_KEY="${EMQX_API_KEY:-linkflow-init}"
EMQX_API_SECRET="${EMQX_API_SECRET:-linkflow-init-secret}"
BACKEND_AUTH_URL="${BACKEND_AUTH_URL:-http://backend:18080/internal/emqx/auth}"
KAFKA_BOOTSTRAP="${KAFKA_BOOTSTRAP:-redpanda:9092}"

CONNECTOR_NAME="linkflow_kafka"

# Kafka 拓扑声明：以后新增 topic 在这里加一行
# 格式: <action_name>|<kafka_topic>
KAFKA_ACTIONS="
device_events|lf.v1.device.events
"
# 例: 后续要加
# device_service|lf.v1.device.service
# audit_log|lf.v1.audit.log

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
RULES_DIR="${SCRIPT_DIR}/rules-engine"

AUTH_HEADER="-u ${EMQX_API_KEY}:${EMQX_API_SECRET}"
CT_HEADER="-H Content-Type:application/json"

# ---------- helpers ----------
wait_for_api() {
  echo "Waiting for EMQX API at ${EMQX_API_URL}"
  i=0
  while [ $i -lt 60 ]; do
    if curl -fsS $AUTH_HEADER "${EMQX_API_URL}/api/v5/nodes" >/dev/null 2>&1; then
      return 0
    fi
    i=$((i+1)); sleep 2
  done
  echo "EMQX API not ready" >&2; exit 1
}

# upsert <api-path> <resource-id-for-PUT> <json-payload>
upsert() {
  path="$1"; id="$2"; payload="$3"
  status=$(curl -sS -o /tmp/emqx-resp.json -w '%{http_code}' \
    $AUTH_HEADER $CT_HEADER -X POST "${EMQX_API_URL}${path}" -d "${payload}")
  case "${status}" in
    200|201|204)
      echo "  created: ${id}"; return 0 ;;
    400|409)
      # EMQX 5 的 connectors/actions PUT 接口要求 identity 由 URL 决定，body 里不能再带 name/type
      case "${path}" in
        */connectors|*/actions)
          put_payload=$(printf '%s' "${payload}" | jq 'del(.name, .type)') ;;
        *)
          put_payload="${payload}" ;;
      esac
      ustatus=$(curl -sS -o /tmp/emqx-resp.json -w '%{http_code}' \
        $AUTH_HEADER $CT_HEADER -X PUT "${EMQX_API_URL}${path}/${id}" -d "${put_payload}")
      case "${ustatus}" in
        200|201|204) echo "  updated: ${id}"; return 0 ;;
        *) echo "update failed ${id}: HTTP ${ustatus}" >&2
           cat /tmp/emqx-resp.json >&2; exit 1 ;;
      esac ;;
    *) echo "create failed ${id}: HTTP ${status}" >&2
       cat /tmp/emqx-resp.json >&2; exit 1 ;;
  esac
}

register_kafka_action() {
  aname="$1"; topic="$2"
  payload=$(cat <<EOF
{
  "type": "kafka_producer",
  "name": "${aname}",
  "connector": "${CONNECTOR_NAME}",
  "parameters": {
    "topic": "${topic}",
    "message": { "key": "\${.tenant_id}", "value": "\${.}" },
    "partition_strategy": "key_dispatch",
    "required_acks": "all_isr"
  }
}
EOF
)
  echo "Upserting Kafka action: ${aname} -> ${topic}"
  upsert "/api/v5/actions" "kafka_producer:${aname}" "${payload}"
}

# register_rule <id> <name> <sql> <action1> [action2 ...]
register_rule() {
  rid="$1"; rname="$2"; rsql="$3"; shift 3
  if [ $# -eq 0 ]; then
    echo "register_rule: ${rid} 缺少 action 参数" >&2; exit 1
  fi
  actions_json="["
  first=1
  for a in "$@"; do
    [ $first -eq 1 ] || actions_json="${actions_json},"
    actions_json="${actions_json}\"kafka_producer:${a}\""
    first=0
  done
  actions_json="${actions_json}]"

  rsql_json=$(printf '%s' "$rsql" | jq -Rs .)
  payload=$(cat <<EOF
{
  "id": "${rid}",
  "name": "${rname}",
  "sql": ${rsql_json},
  "actions": ${actions_json},
  "enable": true
}
EOF
)
  echo "Upserting rule: ${rid} -> ${actions_json}"
  upsert "/api/v5/rules" "${rid}" "${payload}"
}

# 暴露给子脚本（POSIX sh 不支持 export -f，忽略错误即可；子脚本是 . source 进来的，函数自然可见）
export CONNECTOR_NAME KAFKA_BOOTSTRAP EMQX_API_URL EMQX_API_KEY EMQX_API_SECRET

# ---------- 0. 等 EMQX ----------
wait_for_api

# ---------- 1. HTTP authenticator ----------
echo "Upserting HTTP authenticator"
auth_payload=$(cat <<EOF
{
  "mechanism": "password_based",
  "backend": "http",
  "method": "post",
  "url": "${BACKEND_AUTH_URL}",
  "headers": { "Content-Type": "application/json" },
  "body": {
    "username": "\${username}",
    "password": "\${password}",
    "clientid": "\${clientid}"
  }
}
EOF
)
upsert "/api/v5/authentication" "password_based%3Ahttp" "${auth_payload}"

# ---------- 1.5 Authorization (file ACL, 租户隔离) ----------
# 服务端账号 (linkflow-mqtt-gateway) 在 backend 认证响应里是 superuser，自动绕过 ACL。
# 普通设备需要带 client_attrs.{tenant_slug,product_key,device_slug} 才能 publish/subscribe 自己域内的 topic。
echo "Upserting authorization file source"
acl_rules=$(cat <<'ACL'
%% 保留 EMQX 默认的运维/安全规则
{allow, {username, {re, "^dashboard$"}}, subscribe, ["$SYS/#"]}.
{allow, {ipaddr, "127.0.0.1"}, all, ["$SYS/#", "#"]}.
{deny, all, subscribe, ["$SYS/#", {eq, "#"}, {eq, "+/#"}]}.

%% 设备只能 publish 自己的上行 topic
{allow, all, publish, ["lf/v1/${client_attrs.tenant_slug}/${client_attrs.product_key}/${client_attrs.device_slug}/property/up/+"]}.

%% 设备只能订阅自己的下行 topic
{allow, all, subscribe, ["lf/v1/${client_attrs.tenant_slug}/${client_attrs.product_key}/${client_attrs.device_slug}/property/down/+"]}.

%% 兜底拒绝
{deny, all}.
ACL
)
acl_payload=$(jq -nc --arg rules "${acl_rules}" '{type:"file", enable:true, rules:$rules}')
astatus=$(curl -sS -o /tmp/emqx-resp.json -w '%{http_code}' \
  $AUTH_HEADER $CT_HEADER -X PUT "${EMQX_API_URL}/api/v5/authorization/sources/file" -d "${acl_payload}")
case "${astatus}" in
  200|201|204) echo "  updated: authorization/sources/file" ;;
  *) echo "ACL update failed: HTTP ${astatus}" >&2
     cat /tmp/emqx-resp.json >&2; exit 1 ;;
esac

# 默认无匹配 = 拒绝，避免 client_attrs 缺失时意外放行
echo "Setting authorization no_match=deny"
settings_payload='{"no_match":"deny","deny_action":"ignore","cache":{"enable":true,"max_size":32,"ttl":"1m"}}'
sstatus=$(curl -sS -o /tmp/emqx-resp.json -w '%{http_code}' \
  $AUTH_HEADER $CT_HEADER -X PUT "${EMQX_API_URL}/api/v5/authorization/settings" -d "${settings_payload}")
case "${sstatus}" in
  200|201|204) echo "  updated: authorization/settings" ;;
  *) echo "authorization settings update failed: HTTP ${sstatus}" >&2
     cat /tmp/emqx-resp.json >&2; exit 1 ;;
esac

# ---------- 2. Kafka Connector (单实例, 所有 action 共享) ----------
echo "Upserting Kafka connector: ${CONNECTOR_NAME}"
connector_payload=$(cat <<EOF
{
  "type": "kafka_producer",
  "name": "${CONNECTOR_NAME}",
  "bootstrap_hosts": "${KAFKA_BOOTSTRAP}",
  "ssl": { "enable": false },
  "authentication": "none",
  "connect_timeout": "5s"
}
EOF
)
upsert "/api/v5/connectors" "kafka_producer:${CONNECTOR_NAME}" "${connector_payload}"

# ---------- 3. Kafka Actions (一个 topic 一个 action) ----------
echo "${KAFKA_ACTIONS}" | while IFS='|' read -r aname topic; do
  aname=$(printf '%s' "$aname" | tr -d ' \t')
  topic=$(printf '%s' "$topic" | tr -d ' \t')
  [ -z "$aname" ] && continue
  [ -z "$topic" ] && { echo "action ${aname} 缺少 topic" >&2; exit 1; }
  register_kafka_action "$aname" "$topic"
done

# ---------- 4. 加载所有 rule 子脚本 ----------
if [ -d "${RULES_DIR}" ]; then
  for f in "${RULES_DIR}"/*.sh; do
    [ -f "$f" ] || continue
    echo "Loading rule script: $(basename "$f")"
    # shellcheck disable=SC1090
    . "$f"
  done
fi

echo "EMQX bootstrap done."