# OPA Authorization Guide

Dokumentasi ini menjelaskan **Open Policy Agent (OPA)** dari dasar sampai konsep yang lebih advanced, sekaligus membedah policy Rego yang dipakai pada proyek ini.

---

## Overview

OPA adalah **policy engine** untuk authorization. Ia membantu aplikasi memisahkan:

* **Authentication**: siapa user-nya
* **Authorization**: user ini boleh melakukan apa
* **Business logic**: aksi utama aplikasi

Pada stack ini, alurnya biasanya seperti berikut:

```text
React -> Go API -> OPA -> allow / deny
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

### 2. RBAC

RBAC = **Role-Based Access Control**.

Hak akses diberikan berdasarkan role.

Contoh:

* `admin` → semua akses
* `user` → akses terbatas

Di policy ini, RBAC dipakai lewat mapping:

```rego
role_permissions := {
    "admin": [...],
    "user": [...]
}
```

---

### 3. Ownership Check

Selain role, policy ini juga mengecek kepemilikan resource.

Contoh:

* user boleh mengakses todo miliknya sendiri
* admin boleh mengakses semua todo

Ini sudah mulai mendekati **RBAC + ABAC sederhana** karena ada atribut tambahan seperti `resource_owner_id` dan `user_id`.

---

## Arsitektur Sederhana

```text
[React SPA]
    |
    | login / request
    v
[Go API]
    |
    | verify token / build input
    v
[OPA]
    |
    | evaluate Rego policy
    v
[allow / deny]
```

### Tugas tiap komponen

**React**

* hanya untuk UI
* menampilkan menu sesuai permission
* tidak boleh dipercaya untuk security final

**Go API**

* memverifikasi token
* mengambil data resource
* mengirim input ke OPA
* menegakkan keputusan allow/deny

**OPA**

* membaca input
* mengevaluasi policy
* mengembalikan keputusan authorization

---

## Policy yang Dipakai

Berikut policy Rego yang dijelaskan dalam dokumentasi ini:

```rego
package authz

import rego.v1

default allow = false

