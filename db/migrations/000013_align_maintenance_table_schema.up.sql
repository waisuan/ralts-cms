-- Migration 013: Align maintenance table schema with target
-- Changes:
-- 1. Change id from SERIAL to BIGSERIAL
-- 2. Rename machine_serial_number to serialNumber
-- 3. Rename work_order_number to workOrderNumber
-- 4. Rename work_order_date to workOrderDate
-- 5. Rename action_taken to actionTaken
-- 6. Rename reported_by to reportedBy
-- 7. Rename work_order_type to workOrderType
-- 8. Rename created_at to createdAt
-- 9. Rename updated_at to updatedAt
-- 10. Adjust VARCHAR lengths and nullable constraints
-- 11. Update foreign key constraint

BEGIN;

-- First, change the id column type from integer to bigint
ALTER TABLE maintenance ALTER COLUMN id TYPE bigint;

-- Rename columns to match target schema
DO $$
BEGIN
    -- Rename machine_serial_number to serialNumber
    IF EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'maintenance' AND column_name = 'machine_serial_number'
    ) THEN
        ALTER TABLE maintenance RENAME COLUMN machine_serial_number TO "serialNumber";
    END IF;

    -- Rename work_order_number to workOrderNumber
    IF EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'maintenance' AND column_name = 'work_order_number'
    ) THEN
        ALTER TABLE maintenance RENAME COLUMN work_order_number TO "workOrderNumber";
    END IF;

    -- Rename work_order_date to workOrderDate
    IF EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'maintenance' AND column_name = 'work_order_date'
    ) THEN
        ALTER TABLE maintenance RENAME COLUMN work_order_date TO "workOrderDate";
    END IF;

    -- Rename action_taken to actionTaken
    IF EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'maintenance' AND column_name = 'action_taken'
    ) THEN
        ALTER TABLE maintenance RENAME COLUMN action_taken TO "actionTaken";
    END IF;

    -- Rename reported_by to reportedBy
    IF EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'maintenance' AND column_name = 'reported_by'
    ) THEN
        ALTER TABLE maintenance RENAME COLUMN reported_by TO "reportedBy";
    END IF;

    -- Rename work_order_type to workOrderType  
    IF EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'maintenance' AND column_name = 'work_order_type'
    ) THEN
        ALTER TABLE maintenance RENAME COLUMN work_order_type TO "workOrderType";
    END IF;

    -- Rename created_at to createdAt
    IF EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'maintenance' AND column_name = 'created_at'
    ) THEN
        ALTER TABLE maintenance RENAME COLUMN created_at TO "createdAt";
    END IF;

    -- Rename updated_at to updatedAt
    IF EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'maintenance' AND column_name = 'updated_at'
    ) THEN
        ALTER TABLE maintenance RENAME COLUMN updated_at TO "updatedAt";
    END IF;
END $$;

-- Adjust column lengths to match target schema
ALTER TABLE maintenance ALTER COLUMN "serialNumber" TYPE VARCHAR(100);
ALTER TABLE maintenance ALTER COLUMN "workOrderNumber" TYPE VARCHAR(100);
ALTER TABLE maintenance ALTER COLUMN "reportedBy" TYPE VARCHAR(200);
ALTER TABLE maintenance ALTER COLUMN "workOrderType" TYPE VARCHAR(100);
ALTER TABLE maintenance ALTER COLUMN attachment TYPE VARCHAR(200);

-- Update nullable constraints to match target schema
-- Make these fields nullable (target schema shows them as nullable)
ALTER TABLE maintenance ALTER COLUMN "workOrderDate" DROP NOT NULL;
ALTER TABLE maintenance ALTER COLUMN "actionTaken" DROP NOT NULL;
ALTER TABLE maintenance ALTER COLUMN "reportedBy" DROP NOT NULL;
ALTER TABLE maintenance ALTER COLUMN "workOrderType" DROP NOT NULL;

-- Ensure timestamps are NOT NULL to match target
ALTER TABLE maintenance ALTER COLUMN "createdAt" SET NOT NULL;
ALTER TABLE maintenance ALTER COLUMN "updatedAt" SET NOT NULL;

-- Drop existing foreign key constraint
ALTER TABLE maintenance DROP CONSTRAINT IF EXISTS fk_maintenance_serialnumber_serialnumber;

-- Recreate foreign key constraint with new column names
ALTER TABLE maintenance ADD CONSTRAINT fk_maintenance_serialnumber_serialnumber 
    FOREIGN KEY ("serialNumber") REFERENCES machines("serialNumber") ON DELETE CASCADE;

-- Drop search vector functionality (will be added in future migration)
DROP TRIGGER IF EXISTS maintenance_search_vector_update_trigger ON maintenance;
DROP FUNCTION IF EXISTS maintenance_search_vector_update() CASCADE;
DROP INDEX IF EXISTS idx_maintenance_search_vector;
ALTER TABLE maintenance DROP COLUMN IF EXISTS search_vector;

COMMIT;
