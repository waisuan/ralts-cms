-- Fix maintenance search vector to use work_order_type field
-- Drop the existing trigger first, then function with CASCADE to handle dependencies
DROP TRIGGER IF EXISTS maintenance_search_vector_update_trigger ON maintenance;
DROP FUNCTION IF EXISTS maintenance_search_vector_update() CASCADE;

-- Create updated function that uses work_order_type instead of worker_order_type
CREATE OR REPLACE FUNCTION maintenance_search_vector_update() RETURNS trigger AS $$
BEGIN
    NEW.search_vector :=
        setweight(to_tsvector('english', COALESCE(NEW.work_order_number, '')), 'A') ||
        setweight(to_tsvector('english', COALESCE(NEW.reported_by, '')), 'B') ||
        setweight(to_tsvector('english', COALESCE(NEW.work_order_type, '')), 'B');
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Recreate trigger
CREATE TRIGGER maintenance_search_vector_update_trigger
    BEFORE INSERT OR UPDATE ON maintenance
    FOR EACH ROW EXECUTE FUNCTION maintenance_search_vector_update();

-- Update existing records with correct search vector
UPDATE maintenance SET search_vector = 
    setweight(to_tsvector('english', COALESCE(work_order_number, '')), 'A') ||
    setweight(to_tsvector('english', COALESCE(reported_by, '')), 'B') ||
    setweight(to_tsvector('english', COALESCE(work_order_type, '')), 'B');
