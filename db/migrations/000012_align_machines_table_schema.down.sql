-- Rollback migration 012: Revert machines table schema changes

BEGIN;

-- Restore search vector functionality (recreate what existed before)
ALTER TABLE machines ADD COLUMN IF NOT EXISTS search_vector tsvector;
CREATE INDEX IF NOT EXISTS idx_machines_search_vector ON machines USING GIN (search_vector);

CREATE OR REPLACE FUNCTION machines_search_vector_update() RETURNS trigger AS $$
BEGIN
    NEW.search_vector :=
        setweight(to_tsvector('english', COALESCE(NEW.serial_number, '')), 'A') ||
        setweight(to_tsvector('english', COALESCE(NEW.customer, '')), 'B') ||
        setweight(to_tsvector('english', COALESCE(NEW.state, '')), 'C') ||
        setweight(to_tsvector('english', COALESCE(NEW.account_type, '')), 'C') ||
        setweight(to_tsvector('english', COALESCE(NEW.model, '')), 'B') ||
        setweight(to_tsvector('english', COALESCE(NEW.status, '')), 'C') ||
        setweight(to_tsvector('english', COALESCE(NEW.brand, '')), 'B') ||
        setweight(to_tsvector('english', COALESCE(NEW.district, '')), 'C') ||
        setweight(to_tsvector('english', COALESCE(NEW.person_in_charge, '')), 'C') ||
        setweight(to_tsvector('english', COALESCE(NEW.reported_by, '')), 'C');
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER machines_search_vector_update
    BEFORE INSERT OR UPDATE ON machines
    FOR EACH ROW EXECUTE FUNCTION machines_search_vector_update();

-- Add back ppm_status column
ALTER TABLE machines ADD COLUMN ppm_status VARCHAR(255);

-- Revert constraints
ALTER TABLE machines ALTER COLUMN "createdAt" DROP NOT NULL;
ALTER TABLE machines ALTER COLUMN "updatedAt" DROP NOT NULL;

-- Revert column lengths
ALTER TABLE machines ALTER COLUMN "serialNumber" TYPE VARCHAR(255);
ALTER TABLE machines ALTER COLUMN customer TYPE VARCHAR(255);
ALTER TABLE machines ALTER COLUMN state TYPE VARCHAR(255);
ALTER TABLE machines ALTER COLUMN "accountType" TYPE VARCHAR(255);
ALTER TABLE machines ALTER COLUMN model TYPE VARCHAR(255);
ALTER TABLE machines ALTER COLUMN status TYPE VARCHAR(255);
ALTER TABLE machines ALTER COLUMN brand TYPE VARCHAR(255);
ALTER TABLE machines ALTER COLUMN district TYPE VARCHAR(255);
ALTER TABLE machines ALTER COLUMN "personInCharge" TYPE VARCHAR(255);
ALTER TABLE machines ALTER COLUMN "reportedBy" TYPE VARCHAR(255);
ALTER TABLE machines ALTER COLUMN attachment TYPE VARCHAR(255);

-- Rename columns back to snake_case
ALTER TABLE machines RENAME COLUMN "serialNumber" TO serial_number;
ALTER TABLE machines RENAME COLUMN "personInCharge" TO person_in_charge;
ALTER TABLE machines RENAME COLUMN "reportedBy" TO reported_by;
ALTER TABLE machines RENAME COLUMN "additionalNotes" TO additional_notes;
ALTER TABLE machines RENAME COLUMN "accountType" TO account_type;
ALTER TABLE machines RENAME COLUMN "tncDate" TO tnc_date;
ALTER TABLE machines RENAME COLUMN "ppmDate" TO ppm_date;
ALTER TABLE machines RENAME COLUMN "createdAt" TO created_at;
ALTER TABLE machines RENAME COLUMN "updatedAt" TO updated_at;

-- Revert id column type back to integer
ALTER TABLE machines ALTER COLUMN id TYPE integer;

COMMIT;
