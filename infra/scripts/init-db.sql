-- scripts/init-db.sql
-- Executed automatically when the PostgreSQL container is first created.

-- Extensions
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";  -- UUID generation (pgx uses its own, but useful for raw SQL)
CREATE EXTENSION IF NOT EXISTS "pg_trgm";    -- Trigram indexes for ILIKE search performance
CREATE EXTENSION IF NOT EXISTS "citext";     -- Case-insensitive text type

-- Set timezone
SET timezone = 'UTC';

\echo 'Database initialized successfully'
