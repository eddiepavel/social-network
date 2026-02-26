# ---- Stage 1: Build ----
FROM golang:1.24-alpine AS builder

WORKDIR /build

# Install build dependencies for CGO (required by go-sqlite3)
RUN apk add --no-cache gcc musl-dev

# Cache Go module downloads
COPY go.mod go.sum ./
RUN go mod download

# Copy full source
COPY . .

# Accept build args for production config
ARG PRODUCTION=false
ARG ALLOW_ORIGIN=http://localhost:3000

# Generate .env from .env.example with a random SECRET_SIGN and production config
RUN cp .env.example .env \
    && SECRET=$(head -c 48 /dev/urandom | base64 | tr -dc 'a-zA-Z0-9' | head -c 64) \
    && sed -i "s/^SECRET_SIGN=.*/SECRET_SIGN=${SECRET}/" .env \
    && sed -i "s/^PRODUCTION=.*/PRODUCTION=${PRODUCTION}/" .env \
    && sed -i "s|^ALLOW_ORIGIN=.*|ALLOW_ORIGIN=${ALLOW_ORIGIN}|" .env

# Build the Go binary with CGO enabled (needed for sqlite3)
RUN CGO_ENABLED=1 GOOS=linux go build -ldflags="-s -w" -o server ./cmd/server

# ---- Stage 2: Runtime ----
FROM alpine:3.19

WORKDIR /app

# Install minimal runtime dependencies
RUN apk add --no-cache ca-certificates

# Copy compiled binary
COPY --from=builder /build/server ./server

# Copy generated .env
COPY --from=builder /build/.env ./.env

# Copy the full pkg directory (compiled queries are in the binary,
# but migrations are read from disk at startup)
COPY --from=builder /build/pkg ./pkg

# Keep a seed copy of the migrations so they survive volume mounts
RUN cp -r /app/pkg/db/migrations /app/migrations-seed

# Create the empty storage directory (app creates uploads/ and public/ inside)
RUN mkdir -p /app/storage

# Entrypoint: restore migrations into the volume, then start the server
RUN printf '#!/bin/sh\nset -e\nmkdir -p /app/pkg/db/migrations\ncp -r /app/migrations-seed/* /app/pkg/db/migrations/\nexec ./server\n' \
    > /app/entrypoint.sh \
    && chmod +x /app/entrypoint.sh

EXPOSE 8000

ENTRYPOINT ["/app/entrypoint.sh"]
