# Restful Template API

Template RESTful API *production-grade* yang dibangun menggunakan Go 1.26. Template ini mengusung arsitektur **Modular Monolith (Package by Feature)**, *routing* berkinerja tinggi menggunakan Chi, *auto-generation* OpenAPI 3.1 dengan Huma v2, serta menggunakan PostgreSQL 18 untuk men-generate UUID yang aman dan terurut berdasarkan waktu (*time-ordered*).

## Daftar Isi
- [Restful Template API](#restful-template-api)
  - [Daftar Isi](#daftar-isi)
  - [Fitur Utama](#fitur-utama)
  - [Arsitektur Modular Monolith](#arsitektur-modular-monolith)
    - [Arsitektur "Main calls run()"](#arsitektur-main-calls-run)
    - [Aturan Main (Rules of Engagement)](#aturan-main-rules-of-engagement)
    - [Komunikasi Antar Modul (Inter-Module Communication)](#komunikasi-antar-modul-inter-module-communication)
    - [Transaksi Lintas Modul (Cross-Module Transactions)](#transaksi-lintas-modul-cross-module-transactions)
  - [Arsitektur Observability](#arsitektur-observability)
    - [Penjelasan Alur Data (Flow Breakdown)](#penjelasan-alur-data-flow-breakdown)
      - [1. Alur Metrics — *Pull-based via Prometheus*](#1-alur-metrics--pull-based-via-prometheus)
      - [2. Alur Traces — *Push-based via OTLP ke Alloy → Tempo*](#2-alur-traces--push-based-via-otlp-ke-alloy--tempo)
      - [3. Alur Logs — *Pull from Docker via Alloy → Loki*](#3-alur-logs--pull-from-docker-via-alloy--loki)
    - [Komponen-Komponen dalam Stack](#komponen-komponen-dalam-stack)
    - [Unified Error Handling \& Observability](#unified-error-handling--observability)
  - [Struktur Projek](#struktur-projek)
  - [Mulai Menggunakan (Getting Started)](#mulai-menggunakan-getting-started)
    - [Prasyarat (Prerequisites)](#prasyarat-prerequisites)
    - [1. Konfigurasi Awal](#1-konfigurasi-awal)
    - [2. Menjalankan di Lokal (Hanya via Docker Compose)](#2-menjalankan-di-lokal-hanya-via-docker-compose)
    - [3. Menjalankan di Lokal (Go Secara Nativ + Postgres di Docker)](#3-menjalankan-di-lokal-go-secara-nativ--postgres-di-docker)
  - [Panduan Development](#panduan-development)
    - [Migrasi Database (Database Migrations)](#migrasi-database-database-migrations)
    - [Integration Testing](#integration-testing)
  - [Dokumentasi API](#dokumentasi-api)
  - [Konfigurasi Variabel (`.env`)](#konfigurasi-variabel-env)

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

### Arsitektur "Main calls run()"
Semua *dependency injection*, *database wiring*, dan *router setup* tidak dilakukan langsung di `main()`. Sebaliknya, `main.go` sangat tipis dan hanya menginisiasi `app.Setup()`. Ini memungkinkan pengujian integrasi *(integration tests)* memutar infrastruktur yang 100% identik dengan produksi tanpa *bypass* komponen krusial apa pun.

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

Template ini dilengkapi dengan *observability stack* yang lengkap berbasis **Grafana LGTM** (Loki, Grafana, Tempo, Mimir/Prometheus). Tumpukan ini mengimplementasikan tiga pilar observability modern — **Metrics**, **Traces**, dan **Logs** — yang semuanya dapat dikunjungi lewat satu antarmuka terpusat: **Grafana UI**.

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

Diagram di atas dapat dibaca sebagai tiga alur data yang berjalan secara paralel dan independen.

#### 1. Alur Metrics — *Pull-based via Prometheus*

*   Aplikasi Go meng-*expose* data *metrics* (jumlah *request*, latensi, *error rate*, dsb.) dalam format teks Prometheus di *endpoint* **`/metrics`**.
*   **Prometheus** secara aktif "*menarik*" (*scrape*) data dari endpoint tersebut secara berkala (setiap 15 detik, misalnya). Inilah mengapa mekanisme ini disebut **pull-based**.
*   Prometheus menyimpan semua *metrics* tersebut dalam *time-series database* miliknya.
*   **Grafana** kemudian menggunakan bahasa kueri **PromQL** untuk membaca data dari Prometheus dan menampilkannya sebagai grafik dan *dashboard*.

#### 2. Alur Traces — *Push-based via OTLP ke Alloy → Tempo*

*   Ketika sebuah HTTP *request* masuk ke API, *middleware* `otelchi` secara otomatis membuat sebuah **Trace** (jejak distribusi) beserta **Span** (satuan unit kerja) di dalamnya.
*   Secara paralel, *instrumentation* `otelpgx` menciptakan *span* anak (*child span*) untuk setiap *query* yang dikirim ke PostgreSQL, sehingga kamu bisa melihat secara persis berapa lama waktu yang dihabiskan di *layer* database.
*   Semua Trace dikirim secara asinkron dari aplikasi ke **Grafana Alloy** menggunakan protokol **OTLP melalui gRPC** (port `4317`). Ini adalah mekanisme **push-based**.
*   **Alloy** berperan sebagai *telemetry collector* — ia menerima *traces* lalu meneruskannya ke **Grafana Tempo**.
*   **Grafana Tempo** menyimpan *traces* tersebut dan Grafana bisa menelusurinya menggunakan bahasa kueri **TraceQL**.

#### 3. Alur Logs — *Pull from Docker via Alloy → Loki*

*   Aplikasi Go mencetak *log* terstruktur (format JSON) ke `stdout`/`stderr`. Aplikasi **tidak perlu tahu** ke mana log ini akan pergi — ini adalah *concern* dari infrastruktur, bukan aplikasi.
*   Docker menangkap semua output `stdout`/`stderr` tersebut dan menyimpannya melalui mekanisme *logging driver* standarnya.
*   **Grafana Alloy** diberikan akses ke *Docker Socket* (`/var/run/docker.sock`). Melalui akses ini, Alloy secara aktif *menarik* log dari semua *container* yang berjalan (*container discovery*) dan secara otomatis menambahkan label seperti nama *service* (`service="api"`) ke setiap baris log.
*   Alloy kemudian mendorong (*push*) semua log tersebut ke **Grafana Loki** melalui HTTP (port `3100`).
*   **Grafana Loki** menyimpan log dan Grafana dapat menggali informasinya menggunakan bahasa kueri **LogQL**.

### Komponen-Komponen dalam Stack

| Komponen | Peran | Port |
| :--- | :--- | :--- |
| **Grafana Alloy** | *Telemetry Collector*: Menerima OTLP Traces dari aplikasi dan menarik logs dari Docker, lalu meneruskan ke *backend* masing-masing. | `4317` (gRPC), `4318` (HTTP) |
| **Prometheus** | *Metrics Backend*: Menyimpan *time-series metrics* yang di-*scrape* langsung dari *endpoint* `/metrics` aplikasi. | `9090` |
| **Grafana Loki** | *Logs Backend*: Menyimpan log terstruktur yang dikirimkan oleh Alloy. Dioptimalkan untuk penyimpanan log bervolume tinggi. | `3100` |
| **Grafana Tempo** | *Traces Backend*: Menyimpan *distributed traces* dalam format yang mendukung penelusuran (*root cause analysis*) berbasis TraceQL. | `3200` |
| **Grafana UI** | *Visualization Layer*: Satu antarmuka terpusat untuk mengeksplorasi Metrics (PromQL), Logs (LogQL), dan Traces (TraceQL). Mendukung *correlation* antar sinyal. | `3000` |

### Unified Error Handling & Observability

Projek ini menerapkan **Secure Idiomatic Error Handling**.
*   **SafeError (`internal/shared/errors`)**: Kesalahan internal disembunyikan menggunakan wrapper `SafeError` yang mencegah bocornya informasi kredensial, struktur path, atau detail SQL ke pengguna (client). Semua kesalahan internal dikonversi secara aman menjadi kode HTTP seperti 500 atau 400 menggunakan `httpapi.ToHumaErr()`.
*   **Wide Events (`internal/shared/wideevent`)**: Alih-alih membuat banyak statement *log* yang berantakan di setiap *layer*, aplikasi mengumpulkan data logis di sepanjang siklus hidup *request* menggunakan *context*, dan mengeluarkannya sebagai **Satu Baris Log per Request** di layer *middleware*. Format ini sangat cocok untuk mesin analitik seperti Loki atau Elasticsearch.

---

## Struktur Projek

```text
.
├── cmd/
│   └── server/       # Entrypoint aplikasi utama, cukup mendelegasikan eksekusi ke internal/app
├── internal/
│   ├── app/          # Tempat di mana seluruh dependency injection dan wiring disatukan (Setup)
│   ├── config/       # Konfigurasi aplikasi (Viper)
│   ├── delivery/     # Setup root HTTP server (Middleware & Router Utama)
│   ├── modules/      # Seluruh fitur-fitur bisnis berada di sini
│   │   ├── auth/         # Fitur Auth (login, register, user management, dsb)
│   │   └── todos/        # Fitur Todo (operasi CRUD)
│   └── shared/       # Cross-cutting utilities yang bisa digunakan secara aman oleh semua modul (errors, httpapi, observability, database, jwt, redis)
├── tests/            # Test Infrastruktur, berisi Black-box End-to-End Integration Tests
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
Semua skrip SQL untuk migrasi database sudah disematkan ke dalam kode aplikasi (*embedded* melalui paket `embed` bawaan golang) dan otomatis akan dieksekusi begitu server dijalankan atau saat Integration Tests berjalan. Tidak diperlukan command khusus untuk menjalankannya.

### Integration Testing
Pengujian *end-to-end* menggunakan `testcontainers-go`. Semua skrip pengujian (seperti di folder `tests/`) otomatis membangkitkan dan memanajemen container Docker sementara (PostgreSQL & Redis) untuk menjalankan API persis seperti versi produksi melalui eksekusi `app.Setup()`.
Gunakan `make test-integration` untuk mengeksekusi ini.

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
