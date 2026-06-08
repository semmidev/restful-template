# OPA Authorization Guide

Dokumentasi ini menjelaskan **Open Policy Agent (OPA)** dari dasar sampai konsep yang lebih advanced, sekaligus membedah policy Rego dan teknik manajemen policy yang dinamis, thread-safe, dan scalable pada proyek ini.

---

## Overview

OPA adalah **policy engine** untuk authorization. Ia membantu aplikasi memisahkan:

* **Authentication**: siapa user-nya
* **Authorization**: user ini boleh melakukan apa
* **Business logic**: aksi utama aplikasi

Pada stack ini, alurnya biasanya seperti berikut:

```text
React -> Go API -> OPA (Embedded) -> allow / deny
```

Keycloak atau sistem auth lain bertugas mengeluarkan identitas user, sedangkan OPA bertugas memutuskan apakah request tersebut diizinkan.

---

## Kenapa OPA?

Kalau authorization masih sederhana, RBAC biasa di backend sudah cukup. Namun OPA mulai sangat berguna ketika:

* role makin banyak
* aturan akses makin kompleks
* keputusan akses bergantung pada atribut user dan resource
* policy ingin dipisah dari business logic
* aturan harus lebih mudah diuji, dirawat, dan diubah

Contoh kasus yang cocok untuk OPA:

* admin boleh semua
* user biasa hanya boleh milik sendiri
* manager boleh approve jika nominal di bawah batas tertentu
* akses data tergantung branch, department, status, atau ownership

---

## Konsep Dasar

### 1. Authentication vs Authorization

* **Authentication** menjawab: *siapa kamu?*
* **Authorization** menjawab: *kamu boleh apa?*

OPA fokus ke authorization.

---

### 2. RBAC Dinamis (Dynamic RBAC)

RBAC = **Role-Based Access Control**.

Hak akses diberikan berdasarkan role. 

Dalam proyek ini, hak akses tidak lagi ditulis secara hardcoded dalam file Rego. Sebagai gantinya, data permission disimpan secara dinamis dalam file JSON (`permissions.json`) dan dimuat secara dinamis ke dalam ruang memori OPA saat runtime:

```json
{
  "admin": [
    "todo:create", "todo:read", "todo:update", "todo:delete", "todo:list", "todo:stats",
    "auth:delete_account", "auth:switch_role",
    "user:create", "user:read", "user:update", "user:delete", "user:list"
  ],
  "user": [
    "todo:create", "todo:read", "todo:update", "todo:delete", "todo:list", "todo:stats",
    "auth:delete_account", "auth:switch_role"
  ]
}
```

Hal ini memungkinkan penambahan permission atau role baru tanpa perlu mengubah logika rule Rego atau melakukan kompilasi ulang (redeploy) kode policy utama.

---

### 3. Multiple Role Evaluation

Seringkali, seorang user memiliki lebih dari satu role (misalnya `user` sekaligus `billing_admin`). Policy pada proyek ini dirancang agar dapat mengevaluasi **seluruh array role** (`input.roles`) yang dimiliki oleh user, bukan hanya satu role aktif saja. Jika salah satu dari role yang dimiliki user memiliki permission untuk aksi tersebut, maka akses akan diizinkan.

---

### 4. Ownership & ABAC (Attribute-Based Access Control)

Selain role, policy ini juga mengecek atribut resource secara dinamis (ABAC). 

Alih-alih menggunakan variabel flat yang kaku (seperti `input.resource_owner_id`), kita menggunakan struktur nested map `input.resource` yang dapat dikembangkan untuk menampung berbagai atribut resource lainnya di masa depan (misal: `status`, `created_at`, `org_id`):

```rego
is_authorized_for_resource if {
    input.resource.owner_id == input.user_id
}
```

---

## Arsitektur & Manajemen Policy

```text
[React SPA]
    |
    | login / request
    v
[Go API (Embedded OPA)]
    |
    | Build input map (User ID, Roles, Action, Resource Context)
    v
[atomic.Pointer[Evaluator]]
    |
    | evaluates rules against data.permissions.role_permissions
    v
allow / deny
```

### Tugas tiap komponen

