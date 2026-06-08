CREATE TABLE IF NOT EXISTS users (
    id            UUID        NOT NULL DEFAULT uuidv7() PRIMARY KEY,
    email         TEXT        NOT NULL UNIQUE,
    password_hash TEXT,
    google_id     TEXT        UNIQUE,
    active_role   TEXT        REFERENCES roles(name) DEFAULT 'user',
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS user_roles (
    user_id   UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role_name TEXT NOT NULL REFERENCES roles(name) ON DELETE CASCADE,
    PRIMARY KEY (user_id, role_name)
);
