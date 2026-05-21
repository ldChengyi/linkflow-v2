#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
COMPOSE_FILE="${ROOT_DIR}/deploy/docker-compose.yml"

case "${1:-}" in
  --drop-volumes)
    docker compose -f "${COMPOSE_FILE}" down -v --remove-orphans
    ;;
  --keep-volumes)
    docker compose -f "${COMPOSE_FILE}" down --remove-orphans
    ;;
  *)
    echo "usage: $0 --drop-volumes|--keep-volumes" >&2
    exit 2
    ;;
esac

docker compose -f "${COMPOSE_FILE}" up -d --build --force-recreate
