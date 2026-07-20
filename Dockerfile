# ── Stage 1: build ────────────────────────────────────────────────────────────
FROM golang:1.23-alpine AS builder

WORKDIR /app

# Cache dependency downloads separately from source changes.
COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /server ./cmd/server

# ── Stage 2: runtime ──────────────────────────────────────────────────────────
FROM alpine:3.20

# Install CA certificates so outbound HTTPS calls work.
RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app

COPY --from=builder /server .

EXPOSE 8080

CMD ["/app/server"]
