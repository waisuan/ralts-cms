-- Migration 014: Rename indexes to match target schema naming
-- Changes:
-- 1. Rename idx_machines_serial_number to machines_serialnumber_unique
-- 2. Rename idx_maintenance to maintenance_workordernumber_unique
-- 3. Drop ppm_date and work_order_type indexes (will be added in future migrations)
-- Note: Must drop and recreate foreign key to change index dependency

BEGIN;

-- Step 1: Drop the foreign key constraint that depends on the old index
ALTER TABLE maintenance DROP CONSTRAINT IF EXISTS fk_maintenance_serialnumber_serialnumber;

-- Step 2: Drop the old index now that nothing depends on it
DROP INDEX IF EXISTS idx_machines_serial_number;

-- Step 3: Create the new index with the target name
CREATE UNIQUE INDEX IF NOT EXISTS machines_serialnumber_unique ON machines ("serialNumber");

-- Step 4: Recreate the foreign key constraint (it will use the new index)
ALTER TABLE maintenance ADD CONSTRAINT fk_maintenance_serialnumber_serialnumber 
    FOREIGN KEY ("serialNumber") REFERENCES machines("serialNumber") ON DELETE CASCADE;

-- Rename maintenance composite index to match target naming
DROP INDEX IF EXISTS idx_maintenance;
CREATE UNIQUE INDEX IF NOT EXISTS maintenance_workordernumber_unique ON maintenance ("serialNumber", "workOrderNumber");

-- Drop ppm_date and work_order_type indexes (will be added in future migrations)
DROP INDEX IF EXISTS idx_machines_ppm_date;
DROP INDEX IF EXISTS idx_maintenance_work_order_type;

COMMIT;
