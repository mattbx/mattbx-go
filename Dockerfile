# syntax=docker/dockerfile:1

# --- Build ------------------------------------------------------------------
FROM golang:1.26-alpine AS build

WORKDIR /src

# Cache dependencies separately from source. This also fetches the templ and
# air tool dependencies recorded in go.mod's tool directives.
COPY go.mod go.sum ./
RUN go mod download

# Tailwind standalone CLI (no Node). The musl build is required on Alpine, and
# it is dynamically linked against the C++ runtime, which golang:alpine lacks
# (build stage only; the runtime image is unaffected). Keep TAILWIND_VERSION in
# step with the version used locally (`tailwindcss --help` prints it). Fetched
# before the source copy so it stays cached.
RUN apk add --no-cache libstdc++ libgcc
ARG TAILWIND_VERSION=4.3.3
ARG TARGETARCH
RUN case "$TARGETARCH" in \
      amd64) tw=tailwindcss-linux-x64-musl ;; \
      arm64) tw=tailwindcss-linux-arm64-musl ;; \
      *) echo "unsupported arch: $TARGETARCH" >&2; exit 1 ;; \
    esac \
 && wget -qO /usr/local/bin/tailwindcss \
      "https://github.com/tailwindlabs/tailwindcss/releases/download/v${TAILWIND_VERSION}/${tw}" \
 && chmod +x /usr/local/bin/tailwindcss

COPY . .

# .dockerignore excludes *_templ.go, so the generated code is always built
# fresh from the .templ sources rather than trusting whatever was committed.
RUN go tool templ generate

# Compiled before `go build` because the stylesheet is embedded in the binary.
RUN tailwindcss -i internal/ui/tailwind/input.css -o internal/ui/static/tailwind.css --minify

# CGO_ENABLED=0 works because the SQLite driver (modernc.org/sqlite) is pure
# Go. That keeps the runtime image free of a libc/toolchain dependency.
#
# The build date is computed here, inside the build step, rather than passed
# in from outside: .dockerignore excludes .git, so there's no commit to read,
# and this way nothing needs configuring in Disco (or wherever else this ever
# gets built) for the date to be correct.
RUN CGO_ENABLED=0 GOOS=linux go build \
        -trimpath \
        -ldflags="-s -w -X github.com/mattbx/mattbx-go/internal/build.Date=$(date -u +%Y-%m-%d)" \
        -o /out/server ./cmd/web

# --- Runtime ----------------------------------------------------------------
# Alpine rather than distroless/scratch for two concrete reasons: Disco's
# health check runs a shell command inside the container (busybox wget must
# exist), and running as root avoids ownership problems on a freshly created
# Disco volume mounted at /data.
FROM alpine:3.21

RUN apk add --no-cache ca-certificates tzdata

COPY --from=build /out/server /server

ENV PORT=8080 \
    ENV=production \
    DB_PATH=/data/app.db

EXPOSE 8080

# The rest of the configuration (passwords, SESSION_SECRET, BASE_URL) comes
# from `disco env:set` and is deliberately absent from the image.
ENTRYPOINT ["/server"]
