-- Add full-text search column
ALTER TABLE maintenance ADD COLUMN IF NOT EXISTS search_vector tsvector;

-- Create a GIN index for fast full-text search
CREATE INDEX IF NOT EXISTS idx_maintenance_search_vector ON maintenance USING GIN (search_vector);

-- Create a function to update the search vector
CREATE OR REPLACE FUNCTION maintenance_search_vector_update() RETURNS trigger AS $$
BEGIN
    NEW.search_vector :=
        setweight(to_tsvector('english', COALESCE(NEW.work_order_number, '')), 'A') ||
        setweight(to_tsvector('english', COALESCE(NEW.reported_by, '')), 'B') ||
        setweight(to_tsvector('english', COALESCE(NEW.worker_order_type, '')), 'B');
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Create trigger to automatically update search vector
DROP TRIGGER IF EXISTS maintenance_search_vector_update ON maintenance;
CREATE TRIGGER maintenance_search_vector_update
    BEFORE INSERT OR UPDATE ON maintenance
    FOR EACH ROW EXECUTE FUNCTION maintenance_search_vector_update();

-- Update existing records with search vectors
UPDATE maintenance SET search_vector = NULL; 