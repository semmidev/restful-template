# Restful Template API

A production-grade RESTful API template built with Go 1.26, featuring **Clean Architecture**, high performance routing via Chi, OpenAPI 3.1 auto-generation via Huma v2, and secure, time-ordered UUIDs via PostgreSQL 18.

## Table of Contents
- [Key Features](#key-features)
- [Architecture Flow](#architecture-flow)
- [Observability Architecture](#observability-architecture)
- [Project Structure](#project-structure)
- [Getting Started](#getting-started)
  - [Prerequisites](#prerequisites)
  - [1. Configuration](#1-configuration)
  - [2. Running Locally (Docker Compose)](#2-running-locally-docker-compose)
  - [3. Running Locally (Go + Postgres)](#3-running-locally-go--postgres)
- [Development Guide](#development-guide)
  - [Code Quality & Testing](#code-quality--testing)
  - [Building for Production](#building-for-production)
  - [Database Migrations](#database-migrations)
- [API Documentation](#api-documentation)
  - [Adding New Routes](#adding-new-routes)
  - [Search and Filtering Example](#search-and-filtering-example)
- [Configuration Variables (.env)](#configuration-variables-env)

---

## Key Features

*   **Architecture & Design**
    *   Clean Architecture / Domain-Driven Design
    *   Clean dependency flow: `domain` → `usecase` → `delivery` / `repository`
    *   `domain.TxManager` for injecting transactions into the `context.Context` from the usecase layer.
*   **Routing & Documentation**
    *   [Huma v2](https://github.com/danielgtaylor/huma) + [Chi v5](https://github.com/go-chi/chi) integration
    *   Automatic **OpenAPI 3.1** specification generation (`/openapi.json`)
    *   Built-in Swagger UI documentation (`/docs`)
    *   Typed HTTP handlers with auto-validation
*   **Database & Persistence**
    *   PostgreSQL 18 integration via `pgxpool`
    *   **Native UUID v7** support for time-ordered, fragmentation-free primary keys
    *   Embedded SQL migrations (no external CLI needed at runtime)
    *   Pagination and advanced filtering (e.g., status, full-text substring keyword search)
*   **Security & Authentication**
    *   JWT-based Authentication (Access + Refresh tokens with rotation)
    *   Argon2id password hashing (PHC string format, constant-time verification)
    *   CORS configured
    *   Secure HTTP Headers middleware
*   **Observability & Reliability**
    *   Structured JSON/Text Logging via `log/slog`
    *   OpenTelemetry (OTEL) integration for distributed tracing and metrics (`otelchi` and `otelpgx`)
    *   Graceful shutdown with signal handling
    *   Request ID propagation
    *   Panic recovery middleware
    *   RFC 9457 structured error responses (`application/problem+json`)
*   **Configuration & Deployment**
    *   12-Factor App ready
    *   Viper-based configuration loading (`.env` file + OS environment variables)
    *   Multi-stage, distroless Dockerfile for tiny, secure images
    *   `docker-compose` setup with healthchecks and restart policies

---

## Architecture Flow

```mermaid
flowchart TD
    Client([Client]) -- "HTTP Request" --> Delivery

    subgraph Clean Architecture
        Delivery["Delivery Layer<br/>(Chi + Huma)"]
        Usecase["Usecase Layer<br/>(Business Logic)"]
        Domain["Domain Layer<br/>(Entities & Interfaces)"]
        Infrastructure["Infrastructure Layer<br/>(Postgres, JWT)"]

        Delivery -- "Depends on" --> Usecase
        Usecase -- "Depends on" --> Domain
        Infrastructure -. "Implements interfaces" .-> Domain
        Delivery -. "Depends on" .-> Domain
        
        Usecase -. "Runtime execution" .-> Infrastructure
    end

    Infrastructure -- "SQL Queries" --> Database[(PostgreSQL 18)]
```

---

## Observability Architecture

```mermaid
flowchart TD
    subgraph Application
        API["API Service<br/>(Go App)"]
        DB[(PostgreSQL)]
        API -- "Read/Write" --> DB
    end

    subgraph Grafana Observability Stack
        Alloy["Grafana Alloy<br/>(Telemetry Collector)"]
        Prometheus["Prometheus<br/>(Metrics Storage)"]
        Loki["Grafana Loki<br/>(Logs Storage)"]
        Tempo["Grafana Tempo<br/>(Traces Storage)"]
        Grafana["Grafana UI<br/>(Visualization)"]
    end

    %% Metrics Flow
    Prometheus -- "1. Pulls metrics (/metrics)" --> API
    
    %% Traces Flow
    API -- "2. Pushes OTLP Traces (gRPC: 4317)" --> Alloy
    Alloy -- "3. Pushes Traces (gRPC: 4317)" --> Tempo
    
    %% Logs Flow
    DockerSocket[["/var/run/docker.sock"]]
    API -. "Writes stdout/stderr" .-> DockerSocket
    Alloy -- "4. Pulls container logs" --> DockerSocket
    Alloy -- "5. Pushes Logs (HTTP: 3100)" --> Loki

    %% Grafana UI Flow
    Grafana -- "6. Pulls Metrics (PromQL)" --> Prometheus
    Grafana -- "7. Pulls Logs (LogQL)" --> Loki
    Grafana -- "8. Pulls Traces (TraceQL)" --> Tempo
```

---


## Project Structure

```text
.
├── cmd/
│   ├── migrate/      # Standalone database migration utility
│   └── server/       # Main API application entrypoint
├── internal/
│   ├── config/       # Application configuration (Viper)
│   ├── delivery/     # Inbound (HTTP Server, Routers)
│   ├── domain/       # Core business entities, usecase, and repository interfaces
│   ├── infrastructure/
│   │   ├── database/     # DB connection, migrations, TxManager implementation
│   │   ├── jwt/          # External services (e.g. JWT Auth)
│   │   ├── observability/# Logging and OpenTelemetry tracing setup
│   │   └── repository/   # Outbound (Postgres, Redis)
│   ├── shared/       # Cross-cutting utilities (UUID generation, Password hashing)
│   └── usecase/      # Use cases / Business logic implementation
```

---

## Getting Started

### Prerequisites
*   Go 1.23+
*   Docker & Docker Compose (for local development database)
*   Make

### 1. Configuration

Copy the example environment file and customize it if needed.

```bash
cp .env.example .env
```

### 2. Running Locally (Docker Compose)

The easiest way to start the entire stack (API + PostgreSQL) is using Docker Compose.

```bash
make docker-up
```
The API will be available at `http://localhost:8080`.

To view the logs:
```bash
make docker-logs
```

To shut down:
```bash
make docker-down
```

### 3. Running Locally (Go + Postgres)

If you prefer to run the Go application natively while keeping the database in Docker:

1.  Start only the database:
    ```bash
    docker compose up postgres -d
    ```
2.  Run the application (migrations are applied automatically on startup):
    ```bash
    make run
    ```

---

## Development Guide

We use a `Makefile` to simplify common tasks.

### Code Quality & Testing

*   **Format & Tidy**: Clean up dependencies and format code.
    ```bash
    make tidy
    ```
*   **Lint**: Run `golangci-lint` (requires it to be installed).
    ```bash
    make lint
    ```
*   **Vet**: Run Go vet over the project.
    ```bash
    make vet
    ```
*   **Test**: Run the test suite with race detection and coverage.
    ```bash
    make test
    ```
*   **Coverage**: View the test coverage in your browser.
    ```bash
    make coverage
    ```

### Building for Production

To build a standalone, statically linked binary:

```bash
make build
```
The binary will be output to `bin/server`.

To clean up build artifacts:
```bash
make clean
```

### Database Migrations

Migrations are automatically embedded into the binary and executed on application startup. However, if you need to run them manually, you can use the provided migration CLI:

```bash
make migrate-up
make migrate-down
```

---

## API Documentation

Once the server is running, the interactive documentation is served automatically:

*   **Swagger UI**: [http://localhost:8080/docs](http://localhost:8080/docs)
*   **OpenAPI 3.1 JSON**: [http://localhost:8080/openapi.json](http://localhost:8080/openapi.json)

### Adding New Routes

1.  Define your Input/Output structs and the Interface method in `internal/domain/`.
2.  Implement the business logic in `internal/usecase/`.
3.  Register the route in `internal/delivery/http/routes.go` using `huma.Register`. Huma handles request validation and OpenAPI generation automatically based on your structs.

### Search and Filtering Example
The `GET /api/v1/todos` endpoint supports advanced filtering:
*   `status`: Filter by todo status (e.g., `pending`, `in_progress`, `done`).
*   `q`: Case-insensitive substring search across `title` and `description`.
*   `limit` / `offset`: Pagination.

Example Request:
```
GET /api/v1/todos?q=buy&status=pending&limit=10&offset=0
```

---

## Configuration Variables (`.env`)

| Variable | Description | Default |
| :--- | :--- | :--- |
| `APP_ENV` | Environment (`development` or `production`) | `development` |
| `APP_NAME` | Name of the application | `restful-template` |
| `HTTP_PORT` | Port the API listens on | `8080` |
| `DATABASE_DSN` | Postgres connection string | `postgres://todo:todo@localhost:5432/todo?sslmode=disable` |
| `JWT_SECRET` | Secret key for signing JWTs (**MUST** change in prod) | `change-me-in-production-min-32-bytes!` |
| `JWT_ACCESS_TTL` | Access token lifespan | `15m` |
| `JWT_REFRESH_TTL` | Refresh token lifespan | `168h` |
| `LOG_LEVEL` | Logging verbosity (`debug`, `info`, `warn`, `error`) | `info` |
| `LOG_FORMAT` | Log output format (`json`, `text`) | `json` |
| `TELEMETRY_OTLP_ENDPOINT` | OpenTelemetry gRPC endpoint | `localhost:4317` |
