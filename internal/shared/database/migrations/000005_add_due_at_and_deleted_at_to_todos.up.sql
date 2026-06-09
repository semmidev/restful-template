ALTER TABLE todos ADD COLUMN IF NOT EXISTS due_at TIMESTAMPTZ;
ALTER TABLE todos ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMPTZ;

CREATE INDEX IF NOT EXISTS idx_todos_user_deleted_at ON todos (user_id, deleted_at);
