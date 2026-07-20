# Postgres Data Summary Service

![CI](https://github.com/RvShivam/postgres-summary-service/actions/workflows/ci.yml/badge.svg)

A clean, modular Golang microservice that fetches database statistics (schemas, table counts, row counts, and storage sizes) from a remote PostgreSQL instance via an external summary API, stores them locally, and exposes the results through a REST API.

---

## Table of Contents

- [Architecture](#architecture)
- [API Reference](#api-reference)
- [Project Structure](#project-structure)
- [Data Model](#data-model)
- [Prerequisites](#prerequisites)
- [Running Locally (Development)](#running-locally-development)
- [Running with Docker Compose](#running-with-docker-compose)
- [Configuration](#configuration)
- [Running Tests](#running-tests)
- [Security Notes](#security-notes)

---

## Architecture

The service follows a strict **Handler → Service → Repository** layered architecture with clean dependency injection throughout.

```
HTTP Request
     │
     ▼
┌──────────┐     ┌──────────┐     ┌────────────┐
│  Handler │────▶│  Service │────▶│ Repository │──▶ Local PostgreSQL
└──────────┘     └──────────┘     └────────────┘
                      │
                      ▼
               ┌─────────────┐
               │  External   │──▶ POST /api/summary
               │   Client    │    (remote summary API)
               └─────────────┘
```

All cross-layer dependencies are **interface-driven**, making each layer independently testable without real infrastructure.

---

## API Reference

### `POST /summary/sync`

Triggers a sync against a remote PostgreSQL instance via the external summary API and persists the result locally.

**Request Body:**
```json
{
  "host":     "remote-db.example.com",
  "port":     5432,
  "user":     "readonly",
  "password": "pass",
  "dbname":   "sample"
}
```

> ⚠️ Passwords are **never stored** — they are forwarded to the external service and discarded.

**Response `201 Created`:**
```json
{
  "ID":                "d290f1ee-...",
  "ExternalSummaryID": "sum123",
  "Host":              "remote-db.example.com",
  "Port":              5432,
  "User":              "readonly",
  "DBName":            "sample",
  "SyncedAt":          "2026-07-20T10:00:00Z",
  "Schemas": [
    {
      "Name": "public",
      "Tables": [
        { "Name": "users",  "RowCount": 1240, "SizeMB": 12.5 },
        { "Name": "orders", "RowCount": 580,  "SizeMB": 8.3  }
      ]
    }
  ]
}
```

| Status | Condition |
|--------|-----------|
| `201 Created` | Sync successful and persisted |
| `400 Bad Request` | Missing or invalid fields in request body |
| `500 Internal Server Error` | External service unreachable or DB write failed |

---

### `GET /summaries`

Returns a list of all previously synced summaries ordered by most recent first.

**Response `200 OK`:**
```json
[
  {
    "ID":                "d290f1ee-...",
    "ExternalSummaryID": "sum123",
    "Host":              "remote-db.example.com",
    "Port":              5432,
    "User":              "readonly",
    "DBName":            "sample",
    "SyncedAt":          "2026-07-20T10:00:00Z"
  }
]
```

---

### `GET /summaries/{id}`

Returns the full summary for a given sync ID, including all schemas and tables.

**Response `200 OK`:** Full summary object (same as sync response above).

| Status | Condition |
|--------|-----------|
| `200 OK` | Summary found |
| `400 Bad Request` | `{id}` is not a valid UUID |
| `404 Not Found` | No summary exists for the given ID |

---

## Project Structure

```
.
├── cmd/
│   └── server/
│       └── main.go               # Entry point — wires all dependencies
├── internal/
│   ├── config/
│   │   └── config.go             # Env-var based configuration loader
│   ├── domain/
│   │   ├── summary.go            # Core Summary aggregate
│   │   ├── summary_overview.go   # Lightweight list DTO
│   │   ├── schema.go
│   │   └── table.go
│   ├── handler/
│   │   ├── handler.go            # Handler struct + constructor
│   │   ├── sync_summary.go       # POST /summary/sync
│   │   ├── list_summaries.go     # GET /summaries
│   │   ├── get_summary.go        # GET /summaries/:id
│   │   └── handler_test.go       # Handler unit tests
│   ├── service/
│   │   ├── service.go            # Service interface + constructor
│   │   ├── sync_summary.go
│   │   ├── list_summaries.go
│   │   ├── get_summary.go
│   │   └── service_test.go       # Service unit tests
│   ├── repository/
│   │   ├── repository.go         # Repository interface
│   │   ├── postgres.go           # DB interface + PostgresRepository struct
│   │   ├── save_summary.go       # Transactional insert
│   │   ├── list_summaries.go
│   │   ├── get_summary.go
│   │   └── repository_test.go    # Repository unit tests (pgxmock)
│   ├── external/
│   │   ├── client.go             # HTTP client w/ retry logic
│   │   ├── dto.go                # External API request/response types
│   │   └── client_test.go        # External client unit tests
│   ├── local/
│   │   └── postgres.go           # pgxpool connection factory
│   ├── mocks/
│   │   ├── external_client.go    # MockExternalClient (testify/mock)
│   │   ├── repository.go         # MockRepository
│   │   └── service.go            # MockService
│   └── utils/                    # Shared utilities (reserved)
├── migrations/
│   ├── 000001_create_tables.up.sql
│   └── 000001_create_tables.down.sql
├── mockserver/
│   └── main.go                   # Local stand-in for the external summary API
├── router/
│   └── router.go                 # Gin route registration
├── Dockerfile                    # Multi-stage build
├── docker-compose.yml            # App + Postgres
├── .env                          # Local dev environment variables (not committed)
└── go.mod
```

---

## Data Model

```sql
summaries          -- one row per sync
  id               UUID PRIMARY KEY
  external_summary_id TEXT
  host             TEXT
  port             INTEGER
  user_name        TEXT
  db_name          TEXT
  synced_at        TIMESTAMPTZ

schemas            -- one row per schema in a summary
  id               UUID PRIMARY KEY
  summary_id       UUID → summaries(id) ON DELETE CASCADE
  name             TEXT

tables             -- one row per table in a schema
  id               UUID PRIMARY KEY
  schema_id        UUID → schemas(id) ON DELETE CASCADE
  name             TEXT
  row_count        BIGINT
  size_mb          DOUBLE PRECISION
```

---

## Prerequisites

| Tool | Version | Purpose |
|------|---------|---------|
| Go | ≥ 1.23 | Build and run |
| Docker + Compose | any recent | Run Postgres locally |
| `psql` | any | Apply migrations |
| `jq` *(optional)* | any | Pretty-print curl output |

---

## Running Locally (Development)

You need **3 terminals**.

### 1. Start Postgres

```bash
docker-compose up postgres -d

# Verify it's ready
docker exec postgres-summary-db pg_isready -U postgres -d summarydb
```

### 2. Apply migrations (once)

```bash
psql "host=localhost port=5433 user=postgres password=postgres dbname=summarydb sslmode=disable" \
  -f migrations/000001_create_tables.up.sql
```

> To roll back: replace `up.sql` with `down.sql`.

### 3. Start the mock external service (Terminal 1)

The real external service (`external-service.local`) doesn't exist in local dev. Run the bundled mock instead:

```bash
go run ./mockserver
# → [mock] external service listening on :9090
```

### 4. Start the service (Terminal 2)

```bash
go run ./cmd/server
# → server listening on :8080
```

### 5. Test the endpoints (Terminal 3)

```bash
# Trigger a sync
curl -s -X POST http://localhost:8080/summary/sync \
  -H "Content-Type: application/json" \
  -d '{"host":"remote-db.example.com","port":5432,"user":"readonly","password":"pass","dbname":"sample"}' \
  | jq .

# List synced summaries
curl -s http://localhost:8080/summaries | jq .

# Get a specific summary (replace with actual ID from sync response)
curl -s http://localhost:8080/summaries/<id> | jq .
```

---

## Running with Docker Compose

To run the entire stack (app + Postgres) in containers:

```bash
# Build and start everything
docker-compose up --build

# Apply migrations against the containerised DB
docker exec postgres-summary-db psql -U postgres -d summarydb \
  -f /dev/stdin < migrations/000001_create_tables.up.sql
```

The app will be available at `http://localhost:8080`.

> **Note:** In docker-compose the `EXTERNAL_SERVICE_URL` is set to `http://external-service.local/api/summary`. Replace this with a reachable URL in production.

---

## Configuration

All configuration is via environment variables. Copy `.env.example` (or create `.env`) with the following keys:

| Variable | Description | Default (docker-compose) |
|----------|-------------|--------------------------|
| `PORT` | HTTP port the service listens on | `8080` |
| `DB_HOST` | Local Postgres host | `postgres` |
| `DB_PORT` | Local Postgres port | `5432` |
| `DB_USER` | Local Postgres user | `postgres` |
| `DB_PASSWORD` | Local Postgres password | `postgres` |
| `DB_NAME` | Local Postgres database name | `summarydb` |
| `EXTERNAL_SERVICE_URL` | Full URL of the external summary API | `http://external-service.local/api/summary` |

---

## Running Tests

No external services or database needed — all dependencies are mocked.

```bash
go test ./... -v
```

**28 tests across 4 packages:**

| Package | Tests | What's covered |
|---------|-------|----------------|
| `internal/external` | 5 | HTTP success, 4xx permanent error, 5xx retry exhaustion, context cancellation, unreachable host |
| `internal/repository` | 8 | `SaveSummary` (success, begin error, insert error), `ListSummaries` (success, empty, query error), `GetSummary` (success w/ nested data, not found) |
| `internal/service` | 7 | `SyncSummary`, `ListSummaries`, `GetSummary` — success and error paths with mocked dependencies |
| `internal/handler` | 8 | All 3 endpoints — valid requests, missing fields, invalid UUID, 404, service errors |

---

## CI Pipeline

The repository includes a GitHub Actions workflow at [`.github/workflows/ci.yml`](.github/workflows/ci.yml) that runs automatically on every push and pull request to `main`.

**Pipeline steps:**

| Step | Command | Purpose |
|------|---------|--------|
| Checkout | `actions/checkout@v4` | Pull the source code |
| Setup Go | `actions/setup-go@v5` | Install Go version from `go.mod`, with module cache |
| Download deps | `go mod download` | Restore cached dependencies |
| Verify deps | `go mod verify` | Ensure `go.sum` is consistent (no tampering) |
| Build | `go build ./...` | Confirm the entire project compiles |
| Vet | `go vet ./...` | Catch common correctness issues |
| Test | `go test ./... -v -race` | Run all 28 unit tests with race detection |

> No database or external service is required in CI — all dependencies are mocked.

---

## Security Notes

- **Passwords are never stored** — the remote DB password is forwarded to the external API and immediately discarded. It does not appear in the local database or logs.
- **Input validation** — all sync request fields are required and validated at the handler layer via `binding:"required"`.
- **Retry safety** — the external client retries only on network errors and 5xx responses (transient). 4xx responses (client errors) surface immediately without retrying.
- **Context propagation** — all DB and HTTP calls respect request context, ensuring proper cancellation on client disconnect or timeout.
