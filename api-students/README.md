# API Students

REST API untuk pengelolaan data mahasiswa.

## Base URL

```text
http://localhost:3000
```

## Kontrak API

| Metode | Endpoint | Parameter | Contoh Body Permintaan | Status yang Mungkin Dikembalikan | Contoh Respons |
|---|---|---|---|---|---|
| GET | `/api/v1/students` | `page`, `limit`, `search`, `sort`, `order`, `is_active` | Tidak ada | `200`, `400` | `{"success":true,"message":"daftar student berhasil diambil","data":[...],"meta":{...}}` |
| GET | `/api/v1/students/:id` | `id` | Tidak ada | `200`, `400`, `404` | `{"success":true,"message":"student ditemukan","data":{...}}` |
| POST | `/api/v1/students` | Tidak ada | `{"name":"Fahmi","email":"fahmi@gmail.com","nim":"434241097","grade":85,"password":"password123"}` | `201`, `400`, `409`, `415`, `422` | `{"success":true,"message":"student berhasil dibuat","data":{...}}` |
| PUT | `/api/v1/students/:id` | `id` | `{"name":"Fahmi Rizky","email":"fahmirizky@gmail.com","nim":"434241097","grade":90,"is_active":false}` | `200`, `400`, `404`, `409`, `415`, `422` | `{"success":true,"message":"student berhasil diubah","data":{...}}` |
| PATCH | `/api/v1/students/:id` | `id` | `{"grade":95}` | `200`, `400`, `404`, `409`, `415`, `422` | `{"success":true,"message":"student berhasil diubah","data":{...}}` |
| DELETE | `/api/v1/students/:id` | `id` | Tidak ada | `204`, `400`, `404` | Tidak ada body |
