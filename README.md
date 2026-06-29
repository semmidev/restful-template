# Fullstack Web Application Template

[![Go Version](https://img.shields.io/badge/Go-1.26-00ADD8?style=flat-square&logo=go)](https://go.dev/)
[![PostgreSQL](https://img.shields.io/badge/PostgreSQL-18-4169E1?style=flat-square&logo=postgresql&logoColor=white)](https://www.postgresql.org/)
[![React](https://img.shields.io/badge/React-19-61DAFB?style=flat-square&logo=react&logoColor=black)](https://react.dev/)
[![License: Apache 2.0](https://img.shields.io/badge/License-Apache_2.0-blue?style=flat-square)](LICENSE)
[![CI](https://img.shields.io/github/actions/workflow/status/semmidev/restful-template/ci.yml?label=CI&style=flat-square)](https://github.com/semmidev/restful-template/actions/workflows/ci.yml)
[![CD](https://img.shields.io/github/actions/workflow/status/semmidev/restful-template/cd.yml?label=CD&style=flat-square)](https://github.com/semmidev/restful-template/actions/workflows/cd.yml)

Template ini dirancang untuk membangun aplikasi web dengan arsitektur Modular Monolith yang sederhana untuk dikembangkan, mudah dipelihara, dan siap berkembang seiring kebutuhan bisnis. Frontend dan backend terintegrasi dalam satu deployment unit, sehingga proses pengembangan, distribusi, dan operasional menjadi lebih efisien. Template ini juga telah dilengkapi dengan dokumentasi API otomatis serta fondasi yang siap digunakan untuk membangun aplikasi produksi.

Untuk panduan detail bagi AI Agent/Copilot dalam memahami standar engineering dan batasan arsitektur repositori ini, silakan merujuk ke [AGENTS.md](AGENTS.md).

## Preview

### API Documentation

<p align="center">
  <img src="./github/assets/api-docs.png" alt="API Documentation" width="100%">
</p>

## 🎨 Web UI

<p align="center">
  <a href="./github/assets/desktop-demo.webm">
    <img src="./github/assets/desktop-demo-thumbnail.png" alt="Web UI Preview" width="100%">
  </a>
  <br>
  <sub><i>Klik gambar di atas untuk memutar video demonstrasi lengkap.</i></sub>
</p>


### API Monitoring

<p align="center">
  <img src="./github/assets/api-monitoring.png" alt="API Monitoring" width="100%">
</p>

## Tech Stack

| Layer | Teknologi |
| :--- | :--- |
| **Frontend** | React 19 · TypeScript · Zustand 5 · Zod 4 · Vite 8 |
| **Styling & UI** | TailwindCSS v3 · Shadcn UI · Lucide |
| **Backend Language** | Go 1.26 |
| **Router** | [Chi v5](https://github.com/go-chi/chi) |
| **API Framework** | [Huma v2](https://github.com/danielgtaylor/huma) — OpenAPI 3.1 auto-generation |
| **Database** | PostgreSQL 18 via `pgxpool`, UUID v7 native |
| **Query Builder** | [Squirrel](https://github.com/Masterminds/squirrel) |
| **Migrations** | [golang-migrate](https://github.com/golang-migrate/migrate) — embedded ke binary |
| **Cache** | Redis via `go-redis/v9` |
| **Auth & Policy** | JWT Cookie-Based Session (Access + Refresh) · Argon2id password hashing · **Open Policy Agent (OPA)** (policy engine embedded) |
| **Observability** | OpenTelemetry · Prometheus · Grafana LGTM Stack |
| **Async Worker** | [asynq](https://github.com/hibiken/asynq) — Redis-backed task queue (separate binary) |
| **Worker UI** | [asynqmon](https://github.com/hibiken/asynqmon) — Web UI mounted at `/adm/asynq` (Basic Auth) |
| **Config** | [Viper](https://github.com/spf13/viper) — `.env` + OS env vars |
| **Testing** | [testcontainers-go](https://github.com/testcontainers/testcontainers-go) — E2E Integration Tests |

---

## Daftar Isi

- [Quick Start](#quick-start)
- [Fitur Utama](#fitur-utama)
- [Struktur Projek](#struktur-projek)
- [Arsitektur Modular Monolith](#arsitektur-modular-monolith)
- [Arsitektur Frontend (React SPA)](#arsitektur-frontend-react-spa)
- [Arsitektur Observability](#arsitektur-observability)
- [Panduan Development](#panduan-development)
- [Dokumentasi API](#dokumentasi-api)
- [Konfigurasi Variabel (.env)](#konfigurasi-variabel-env)
- [Panduan AI Agent (AGENTS.md)](AGENTS.md)
- [Panduan Infrastruktur & Deployment](infra/README.md)
- [Contributing](#contributing)
- [License](#license)

---

## Quick Start

> **Prasyarat**: Go 1.26+, Node.js 20+, Docker & Docker Compose, Make

```bash
# 1. Clone & masuk ke direktori
git clone https://github.com/semmidev/restful-template.git
cd restful-template

# 2. Salin file konfigurasi
cp .env.example .env

# 3. Jalankan seluruh stack (API + DB + Observability)
make docker-up
```

API dan SPA Frontend yang tersemat (*embedded*) akan aktif di **`http://localhost:8080`**.
Dokumentasi interaktif tersedia di **`http://localhost:8080/docs`**.

### Panduan Development Frontend (Lokal)

Jika kamu ingin menjalankan server frontend secara terpisah untuk memantau perubahan secara langsung (*Hot Module Replacement*):

```bash
# Masuk ke folder frontend dan install dependensi
cd frontend
npm install

# Jalankan Vite dev server
npm run dev
```
Dev server frontend akan aktif di **`http://localhost:5173`** dan melakukan proxy otomatis ke backend di port `8080`.

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
  - **Async Background Worker**: Menjalankan task asinkron (seperti kirim email) menggunakan `asynq` — Redis-backed, separate binary (`cmd/worker`).

- **Routing & Documentation**
  - Integrasi Huma v2 + Chi v5 dengan *typed HTTP handlers* dan validasi otomatis.
  - *Auto-generation* spesifikasi **OpenAPI 3.1** (`/openapi.json`) & Swagger UI (`/docs`).

- **Database & Persistence**
  - PostgreSQL 18 melalui `pgxpool` + **Native UUID v7** (*time-ordered*, anti-fragmentasi).
  - Sistem migrasi *embedded* ke binary — tidak butuh CLI eksternal di *runtime*.
  - Pagination dan *advanced filtering* (*full-text substring keyword search*).

- **Security & Authentication**
  - JWT dengan Access + Refresh cookies (rotasi otomatis, HTTP-Only, secure).
  - *Password hashing* Argon2id (format PHC, verifikasi *constant-time*).
  - **Otorisasi Terpusat OPA (Open Policy Agent)**: Integrasi RBAC (Role-Based Access Control) dan ABAC (Attribute-Based Access Control) dinamis berbasis policy Rego (`policy.rego`).
  - **User Management CRUD (Admin Only)**: Panel administrasi lengkap untuk manajemen pengguna (CRUD) terlindung dari self-lockout.
  - CORS siap pakai + Middleware *Secure HTTP Headers*.

- **Observability & Reliability**
  - *Structured JSON/Text Logging* dengan `log/slog` dan pola **Wide Event** (satu baris log per request).
  - **OpenTelemetry** (OTEL) untuk *distributed tracing* (`otelchi`) dan DB traces (`otelpgx`).
  - *Graceful shutdown*, propagasi Request ID, *panic recovery middleware*.
  - Format error RFC 9457 (`application/problem+json`).
  - **Asynqmon Web UI** (`/adm/asynq`) — pantau antrian task, worker, dan job history secara visual, dilindungi dengan Basic Auth.

- **Configuration & Deployment**
  - Standar *12-Factor App* via Viper (`.env` + OS environment variables).
  - Dockerfile *multi-stage distroless* untuk API, Scheduler, **dan Worker** — image kecil dan aman.
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
7. **Async Background Worker (Event-Driven Tasks)**
   - Untuk tugas *fire-and-forget* (seperti kirim email selamat datang setelah registrasi), proyek ini menggunakan pola **Task Distributor / Task Processor** berbasis `asynq`.
   - **Distributor** (`internal/shared/asynqtask`) memiliki peran sebagai *Producer* — modul memanggil `distributor.DistributeTask...()` via interface (`TaskDistributor`), lalu task diserialisasi dan dikirim ke Redis queue tanpa blocking request HTTP.
   - **Processor**: Handler task diimplementasikan langsung di dalam modul masing-masing (misalnya `auth_worker.go`) — `cmd/worker` binary mendengarkan Redis queue dan mengeksekusi handler task secara konkuren.
   - **Asynqmon Web UI** (`/admin/asynq`) built-in untuk memantau queue, retry, dan dead-letter tasks secara real-time.
8. **Google Login dengan OAuth 2.0 + PKCE (Proof Key for Code Exchange)**
   - Menggunakan alur otorisasi modern dan aman untuk aplikasi Single Page Application (SPA).
   - **Alur Kerja**:
     1. React SPA meminta konfigurasi publik (Client ID & Redirect URI) dari Go backend.
     2. React SPA men-generate `state` (anti-CSRF), `code_verifier`, dan `code_challenge` (SHA-256), lalu mengarahkan user ke halaman Google Login.
     3. Google mengautentikasi user dan mengarahkan kembali ke SPA callback (`/login/google/callback`) dengan membawa `code` and `state`.
     4. SPA memverifikasi `state`, mengambil `code_verifier`, lalu mengirimkan `code` dan `code_verifier` ke Go backend.
     5. Go backend melakukan pertukaran kode (*token exchange*) secara aman ke Google Token API. Google memverifikasi `code_verifier` terhadap `code_challenge` awal.
     6. Setelah divalidasi oleh Google, Go backend mengambil data profil user, mencocokkannya ke database (membuat user baru jika belum ada), dan menghasilkan JWT Session token (Access & Refresh) yang disimpan secara aman di dalam cookie HTTP-Only (Set-Cookie) untuk dikembalikan ke React SPA.

   ```mermaid
   sequenceDiagram
       autonumber
       actor User
       participant SPA as React SPA
       participant Backend as Go Backend
       participant Google as Google Auth Server

       User->>SPA: Klik "Continue with Google"
       SPA->>Backend: GET /api/v1/auth/google/config
       Backend->>SPA: Return Client ID & Redirect URI
       Note over SPA: Generate state, code_verifier & code_challenge (S256)
       SPA->>Google: Redirect dengan Client ID, Redirect URI & code_challenge
       Google->>User: Form login & consent Google
       User->>Google: Setujui otorisasi
       Google->>SPA: Redirect ke /login/google/callback?code=CODE&state=STATE
       Note over SPA: Validasi state (anti-CSRF)
       SPA->>Backend: POST /api/v1/auth/google (code & code_verifier)
       Backend->>Google: Post /token (exchange code & code_verifier)
       Google->>Backend: Return Access Token Google
       Backend->>Google: GET /userinfo (dengan Access Token)
       Google->>Backend: Return Profil User (Email & Google ID)
       Note over Backend: Transaksi DB: Create/Link User & Issue JWT Session (Cookies)
       Backend->>SPA: Return Cookie Set-Cookie (access_token & refresh_token)
       SPA->>User: Akses Workspace (Dashboard)
   ```

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
├── frontend/         # React SPA (Vite + TypeScript + Zustand + Zod)
│   ├── src/
│   │   ├── components/  # Global UI components
│   │   └── features/    # Domain modules (auth, todos)
├── cmd/
│   ├── server/       # Entrypoint utama API
│   ├── scheduler/    # Entrypoint terpisah untuk background cron jobs (gocron/v2)
│   └── worker/       # Entrypoint terpisah untuk async task worker (asynq)
├── config/           # Konfigurasi infrastruktur (Prometheus, Grafana, Loki, Tempo, Alloy)
├── internal/
│   ├── app/          # Dependency injection & wiring terpusat (Setup)
│   ├── config/       # Konfigurasi aplikasi (Viper)
│   ├── http/         # Setup root HTTP server (Middleware & Router Utama)
│   ├── auth/         # Auth: login, register, user management
│   ├── todo/         # Todo: operasi CRUD
│   ├── web/          # Embedded SPA server handler (go:embed)

│   └── shared/       # Cross-cutting utilities
│       ├── asynqtask/    # Task type definitions & Distributor (asynq producer)
│       └── ...           # errors, httpapi, observability, database, jwt, redis, wideevent
└── tests/            # Black-box End-to-End Integration Tests
```

> **Pola "Main calls run()"**: Semua *dependency injection* dan *wiring* tidak dilakukan di `main()`. Fungsi `main.go` sangat tipis dan hanya menginisiasi `app.Setup()`, sehingga *integration tests* bisa memutar infrastruktur yang **100% identik** dengan produksi.

---

## Arsitektur Modular Monolith

Projek ini menggunakan *pattern* **Package by Feature** — kode dikelompokkan berdasarkan **fitur bisnis** (contoh: `auth`, `todos`), bukan fungsi teknisnya. Pendekatan ini mendorong *high cohesion* dan *low coupling* yang optimal untuk aplikasi berskala besar.

### Aturan Main (Rules of Engagement)

1. **Tidak ada *direct imports* antar modul** — `internal/auth` dilarang meng-*import* `internal/todo` secara langsung.
2. **Shared Kernel** — Kode generik yang dibutuhkan lebih dari satu modul diletakkan di `internal/shared`.
3. **Encapsulation** — Semua struktur internal modul (handlers, service, repositories) bersifat *private*. Hanya *public constructors* (seperti `NewAuthService`) yang boleh di-*wire* dari `cmd/server/main.go`.

### Komunikasi Antar Modul (Inter-Module Communication)

Ketika modul perlu berinteraksi dengan modul lain, kita gunakan **Synchronous Interface Injection** (mengacu pada **Consumer-Driven Contracts**).

**Contoh — Fitur *Account Deletion*:**

Modul `auth` harus menghapus data `todos` milik *user* saat akun dihapus. Alih-alih meng-*import* modul `todos`, modul `auth` mendefinisikan kontrak yang ia butuhkan:

```go
// internal/auth/auth_domain.go
type TodoService interface {
    DeleteAllByUserID(ctx context.Context, userID uuid.UUID) error
}
```

Modul `todos` mengimplementasikan kontrak ini, lalu di-*inject* ke `auth` saat inisialisasi di `cmd/server/main.go`. Hasilnya: `auth` memanggil `todos` tanpa pernah terhubung secara statis.

**Mengapa Service/Usecase diekspos sebagai Interface?**
- **Isolated Unit Testing**: Layer HTTP (`internal/http`) bergantung pada *interface*, bukan *concrete struct*. Ini memungkinkan pembuatan unit test untuk HTTP Handler secara terisolasi menggunakan *mock* (tanpa setup database/Redis).
- **Mencegah Circular Dependencies**: Jika modul `auth` bergantung pada konkrit struktur `todos`, dan suatu saat `todos` butuh fungsi dari `auth`, akan terjadi *circular import* (yang dilarang keras di Go). Interface menjaga kedua modul tetap terisolasi.
- **Architectural Boundary**: Compiler Go menjamin bahwa layer luar (seperti router/HTTP) hanya bisa memanggil fungsi yang terdaftar secara eksplisit di kontrak *interface*, mencegah kebocoran logika internal *service*.

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

## Arsitektur Frontend (React SPA)

Frontend aplikasi dibangun sebagai Single Page Application (SPA) modern yang tangguh dengan performa tinggi dan desain pixel-perfect ala developer tools (Linear App).

### 1. Struktur Folder & Scoping Fitur
Mengikuti pola modular yang bersih, semua kode frontend berada di direktori `frontend/src/` dengan struktur sebagai berikut:
- **`components/`**: Berisi UI components global dan reusable yang berbasis Radix UI melalui Shadcn UI (seperti button, input, dialog, sidebar).
- **`lib/`**: Utilitas bersama yang digunakan lintas modul, termasuk Axios client (`client.ts`), skema Zod global (`schemas.ts`), dan utilitas styling Tailwind (`utils.ts`).
- **`features/`**: Folder modular berbasis domain bisnis (misal: `auth`, `todos`). Setiap modul fitur memiliki struktur flat tanpa subdirektori:
  - `api.ts`: Kumpulan fungsi pemanggilan HTTP request khusus untuk fitur tersebut.
  - `store.ts`: Zustand store untuk mengelola state dan async action terkait fitur.
  - `pages/`: Halaman (containers) utama yang didaftarkan pada router (misal: `Todos.tsx`, `Dashboard.tsx`).
  - `components/`: UI components lokal yang hanya digunakan di dalam modul fitur tersebut.

### 2. Autentikasi JWT & Queue Refresh Token Otomatis
Autentikasi terintegrasi dengan backend secara aman melalui cookie HTTP-Only dan rotasi otomatis di `frontend/src/lib/client.ts`:
- **Cookie-Based**: Browser secara otomatis mengelola cookie `access_token` dan `refresh_token` dengan flag `HttpOnly`, `Secure`, dan `SameSite=Lax` tanpa akses dari skrip client-side (mencegah serangan XSS).
- **Queue Refresh Token (Penanganan Error 401)**:
  - Jika request gagal dengan status `401 Unauthorized`, response interceptor akan mencoba memperbarui token secara otomatis melalui `/auth/refresh` (browser otomatis mengirimkan cookie `refresh_token` untuk path `/api/v1/auth`).
  - **Mekanisme Antrean (Queue)**: Untuk mencegah terjadinya beberapa request refresh token secara konkuren (yang bisa merusak rotasi token), flag `isRefreshing` digunakan. Request gagal berikutnya akan ditahan dalam Promise dan dimasukkan ke dalam `failedQueue`.
  - Jika refresh berhasil, seluruh antrean request dijalankan ulang secara otomatis menggunakan session yang diperbarui.
  - Jika refresh gagal, state auth lokal dibersihkan dan user diarahkan kembali ke `/login`.

### 3. Optimistic Locking & ETag di Frontend
Untuk menghindari konflik penulisan konkuren data (*lost update*):
- Setiap entri resource memiliki metadata `updated_at`.
- Saat frontend mengirimkan request update (seperti `PATCH /todos/:id`), Axios client secara otomatis melampirkan nilai `updated_at` tersebut ke dalam header `If-Match` (misal: `If-Match: "2026-06-07T10:49:43Z"`).
- Jika ada user lain yang telah mengedit todo tersebut terlebih dahulu di database, server akan merespons dengan `412 Precondition Failed`. Frontend menangkap error ini, menginfokan terjadinya konflik konkuren ke user, dan memicu re-fetch otomatis agar tampilan sinkron dengan server.

### 4. Standar Desain Minimalis (Linear Style)
Desain visual mengusung gaya minimalis, high-contrast, dan tajam:
- **HSL Semantic Variables**: Pewarnaan elemen UI dilarang keras menggunakan warna statis (seperti `#ffffff` atau `bg-slate-100`). Seluruh komponen menggunakan variabel HSL (seperti `bg-background`, `border-border`, `text-primary`) untuk memastikan adaptasi tema light/dark berjalan mulus.
- **Border & Shadow**: Penggunaan border dibuat sangat tipis (`border-border/80`) dengan sudut tumpul yang rapat (`rounded-md` atau `--radius: 0.375rem`) dan shadow minimalis (`shadow-none` atau `shadow-sm`) demi meniru estetika panel developer tools.
- **Grafik Recharts**: Seluruh bagan dibungkus dalam `ResponsiveContainer` agar responsif. Kustomisasi tooltip diselaraskan dengan visual kartu dashboard (`bg-card/95 border-border/80`).

### 5. Vite Dev Proxy & Code Splitting
- **Proxy Lokal**: Vite dev server dikonfigurasi untuk meneruskan semua request dengan prefix `/api` dan `/todos` ke backend server di port `8080`, menghindari isu CORS di lingkungan lokal.
- **Code Splitting (Lazy Loading)**: Semua halaman utama di-load menggunakan `React.lazy()` dan dibungkus di dalam `<Suspense>` di `App.tsx` untuk memastikan load awal yang sangat cepat dengan memisahkan bundle halaman.

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
| `make migrate-create` | Membuat file migrasi baru (meminta input nama migrasi) |
| `make migrate-up` | Menjalankan seluruh migrasi ke atas (up) |
| `make migrate-down` | Menjalankan seluruh migrasi ke bawah (down) |
| `make migrate-rollback` | Membatalkan satu migrasi terakhir (down 1) |
| `make migrate-force` | Memaksa database ke versi migrasi tertentu (meminta input versi) |
| `make migrate-version` | Melihat versi migrasi saat ini di database |

### Linter & Isolasi Workspace

Projek Go ini menggunakan `golangci-lint` untuk menjaga kualitas kode backend.
- **Isolasi Node Modules**: Untuk mencegah linter memindai direktori `frontend/node_modules/` secara rekursif (yang dapat menyebabkan kegagalan linter), sebuah file dummy `frontend/go.mod` diletakkan di dalam folder frontend. Hal ini secara efektif mengisolasi folder frontend dari modul utama Go saat linter menjalankan perintah `golangci-lint run ./...` dari root workspace.

### Migrasi Database

Semua skrip SQL migrasi sudah di-*embed* langsung ke dalam binary aplikasi (menggunakan paket `embed` bawaan Go) dan **otomatis dieksekusi** saat server dijalankan (`DATABASE_RUN_MIGRATIONS=true` pada `.env`) atau *integration tests* berjalan.

Untuk manajemen migrasi secara manual atau pembuatan file migrasi baru pada environment development lokal, Anda dapat menggunakan perintah Make berikut (secara default terhubung ke database di `localhost:5432`):

- **Membuat Migrasi Baru**: `make migrate-create` (akan meminta input nama file migrasi)
- **Menjalankan Migrasi (Up)**: `make migrate-up` (menjalankan semua migrasi pending)
- **Membatalkan Semua Migrasi (Down)**: `make migrate-down`
- **Membatalkan Satu Migrasi Terakhir**: `make migrate-rollback`
- **Force Version**: `make migrate-force` (memaksa database ke versi tertentu jika terjadi status dirty)
- **Cek Versi Saat Ini**: `make migrate-version`

*Tips: Jika DSN database berbeda dari default, Anda dapat mengirimkannya sebagai parameter, contoh: `make migrate-up DB_DSN=postgres://user:pass@host:port/dbname?sslmode=disable`*

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
| **Asynqmon Worker UI** | [http://localhost:8080/admin/asynq](http://localhost:8080/admin/asynq) (login: `ASYNQMON_USERNAME` / `ASYNQMON_PASSWORD`) |
| **Grafana Dashboard** | [http://localhost:3000](http://localhost:3000) |

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
