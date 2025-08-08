-- Drop trigger
DROP TRIGGER IF EXISTS maintenance_search_vector_update ON maintenance;

-- Drop function
DROP FUNCTION IF EXISTS maintenance_search_vector_update();

-- Drop index
DROP INDEX IF EXISTS idx_maintenance_search_vector;

-- Drop column
ALTER TABLE maintenance DROP COLUMN IF EXISTS search_vector; 