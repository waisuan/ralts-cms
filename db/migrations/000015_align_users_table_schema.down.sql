-- Rollback migration 015: Revert users table schema changes

BEGIN;

-- Drop the new indexes
DROP INDEX IF EXISTS users_email_unique;
DROP INDEX IF EXISTS users_username_unique;

-- Recreate the original indexes
CREATE UNIQUE INDEX IF NOT EXISTS idx_users_email ON users (email);
CREATE INDEX IF NOT EXISTS idx_users_role ON users (role);
CREATE INDEX IF NOT EXISTS idx_users_status ON users (status);

-- Revert created_at constraints and default
ALTER TABLE users ALTER COLUMN created_at DROP NOT NULL;
ALTER TABLE users ALTER COLUMN created_at SET DEFAULT CURRENT_TIMESTAMP;

-- Revert role default value
ALTER TABLE users ALTER COLUMN role DROP DEFAULT;
ALTER TABLE users ALTER COLUMN role SET DEFAULT 'user';

-- Remove approved column
ALTER TABLE users DROP COLUMN IF EXISTS approved;

-- Revert column lengths and types
ALTER TABLE users ALTER COLUMN username TYPE VARCHAR(255);
ALTER TABLE users ALTER COLUMN email TYPE VARCHAR(255);
ALTER TABLE users ALTER COLUMN password TYPE VARCHAR(255);
ALTER TABLE users ALTER COLUMN salt TYPE VARCHAR(255);

-- Rename username back to name
ALTER TABLE users RENAME COLUMN username TO name;

-- Add back the removed columns with their original specifications
ALTER TABLE users ADD COLUMN status VARCHAR(50) NOT NULL DEFAULT 'active';
ALTER TABLE users ADD COLUMN avatar VARCHAR(500);
ALTER TABLE users ADD COLUMN last_login TIMESTAMP;
ALTER TABLE users ADD COLUMN updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP;

-- Revert id column type back to integer
ALTER TABLE users ALTER COLUMN id TYPE integer;

COMMIT;
