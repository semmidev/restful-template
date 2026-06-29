DROP INDEX IF EXISTS idx_todos_user_created_at;
DROP INDEX IF EXISTS idx_todos_user_status;
DROP INDEX IF EXISTS idx_todos_user_id;
DROP TABLE IF EXISTS todos CASCADE;
DROP TYPE IF EXISTS todo_status CASCADE;
