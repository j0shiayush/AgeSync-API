# User API

A production-ready RESTful API built in Go for managing users, featuring dynamic age calculation, structured logging, request tracing, and full containerisation.

---

## Table of Contents

1. [Architecture & Design Choices](#architecture--design-choices)
2. [Prerequisites](#prerequisites)
3. [Running with Docker (recommended)](#running-with-docker-recommended)
4. [Running without Docker (local Postgres)](#running-without-docker-local-postgres)
5. [Running Unit Tests](#running-unit-tests)
6. [SQLC Code Generation](#sqlc-code-generation)
7. [API Reference & cURL Examples](#api-reference--curl-examples)
8. [Environment Variables](#environment-variables)

---

## Architecture & Design Choices

### Directory Layout

```
.
├── cmd/server/main.go          # Entrypoint — wires all layers together
├── config/                     # Typed, env-driven configuration
├── db/
│   ├── migrations/             # Plain SQL migration files (up + down)
│   └── sqlc/                   # SQLC schema, query, and generated Go code
├── internal/
│   ├── handler/                # HTTP transport — parses/validates requests, writes responses
│   ├── repository/             # Database access — wraps SQLC Querier
│   ├── service/                # Business logic — age calculation, orchestration
│   ├── routes/                 # Route & global middleware registration
│   ├── middleware/             # RequestID and Logger middleware
│   ├── models/                 # Domain models and DTO types
│   └── logger/                 # Uber Zap initialisation
├── Dockerfile                  # Multi-stage build (builder → scratch)
├── docker-compose.yml          # PostgreSQL + app, health-checked startup
└── sqlc.yaml                   # SQLC generation config
```

### Layer Responsibilities

| Layer | Responsibility |
|---|---|
| `handler` | HTTP concerns only: body parsing, input validation, status codes |
| `service` | Pure business logic; age calculation; no HTTP or DB knowledge |
| `repository` | Database I/O; adapts between SQLC types and domain types |
| `db/sqlc` | Auto-generated type-safe SQL (never hand-edit these files) |

### Key Design Decisions

**Clean Architecture** — Dependencies point inward. Handlers depend on service interfaces; services depend on repository interfaces. This means the database can be swapped (or mocked in tests) without touching the HTTP layer.

**SQLC** — All SQL is written by hand in `db/sqlc/query.sql` and compiled into type-safe Go code. There is no ORM, no reflection-based query builder, and no runtime SQL construction. The Go compiler catches schema/query mismatches at compile time.

**Interface-driven layers** — Both `UserService` and `UserRepository` are Go interfaces. This enables the unit tests for `CalculateAge` to run without a database, and makes future integration tests straightforward with a mock or an in-memory implementation.

**Age calculation** — Age is computed in the service layer using pure Go `time` arithmetic (`CalculateAge` in `internal/service/user_service.go`). It is never persisted, because it changes every day. The algorithm correctly handles leap-year birthdays.

**Graceful shutdown** — The server listens for `SIGINT`/`SIGTERM`, drains in-flight connections with a 5-second deadline, then exits cleanly — essential for Kubernetes or ECS deployments.

**Multi-stage Docker build** — The final image is built `FROM scratch`. It contains only the statically-linked binary, TLS certificates, and timezone data. The image is typically under 15 MB.

**Structured JSON logging** — Uber Zap is used throughout. In production mode all logs are JSON; in development mode they are coloured and human-readable. Every request log includes the `X-Request-Id` for end-to-end tracing.

---

## Prerequisites

### With Docker

- [Docker](https://docs.docker.com/get-docker/) ≥ 24
- [Docker Compose](https://docs.docker.com/compose/install/) v2 (`docker compose`, not `docker-compose`)

### Without Docker

- Go ≥ 1.22
- PostgreSQL ≥ 14
- (Optional) [sqlc](https://docs.sqlc.dev/en/latest/overview/install.html) ≥ 1.26 if you want to regenerate the DB layer

---

## Running with Docker (recommended)

### 1. Clone and enter the repository

```bash
git clone https://AgeSync-API.git
cd userapi
```

### 2. (Optional) Copy and edit environment overrides

```bash
cp .env.example .env   # edit DB_PASSWORD etc. if desired
```

### 3. Build and start all services

```bash
docker compose up --build
```

Docker Compose will:

1. Start a `postgres:16-alpine` container and wait until it is healthy.
2. Build the Go binary using a two-stage Dockerfile.
3. Start the API server, which connects to Postgres over the internal Docker network.

The API is available at **http://localhost:8080**.

### 4. Stop and remove containers

```bash
docker compose down          # keep the postgres volume
docker compose down -v       # also remove the volume (destroys all data)
```

---

## Running without Docker (local Postgres)

### 1. Create the database and apply the migration

```bash
psql -U postgres -c "CREATE DATABASE userapi;"
psql -U postgres -d userapi -f db/migrations/000001_create_users_table.up.sql
```

### 2. Export environment variables

```bash
export APP_ENV=development
export APP_PORT=8080
export DB_HOST=localhost
export DB_PORT=5432
export DB_USER=postgres
export DB_PASSWORD=secret
export DB_NAME=userapi
export DB_SSLMODE=disable
```

### 3. Download dependencies and run

```bash
go mod download
go run ./cmd/server/main.go
```

The server starts at **http://localhost:8080**.

---

## Running Unit Tests

The unit tests for the age calculation logic are pure (no I/O, no database), and run instantly:

```bash
# Run all tests with the race detector enabled
go test -race ./internal/service/...

# Verbose output (see each sub-test name)
go test -race -v ./internal/service/...

# Run all tests across the whole project
go test -race ./...
```

Expected output:

```
=== RUN   TestCalculateAge
=== RUN   TestCalculateAge/exact_birthday_today_returns_correct_age
=== RUN   TestCalculateAge/birthday_has_already_passed_this_year
...
--- PASS: TestCalculateAge (0.00s)
=== RUN   TestCalculateAge_Idempotent
--- PASS: TestCalculateAge_Idempotent (0.00s)
=== RUN   TestCalculateAge_DoesNotMutateInputs
--- PASS: TestCalculateAge_DoesNotMutateInputs (0.00s)
PASS
ok      AgeSync-API/internal/service
```

---

## SQLC Code Generation

The files under `db/sqlc/*.go` are **auto-generated** from `db/sqlc/query.sql` and `db/sqlc/schema.sql`. You should not edit them by hand.

### Install sqlc

```bash
# macOS
brew install sqlc

# Linux (download latest binary)
curl -fsSL https://releases.sqlc.dev/linux/sqlc_1.26.0_linux_amd64.tar.gz \
  | tar -xz -C /usr/local/bin

# Go install
go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
```

### Regenerate

```bash
sqlc generate
```

This reads `sqlc.yaml` at the project root and overwrites the `db/sqlc/*.go` files. Run this after any change to `query.sql` or `schema.sql`.

---

## API Reference & cURL Examples

> Base URL: `http://localhost:8080`

### Health Check

```bash
curl -s http://localhost:8080/health | jq
# {"status":"ok"}
```

---

### 1. Create a User — `POST /users`

**Request**

```bash
curl -s -X POST http://localhost:8080/users \
  -H "Content-Type: application/json" \
  -d '{"name": "Alice", "dob": "1990-05-10"}' | jq
```

**Response `201 Created`**

```json
{
  "id": 1,
  "name": "Alice",
  "dob": "1990-05-10"
}
```

**Validation error `422 Unprocessable Entity`**

```bash
curl -s -X POST http://localhost:8080/users \
  -H "Content-Type: application/json" \
  -d '{"name": "", "dob": "not-a-date"}' | jq
# {"error":"Name: required validation failed"}
```

---

### 2. Get User by ID — `GET /users/:id`

```bash
curl -s http://localhost:8080/users/1 | jq
```

**Response `200 OK`**

```json
{
  "id": 1,
  "name": "Alice",
  "dob": "1990-05-10",
  "age": 35
}
```

**Not found `404`**

```bash
curl -s http://localhost:8080/users/999 | jq
# {"error":"user not found"}
```

---

### 3. Update User — `PUT /users/:id`

```bash
curl -s -X PUT http://localhost:8080/users/1 \
  -H "Content-Type: application/json" \
  -d '{"name": "Alice Updated", "dob": "1991-03-15"}' | jq
```

**Response `200 OK`**

```json
{
  "id": 1,
  "name": "Alice Updated",
  "dob": "1991-03-15"
}
```

---

### 4. Delete User — `DELETE /users/:id`

```bash
curl -s -o /dev/null -w "%{http_code}" -X DELETE http://localhost:8080/users/1
# 204
```

---

### 5. List Users — `GET /users?page=1&limit=10`

```bash
curl -s "http://localhost:8080/users?page=1&limit=10" | jq
```

**Response `200 OK`**

```json
[
  {
    "id": 2,
    "name": "Bob",
    "dob": "1985-08-22",
    "age": 39
  },
  {
    "id": 3,
    "name": "Carol",
    "dob": "2000-01-01",
    "age": 25
  }
]
```

---

### Inspecting the X-Request-Id header

```bash
curl -si http://localhost:8080/health | grep -i x-request-id
# X-Request-Id: f47ac10b-58cc-4372-a567-0e02b2c3d479
```

You can also supply your own:

```bash
curl -si http://localhost:8080/health -H "X-Request-Id: my-trace-id" | grep -i x-request-id
# X-Request-Id: my-trace-id
```

---

## Environment Variables

| Variable | Default | Description |
|---|---|---|
| `APP_ENV` | `development` | `development` uses coloured console logging; anything else uses JSON |
| `APP_PORT` | `8080` | TCP port the HTTP server binds to |
| `DB_HOST` | `localhost` | PostgreSQL hostname |
| `DB_PORT` | `5432` | PostgreSQL port |
| `DB_USER` | `postgres` | Database user |
| `DB_PASSWORD` | `secret` | Database password |
| `DB_NAME` | `userapi` | Database name |
| `DB_SSLMODE` | `disable` | `disable` \| `require` \| `verify-full` |