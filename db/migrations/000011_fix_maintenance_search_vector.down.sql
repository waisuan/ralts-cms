-- Revert maintenance search vector fix
-- Drop the updated trigger and function
DROP TRIGGER IF EXISTS maintenance_search_vector_update_trigger ON maintenance;
DROP FUNCTION IF EXISTS maintenance_search_vector_update();

-- Recreate the original function that uses worker_order_type
-- NOTE: This will fail if worker_order_type column doesn't exist after migration 009
CREATE OR REPLACE FUNCTION maintenance_search_vector_update() RETURNS trigger AS $$
BEGIN
    NEW.search_vector :=
        setweight(to_tsvector('english', COALESCE(NEW.work_order_number, '')), 'A') ||
        setweight(to_tsvector('english', COALESCE(NEW.reported_by, '')), 'B') ||
        setweight(to_tsvector('english', COALESCE(NEW.worker_order_type, '')), 'B');
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Recreate original trigger
CREATE TRIGGER maintenance_search_vector_update_trigger
    BEFORE INSERT OR UPDATE ON maintenance
    FOR EACH ROW EXECUTE FUNCTION maintenance_search_vector_update();
