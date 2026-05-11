#!/usr/bin/env sh
set -eu

exec tredis --host 127.0.0.1 --port 6379 --db 0
