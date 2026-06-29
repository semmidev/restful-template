DO $$ BEGIN
    CREATE TYPE todo_status AS ENUM ('pending', 'in_progress', 'done');
EXCEPTION
    WHEN duplicate_object THEN null;
END $$;

CREATE TABLE IF NOT EXISTS todos (
    id          UUID        NOT NULL DEFAULT uuidv7() PRIMARY KEY,
    user_id     UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    title       TEXT        NOT NULL,
    description TEXT,
    cover       TEXT,
    status      todo_status NOT NULL DEFAULT 'pending',
    importance  BOOLEAN     NOT NULL DEFAULT false,
    urgency     BOOLEAN     NOT NULL DEFAULT false,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_todos_user_id        ON todos (user_id);
CREATE INDEX IF NOT EXISTS idx_todos_user_status    ON todos (user_id, status);
CREATE INDEX IF NOT EXISTS idx_todos_user_created_at ON todos (user_id, created_at);
