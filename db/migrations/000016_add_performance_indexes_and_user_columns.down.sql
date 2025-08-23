-- Rollback migration 016: Remove performance indexes and user table columns

BEGIN;

-- Remove the added columns from users table
ALTER TABLE users DROP COLUMN IF EXISTS status;
ALTER TABLE users DROP COLUMN IF EXISTS avatar;
ALTER TABLE users DROP COLUMN IF EXISTS last_login;
ALTER TABLE users DROP COLUMN IF EXISTS updated_at;

-- Drop the performance indexes
DROP INDEX IF EXISTS idx_maintenance_work_order_type;
DROP INDEX IF EXISTS idx_machines_ppm_date;

COMMIT;
