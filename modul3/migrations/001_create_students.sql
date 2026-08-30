CREATE TABLE IF NOT EXISTS students (
    id         SERIAL PRIMARY KEY,
    nim        VARCHAR(50) NOT NULL,
    name       VARCHAR(100) NOT NULL,
    grade      NUMERIC(5,2) NOT NULL,
    is_active  BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- nim unik 
CREATE UNIQUE INDEX IF NOT EXISTS students_nim_key
ON students (nim);  

CREATE INDEX IF NOT EXISTS students_name_idx
ON students (name);