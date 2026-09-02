# syntax=docker/dockerfile:1

# --- Build ------------------------------------------------------------------
FROM golang:1.26-alpine AS build

WORKDIR /src

# Cache dependencies separately from source. This also fetches the templ and
# air tool dependencies recorded in go.mod's tool directives.
COPY go.mod go.sum ./
RUN go mod download

COPY . .

# .dockerignore excludes *_templ.go, so the generated code is always built
# fresh from the .templ sources rather than trusting whatever was committed.
RUN go tool templ generate

# CGO_ENABLED=0 works because the SQLite driver (modernc.org/sqlite) is pure
# Go. That keeps the runtime image free of a libc/toolchain dependency.
RUN CGO_ENABLED=0 GOOS=linux go build \
        -trimpath \
        -ldflags="-s -w" \
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