# Role-permission mapping (RBAC)
role_permissions := {
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

# Check if the active role has the permission
role_has_permission(role, perm) if {
    role_permissions[role][_] == perm
}

# Allow evaluation
allow if {
    role_has_permission(input.active_role, input.action)
    is_authorized_for_resource
}

is_authorized_for_resource if {
    # If no resource owner is specified, ownership check is not applicable
    not input.resource_owner_id
}

is_authorized_for_resource if {
    # Admin can access any resource
    input.active_role == "admin"
}

is_authorized_for_resource if {
    # Owner can access their own resource
    input.resource_owner_id == input.user_id
}
```

---

## Penjelasan Struktur Policy

### `package authz`

```rego
package authz
```

Ini menentukan namespace policy.

Artinya rule di dalam file ini nanti bisa diakses sebagai:

```text
data.authz.allow
```

---

### `import rego.v1`

Baris ini mengaktifkan sintaks Rego modern.

Biasanya dipakai agar penulisan policy lebih rapi dan mengikuti gaya terbaru.

---

### `default allow = false`

Ini berarti:

* jika tidak ada rule `allow` yang terpenuhi
* maka hasil akhirnya adalah `false`

Konsep ini penting karena authorization sebaiknya **default deny**.

---

### `role_permissions`

```rego
role_permissions := {
    "admin": [...],
    "user": [...]
}
```

Ini adalah mapping dari role ke daftar permission.

Contoh:

* role `admin` punya permission untuk todo dan user management
* role `user` hanya punya permission dasar

---

### `role_has_permission(role, perm)`

```rego
role_has_permission(role, perm) if {
    role_permissions[role][_] == perm
}
```

Rule ini mengecek apakah sebuah role punya permission tertentu.

Kalau disederhanakan ke gaya Go, logikanya mirip seperti:

```go
func RoleHasPermission(role string, perm string) bool {
    for _, p := range rolePermissions[role] {
        if p == perm {
            return true
        }
    }
    return false
}
```

---

### `allow if { ... }`

```rego
allow if {
    role_has_permission(input.active_role, input.action)
    is_authorized_for_resource
}
```

Ini rule utama yang menentukan request boleh atau tidak.

Agar `allow` bernilai `true`, dua syarat harus terpenuhi:

1. role aktif punya permission untuk aksi itu
2. resource juga lolos pemeriksaan akses

---

### `is_authorized_for_resource`

Rule ini punya beberapa kemungkinan benar. Artinya cukup salah satu bernilai true.

#### 1. Tidak ada resource owner

```rego
is_authorized_for_resource if {
    not input.resource_owner_id
}
```

Kalau resource tidak punya owner spesifik, ownership check tidak dilakukan.

Contoh:

* list todo
* stats
* endpoint umum yang tidak terkait satu resource tertentu

---

#### 2. Admin boleh akses semua

```rego
is_authorized_for_resource if {
    input.active_role == "admin"
}
```

Admin tidak dibatasi oleh ownership.

---

#### 3. Owner boleh akses resource miliknya sendiri

```rego
is_authorized_for_resource if {
    input.resource_owner_id == input.user_id
}
```

Kalau resource milik user itu sendiri, maka akses diizinkan.

---

## Cara Membaca Policy Ini

Policy ini bisa dibaca seperti kalimat berikut:

> User boleh melakukan aksi jika role-nya punya permission untuk aksi tersebut, dan resource yang diakses sesuai dengan aturan ownership atau admin override.

Dengan kata lain:

```text
ALLOW = Role Permission Match AND Resource Access Match
```

---

## Cara Kerja `allow`

Banyak yang mengira `allow` adalah variable biasa. Dalam Rego, `allow` lebih tepat dianggap sebagai **rule**.

Contoh:

```rego
default allow = false

allow if {
    ...
}
```

Artinya:

* default nilainya `false`
* jika ada rule yang berhasil, hasilnya menjadi `true`

Jadi OPA tidak sekadar membaca variable, tetapi **mencari bukti** apakah rule `allow` bisa dipenuhi.

---

## Contoh Input

### 1. User biasa mengedit todo miliknya sendiri

```json
{
  "user_id": "123",
  "active_role": "user",
  "action": "todo:update",
  "resource_owner_id": "123"
}
```

Hasil:

* permission ada
* resource milik sendiri
* **allow = true**

---

### 2. User biasa mengedit todo milik orang lain

```json
{
  "user_id": "123",
  "active_role": "user",
  "action": "todo:update",
  "resource_owner_id": "999"
}
```

Hasil:

* permission mungkin ada
* resource bukan miliknya
* **allow = false**

---

### 3. Admin mengakses resource apa pun

```json
{
  "user_id": "1",
  "active_role": "admin",
  "action": "user:delete",
  "resource_owner_id": "999"
}
```

Hasil:

* admin punya permission
* admin boleh akses resource apa pun
* **allow = true**

---

## Kenapa Ada `404 Not Found` di Implementasi Go?

Dalam implementasi Go, kadang hasil deny untuk resource tertentu dikembalikan sebagai `404 Not Found`.

Tujuannya adalah untuk menghindari:

* resource enumeration
* user mengetahui resource ada tetapi tidak boleh diakses

Dengan pendekatan ini, aplikasi tidak membocorkan informasi bahwa resource tersebut sebenarnya ada.

---

## Flow Request End-to-End

```text
1. User login via Keycloak
2. React menerima access token
3. React memanggil Go API
4. Go API memverifikasi token
5. Go API membangun input authorization
6. Go API mengirim input ke OPA
7. OPA mengevaluasi policy
8. OPA mengembalikan allow / deny
9. Go API mengeksekusi atau menolak request
```

---

## Kapan Policy Ini Cocok Dipakai?

Policy ini cocok untuk:

* todo app
* admin dashboard
* CMS internal
* sistem approval sederhana
* aplikasi dengan role + ownership

Tidak terlalu kompleks, tetapi sudah lebih kuat daripada RBAC murni karena ada ownership awareness.

---

## Kelebihan Pendekatan Ini

* policy terpisah dari business logic
* mudah dibaca dan dirawat
* default deny
* bisa berkembang ke ABAC atau policy yang lebih kompleks
* cocok untuk aplikasi Go yang butuh authorization rapi

---

## Hal yang Bisa Dikembangkan

Setelah policy ini stabil, beberapa pengembangan yang umum dilakukan:

### 1. Permission per resource type

Contoh:

* `todo:create`
* `user:update`
* `report:export`

### 2. ABAC tambahan

Contoh:

* branch harus sama
* department harus sama
* status resource harus `draft` atau `pending`

### 3. Decision object

Bukan hanya `allow` atau `deny`, tetapi juga alasan keputusan.

Contoh:

```rego
decision := {
  "allow": true,
  "reason": "owner"
}
```

### 4. Policy per domain

Bisa dipisah menjadi:

* `authz.rego`
* `todo.rego`
* `user.rego`
* `report.rego`

---

## Integrasi ke Go

Secara umum, Go akan:

* membaca user identity dari context
* membaca role aktif dan roles list
* mengambil resource owner id jika diperlukan
* mengirim semua itu ke OPA
* mengecek hasil `allow`

Pattern ini membuat handler tetap bersih dan authorization logic tidak tersebar di banyak tempat.

---

## Ringkasan Konsep

```text
package authz          -> namespace policy
allow                  -> nama rule keputusan
default allow = false  -> default deny
role_permissions       -> mapping role ke permission
role_has_permission    -> cek permission berdasarkan role
is_authorized_for_resource -> cek ownership / admin override
```

---

## Kesimpulan

Policy ini adalah contoh authorization yang sudah cukup matang untuk aplikasi modern:

* memakai RBAC sebagai dasar
* menambahkan ownership check
* mengikuti prinsip default deny
* mudah dikembangkan ke ABAC atau policy yang lebih advanced

Kalau kamu paham policy ini, kamu sudah punya pondasi yang kuat untuk membangun authorization yang rapi di Go + React + Keycloak + OPA.

---

## Next Step

Langkah berikut yang paling berguna biasanya adalah:

1. menambahkan diagram arsitektur
2. menambahkan contoh request/response API
3. menambahkan contoh implementasi middleware Go
4. menambahkan contoh policy testing dengan OPA
