# Restful Template API

[![Go Version](https://img.shields.io/badge/Go-1.26-00ADD8?style=flat-square&logo=go)](https://go.dev/)
[![License: Apache 2.0](https://img.shields.io/badge/License-Apache_2.0-blue?style=flat-square)](LICENSE)
[![OpenAPI](https://img.shields.io/badge/OpenAPI-3.1-6BA539?style=flat-square&logo=swagger)](http://localhost:8080/docs)
[![OpenTelemetry](https://img.shields.io/badge/OpenTelemetry-enabled-f5a800?style=flat-square&logo=opentelemetry)](https://opentelemetry.io/)
[![CI](https://img.shields.io/github/actions/workflow/status/semmidev/restful-template/ci.yml?label=CI&style=flat-square)](https://github.com/semmidev/restful-template/actions/workflows/ci.yml)
[![CD](https://img.shields.io/github/actions/workflow/status/semmidev/restful-template/cd.yml?label=CD&style=flat-square)](https://github.com/semmidev/restful-template/actions/workflows/cd.yml)

Template ini mengusung arsitektur **Modular Monolith (Package by Feature)**, *routing* berkinerja tinggi menggunakan Chi, *auto-generation* OpenAPI 3.1 dengan Huma v2, serta menggunakan PostgreSQL 18 untuk men-generate UUID yang aman dan terurut berdasarkan waktu (*time-ordered*).

## Preview

![API Documentation](assets/api-docs.png)
![API Monitoring](assets/api-monitoring.png)

## Tech Stack

| Layer | Teknologi |
| :--- | :--- |
| **Language** | Go 1.26 |
| **Router** | [Chi v5](https://github.com/go-chi/chi) |
| **API Framework** | [Huma v2](https://github.com/danielgtaylor/huma) — OpenAPI 3.1 auto-generation |
| **Database** | PostgreSQL 18 via `pgxpool`, UUID v7 native |
| **Query Builder** | [Squirrel](https://github.com/Masterminds/squirrel) |
| **Migrations** | [golang-migrate](https://github.com/golang-migrate/migrate) — embedded ke binary |
| **Cache** | Redis via `go-redis/v9` |
| **Auth** | JWT (Access + Refresh) · Argon2id password hashing |
| **Observability** | OpenTelemetry · Prometheus · Grafana LGTM Stack |
| **Config** | [Viper](https://github.com/spf13/viper) — `.env` + OS env vars |
| **Testing** | [testcontainers-go](https://github.com/testcontainers/testcontainers-go) — E2E Integration Tests |

---

## Daftar Isi

- [Quick Start](#quick-start)
- [Fitur Utama](#fitur-utama)
- [Struktur Projek](#struktur-projek)
- [Arsitektur Modular Monolith](#arsitektur-modular-monolith)
- [Arsitektur Observability](#arsitektur-observability)
- [Panduan Development](#panduan-development)
- [Dokumentasi API](#dokumentasi-api)
- [Konfigurasi Variabel (.env)](#konfigurasi-variabel-env)
- [Panduan Infrastruktur & Deployment](infra/README.md)
- [Contributing](#contributing)
- [License](#license)

---

## Quick Start

> **Prasyarat**: Go 1.26+, Docker & Docker Compose, Make

```bash
# 1. Clone & masuk ke direktori
git clone https://github.com/semmidev/restful-template.git
cd restful-template

# 2. Salin file konfigurasi
cp .env.example .env

# 3. Jalankan seluruh stack (API + DB + Observability)
make docker-up
```

API akan aktif di **`http://localhost:8080`**.
Dokumentasi interaktif tersedia di **`http://localhost:8080/docs`**.

### Alternatif: Jalankan Go Secara Nativ

Jika kamu lebih suka menjalankan proses Go langsung di terminal (lebih cepat untuk iterasi development):

```bash
# Nyalakan database saja via Docker
docker compose up postgres redis -d

# Jalankan server (migrasi DB berjalan otomatis saat startup)
make run
```

---

## Fitur Utama

- **Architecture & Design**
  - **Modular Monolith (Package by Feature)**: Batasan (*boundaries*) yang tegas mencegah *technical debt* dan arsitektur *ball-of-mud*.
  - *Synchronous Interface Injection* untuk *Consumer-Driven Contracts* yang bersih antar modul.
  - Transaksi lintas modul dikelola secara aman menggunakan `database.TxManager`.
  - **Dedicated Background Scheduler**: Menjalankan cron jobs terpisah dari API utama menggunakan `gocron/v2`.

- **Routing & Documentation**
  - Integrasi Huma v2 + Chi v5 dengan *typed HTTP handlers* dan validasi otomatis.
  - *Auto-generation* spesifikasi **OpenAPI 3.1** (`/openapi.json`) & Swagger UI (`/docs`).

- **Database & Persistence**
  - PostgreSQL 18 melalui `pgxpool` + **Native UUID v7** (*time-ordered*, anti-fragmentasi).
  - Sistem migrasi *embedded* ke binary — tidak butuh CLI eksternal di *runtime*.
  - Pagination dan *advanced filtering* (*full-text substring keyword search*).

- **Security & Authentication**
  - JWT dengan Access + Refresh tokens (rotasi otomatis).
  - *Password hashing* Argon2id (format PHC, verifikasi *constant-time*).
  - CORS siap pakai + Middleware *Secure HTTP Headers*.

- **Observability & Reliability**
  - *Structured JSON/Text Logging* dengan `log/slog` dan pola **Wide Event** (satu baris log per request).
  - **OpenTelemetry** (OTEL) untuk *distributed tracing* (`otelchi`) dan DB traces (`otelpgx`).
  - *Graceful shutdown*, propagasi Request ID, *panic recovery middleware*.
  - Format error RFC 9457 (`application/problem+json`).

- **Configuration & Deployment**
  - Standar *12-Factor App* via Viper (`.env` + OS environment variables).
  - Dockerfile *multi-stage distroless* — image kecil dan aman.
  - `docker-compose` lengkap dengan *healthchecks*, *restart policies*, dan full observability stack.
  - **CI/CD** otomatis via GitHub Actions (golangci-lint & E2E Testing).

---

## Arsitektur Enterprise-Grade (Deep Dive)

Proyek ini melampaui sekadar kerangka kerja RESTful biasa dan menerapkan pola *Enterprise-Grade* yang kokoh:

1. **Pola "Canonical Log Line" (Wide Events)**
   - Tidak ada lagi ribuan *log statements* yang tercecer. Middleware secara cerdas mengumpulkan metadata (durasi, trace ID, user ID, status HTTP, error) ke dalam *context*, dan menghasilkan tepat **satu baris log terstruktur per request**. Ini adalah *best-practice* industri untuk menunjang mesin analitik (*log aggregator*) seperti Loki atau Elasticsearch.
2. **Injeksi Kontrak Sinkron (Decoupled Modular Monolith)**
   - Ketergantungan (dependencies) antar modul sepenuhnya dilonggarkan melalui abstraksi *interface* (Consumer-Driven Contracts). Modul `auth` berinteraksi dengan `todos` tanpa impor langsung. Jika kelak aplikasi perlu dipecah menjadi *Microservices*, *overhead* migrasi kodenya mendekati nol.
3. **Cross-Module Transaction Manager**
   - Mendukung eksekusi operasi secara atomik antar modul yang terisolasi. Jika terjadi kegagalan (misalnya gagal menghapus `todos` saat akun `user` dihapus), `TxManager` akan me-*rollback* seluruh rentetan transaksi database tanpa *leakage* data.
4. **Fail-Open Rate Limiter & Security Suite**
   - Pembatasan laju (*Rate Limiter*) berbasis Redis dirancang agar **Fail-Open**. Artinya, jika server Redis mengalami *down*, permintaan klien tetap diproses agar API tidak ikut lumpuh total. Didukung dengan lapisan *Security Headers* (HSTS, nosniff, frame-options) bawaan.
5. **Observability Zero-Configuration**
   - Jejak terdistribusi (*Trace ID*) dari OpenTelemetry secara cerdas disuntikkan kembali ke dalam *HTTP Response Header* (`X-Trace-Id`). Klien atau *Frontend Developer* yang mendapatkan error bisa melaporkan ID tersebut, dan Anda bisa langsung melihat urutan *SQL query* mana yang memicu *bug* di dasbor Grafana Tempo.
6. **Dedicated Background Scheduler (Cron Jobs) & Distributed Locking**
   - Menjalankan tugas di latar belakang (seperti membersihkan *refresh token* kedaluwarsa) langsung dari web API rentan terhadap isu *race conditions* dan pemborosan CPU ketika aplikasi di-*scale* secara horizontal (banyak replika).
   - Proyek ini memisahkan *scheduler* menjadi *binary* dan *container* independen (`cmd/scheduler`). *Logic* pekerjaannya (*job logic*) tetap terenkapsulasi secara modular di dalam domainnya (contoh: `AuthJob` di modul `auth`) dengan arsitektur yang bersih tanpa perlu menyentuh *query* database mentah secara langsung.
   - **Distributed Locking via Redis**: Scheduler dilengkapi dengan implementasi `gocron.Locker` kustom (`internal/shared/redis/locker.go`) yang menggunakan Redis `SetNX`. Mekanisme ini menjamin bahwa sebuah tugas tidak akan pernah tumpang tindih (*overlap*) dengan eksekusi sebelumnya, dan menjaga sistem tetap aman meskipun *scheduler* di-*scale* ke beberapa *instance*. Terdapat *TTL auto-expire* untuk mencegah *deadlock* jika *worker* mendadak lumpuh saat memegang kunci.

---

## Integrasi Huma v2 Tingkat Lanjut

Template ini tidak sekadar menggunakan Huma sebagai generator OpenAPI, melainkan memanfaatkan kapabilitas optimal Huma v2 untuk performa, concurrency, dan Developer Experience (DX):

1. **Conditional Requests & Optimistic Locking (`If-Match` / `If-None-Match`)**
   - API mendukung HTTP standar ETag dan header Last-Modified (misalnya pada operasi `Todo`).
   - Mencegah masalah *lost update* di data *concurrent* (Huma otomatis merespons `412 Precondition Failed` jika ETag pada *request PATCH* tidak sesuai dengan database).
   - Menghemat bandwidth: merespons dengan `304 Not Modified` jika data client masih sinkron (`If-None-Match`).
2. **CBOR Content Negotiation**
   - Mendukung format respons `application/cbor` secara otomatis.
   - Tanpa merombak kode *handler*, client yang mengirimkan header `Accept: application/cbor` akan mendapatkan respons biner yang ukurannya ~30% lebih kecil dibanding JSON tradisional.
3. **Pagination Tipe RFC 8288 (`Link` Header)**
   - Format pagination data-list (misal, GET `/todos`) memanfaatkan HTTP Header `Link` bawaan, membuat parsing link `next`/`prev`/`last` lebih bersih layaknya GitHub API.
   - Header Pagination dideklarasikan secara eksplisit dalam respons struct dan dapat dilihat di Swagger UI!
4. **Schema Discovery (`SchemaLinkTransformer`)**
   - Semua body respons menyertakan header `Link: <url>; rel="describedby"`, mengarahkan ke definisi JSON Schema.
   - Hal ini membuat API kita sepenuhnya *self-descriptive* sehingga *tools* seperti VSCode atau IDE bisa memberikan *as-you-type validation* dan *autocomplete* untuk integrasi API di sisi frontend.
5. **Ultra-Fast Handler Unit Tests (`humatest`)**
   - Handler API memiliki unit test yang sepenuhnya terisolasi dan berjalan dalam hitungan *milidetik* via router lokal Huma `humatest.New()`, sehingga mempermudah TDD (Test-Driven Development) tanpa ketergantungan Docker container.

---

## Struktur Projek

```text
.
├── cmd/
│   ├── server/       # Entrypoint utama API
│   └── scheduler/    # Entrypoint terpisah untuk background cron jobs (gocron/v2)
├── config/           # Konfigurasi infrastruktur (Prometheus, Grafana, Loki, Tempo, Alloy)
├── internal/
│   ├── app/          # Dependency injection & wiring terpusat (Setup)
│   ├── config/       # Konfigurasi aplikasi (Viper)
│   ├── delivery/     # Setup root HTTP server (Middleware & Router Utama)
│   ├── modules/      # Seluruh fitur bisnis
│   │   ├── auth/         # Auth: login, register, user management
│   │   └── todos/        # Todo: operasi CRUD
│   └── shared/       # Cross-cutting utilities (errors, httpapi, observability, database, jwt, redis, wideevent)
└── tests/            # Black-box End-to-End Integration Tests
```

> **Pola "Main calls run()"**: Semua *dependency injection* dan *wiring* tidak dilakukan di `main()`. Fungsi `main.go` sangat tipis dan hanya menginisiasi `app.Setup()`, sehingga *integration tests* bisa memutar infrastruktur yang **100% identik** dengan produksi.

---

## Arsitektur Modular Monolith

Projek ini menggunakan *pattern* **Package by Feature** — kode dikelompokkan berdasarkan **fitur bisnis** (contoh: `auth`, `todos`), bukan fungsi teknisnya. Pendekatan ini mendorong *high cohesion* dan *low coupling* yang optimal untuk aplikasi berskala besar.

### Aturan Main (Rules of Engagement)

1. **Tidak ada *direct imports* antar modul** — `internal/modules/auth` dilarang meng-*import* `internal/modules/todos` secara langsung.
2. **Shared Kernel** — Kode generik yang dibutuhkan lebih dari satu modul diletakkan di `internal/shared`.
3. **Encapsulation** — Semua struktur internal modul (handlers, usecase, repositories) bersifat *private*. Hanya *public constructors* (seperti `NewAuth`) yang boleh di-*wire* dari `cmd/server/main.go`.

### Komunikasi Antar Modul (Inter-Module Communication)

Ketika modul perlu berinteraksi dengan modul lain, kita gunakan **Synchronous Interface Injection** (mengacu pada **Consumer-Driven Contracts**).

**Contoh — Fitur *Account Deletion*:**

Modul `auth` harus menghapus data `todos` milik *user* saat akun dihapus. Alih-alih meng-*import* modul `todos`, modul `auth` mendefinisikan kontrak yang ia butuhkan:

```go
// internal/modules/auth/auth_domain.go
type TodoService interface {
    DeleteAllByUserID(ctx context.Context, userID uuid.UUID) error
}
```

Modul `todos` mengimplementasikan kontrak ini, lalu di-*inject* ke `auth` saat inisialisasi di `cmd/server/main.go`. Hasilnya: `auth` memanggil `todos` tanpa pernah terhubung secara statis.

**Mengapa Service/Usecase diekspos sebagai Interface?**
- **Isolated Unit Testing**: Layer HTTP (`delivery/http`) bergantung pada *interface*, bukan *concrete struct*. Ini memungkinkan pembuatan unit test untuk HTTP Handler secara terisolasi menggunakan *mock* (tanpa setup database/Redis).
- **Mencegah Circular Dependencies**: Jika modul `auth` bergantung pada konkrit struktur `todos`, dan suatu saat `todos` butuh fungsi dari `auth`, akan terjadi *circular import* (yang dilarang keras di Go). Interface menjaga kedua modul tetap terisolasi.
- **Architectural Boundary**: Compiler Go menjamin bahwa layer luar (seperti router/HTTP) hanya bisa memanggil fungsi yang terdaftar secara eksplisit di kontrak *interface*, mencegah kebocoran logika internal *usecase*.

### Transaksi Lintas Modul (Cross-Module Transactions)

Integritas data lintas modul dijaga menggunakan `TxManager` sebagai *cross-cutting concern*:

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

Semua *query* dalam blok tersebut dieksekusi secara **atomik** — sukses semua atau *rollback* semua.

---

## Arsitektur Observability

Template ini dilengkapi *observability stack* lengkap berbasis **Grafana LGTM** (Loki, Grafana, Tempo, Prometheus) yang mengimplementasikan tiga pilar observability modern — **Metrics**, **Traces**, dan **Logs** — dalam satu antarmuka terpusat: **Grafana UI** (`http://localhost:3000`).

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

### Penjelasan Alur Data (Flow Breakdown)

Diagram di atas merepresentasikan tiga alur data yang berjalan secara paralel dan independen.

#### 1. Alur Metrics — *Pull-based via Prometheus*

- Aplikasi meng-*expose* metrics (jumlah *request*, latensi, *error rate*, dsb.) di endpoint **`/metrics`** dalam format teks Prometheus.
- **Prometheus** secara aktif men-*scrape* endpoint tersebut secara berkala (**pull-based**) dan menyimpannya dalam *time-series database*.
- **Grafana** membaca data via **PromQL** dan menampilkannya sebagai grafik di *dashboard*.

#### 2. Alur Traces — *Push-based via OTLP ke Alloy → Tempo*

- Setiap HTTP *request* yang masuk diinstrumentasi oleh middleware `otelchi` yang otomatis membuat **Trace** + **Span**.
- `otelpgx` menambahkan *child span* untuk setiap *query* ke PostgreSQL — sehingga kamu bisa melihat persis berapa waktu yang dihabiskan di *layer* database.
- Traces dikirim secara asinkron dari aplikasi ke **Grafana Alloy** via **OTLP/gRPC** (port `4317`) — mekanisme **push-based**.
- **Alloy** meneruskan traces ke **Grafana Tempo** untuk disimpan dan ditelusuri via **TraceQL**.

#### 3. Alur Logs — *Pull from Docker via Alloy → Loki*

- Aplikasi hanya menulis *structured log* (JSON) ke `stdout`/`stderr` — ia tidak perlu tahu ke mana log akan dikirim.
- Docker menangkap output tersebut melalui *logging driver* standarnya.
- **Grafana Alloy** (yang memiliki akses ke `/var/run/docker.sock`) secara aktif *menarik* log dari semua *container* yang berjalan (*container discovery*) dan menambahkan label seperti `service="api"` secara otomatis.
- Alloy mendorong (*push*) log tersebut ke **Grafana Loki** (HTTP port `3100`) untuk digali via **LogQL**.

### Komponen Stack

| Komponen | Peran | Port |
| :--- | :--- | :--- |
| **Grafana Alloy** | *Telemetry Collector*: menerima OTLP Traces & menarik logs dari Docker | `4317` (gRPC), `4318` (HTTP) |
| **Prometheus** | *Metrics Backend*: *time-series database* untuk metrics aplikasi | `9090` |
| **Grafana Loki** | *Logs Backend*: penyimpanan log bervolume tinggi | `3100` |
| **Grafana Tempo** | *Traces Backend*: *distributed tracing* dengan TraceQL | `3200` |
| **Grafana UI** | *Visualization Layer*: eksplorasi Metrics + Logs + Traces dalam satu UI | `3000` |

### Unified Error Handling & Observability

Projek ini menerapkan **Secure Idiomatic Error Handling**:

- **`SafeError` (`internal/shared/errors`)**: Kesalahan internal disembunyikan menggunakan wrapper `SafeError` yang mencegah bocornya informasi kredensial, path, atau detail SQL ke *client*. Semua kesalahan internal dikonversi aman menjadi kode HTTP (500/400) via `httpapi.ToHumaErr()`.
- **Wide Events (`internal/shared/wideevent`)**: Alih-alih banyak statement *log* yang tersebar di tiap *layer*, aplikasi mengumpulkan data logis sepanjang siklus *request* menggunakan `context`, lalu mengeluarkannya sebagai **satu baris log per request** di *middleware*. Format ini optimal untuk mesin analitik seperti Loki atau Elasticsearch.

---

## Panduan Development

### Perintah Make

| Perintah | Deskripsi |
| :--- | :--- |
| `make run` | Menjalankan server secara lokal |
| `make docker-up` | Menjalankan seluruh stack via Docker Compose |
| `make tidy` | Format kode & tidy dependencies |
| `make lint` | Menjalankan linter |
| `make vet` | Menjalankan `go vet` |
| `make test` | Menjalankan unit tests |
| `make test-integration` | Menjalankan E2E integration tests (membutuhkan Docker) |
| `make coverage` | Menghasilkan laporan code coverage |
| `make build` | Build binary produksi |

### Migrasi Database

Semua skrip SQL migrasi sudah di-*embed* langsung ke dalam binary aplikasi (menggunakan paket `embed` bawaan Go) dan **otomatis dieksekusi** saat server dijalankan atau *integration tests* berjalan. Tidak diperlukan perintah atau CLI eksternal tambahan.

### Integration Testing

Pengujian *end-to-end* menggunakan `testcontainers-go`. Setiap test otomatis:
1. Membangkitkan container Docker sementara (PostgreSQL & Redis).
2. Menjalankan API via `app.Setup()` — **100% identik** dengan setup produksi.
3. Membersihkan container setelah test selesai.

```bash
make test-integration
```

---

## Dokumentasi API

Selama server berjalan, dokumentasi interaktif dapat diakses di:

| Interface | URL |
| :--- | :--- |
| **Swagger UI** | [http://localhost:8080/docs](http://localhost:8080/docs) |
| **OpenAPI 3.1 JSON** | [http://localhost:8080/openapi.json](http://localhost:8080/openapi.json) |
| **Grafana Dashboard** | [http://localhost:3000](http://localhost:3000) |

---

## Konfigurasi Variabel (`.env`)

Salin `.env.example` ke `.env` dan sesuaikan nilai-nilainya.

| Variabel | Deskripsi | Default |
| :--- | :--- | :--- |
| `APP_ENV` | Mode operasi (`development` / `production`) | `development` |
| `APP_NAME` | Nama aplikasi | `restful-template` |
| `APP_DESCRIPTION` | Deskripsi singkat aplikasi | `Template RESTful API menggunakan Go 1.26` |
| `HTTP_PORT` | Port server HTTP | `8080` |
| `DATABASE_DSN` | Connection string PostgreSQL | `postgres://todo:todo@localhost:5432/todo?sslmode=disable` |
| `JWT_SECRET` | Secret key untuk signing JWT (**WAJIB** diganti di production) | `change-me-in-production-min-32-bytes!` |
| `JWT_ACCESS_TTL` | Masa berlaku Access Token | `15m` |
| `JWT_REFRESH_TTL` | Masa berlaku Refresh Token | `168h` |
| `LOG_LEVEL` | Verbositas log (`debug`, `info`, `warn`, `error`) | `info` |
| `LOG_FORMAT` | Format log (`json` untuk production, `text` untuk lokal) | `json` |
| `TELEMETRY_OTLP_ENDPOINT` | OpenTelemetry gRPC target endpoint | `localhost:4317` |

---

## Contributing

Kontribusi sangat disambut! Silakan buka *issue* untuk melaporkan bug atau mendiskusikan fitur baru sebelum membuka *Pull Request*.

1. Fork repositori ini.
2. Buat branch baru: `git checkout -b feat/nama-fitur`
3. Commit perubahan: `git commit -m 'feat: tambahkan fitur X'`
4. Push branch: `git push origin feat/nama-fitur`
5. Buka Pull Request.

---

## License

Didistribusikan di bawah lisensi **Apache 2.0**. Lihat file [LICENSE](LICENSE) untuk detail lebih lanjut.
