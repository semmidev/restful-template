DELETE FROM users WHERE password_hash IS NULL;
ALTER TABLE users DROP COLUMN google_id;
ALTER TABLE users ALTER COLUMN password_hash SET NOT NULL;
