CREATE TABLE IF NOT EXISTS users (
    id           UUID        NOT NULL DEFAULT uuidv7() PRIMARY KEY,
    email        TEXT        NOT NULL UNIQUE,
    password_hash TEXT       NOT NULL,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
