-- Migration 015: Align users table schema with target
-- Changes:
-- 1. Change id from SERIAL to BIGSERIAL  
-- 2. Rename name to username and adjust length
-- 3. Adjust email and password lengths
-- 4. Change salt from VARCHAR to TEXT
-- 5. Add approved boolean column
-- 6. Change role default from 'user' to 'NON_ADMIN'
-- 7. Remove status, avatar, last_login, updated_at columns
-- 8. Make created_at NOT NULL and change default to now()
-- 9. Update indexes to match target schema

BEGIN;

-- First, change the id column type from integer to bigint
ALTER TABLE users ALTER COLUMN id TYPE bigint;

-- Remove columns that don't exist in target schema
DO $$
BEGIN
    -- Remove status column if it exists
    IF EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'users' AND column_name = 'status'
    ) THEN
        ALTER TABLE users DROP COLUMN status;
    END IF;

    -- Remove avatar column if it exists
    IF EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'users' AND column_name = 'avatar'
    ) THEN
        ALTER TABLE users DROP COLUMN avatar;
    END IF;

    -- Remove last_login column if it exists
    IF EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'users' AND column_name = 'last_login'
    ) THEN
        ALTER TABLE users DROP COLUMN last_login;
    END IF;

    -- Remove updated_at column if it exists
    IF EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'users' AND column_name = 'updated_at'
    ) THEN
        ALTER TABLE users DROP COLUMN updated_at;
    END IF;
END $$;

-- Rename and adjust column specifications
DO $$
BEGIN
    -- Rename name to username if it exists
    IF EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'users' AND column_name = 'name'
    ) THEN
        ALTER TABLE users RENAME COLUMN name TO username;
    END IF;
END $$;

-- Adjust column lengths and types
ALTER TABLE users ALTER COLUMN username TYPE VARCHAR(50);
ALTER TABLE users ALTER COLUMN email TYPE VARCHAR(100);
ALTER TABLE users ALTER COLUMN password TYPE VARCHAR(200);
ALTER TABLE users ALTER COLUMN salt TYPE TEXT;

-- Add approved column if it doesn't exist
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'users' AND column_name = 'approved'
    ) THEN
        ALTER TABLE users ADD COLUMN approved boolean NOT NULL DEFAULT false;
    END IF;
END $$;

-- Update role default value (need to drop and recreate default)
ALTER TABLE users ALTER COLUMN role DROP DEFAULT;
ALTER TABLE users ALTER COLUMN role SET DEFAULT 'NON_ADMIN';

-- Update created_at constraints and default
ALTER TABLE users ALTER COLUMN created_at SET NOT NULL;
ALTER TABLE users ALTER COLUMN created_at SET DEFAULT now();

-- Update indexes to match target schema
-- Drop existing indexes
DROP INDEX IF EXISTS idx_users_role;
DROP INDEX IF EXISTS idx_users_status;
DROP INDEX IF EXISTS idx_users_email;

-- Create new indexes matching target schema
CREATE UNIQUE INDEX IF NOT EXISTS users_email_unique ON users (email);
CREATE UNIQUE INDEX IF NOT EXISTS users_username_unique ON users (username);

COMMIT;
