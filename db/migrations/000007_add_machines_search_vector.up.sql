-- Add full-text search column
ALTER TABLE machines ADD COLUMN IF NOT EXISTS search_vector tsvector;

-- Create a GIN index for fast full-text search
CREATE INDEX IF NOT EXISTS idx_machines_search_vector ON machines USING GIN (search_vector);

-- Create a function to update the search vector
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

-- Create trigger to automatically update search vector
DROP TRIGGER IF EXISTS machines_search_vector_update ON machines;
CREATE TRIGGER machines_search_vector_update
    BEFORE INSERT OR UPDATE ON machines
    FOR EACH ROW EXECUTE FUNCTION machines_search_vector_update();

-- Update existing records with search vectors
UPDATE machines SET search_vector = 
    setweight(to_tsvector('english', COALESCE(serial_number, '')), 'A') ||
    setweight(to_tsvector('english', COALESCE(customer, '')), 'B') ||
    setweight(to_tsvector('english', COALESCE(state, '')), 'C') ||
    setweight(to_tsvector('english', COALESCE(account_type, '')), 'C') ||
    setweight(to_tsvector('english', COALESCE(model, '')), 'B') ||
    setweight(to_tsvector('english', COALESCE(status, '')), 'C') ||
    setweight(to_tsvector('english', COALESCE(brand, '')), 'B') ||
    setweight(to_tsvector('english', COALESCE(district, '')), 'C') ||
    setweight(to_tsvector('english', COALESCE(person_in_charge, '')), 'C') ||
    setweight(to_tsvector('english', COALESCE(reported_by, '')), 'C');