* **React SPA**: Mengirim token access dalam HTTP request.
* **Go API Server**:
  * Mengekstrak identitas user dan daftar roles dari context request (didapat dari JWT token).
  * Memanggil evaluator OPA secara thread-safe menggunakan `atomic.Pointer`.
  * Mendukung dynamic **Hot-Reloading** lewat fungsi `Reload(ctx, permissionsJSON)` untuk mengganti aturan role-permission saat aplikasi berjalan tanpa downtime.
* **OPA Engine (In-Process)**: Mengevaluasi rule Rego berdasarkan input dan dynamic data permissions.

---

## Policy yang Dipakai

Berikut policy Rego yang digunakan dalam proyek ini:

```rego
package authz

import rego.v1

default allow = false

# Mengambil mapping permission dari data dynamic permissions package
role_permissions := data.permissions.role_permissions

# Mengecek apakah salah satu dari daftar roles milik user memiliki permission tertentu
role_has_permission(roles, perm) if {
    some role in roles
    role_permissions[role][_] == perm
}

# Evaluasi perizinan utama
allow if {
    role_has_permission(input.roles, input.action)
    is_authorized_for_resource
}

is_authorized_for_resource if {
    # Jika tidak ada resource owner yang dispesifikasikan, ownership check dilewati
    not input.resource.owner_id
    not input.resource_owner_id
}

is_authorized_for_resource if {
    # Admin dibolehkan mengakses resource apapun
    "admin" in input.roles
}

is_authorized_for_resource if {
    # Owner diperbolehkan mengakses resourcenya sendiri (legacy)
    input.resource_owner_id == input.user_id
}

is_authorized_for_resource if {
    # Owner diperbolehkan mengakses resourcenya sendiri (nested resource map)
    input.resource.owner_id == input.user_id
}
```

---

## Penjelasan Struktur Policy

### `package authz`

Menentukan namespace policy. Di Go, evaluator menunjuk namespace ini via query:
```go
rego.Query("data.authz.allow")
```

### `role_permissions`

Mengambil permission list dari modul data dinamis `permissions` yang di-inject dari memory store Go pada runtime.

### `role_has_permission(roles, perm)`

Menggunakan loop/quantifier `some role in roles` untuk mengecek apakah setidaknya salah satu role user mengandung permission target. Logika Go-nya setara dengan:

```go
func RoleHasPermission(roles []string, perm string) bool {
    for _, role := range roles {
        for _, p := range rolePermissions[role] {
            if p == perm {
                return true
            }
        }
    }
    return false
}
```

### `is_authorized_for_resource`

Fungsi ABAC yang memvalidasi ownership resource secara aman. Jika resource memiliki metadata `owner_id`, OPA akan membandingkannya dengan `input.user_id`.

---

## Contoh Input OPA

### 1. User biasa mengedit todo miliknya sendiri
```json
{
  "user_id": "user-123",
  "roles": ["user"],
  "action": "todo:update",
  "resource": {
    "owner_id": "user-123"
  }
}
```
**Hasil**: `allow = true` (karena permission `todo:update` ada pada role `user`, dan user_id cocok dengan owner_id).

### 2. User biasa mengedit todo milik orang lain
```json
{
  "user_id": "user-123",
  "roles": ["user"],
  "action": "todo:update",
  "resource": {
    "owner_id": "user-999"
  }
}
```
**Hasil**: `allow = false` (karena owner_id tidak cocok dengan user_id).

---

## Kelebihan Teknik Dynamic & Thread-Safe OPA di Go

1. **Zero-Downtime Hot-Reloading**: Melalui `atomic.Pointer` di Go, kita dapat memanggil `Reload` untuk mengganti memori dynamic permissions tanpa memblokir request yang sedang berjalan (Lock-free Read).
2. **Flexible ABAC Expansion**: Variabel nested `resource` map pada input OPA memungkinkan penulisan policy berdasarkan status task, department, IP range, dll., tanpa perlu mengubah tipe struct di Go.
3. **No Redundant Re-compilation**: Dengan memisahkan deklarasi policy (.rego) dari skema permission (.json), perubahan pemetaan permission dapat dilakukan secara runtime (misal ditarik dari Redis cache atau database PostgreSQL).
