CREATE TABLE IF NOT EXISTS roles (
    name        TEXT        PRIMARY KEY,
    description TEXT,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

INSERT INTO roles (name, description) VALUES
('admin', 'Administrator with full access'),
('user', 'Standard user with default access')
ON CONFLICT (name) DO NOTHING;
