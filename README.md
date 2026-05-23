# Restful Template API

Template RESTful API *production-grade* yang dibangun menggunakan Go 1.26. Template ini mengusung arsitektur **Modular Monolith (Package by Feature)**, *routing* berkinerja tinggi menggunakan Chi, *auto-generation* OpenAPI 3.1 dengan Huma v2, serta menggunakan PostgreSQL 18 untuk men-generate UUID yang aman dan terurut berdasarkan waktu (*time-ordered*).

## Daftar Isi
- [Fitur Utama](#fitur-utama)
- [Arsitektur Modular Monolith](#arsitektur-modular-monolith)
  - [Aturan Main (Rules of Engagement)](#aturan-main-rules-of-engagement)
  - [Komunikasi Antar Modul (Inter-Module Communication)](#komunikasi-antar-modul-inter-module-communication)
  - [Transaksi Lintas Modul (Cross-Module Transactions)](#transaksi-lintas-modul-cross-module-transactions)
- [Arsitektur Observability](#arsitektur-observability)
- [Struktur Projek](#struktur-projek)
- [Mulai Menggunakan (Getting Started)](#mulai-menggunakan-getting-started)
- [Panduan Development](#panduan-development)
- [Dokumentasi API](#dokumentasi-api)
- [Konfigurasi Variabel (.env)](#konfigurasi-variabel-env)

---

## Fitur Utama

*   **Architecture & Design**
    *   **Modular Monolith (Package by Feature)**: Memiliki batasan (*boundaries*) yang tegas untuk mencegah *technical debt* dan arsitektur yang berantakan (*ball-of-mud*).
    *   *Synchronous Interface Injection* untuk memfasilitasi *Consumer-Driven Contracts* yang bersih antar modul.
    *   Transaksi lintas modul (*Cross-module transactions*) dikelola secara aman menggunakan `database.TxManager`.
*   **Routing & Documentation**
    *   Integrasi [Huma v2](https://github.com/danielgtaylor/huma) + [Chi v5](https://github.com/go-chi/chi)
    *   *Auto-generation* spesifikasi **OpenAPI 3.1** (`/openapi.json`)
    *   Dokumentasi interaktif menggunakan Swagger UI bawaan (`/docs`)
    *   *Typed HTTP handlers* yang memiliki fitur validasi otomatis (*auto-validation*)
*   **Database & Persistence**
    *   Integrasi PostgreSQL 18 melalui `pgxpool`
    *   Dukungan **Native UUID v7** untuk *primary keys* yang kebal terhadap fragmentasi dan terurut secara kronologis.
    *   Sistem migrasi database yang di-*embed* langsung ke dalam binary (tidak butuh *external CLI* tambahan di *runtime*).
    *   Pagination dan *advanced filtering* (misalnya filter status dan pencarian dengan kata kunci / *full-text substring keyword search*)
*   **Security & Authentication**
    *   Autentikasi berbasis JWT (Access + Refresh tokens dengan rotasi)
    *   *Hashing* kata sandi menggunakan Argon2id (format *string* PHC, verifikasi menggunakan *constant-time*)
    *   CORS yang siap dipakai
    *   Middleware untuk *Secure HTTP Headers*
*   **Observability & Reliability**
    *   *Structured JSON/Text Logging* menggunakan `log/slog`
    *   Integrasi OpenTelemetry (OTEL) untuk *distributed tracing* dan *metrics* (`otelchi` dan `otelpgx`)
    *   *Graceful shutdown* dengan penanganan sinyal OS
    *   Propagasi Request ID
    *   Middleware *Panic recovery*
    *   Format error yang terstruktur mengikuti standar RFC 9457 (`application/problem+json`)
*   **Configuration & Deployment**
    *   Memenuhi standar *12-Factor App*
    *   Sistem konfigurasi menggunakan Viper (otomatis membaca file `.env` + *environment variables* dari OS)
    *   Dockerfile *multi-stage* berbasis *distroless* untuk menghasilkan *image* yang mungil dan aman
    *   Setup `docker-compose` yang dilengkapi dengan fitur *healthchecks* dan *restart policies*

---

## Arsitektur Modular Monolith

Projek ini dibangun menggunakan *pattern* arsitektur **Package by Feature** / Modular Monolith.

Ketimbang mengelompokkan kode berdasarkan fungsi teknisnya (seperti menggabungkan semua *controllers* di satu folder, lalu semua *repositories* di folder lain), kode di projek ini dikelompokkan berdasarkan **fitur bisnis** (contoh: `auth`, `todos`). Pendekatan ini mendorong tercapainya *high cohesion* dan *low coupling* yang optimal pada aplikasi yang berskala besar.

### Aturan Main (Rules of Engagement)

Agar *codebase* tetap rapi dan mudah dikelola seiring pertumbuhannya, kita menerapkan beberapa aturan ketat:
1.  **Tidak ada *direct imports* antar modul**: Modul `internal/modules/auth` dilarang meng-*import* `internal/modules/todos` secara langsung.
2.  **Shared Kernel**: Kode apapun yang generik dan dibutuhkan oleh lebih dari satu modul harus diletakkan di `internal/shared` (misalnya: `errors`, `httpapi`, `database`).
3.  **Encapsulation**: Semua struktur yang ada di dalam sebuah modul (handlers, *business logic*, repositories) sifatnya *internal* di modul tersebut. Hanya *public constructors* (seperti `NewAuth`) yang boleh diakses dan disatukan (*wiring*) di *entrypoint* utama yaitu `cmd/server/main.go`.

### Komunikasi Antar Modul (Inter-Module Communication)

Ketika satu modul butuh berinteraksi dengan modul lain, kita menggunakan teknik **Synchronous Interface Injection** yang mengacu pada konsep **Consumer-Driven Contracts**.

Contoh Kasus (Fitur *Account Deletion*):
*   Modul `auth` harus menghapus semua data `todos` milik seorang *user* ketika akun *user* tersebut dihapus.
*   Alih-alih langsung mengimpor modul `todos`, modul `auth` hanya mendefinisikan apa yang ia butuhkan lewat sebuah *interface* di dalam `internal/modules/auth/auth_domain.go`:
    ```go
    type TodoService interface {
        DeleteAllByUserID(ctx context.Context, userID uuid.UUID) error
    }
    ```
*   Secara paralel, modul `todos` mengimplementasikan fungsi tersebut di dalam *Usecase* miliknya.
*   Pada saat proses inisialisasi aplikasi di `cmd/server/main.go`, Usecase dari `todos` akan di-*inject* ke dalam fungsi pembuat (*constructor*) modul `auth`. Dengan begitu, modul `auth` bisa menjalankan fungsi hapus *todo* tanpa pernah terhubung langsung dengan *package* `todos` secara statis.

### Transaksi Lintas Modul (Cross-Module Transactions)

Untuk menjaga integritas data lintas modul tanpa membocorkan logika spesifik *database*, kita memanfaatkan struktur `TxManager` sebagai *cross-cutting concern*.

Ketika sebuah aksi lintas modul berjalan (seperti *Account Deletion* di atas), *parent module* hanya perlu membungkus semua logikanya ke dalam sebuah blok transaksi:
```go
func (s *Usecase) DeleteAccount(ctx context.Context, userID uuid.UUID) error {
    return s.txManager.RunInTx(ctx, func(txCtx context.Context) error {
        if err := s.todos.DeleteAllByUserID(txCtx, userID); err != nil {
            return err
        }
        if err := s.users.Delete(txCtx, userID); err != nil {
            return err
        }
        return nil
    })
}
```
Setiap *query* di dalam blok tersebut—dari *repository* maupun modul mana saja—akan dieksekusi secara atomik menggunakan koneksi yang disematkan pada `txCtx`. Hal ini menjamin bahwa seluruh data (*user* maupun *todos*) sukses terhapus, atau akan ter-*rollback* semuanya secara bersamaan jika terjadi kegagalan.

---

## Arsitektur Observability

```mermaid
flowchart TD
    subgraph Application
        API["API Service<br/>(Aplikasi Go)"]
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

## Struktur Projek

```text
.
├── cmd/
│   ├── migrate/      # Utility stand-alone untuk proses migrasi database
│   └── server/       # Entrypoint aplikasi utama tempat di mana modul-modul di-wire
├── internal/
│   ├── config/       # Konfigurasi aplikasi (Viper)
│   ├── delivery/     # Setup root HTTP server (Middleware & Router Utama)
│   ├── modules/      # Seluruh fitur-fitur bisnis berada di sini
│   │   ├── auth/         # Fitur Auth (login, register, user management, dsb)
│   │   └── todos/        # Fitur Todo (operasi CRUD)
│   └── shared/       # Cross-cutting utilities yang bisa digunakan secara aman oleh semua modul (errors, httpapi, observability, database, jwt, redis)
```

---

## Mulai Menggunakan (Getting Started)

### Prasyarat (Prerequisites)
*   Go 1.26+
*   Docker & Docker Compose (untuk *database* di tahap *local development*)
*   Make

### 1. Konfigurasi Awal

Lakukan duplikasi dari file contoh *environment* dan sesuaikan nilainya dengan setup komputermu.
```bash
cp .env.example .env
```

### 2. Menjalankan di Lokal (Hanya via Docker Compose)
Cara termudah adalah menggunakan Docker Compose. Aplikasi API beserta databasenya akan langsung tereksekusi.
```bash
make docker-up
```
Aplikasi API akan aktif dan tersedia di `http://localhost:8080`.

### 3. Menjalankan di Lokal (Go Secara Nativ + Postgres di Docker)
Kalau kamu lebih suka menjalankan proses Go-nya secara langsung di terminal, kamu bisa menyalakan hanya databasenya saja dari *docker*.
1.  Nyalakan *database* saja secara daemon (*background*):
    ```bash
    docker compose up postgres -d
    ```
2.  Jalankan aplikasinya (proses migrasi *database* sudah tereksekusi secara otomatis saat aplikasi menyala):
    ```bash
    make run
    ```

---

## Panduan Development

Terdapat banyak *shorthand* pada `Makefile` untuk mempermudah rutinitas pengembangan sistem.

*   **Format & Tidy**: `make tidy`
*   **Lint**: `make lint`
*   **Vet**: `make vet`
*   **Test**: `make test`
*   **Coverage**: `make coverage`
*   **Build**: `make build`

### Migrasi Database (Database Migrations)
Semua skrip SQL untuk migrasi database sudah disematkan ke dalam kode aplikasi (*embedded*) dan otomatis akan dieksekusi begitu server dijalankan. Walau begitu, jika kamu butuh memicu migrasi secara manual dari CLI, kamu bisa memanggil:
```bash
make migrate-up
make migrate-down
```

---

## Dokumentasi API

Selama server aplikasi berjalan dalam kondisi normal, dokumentasi interaktif API dapat diakses melalui link-link berikut:

*   **Swagger UI**: [http://localhost:8080/docs](http://localhost:8080/docs)
*   **OpenAPI 3.1 JSON**: [http://localhost:8080/openapi.json](http://localhost:8080/openapi.json)

---

## Konfigurasi Variabel (`.env`)

| Variabel | Deskripsi | Default |
| :--- | :--- | :--- |
| `APP_ENV` | Mode operasi server (`development` atau `production`) | `development` |
| `APP_NAME` | Nama dari aplikasi | `restful-template` |
| `APP_DESCRIPTION` | Deskripsi singkat dari aplikasi | `Template RESTful API menggunakan Go 1.26` |
| `HTTP_PORT` | Port tujuan dimana aplikasi menerima *request* API | `8080` |
| `DATABASE_DSN` | Connection string yang menuju ke database PostgreSQL | `postgres://todo:todo@localhost:5432/todo?sslmode=disable` |
| `JWT_SECRET` | Kunci sandi / Secret key untuk men-sign payload JWT (**WAJIB** diganti pada environment production) | `change-me-in-production-min-32-bytes!` |
| `JWT_ACCESS_TTL` | Umur maksimal Access Token sebelum *expired* | `15m` |
| `JWT_REFRESH_TTL` | Umur maksimal Refresh Token sebelum *expired* | `168h` |
| `LOG_LEVEL` | Tingkat prioritas verbositas keluaran Log (`debug`, `info`, `warn`, `error`) | `info` |
| `LOG_FORMAT` | Tipe *formatting log* (`json` direkomendasikan untuk *production*, `text` untuk lokal) | `json` |
| `TELEMETRY_OTLP_ENDPOINT` | OpenTelemetry gRPC target endpoint url | `localhost:4317` |
