-- Rollback migration 013: Revert maintenance table schema changes

BEGIN;

-- Drop the new foreign key constraint
ALTER TABLE maintenance DROP CONSTRAINT IF EXISTS fk_maintenance_serialnumber_serialnumber;

-- Recreate the old foreign key constraint
ALTER TABLE maintenance ADD CONSTRAINT fk_maintenance_serialnumber_serialnumber 
    FOREIGN KEY (machine_serial_number) REFERENCES machines(serial_number) ON DELETE CASCADE;

-- Revert nullable constraints
ALTER TABLE maintenance ALTER COLUMN "workOrderDate" SET NOT NULL;
ALTER TABLE maintenance ALTER COLUMN "actionTaken" SET NOT NULL;
ALTER TABLE maintenance ALTER COLUMN "reportedBy" SET NOT NULL;
ALTER TABLE maintenance ALTER COLUMN "workOrderType" SET NOT NULL;

-- Revert constraints
ALTER TABLE maintenance ALTER COLUMN "createdAt" DROP NOT NULL;
ALTER TABLE maintenance ALTER COLUMN "updatedAt" DROP NOT NULL;

-- Revert column lengths
ALTER TABLE maintenance ALTER COLUMN "serialNumber" TYPE VARCHAR(255);
ALTER TABLE maintenance ALTER COLUMN "workOrderNumber" TYPE VARCHAR(255);
ALTER TABLE maintenance ALTER COLUMN "reportedBy" TYPE VARCHAR(255);
ALTER TABLE maintenance ALTER COLUMN "workOrderType" TYPE VARCHAR(255);
ALTER TABLE maintenance ALTER COLUMN attachment TYPE VARCHAR(255);

-- Rename columns back to snake_case
ALTER TABLE maintenance RENAME COLUMN "serialNumber" TO machine_serial_number;
ALTER TABLE maintenance RENAME COLUMN "workOrderNumber" TO work_order_number;
ALTER TABLE maintenance RENAME COLUMN "workOrderDate" TO work_order_date;
ALTER TABLE maintenance RENAME COLUMN "actionTaken" TO action_taken;
ALTER TABLE maintenance RENAME COLUMN "reportedBy" TO reported_by;
ALTER TABLE maintenance RENAME COLUMN "workOrderType" TO work_order_type;
ALTER TABLE maintenance RENAME COLUMN "createdAt" TO created_at;
ALTER TABLE maintenance RENAME COLUMN "updatedAt" TO updated_at;

-- Revert id column type back to integer
ALTER TABLE maintenance ALTER COLUMN id TYPE integer;

-- Restore search vector functionality (recreate what existed before)
ALTER TABLE maintenance ADD COLUMN IF NOT EXISTS search_vector tsvector;
CREATE INDEX IF NOT EXISTS idx_maintenance_search_vector ON maintenance USING GIN (search_vector);

CREATE OR REPLACE FUNCTION maintenance_search_vector_update() RETURNS trigger AS $$
BEGIN
    NEW.search_vector :=
        setweight(to_tsvector('english', COALESCE(NEW.work_order_number, '')), 'A') ||
        setweight(to_tsvector('english', COALESCE(NEW.reported_by, '')), 'B') ||
        setweight(to_tsvector('english', COALESCE(NEW.work_order_type, '')), 'B');
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER maintenance_search_vector_update_trigger
    BEFORE INSERT OR UPDATE ON maintenance
    FOR EACH ROW EXECUTE FUNCTION maintenance_search_vector_update();

COMMIT;
