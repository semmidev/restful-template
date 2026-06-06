## Project Overview

**restful-template** is a production-grade RESTful API template written in Go.
It demonstrates clean architecture, observability best practices, and modern API design
using [Huma v2](https://huma.rocks) + [Chi v5](https://github.com/go-chi/chi).

The primary business domain is a **multi-user Todo service with JWT authentication**,
intentionally kept simple so the *infrastructure and engineering patterns* take center stage.

### Tech Stack

| Layer            | Technology                                            |
|------------------|-------------------------------------------------------|
| HTTP framework   | `go-chi/chi v5` + `danielgtaylor/huma v2`            |
| Database         | PostgreSQL 18 via `jackc/pgx v5`                     |
| Query builder    | `Masterminds/squirrel`                                |
| Cache            | Redis 7 via `redis/go-redis v9`                      |
| Rate limiting    | `go-redis/redis_rate v10`                             |
| Auth             | `golang-jwt/jwt v5` (access + refresh token rotation)|
| Migrations       | `golang-migrate/migrate v4` (embedded SQL files)     |
| Observability    | OpenTelemetry → Grafana Alloy → Tempo (traces)       |
|                  | Prometheus → Grafana (metrics)                       |
|                  | Grafana Loki (logs via structured `slog`)             |
| Scheduler        | `go-co-op/gocron v2` (separate binary)               |
| Async Worker     | `hibiken/asynq` (separate binary, Redis-backed queue)|
| Worker UI        | `hibiken/asynqmon` (mounted at `/admin/asynq`, Basic Auth) |
| Testing          | `smartystreets/goconvey` + `testcontainers-go`       |
| Linting          | `golangci-lint v2` (see `.golangci.yml`)             |

---

## Architecture

This project follows **Clean / Hexagonal Architecture** with strict dependency direction:

```
cmd/
└── server/          ← entry point — wires config, logger, calls app.Setup()
└── scheduler/       ← separate binary for background cron jobs
└── worker/          ← separate binary for asynq async task processing

internal/
├── app/             ← dependency injection root (app.Setup)
├── config/          ← Viper config loaded from .env + env vars
├── delivery/http/   ← HTTP layer: server, routes, middleware (driven adapters)
├── modules/
│   ├── auth/        ← auth domain, repository, service, HTTP handler, middleware
│   └── todos/       ← todos domain, repository, service, HTTP handler

└── shared/
    ├── asynqtask/   ← task type constants, payload structs, TaskDistributor (producer)
    ├── cache/       ← CacheRepository interface
    ├── database/    ← pgxpool, squirrel QB, TxManager, migrations
    ├── errors/      ← SafeError (never leaks internals to clients)
    ├── httpapi/     ← error mapping, user ID extraction
    ├── jwt/         ← JWTService (access + refresh, iss/aud claims)
    ├── middleware/   ← CORS, rate limiter, logger, Prometheus, security headers
    ├── observability/← OtelTracer adapter (interface-based)
    ├── password/    ← bcrypt helpers
    ├── redis/       ← Redis client + CacheRepository impl
    ├── uuidgen/     ← deterministic UUID generation (testable)
    └── wideevent/   ← canonical wide log event enrichment
```

### Dependency Rules (never violate these)

1. **`internal/modules/*`** may only import from `internal/shared/*`. Modules are **never** imported by other modules directly — cross-module calls happen via **interfaces defined in the domain file**.
2. **`internal/delivery/http`** depends on module interfaces (`AuthService`, `TodoService`), not on concrete `*Service` structs — except in `NewServer()` where wiring happens explicitly.
3. **`internal/shared/*`** has **zero imports** from `internal/modules/*` or `internal/delivery/*`.
4. **`internal/app/app.go`** is the **only place** where concrete types are wired together. All other packages consume interfaces.
5. **`cmd/*`** only calls `app.Setup()` and manages the OS signal lifecycle.

---

## File Structure Conventions

### Module Layout (follow exactly for new modules)

Each module in `internal/modules/<name>/` has exactly these files:

| File                         | Responsibility                                                                  |
|------------------------------|---------------------------------------------------------------------------------|
| `<name>_domain.go`           | Domain entity, value objects, domain constants, **repository interface**, **service interface** |
| `<name>_repository.go`       | PostgreSQL-backed repository struct implementing the repository interface       |
| `<name>_service.go`          | Business logic struct implementing the service interface                        |
| `<name>_service_types.go`    | Input/Output structs for the service methods                                    |
| `<name>_http_handlers.go`    | Huma handler implementations                                                    |
| `<name>_http_routes.go`      | Route registration using Huma                                                   |
| `<name>_http_types.go`       | Huma request/response structs, path/query params                                |
| `<name>_http_middleware.go`  | (optional) Module-specific HTTP middleware (e.g. auth middleware)               |
| `<name>_constant.go`         | (optional) Service/HTTP level constants                                         |
| `<name>_distributor_types.go`| (optional) Asynq task scheduling types and payloads                             |
| `<name>_job.go`              | (optional) Scheduled job definitions for the scheduler binary                   |

> **Do not create subdirectories inside a module.** All module files live flat in the module directory.

### Naming Conventions

- Interfaces: named after their role, **not** their implementation — `TodoRepository`, `TodoService`, `TokenService`, `TxManager`.
- Service struct: always named `Service` within its package (`todos.Service`, `auth.Service`).
- Repository struct: always unexported (`todoRepository`, `userRepository`) — consumers only hold the interface.
- Constructor: `New<Name>(deps...) <Interface>` for repositories, `New<Name>Service(deps...) *Service` for services.
- HTTP handler struct: unexported (`todoHandler`), only `Register<Name>Routes(api, service)` is exported.

---

## Engineering Standards

### Error Handling

**Always use `internal/shared/errors` — never `fmt.Errorf` or `errors.New` in business code.**

```go
// CORRECT
return nil, apperrors.NewNotFound("The requested todo does not exist", err)
return nil, apperrors.NewInvalidInput("Invalid todo data", err)
return nil, apperrors.NewInternal("Failed to create todo", err)
return nil, apperrors.NewConflict("Email is already registered", err)

// WRONG — leaks internal details, breaks HTTP mapping
return nil, fmt.Errorf("pgx query failed: %w", err)
```

- `SafeError.Error()` returns only the user-safe message — **never** expose `Internal` to clients.
- `SafeError.Unwrap()` allows `errors.Is(err, apperrors.ErrNotFound)` to work across layers.
- HTTP mapping lives in `internal/shared/httpapi.ToHumaErr()` — never manually set HTTP status codes in handlers.

### Repository Pattern

- Use `database.QB` (the shared `squirrel.StatementBuilderType`) for all query building — no raw string SQL.
- Use `database.GetDB(ctx, r.db)` in every repository method — this enables transparent transaction propagation via `database.ExtractTx(ctx)`.
- Sort columns **must** go through an explicit allowlist `map[string]string` (see `todo_repository.go`) — never interpolate user input directly into ORDER BY.
- Use `COUNT(*) OVER()` window function to avoid a second round-trip for pagination counts.
- On UPDATE, always check `res.RowsAffected() == 0` and return `apperrors.ErrNotFound` — never silently no-op.

### Transaction Management

Use `database.TxManager.RunInTx()` for multi-repository operations. The transaction is propagated via context — repositories **automatically** use it through `database.GetDB(ctx, pool)`.

```go
// CORRECT — atomic delete across two tables
return s.txManager.RunInTx(ctx, func(txCtx context.Context) error {
    if err := s.todos.DeleteAllByUserID(txCtx, userID); err != nil {
        return err
    }
    return s.users.Delete(txCtx, userID)
})
```

### Caching (Redis)

- Cache key format: `<entity>:<ownerID>:<entityID>` (e.g. `todo:userUUID:todoUUID`).
- TTL constant defined at the top of the service or constant file: `const todoCacheTTL = 5 * time.Minute`.
- Pattern: **read-through** on Get, **write-through + invalidate-then-repopulate** on Update, **invalidate** on Delete.
- Cache writes are **best-effort** — never fail the request if `cache.Set` fails (use `_ =`).
- Do **not** SCAN Redis for bulk invalidation on collection deletes — let entries expire naturally.

### Observability

Every service method **must** start with a span:

```go
func (s *Service) Create(ctx context.Context, in CreateTodoInput) (*Todo, error) {
    ctx, span := s.tracer.Start(ctx, "todo.Create")
    defer span.End()
    // ...
}
```

Span names follow the format `"<module>.<Method>"` (e.g. `"auth.Register"`, `"todo.Update"`).

Use `wideevent.Add(ctx, key, value)` in HTTP handlers to enrich the canonical log event with domain-relevant fields (e.g. `todo_id`, `todo_title`, `todo_status`). This feeds the Loki-based structured logs.

### Configuration

All configuration flows through `internal/config.Config` (loaded via Viper from `.env` + env vars).

- Env var format: `SECTION_KEY` maps to `Config.Section.Key` (e.g. `JWT_SECRET` → `Config.JWT.Secret`).
- Timeouts use `mustDuration()` — it panics on misconfiguration at startup (fail-fast is intentional).
- `DATABASE_RUN_MIGRATIONS=false` in multi-replica deploys to avoid advisory-lock contention.
- Never access `os.Getenv` directly in business code — always inject `config.Config`.

### Security Constraints

- **Never** return raw database errors to the client — always wrap with `SafeError`.
- Refresh tokens are stored as `SHA-256(token)` in the database — plaintext tokens never persist.
- Image uploads: enforce `io.LimitReader(f, maxCoverSize+1)` before `io.ReadAll`. Use `http.DetectContentType()` to verify MIME — never trust the client's `Content-Type` header.
- CORS: default is `"*"` for dev only — **must** be restricted via `CORS_ALLOWED_ORIGINS` in production.
- SQL injection prevention: sort columns go through an explicit allowlist map; all query parameters go through squirrel's parameterized builder.

### HTTP / Huma Conventions

- All routes live under `/api/v1/`.
- Use `huma.Register()` with explicit `OperationID`, `Tags`, `Security`, `Summary` — these populate the OpenAPI spec.
- Protected routes declare `Security: []map[string][]string{{"bearerAuth": {}}}`.
- ETag / optimistic locking: handlers fetch the entity, compute `ETag = updated_at` in RFC3339Nano, validate with `conditional.Params.PreconditionFailed()`, then pass the pre-fetched entity to the service. This eliminates a redundant DB round-trip.
- Pagination: use `page`/`per_page` query params; return `X-Total-Count` header and RFC 8288 `Link` header.
- Partial updates use `update_mask` query param following AIP-134.

---

## Development Workflow

### Quick Start

```bash
# Start all infrastructure (Postgres, Redis, Prometheus, Grafana, Loki, Tempo, Alloy)
make docker-up

# Run the API server locally
make run

# Run unit tests (with race detector)
make test

# Run integration tests (uses testcontainers — requires Docker)
make test-integration

# Run linter
make lint

# Format code
make format
```

### Adding a New Module

Follow these steps in order:

1. Create `internal/modules/<name>/` with the standard files (`_domain.go`, `_repository.go`, `_service.go`, etc).
2. Define the domain entity and business methods in `<name>_domain.go`. Define the **repository interface** and **service interface** in the same file.
3. Implement the repository in `<name>_repository.go` using `database.QB` and `database.GetDB(ctx, r.db)`.
4. Define service input/output types in `<name>_service_types.go` and implement the service in `<name>_service.go`. Inject `cache.CacheRepository` and `observability.Tracer`. Add spans to every public method.
5. Define HTTP request/response types in `<name>_http_types.go`, handlers in `<name>_http_handlers.go`, and register routes in `<name>_http_routes.go` using `huma.Register()`.
6. Wire dependencies in `internal/app/app.go`: create the repository, then the service, then pass it to `delivery.NewServer()`.
7. Add the route registration call to `internal/delivery/http/routes.go`.
8. Create SQL migrations in `internal/shared/database/migrations/` using sequential numbering (`000004_...`).

### Writing Tests

- Unit tests live **inside** the module package (same directory as the code).
- Integration tests live in `tests/` and use `testcontainers-go` to spin up real Postgres and Redis instances.
- Use `goconvey` assertions for BDD-style test descriptions.
- Mock the `TodoService` / `AuthService` interfaces in handler tests — never start a real server.
- Test table names in testcontainers: use the same schema as production migrations.

---

## What to Avoid

The following patterns are **explicitly prohibited** in this codebase:

| Anti-pattern                                      | Reason                                                         |
|---------------------------------------------------|----------------------------------------------------------------|
| Raw `fmt.Errorf` in business logic                | Bypasses `SafeError`; leaks internal details to clients        |
| Direct SQL strings in repositories                | Use squirrel; prevents SQL injection and aids readability      |
| Importing one module from another                 | Violates clean arch; creates circular dependencies             |
| Calling `os.Getenv` outside `config/`             | Config must be centralized and testable                        |
| `io.ReadAll` without a `LimitReader` on uploads   | Memory exhaustion risk                                         |
| Trusting `Content-Type` headers on file uploads   | MIME sniff the actual bytes with `http.DetectContentType`      |
| Adding a logger to route registration functions   | Logging belongs to middleware (wide events), not route setup   |
| Failing requests on cache write errors            | Cache is best-effort; don't degrade availability               |
| SCAN Redis for bulk cache invalidation            | O(n) Redis operation; let entries expire via TTL instead       |
| Returning `nil` error when a non-nil was checked  | Caught by `nilerr` linter — always propagate or wrap errors   |

---

## Observability Stack (local)

| Service    | URL                       | Purpose                          |
|------------|---------------------------|----------------------------------|
| API        | http://localhost:8080     | Main application                 |
| OpenAPI    | http://localhost:8080/docs| Huma auto-generated Swagger UI   |
| Metrics    | http://localhost:8080/metrics | Prometheus scrape endpoint   |
| Asynqmon   | http://localhost:8080/admin/asynq | Worker queue monitor (Basic Auth) |
| Prometheus | http://localhost:9090     | Metrics storage                  |
| Grafana    | http://localhost:3000     | Dashboards (metrics + logs + traces) |
| Loki       | http://localhost:3100     | Log aggregation                  |
| Tempo      | http://localhost:3200     | Distributed tracing              |
| Alloy      | http://localhost:12345    | Grafana collector (OTLP gRPC: 4317) |

Traces are exported from the app via OTLP gRPC to Alloy → Tempo.
Metrics are scraped by Prometheus from `/metrics`.
Logs are structured JSON via `log/slog`, collected by Alloy and forwarded to Loki.

---

## Key Design Decisions (rationale for agents)

1. **Huma v2 over raw `net/http`**: Auto-generates OpenAPI 3.1 spec, handles input validation, and decodes request/response types — reducing handler boilerplate significantly.
2. **`ETag` based on `updated_at`**: Avoids a dedicated version column while still providing optimistic concurrency control. The handler fetches the entity, validates the precondition, and passes the pre-loaded entity into the service — reducing PATCH from 3 DB calls to 2.
3. **SHA-256 hashed refresh tokens**: Token rotation on every refresh (old token deleted, new pair issued). Hash storage means a DB breach doesn't expose valid tokens.
4. **`DATABASE_RUN_MIGRATIONS` flag**: Multi-replica deployments must set this to `false` and run migrations as an init-container. The default is `false` for safety.
5. **Separate `cmd/scheduler` binary**: Background jobs run in a separate process to allow independent scaling, restarts, and resource isolation. The scheduler re-uses the same domain and repository code.
6. **Wide events via `wideevent`**: A single structured log line per request carries all domain context (user ID, todo ID, counts) rather than multiple log statements scattered through the call stack. This makes Loki queries dramatically more useful.
7. **`SafeError` with `Unwrap()`**: Allows `errors.Is(err, apperrors.ErrNotFound)` to work through the stack while keeping the public-facing message safe and the internal cause available for structured logging.
8. **Service Interfaces over Concrete Structs**: Services (`AuthService`, `TodoService`) are exposed as interfaces to the HTTP layer and other modules. This enables isolated unit testing of HTTP handlers (mocking the service without DB setup), prevents strict cross-module coupling (circular dependencies), and enforces architectural boundaries.
9. **Separate `cmd/worker` binary (asynq)**: Async task processing runs in its own isolated binary. The `TaskDistributor` interface (defined in the calling module's domain file) points to `internal/shared/asynqtask.Distributor` — this keeps the boundary clean. Worker handlers are implemented directly inside their respective modules (e.g., `auth_worker.go`) and registered in `cmd/worker`. The `ASYNQMON_USERNAME`/`ASYNQMON_PASSWORD` env vars guard the built-in Web UI at `/admin/asynq`.
