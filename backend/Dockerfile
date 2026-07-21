# syntax=docker/dockerfile:1

# ---------------------------------------------------------------------
# Stage 1: build a static binary
# ---------------------------------------------------------------------
FROM golang:1.22-alpine AS builder

WORKDIR /src

# Cache dependencies separately from source for faster rebuilds.
# (go.sum is generated here if the repo doesn't already ship one.)
COPY go.mod ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /out/task-manager-api ./cmd/api

# ---------------------------------------------------------------------
# Stage 2: minimal runtime image
# ---------------------------------------------------------------------
FROM alpine:3.20

RUN apk add --no-cache ca-certificates && \
    adduser -D -u 10001 appuser

WORKDIR /app
COPY --from=builder /out/task-manager-api ./task-manager-api

USER appuser

EXPOSE 8080

HEALTHCHECK --interval=15s --timeout=3s --start-period=10s --retries=5 \
  CMD wget -qO- http://127.0.0.1:8080/health || exit 1

ENTRYPOINT ["./task-manager-api"]
