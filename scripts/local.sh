#!/usr/bin/env bash
# Launch the site locally with hot reload (templ regen + Go rebuild + browser reload).
#
#   ./scripts/local.sh   →   http://localhost:8383
#
# templ and air are pinned as Go tool dependencies in go.mod, so this works on a
# clean machine with nothing but the Go toolchain installed.
set -euo pipefail
cd "$(dirname "$0")/.."

if [ ! -f .env ]; then
  cp .env.example .env
  echo "==> Created .env from .env.example — edit the passwords before sharing anything."
fi

# A blank SESSION_SECRET is a hard startup failure, so fill it in rather than
# making every new checkout debug the same crash.
if ! grep -qE '^SESSION_SECRET=.{32,}' .env; then
  secret="$(openssl rand -hex 32)"
  # Replace the existing (blank) line if present, otherwise append.
  if grep -qE '^SESSION_SECRET=' .env; then
    tmp="$(mktemp)"
    sed "s|^SESSION_SECRET=.*|SESSION_SECRET=${secret}|" .env > "$tmp" && mv "$tmp" .env
  else
    printf 'SESSION_SECRET=%s\n' "$secret" >> .env
  fi
  echo "==> Generated SESSION_SECRET in .env"
fi

if ! grep -qE '^MICROPUB_TOKEN=.{32,}' .env; then
  token="$(openssl rand -hex 32)"
  if grep -qE '^MICROPUB_TOKEN=' .env; then
    tmp="$(mktemp)"
    sed "s|^MICROPUB_TOKEN=.*|MICROPUB_TOKEN=${token}|" .env > "$tmp" && mv "$tmp" .env
  else
    printf 'MICROPUB_TOKEN=%s\n' "$token" >> .env
  fi
  echo "==> Generated MICROPUB_TOKEN in .env"
fi

mkdir -p data tmp
go mod download
go tool templ generate

echo "==> http://localhost:8383  (app on :8080, air proxy adds live reload)"
exec go tool air -c .air.toml
