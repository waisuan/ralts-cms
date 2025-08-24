-- Migration 012: Align machines table schema with target
-- Changes:
-- 1. Change id from SERIAL to BIGSERIAL
-- 2. Rename columns from snake_case to camelCase
-- 3. Adjust VARCHAR lengths to match target
-- 4. Remove ppm_status column
-- 5. Update constraints and defaults

BEGIN;

-- First, change the id column type from integer to bigint
ALTER TABLE machines ALTER COLUMN id TYPE bigint;

-- Rename columns from snake_case to camelCase
DO $$
BEGIN
    -- Rename serial_number to serialNumber
    IF EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'machines' AND column_name = 'serial_number'
    ) THEN
        ALTER TABLE machines RENAME COLUMN serial_number TO "serialNumber";
    END IF;

    -- Rename person_in_charge to personInCharge  
    IF EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'machines' AND column_name = 'person_in_charge'
    ) THEN
        ALTER TABLE machines RENAME COLUMN person_in_charge TO "personInCharge";
    END IF;

    -- Rename reported_by to reportedBy
    IF EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'machines' AND column_name = 'reported_by'
    ) THEN
        ALTER TABLE machines RENAME COLUMN reported_by TO "reportedBy";
    END IF;

    -- Rename additional_notes to additionalNotes
    IF EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'machines' AND column_name = 'additional_notes'
    ) THEN
        ALTER TABLE machines RENAME COLUMN additional_notes TO "additionalNotes";
    END IF;

    -- Rename account_type to accountType
    IF EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'machines' AND column_name = 'account_type'
    ) THEN
        ALTER TABLE machines RENAME COLUMN account_type TO "accountType";
    END IF;

    -- Rename tnc_date to tncDate
    IF EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'machines' AND column_name = 'tnc_date'
    ) THEN
        ALTER TABLE machines RENAME COLUMN tnc_date TO "tncDate";
    END IF;

    -- Rename ppm_date to ppmDate
    IF EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'machines' AND column_name = 'ppm_date'
    ) THEN
        ALTER TABLE machines RENAME COLUMN ppm_date TO "ppmDate";
    END IF;

    -- Rename created_at to createdAt
    IF EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'machines' AND column_name = 'created_at'
    ) THEN
        ALTER TABLE machines RENAME COLUMN created_at TO "createdAt";
    END IF;

    -- Rename updated_at to updatedAt
    IF EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'machines' AND column_name = 'updated_at'
    ) THEN
        ALTER TABLE machines RENAME COLUMN updated_at TO "updatedAt";
    END IF;
END $$;

-- Adjust column lengths to match target schema
ALTER TABLE machines ALTER COLUMN "serialNumber" TYPE VARCHAR(100);
ALTER TABLE machines ALTER COLUMN customer TYPE VARCHAR(200);
ALTER TABLE machines ALTER COLUMN state TYPE VARCHAR(100);
ALTER TABLE machines ALTER COLUMN "accountType" TYPE VARCHAR(100);
ALTER TABLE machines ALTER COLUMN model TYPE VARCHAR(100);
ALTER TABLE machines ALTER COLUMN status TYPE VARCHAR(100);
ALTER TABLE machines ALTER COLUMN brand TYPE VARCHAR(100);
ALTER TABLE machines ALTER COLUMN district TYPE VARCHAR(100);
ALTER TABLE machines ALTER COLUMN "personInCharge" TYPE VARCHAR(200);
ALTER TABLE machines ALTER COLUMN "reportedBy" TYPE VARCHAR(200);
ALTER TABLE machines ALTER COLUMN attachment TYPE VARCHAR(200);

-- Remove ppm_status column if it exists
DO $$
BEGIN
    IF EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'machines' AND column_name = 'ppm_status'
    ) THEN
        ALTER TABLE machines DROP COLUMN ppm_status;
    END IF;
END $$;

ALTER TABLE machines ALTER COLUMN "ppmDate" DROP NOT NULL;
ALTER TABLE machines ALTER COLUMN "tncDate" DROP NOT NULL;

-- Update constraints - make timestamps NOT NULL to match target
ALTER TABLE machines ALTER COLUMN "createdAt" SET NOT NULL;
ALTER TABLE machines ALTER COLUMN "updatedAt" SET NOT NULL;

-- Drop search vector functionality (will be added in future migration)
DROP TRIGGER IF EXISTS machines_search_vector_update ON machines;
DROP FUNCTION IF EXISTS machines_search_vector_update();
DROP INDEX IF EXISTS idx_machines_search_vector;
ALTER TABLE machines DROP COLUMN IF EXISTS search_vector;

COMMIT;
