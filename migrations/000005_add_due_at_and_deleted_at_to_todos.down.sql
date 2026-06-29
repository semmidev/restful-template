DROP INDEX IF EXISTS idx_todos_user_deleted_at;

ALTER TABLE todos DROP COLUMN IF EXISTS due_at;
ALTER TABLE todos DROP COLUMN IF EXISTS deleted_at;
