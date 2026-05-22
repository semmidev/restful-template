# Cloud-Native & Production-Ready Checklist

Based on an analysis of your project, you've done a fantastic job establishing a rock-solid foundation. You are already utilizing many best practices (Distroless containers, Clean Architecture, OpenTelemetry, Structured Logging, Graceful Shutdown, and 12-Factor config). 

Below is a checklist dividing what you **already have** (green flags) and what you should **implement next** before going to a cloud-native production environment (like Kubernetes or AWS ECS).

---

## ✅ 1. Current Strengths (Already Implemented)
Your project already meets these critical production standards:

- [x] **Distroless & Non-Root Containers**: Dockerfile builds a statically linked Go binary on a distroless image (`gcr.io/distroless/static-debian12`) running as `nonroot`. This provides a minimal attack surface.
- [x] **12-Factor Configuration**: Environment variables dictate configuration (via `viper`), cleanly falling back to `.env` for local dev.
- [x] **Graceful Shutdown**: The server properly listens for `SIGINT`/`SIGTERM` and drains HTTP connections with a timeout.
- [x] **Observability Core**: OpenTelemetry is configured for traces/metrics, and structured JSON logging (`slog`) is integrated.
- [x] **Defensive HTTP Server**: Server timeouts (`ReadTimeout`, `WriteTimeout`, `IdleTimeout`) are strictly configured to mitigate Slowloris attacks.
- [x] **Connection Pooling**: `pgxpool` is explicitly configured (`DATABASE_MAX_OPEN_CONNS`, `ConnMaxLifetime`) preventing DB connection exhaustion.
- [x] **Self-Contained Deployments**: Database migrations (`golang-migrate`) are embedded into the Go binary and run automatically on startup.
- [x] **OpenAPI & Standard Errors**: Using `huma` provides automatic OpenAPI docs and standard RFC-9457 error responses.

---

## 🚀 2. Infrastructure & Orchestration (To Do)
To be truly "cloud-native," the application needs to be ready for container orchestration platforms like Kubernetes.

- [ ] **Kubernetes Manifests / Helm Chart**: Create `Deployment`, `Service`, `ConfigMap`, and `Secret` manifests.
- [ ] **Liveness & Readiness Probes**: 
  - Enhance `/api/v1/health` (or add `/readyz`) to ping the Postgres database connection pool (`pool.Ping()`).
  - Keep a basic `/livez` that returns 200 OK instantly.
- [ ] **Horizontal Pod Autoscaling (HPA)**: Configure CPU/Memory requests and limits in K8s, and set up HPA rules.
- [ ] **OpenTelemetry Collector Sidecar**: Rather than sending traces directly to Jaeger, send them to a local `otel-collector` sidecar which batches and forwards to cloud providers (e.g., Datadog, Honeycomb).

---

## 🛡️ 3. Security & Resilience (To Do)
Enhance the edge security and database performance for high traffic.

- [ ] **Rate Limiting**: Add a rate-limiting middleware (e.g., Token Bucket via Redis or memory) to prevent abuse of the `/api/v1/auth/login` and `/register` endpoints.
- [ ] **CORS & Security Headers**: Add strict Cross-Origin Resource Sharing (CORS) rules and security headers (HSTS, X-Frame-Options, X-Content-Type-Options).
- [ ] **Database Indexing**: The current `000002_create_todos.up.sql` needs indices on `(user_id, status)` and `(user_id, created_at)` to support the new sorting features efficiently under load.
- [ ] **Secret Management**: Do not pass `JWT_SECRET` or `DATABASE_DSN` as plain text in orchestration. Use HashiCorp Vault, AWS Secrets Manager, or Kubernetes External Secrets.

---

## ⚙️ 4. CI/CD & Code Quality (To Do)
Automate the path to production.

- [ ] **CI Pipeline (GitHub Actions / GitLab CI)**:
  - Run `golangci-lint`.
  - Run `go test -race -cover`.
  - Build the Docker image and scan it for vulnerabilities (e.g., using `trivy`).
- [ ] **CD Pipeline**: Automate the pushing of the Docker image to a registry (GHCR, ECR) and update Kubernetes manifests (e.g., using ArgoCD or Flux).

---

## 📡 5. Application Logic Tweaks (To Do)
Minor code improvements for enterprise readiness.

- [ ] **Pagination Metadata**: Enhance `ListData` to return `next_cursor` or HAL links for standard REST API traversal instead of just `page` and `per_page`.
- [ ] **Refresh Token Rotation**: Currently, `RefreshBody` accepts a token. Implement Token Rotation (invalidating the old refresh token upon use) and store active refresh token hashes in the database/Redis to allow remote revocation.
