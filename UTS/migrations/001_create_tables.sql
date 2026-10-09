-- ============================================================
-- UTS Pemrograman Back-End: SIAKAD Mini Database Migration
-- Database: siakad_mini_fahmi
-- ============================================================

-- 1. Tabel users
CREATE TABLE IF NOT EXISTS users (
    id          BIGSERIAL PRIMARY KEY,
    email       VARCHAR(255) NOT NULL UNIQUE,
    password    VARCHAR(255) NOT NULL,
    role        VARCHAR(20)  NOT NULL DEFAULT 'mahasiswa'
                CHECK (role IN ('admin', 'mahasiswa')),
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

-- 2. Tabel students
CREATE TABLE IF NOT EXISTS students (
    id             BIGSERIAL PRIMARY KEY,
    user_id        BIGINT       NOT NULL UNIQUE REFERENCES users(id) ON DELETE CASCADE,
    nim            VARCHAR(12)  NOT NULL UNIQUE,
    nama           VARCHAR(255) NOT NULL,
    prodi          VARCHAR(255) NOT NULL,
    angkatan       INTEGER      NOT NULL,
    ipk_terakhir   NUMERIC(3,2) DEFAULT 0.00
                   CHECK (ipk_terakhir >= 0.00 AND ipk_terakhir <= 4.00),
    deleted_at     TIMESTAMPTZ,
    created_at     TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at     TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

-- 3. Tabel courses
CREATE TABLE IF NOT EXISTS courses (
    id         BIGSERIAL PRIMARY KEY,
    kode_mk    VARCHAR(20)  NOT NULL UNIQUE,
    nama_mk    VARCHAR(255) NOT NULL,
    sks        INTEGER      NOT NULL CHECK (sks > 0),
    semester   INTEGER      NOT NULL CHECK (semester > 0),
    kuota      INTEGER      NOT NULL CHECK (kuota > 0),
    created_at TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

-- 4. Tabel enrollments
CREATE TABLE IF NOT EXISTS enrollments (
    id             BIGSERIAL PRIMARY KEY,
    student_id     BIGINT       NOT NULL REFERENCES students(id) ON DELETE CASCADE,
    course_id      BIGINT       NOT NULL REFERENCES courses(id) ON DELETE CASCADE,
    tahun_akademik VARCHAR(50)  NOT NULL,
    created_at     TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    UNIQUE (student_id, course_id, tahun_akademik)
);

-- Indexes untuk optimasi query
CREATE INDEX IF NOT EXISTS idx_students_nim ON students(nim);
CREATE INDEX IF NOT EXISTS idx_students_prodi ON students(prodi);
CREATE INDEX IF NOT EXISTS idx_students_deleted_at ON students(deleted_at);
CREATE INDEX IF NOT EXISTS idx_courses_kode_mk ON courses(kode_mk);
CREATE INDEX IF NOT EXISTS idx_enrollments_student ON enrollments(student_id);
CREATE INDEX IF NOT EXISTS idx_enrollments_course ON enrollments(course_id);
