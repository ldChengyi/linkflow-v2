#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
SQL_DIR="${ROOT_DIR}/deploy/init-scripts/timescaledb"

CONTAINER="${TIMESCALEDB_CONTAINER:-linkflow-timescaledb}"
DB_USER="${LINKFLOW_DB_USER:-linkflow}"
DB_NAME="${LINKFLOW_DB_NAME:-linkflow}"

if [[ ! -d "${SQL_DIR}" ]]; then
    echo "TimescaleDB SQL directory not found: ${SQL_DIR}" >&2
    exit 1
fi

shopt -s nullglob
sql_files=("${SQL_DIR}"/*.sql)
shopt -u nullglob

if [[ ${#sql_files[@]} -eq 0 ]]; then
    echo "No SQL files found in: ${SQL_DIR}" >&2
    exit 1
fi

for sql_file in "${sql_files[@]}"; do
    echo "Applying $(basename "${sql_file}")"
    docker exec -i "${CONTAINER}" \
        psql -v ON_ERROR_STOP=1 -U "${DB_USER}" -d "${DB_NAME}" \
        < "${sql_file}"
done

echo "TimescaleDB schema apply complete."
