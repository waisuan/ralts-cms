-- Rollback migration 017: Remove search vectors from machines and maintenance tables

-- =====================================================
-- REMOVE MAINTENANCE TABLE SEARCH VECTOR
-- =====================================================

-- Drop the maintenance trigger
DROP TRIGGER IF EXISTS maintenance_search_vector_update ON maintenance;

-- Drop the maintenance function
DROP FUNCTION IF EXISTS maintenance_search_vector_update();

-- Drop the maintenance search vector index
DROP INDEX IF EXISTS idx_maintenance_search_vector;

-- Drop the maintenance search vector column
ALTER TABLE maintenance DROP COLUMN IF EXISTS search_vector;

-- =====================================================
-- REMOVE MACHINES TABLE SEARCH VECTOR
-- =====================================================

-- Drop the machines trigger
DROP TRIGGER IF EXISTS machines_search_vector_update ON machines;

-- Drop the machines function
DROP FUNCTION IF EXISTS machines_search_vector_update();

-- Drop the machines search vector index
DROP INDEX IF EXISTS idx_machines_search_vector;

-- Drop the machines search vector column
ALTER TABLE machines DROP COLUMN IF EXISTS search_vector;

