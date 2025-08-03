-- Drop the trigger
DROP TRIGGER IF EXISTS machines_search_vector_update ON machines;

-- Drop the function
DROP FUNCTION IF EXISTS machines_search_vector_update();

-- Drop the index
DROP INDEX IF EXISTS idx_machines_search_vector;

-- Drop the search vector column
ALTER TABLE machines DROP COLUMN IF EXISTS search_vector;