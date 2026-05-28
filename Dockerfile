# syntax=docker/dockerfile:1
# ─── Stage 1: Build ───────────────────────────────────────────────────────────
FROM golang:1.26-alpine AS builder

WORKDIR /app

# Install ca-certificates for HTTPS and git for go get
RUN apk add --no-cache git ca-certificates tzdata

# Cache deps separately from source
COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Build a fully static binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags="-s -w -extldflags '-static'" \
    -trimpath \
    -o /bin/server ./cmd/server

# ─── Stage 2: Distroless runtime ─────────────────────────────────────────────
FROM gcr.io/distroless/static-debian12

WORKDIR /

# Copy binary and runtime config (migrations are embedded in the binary)
COPY --from=builder /bin/server /server
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo

EXPOSE 8080

USER nonroot:nonroot

ENTRYPOINT ["/server"]
