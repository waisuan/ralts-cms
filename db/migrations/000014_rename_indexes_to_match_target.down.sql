-- Rollback migration 014: Revert index names back to original

BEGIN;

-- Step 1: Drop the foreign key constraint
ALTER TABLE maintenance DROP CONSTRAINT IF EXISTS fk_maintenance_serialnumber_serialnumber;

-- Step 2: Drop the new index
DROP INDEX IF EXISTS machines_serialnumber_unique;

-- Step 3: Create the old index with the original name
CREATE UNIQUE INDEX IF NOT EXISTS idx_machines_serial_number ON machines (serial_number);

-- Step 4: Recreate the foreign key constraint with old column names
ALTER TABLE maintenance ADD CONSTRAINT fk_maintenance_serialnumber_serialnumber 
    FOREIGN KEY (machine_serial_number) REFERENCES machines(serial_number) ON DELETE CASCADE;

-- Restore the indexes that were dropped
CREATE INDEX IF NOT EXISTS idx_maintenance_work_order_type ON maintenance (work_order_type);
CREATE INDEX IF NOT EXISTS idx_machines_ppm_date ON machines (ppm_date);

-- Revert maintenance composite index name and column references
DROP INDEX IF EXISTS maintenance_workordernumber_unique;
CREATE UNIQUE INDEX IF NOT EXISTS idx_maintenance ON maintenance (machine_serial_number, work_order_number);

COMMIT;
