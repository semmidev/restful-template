CREATE TABLE IF NOT EXISTS roles (
    name        TEXT        PRIMARY KEY,
    description TEXT,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

INSERT INTO roles (name, description) VALUES
('admin', 'Administrator with full access'),
('user', 'Standard user with default access')
ON CONFLICT (name) DO NOTHING;

CREATE TABLE IF NOT EXISTS user_roles (
    user_id   UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role_name TEXT NOT NULL REFERENCES roles(name) ON DELETE CASCADE,
    PRIMARY KEY (user_id, role_name)
);

ALTER TABLE users ADD COLUMN active_role TEXT REFERENCES roles(name) DEFAULT 'user';

-- Migrate existing users to have default role and active role
UPDATE users SET active_role = 'user' WHERE active_role IS NULL;

INSERT INTO user_roles (user_id, role_name)
SELECT id, 'user' FROM users
ON CONFLICT DO NOTHING;
