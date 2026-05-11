#!/usr/bin/env sh
set -eu

repo_root="$(CDPATH= cd -- "$(dirname -- "$0")/../.." && pwd)"

unset ALL_PROXY
unset all_proxy
unset HTTP_PROXY
unset http_proxy
unset HTTPS_PROXY
unset https_proxy

exec env XDG_CONFIG_HOME=/home/ldchengyi/.local/share/xdg-config \
  NO_PROXY=127.0.0.1,localhost,::1 \
  no_proxy=127.0.0.1,localhost,::1 \
  posting \
  --collection "$repo_root/api/posting/linkflow" \
  --env "$repo_root/api/posting/dev.env"
