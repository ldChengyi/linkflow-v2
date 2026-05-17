#!/usr/bin/env sh
set -eu

EMQX_API_URL="${EMQX_API_URL:-http://emqx:18083}"
EMQX_API_KEY="${EMQX_API_KEY:-linkflow-init}"
EMQX_API_SECRET="${EMQX_API_SECRET:-linkflow-init-secret}"
BACKEND_AUTH_URL="${BACKEND_AUTH_URL:-http://host.docker.internal:18080/internal/emqx/auth}"
AUTH_ID="password_based%3Ahttp"

payload="$(cat <<EOF
{
  "mechanism": "password_based",
  "backend": "http",
  "method": "post",
  "url": "${BACKEND_AUTH_URL}",
  "headers": {
    "Content-Type": "application/json"
  },
  "body": {
    "username": "\${username}",
    "password": "\${password}",
    "clientid": "\${clientid}"
  }
}
EOF
)"

echo "Waiting for EMQX API at ${EMQX_API_URL}"
for i in $(seq 1 60); do
  if curl -fsS -u "${EMQX_API_KEY}:${EMQX_API_SECRET}" "${EMQX_API_URL}/api/v5/nodes" >/dev/null; then
    break
  fi
  if [ "$i" = "60" ]; then
    echo "EMQX API did not become ready" >&2
    exit 1
  fi
  sleep 2
done

echo "Creating EMQX HTTP password authenticator"
status="$(
  curl -sS -o /tmp/emqx-auth-create-response.json -w '%{http_code}' \
    -u "${EMQX_API_KEY}:${EMQX_API_SECRET}" \
    -H 'Content-Type: application/json' \
    -X POST \
    "${EMQX_API_URL}/api/v5/authentication" \
    -d "${payload}"
)"

case "${status}" in
  200|201|204)
    echo "EMQX HTTP authenticator created"
    ;;
  409)
    echo "EMQX HTTP authenticator already exists; updating"
    update_status="$(
      curl -sS -o /tmp/emqx-auth-update-response.json -w '%{http_code}' \
        -u "${EMQX_API_KEY}:${EMQX_API_SECRET}" \
        -H 'Content-Type: application/json' \
        -X PUT \
        "${EMQX_API_URL}/api/v5/authentication/${AUTH_ID}" \
        -d "${payload}"
    )"
    case "${update_status}" in
      200|201|204)
        echo "EMQX HTTP authenticator updated"
        ;;
      *)
        echo "Failed to update EMQX HTTP authenticator: HTTP ${update_status}" >&2
        cat /tmp/emqx-auth-update-response.json >&2 || true
        exit 1
        ;;
    esac
    ;;
  *)
    echo "Failed to create EMQX HTTP authenticator: HTTP ${status}" >&2
    cat /tmp/emqx-auth-create-response.json >&2 || true
    exit 1
    ;;
esac
