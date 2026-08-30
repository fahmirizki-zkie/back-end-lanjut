# REST API Student

## Skema Tabel

Tabel yang digunakan adalah `students`.

| Kolom | Tipe | Keterangan |
|---|---|---|
| `id` | integer | Primary Key |
| `nim` | varchar | NIM dan harus unik |
| `name` | varchar | Nama student |
| `grade` | numeric | Nilai student |
| `is_active` | boolean | Status aktif student |
| `created_at` | timestamptz | Waktu data dibuat |

Index tambahan:
- `students_name_idx` pada kolom `name`.

## Cara Menyiapkan Basis Data dari Nol

1. Pastikan PostgreSQL sudah terpasang dan berjalan.
2. Buat database PostgreSQL.
3. Jalankan SQL/migration pada project untuk membuat tabel `students`.
4. Buat file `.env` berdasarkan `.env.example`.
5. Isi kredensial database pada file `.env`.
6. Jalankan aplikasi dengan:

```bash
go run .
```

## Environment Variables

Variabel environment yang diperlukan:

```env
DB_HOST=
DB_PORT=
DB_USER=
DB_PASSWORD=
DB_NAME=
DB_SSLMODE=
```

Keterangan:

- `DB_HOST` = host PostgreSQL
- `DB_PORT` = port PostgreSQL
- `DB_USER` = username PostgreSQL
- `DB_PASSWORD` = password PostgreSQL
- `DB_NAME` = nama database
- `DB_SSLMODE` = mode SSL PostgreSQL


