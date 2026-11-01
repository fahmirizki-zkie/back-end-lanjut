-- Migrasi 005: Index komposit untuk mendukung keyset pagination (cursor) pada tabel students.
-- Index ini mencocokkan urutan query ORDER BY created_at DESC, id DESC
-- sehingga database dapat langsung melompat ke posisi kursor menggunakan Index Scan.

CREATE INDEX IF NOT EXISTS students_created_at_id_desc_idx
ON students (created_at DESC, id DESC);
