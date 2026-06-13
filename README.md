# AgeSync API

![Hero Image Placeholder](https://via.placeholder.com/1000x300.png?text=Replace+with+a+screenshot+of+your+running+terminal+or+API+response)

A production-ready RESTful API built in Go for managing users, featuring dynamic age calculation, structured logging, request tracing, and full containerisation. Designed with clean architecture principles to ensure maintainability and high performance.

---

## Table of Contents

1. [Architecture & Design Choices](#architecture--design-choices)
2. [Prerequisites](#prerequisites)
3. [Running with Docker (Recommended)](#running-with-docker-recommended)
4. [Running without Docker (Local Postgres)](#running-without-docker-local-postgres)
5. [Running Unit Tests](#running-unit-tests)
6. [API Reference & PowerShell Testing](#api-reference--powershell-testing)
7. [SQLC Code Generation](#sqlc-code-generation)
8. [Environment Variables](#environment-variables)

---

## Architecture & Design Choices

![Architecture Diagram Placeholder](https://via.placeholder.com/800x400.png?text=Replace+with+a+simple+system+architecture+or+folder+structure+diagram)

### Layer Responsibilities

| Layer | Responsibility |
|---|---|
| `handler` | HTTP concerns only: body parsing, input validation, status codes. |
| `service` | Pure business logic; dynamic age calculation; no HTTP or DB knowledge. |
| `repository` | Database I/O; adapts between SQLC generated types and domain types. |
| `db/sqlc` | Auto-generated, type-safe SQL access layer (never hand-edited). |

### Key Design Decisions

* **Clean Architecture:** Dependencies point inward. Handlers depend on service interfaces; services depend on repository interfaces. This enables the database to be swapped or mocked in tests without touching the HTTP layer.
* **SQLC over ORMs:** All SQL is written by hand and compiled into type-safe Go code. This eliminates runtime reflection overhead and catches schema mismatches at compile time.
* **Dynamic Age Calculation:** Age is computed dynamically in the service layer using pure Go `time` calendar arithmetic. It is never persisted to the database to ensure accuracy as time passes.
* **Graceful Shutdown:** The server listens for `SIGINT`/`SIGTERM` and drains in-flight connections with a 5-second deadline before exiting cleanly.
* **Multi-Stage Docker Build:** The final image is built `FROM scratch`, containing only the statically-linked binary. This results in an incredibly lightweight and secure image.
* **Structured Observability:** Uber Zap is used for zero-allocation JSON logging in production, injecting an `X-Request-Id` into every request for end-to-end tracing.

---

## Prerequisites

### With Docker (Recommended)
* [Docker Desktop](https://docs.docker.com/get-docker/) ≥ 24
* [Docker Compose](https://docs.docker.com/compose/install/) v2

### Without Docker (Local Setup)
* Go ≥ 1.22
* PostgreSQL ≥ 14

---

## Running with Docker (Recommended)

### 1. Clone the repository
```bash
git clone https://github.com/yourusername/AgeSync-API.git
cd AgeSync-API
```

### 2. Build and start all services
```bash
docker compose up --build -d
```
Docker Compose will automatically start a `postgres:16-alpine` container, wait until it is healthy, build the Go binary, and start the API server. 

The API is now available at **http://localhost:8080**.

### 3. View live logs
```bash
docker compose logs -f
```

---

## Running without Docker (Local Postgres)

1. **Create the database and apply the migration:**
   ```bash
   psql -U postgres -c "CREATE DATABASE userapi;"
   psql -U postgres -d userapi -f db/migrations/000001_create_users_table.up.sql
   ```
2. **Download dependencies and run:**
   ```bash
   go mod tidy
   go run ./cmd/server/main.go
   ```

---

## Running Unit Tests

The unit tests for the age calculation logic are pure (no I/O, no database) and run instantly.

```bash
# Run all tests with the race detector enabled
go test -race ./internal/service/...

# Run all tests across the whole project
go test -race ./...
```

---

## API Reference & PowerShell Testing

![API Testing Placeholder](https://via.placeholder.com/800x300.png?text=Replace+with+a+screenshot+of+successful+PowerShell+API+responses)

The following examples use native Windows PowerShell (`Invoke-RestMethod`) for seamless JSON payload handling.

> **Base URL:** `http://localhost:8080`

### 1. Create a User (POST)
**Command:**
```powershell
Invoke-RestMethod -Uri http://localhost:8080/users -Method POST -Headers @{"Content-Type"="application/json"} -Body '{"name": "Ayush Joshi", "dob": "2000-08-15"}'
```
**Expected Output:**
```json
{
  "id": 1,
  "name": "Ayush Joshi",
  "dob": "2000-08-15"
}
```

### 2. Fetch the User (GET)
*Watch the age calculate dynamically based on today's date.*
**Command:**
```powershell
Invoke-RestMethod -Uri http://localhost:8080/users/1 -Method GET
```
**Expected Output:**
```json
{
  "id": 1,
  "name": "Ayush Joshi",
  "dob": "2000-08-15",
  "age": 25
}
```

### 3. Update the User (PUT)
**Command:**
```powershell
Invoke-RestMethod -Uri http://localhost:8080/users/1 -Method PUT -Headers @{"Content-Type"="application/json"} -Body '{"name": "Ayush Joshi", "dob": "2001-01-20"}'
```

### 4. List All Users (GET with Pagination)
**Command:**
```powershell
Invoke-RestMethod -Uri "http://localhost:8080/users?page=1&limit=10" -Method GET
```

### 5. Delete the User (DELETE)
**Command:**
```powershell
Invoke-RestMethod -Uri http://localhost:8080/users/1 -Method DELETE
```
*(Returns an empty response with HTTP 204 No Content upon success).*

---

## SQLC Code Generation

The files under `db/sqlc/*.go` are **auto-generated**. If you modify the schema or queries, regenerate the Go code using SQLC:

```bash
sqlc generate
```

---

## Environment Variables

| Variable | Default | Description |
|---|---|---|
| `APP_ENV` | `development` | `development` uses coloured console logging; anything else defaults to structured JSON. |
| `APP_PORT` | `8080` | TCP port the HTTP server binds to. |
| `DB_HOST` | `localhost` | PostgreSQL hostname. |
| `DB_PORT` | `5432` | PostgreSQL port. |
| `DB_USER` | `postgres` | Database user. |
| `DB_PASSWORD` | `secret` | Database password. |
| `DB_NAME` | `userapi` | Database name. |
| `DB_SSLMODE` | `disable` | Database connection security. |
