-- Migration 016: Add back performance indexes and user table columns
-- Changes:
-- 1. Add performance indexes back to machines and maintenance tables
-- 2. Add status, avatar, last_login, updated_at columns to users table

BEGIN;

-- Add performance indexes back to machines table
CREATE INDEX IF NOT EXISTS idx_machines_ppm_date ON machines ("ppmDate");

-- Add performance indexes back to maintenance table  
CREATE INDEX IF NOT EXISTS idx_maintenance_work_order_type ON maintenance ("workOrderType");

-- Add missing columns to users table
DO $$
BEGIN
    -- Add status column if it doesn't exist
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'users' AND column_name = 'status'
    ) THEN
        ALTER TABLE users ADD COLUMN status VARCHAR(50);
    END IF;

    -- Add avatar column if it doesn't exist
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'users' AND column_name = 'avatar'
    ) THEN
        ALTER TABLE users ADD COLUMN avatar VARCHAR(500);
    END IF;

    -- Add last_login column if it doesn't exist
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'users' AND column_name = 'last_login'
    ) THEN
        ALTER TABLE users ADD COLUMN last_login TIMESTAMP;
    END IF;

    -- Add updated_at column if it doesn't exist
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'users' AND column_name = 'updated_at'
    ) THEN
        ALTER TABLE users ADD COLUMN updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP;
    END IF;
END $$;

COMMIT;
