

-- Tabel users menyimpan akun yang bisa login ke API.
CREATE TABLE IF NOT EXISTS users (
    id         SERIAL      PRIMARY KEY,
    username   VARCHAR(50) NOT NULL,
    email      VARCHAR(100) NOT NULL,
    password   VARCHAR(255) NOT NULL,
    role       VARCHAR(20) NOT NULL DEFAULT 'user',
    is_active  BOOLEAN     NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Username unik, case-insensitive.
CREATE UNIQUE INDEX IF NOT EXISTS users_username_key
    ON users (LOWER(username));

-- Email unik.
CREATE UNIQUE INDEX IF NOT EXISTS users_email_key
    ON users (LOWER(email));

-- Refresh token disimpan sebagai HASH SHA-256, bukan nilai aslinya.
CREATE TABLE IF NOT EXISTS refresh_tokens (
    id         BIGSERIAL   PRIMARY KEY,
    user_id    INTEGER     NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash TEXT        NOT NULL UNIQUE,
    expires_at TIMESTAMPTZ NOT NULL,
    revoked_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Index untuk mempercepat pencarian token berdasarkan user.
CREATE INDEX IF NOT EXISTS refresh_tokens_user_id_idx
    ON refresh_tokens (user_id);
