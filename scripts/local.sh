#!/usr/bin/env bash
# Launch the site locally with hot reload (templ regen + Go rebuild + browser reload).
#
#   ./scripts/local.sh   →   http://localhost:8383
#
# templ and air are pinned as Go tool dependencies in go.mod, so this works on a
# clean machine with the Go toolchain and the Tailwind standalone CLI at
# TAILWIND_VERSION (same pin as ARG TAILWIND_VERSION in the Dockerfile).
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

# Same pin as ARG TAILWIND_VERSION in the Dockerfile.
TAILWIND_VERSION=4.3.3

if ! command -v tailwindcss >/dev/null 2>&1; then
  echo "==> tailwindcss not found. Install the standalone CLI v${TAILWIND_VERSION} from" >&2
  echo "    https://github.com/tailwindlabs/tailwindcss/releases/tag/v${TAILWIND_VERSION}" >&2
  exit 1
fi

# `tailwindcss --help` prints the version in its header (see Dockerfile).
tw_help="$(tailwindcss --help)"
if ! grep -qE "v${TAILWIND_VERSION}([^0-9]|$)" <<<"$tw_help"; then
  echo "==> tailwindcss must be v${TAILWIND_VERSION} (same pin as the Docker build)." >&2
  echo "    https://github.com/tailwindlabs/tailwindcss/releases/tag/v${TAILWIND_VERSION}" >&2
  exit 1
fi

tw_in=internal/ui/tailwind/input.css
tw_out=internal/ui/static/tailwind.css
tailwindcss -i "$tw_in" -o "$tw_out"

# --watch=always: Tailwind v4 exits watch mode when stdin closes, which is the
# case for a background job. Air notices the rewritten CSS and reloads.
tailwindcss -i "$tw_in" -o "$tw_out" --watch=always &
tailwind_pid=$!
trap 'kill "$tailwind_pid" 2>/dev/null || true' EXIT

echo "==> http://localhost:8383  (app on :8080, air proxy adds live reload)"
go tool air -c .air.toml
