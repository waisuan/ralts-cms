-- Migration 020: Replace full-text search (tsvector) with trigram-based fuzzy search (pg_trgm)
--
-- pg_trgm provides word_similarity() for fuzzy matching — handles partial words,
-- typos, and proper nouns (e.g. Malaysian place names) that the English stemmer cannot.

-- Enable the pg_trgm extension
CREATE EXTENSION IF NOT EXISTS pg_trgm;

-- =====================================================
-- MACHINES TABLE
-- =====================================================

-- Add generated column that concatenates all searchable fields
ALTER TABLE machines ADD COLUMN search_text text GENERATED ALWAYS AS (
    COALESCE("serialNumber", '') || ' ' ||
    COALESCE(customer, '') || ' ' ||
    COALESCE(state, '') || ' ' ||
    COALESCE("accountType", '') || ' ' ||
    COALESCE(model, '') || ' ' ||
    COALESCE(status, '') || ' ' ||
    COALESCE(brand, '') || ' ' ||
    COALESCE(district, '') || ' ' ||
    COALESCE("personInCharge", '') || ' ' ||
    COALESCE("reportedBy", '')
) STORED;

-- GIN index for fast trigram lookups
CREATE INDEX idx_machines_trgm_search ON machines USING GIN (search_text gin_trgm_ops);

-- Drop old full-text search infrastructure
DROP TRIGGER IF EXISTS machines_search_vector_update ON machines;
DROP FUNCTION IF EXISTS machines_search_vector_update();
DROP INDEX IF EXISTS idx_machines_search_vector;
ALTER TABLE machines DROP COLUMN IF EXISTS search_vector;

-- =====================================================
-- MAINTENANCE TABLE
-- =====================================================

-- Add generated column that concatenates all searchable fields
ALTER TABLE maintenance ADD COLUMN search_text text GENERATED ALWAYS AS (
    COALESCE("workOrderNumber", '') || ' ' ||
    COALESCE("reportedBy", '') || ' ' ||
    COALESCE("workOrderType", '')
) STORED;

-- GIN index for fast trigram lookups
CREATE INDEX idx_maintenance_trgm_search ON maintenance USING GIN (search_text gin_trgm_ops);

-- Drop old full-text search infrastructure
DROP TRIGGER IF EXISTS maintenance_search_vector_update ON maintenance;
DROP FUNCTION IF EXISTS maintenance_search_vector_update();
DROP INDEX IF EXISTS idx_maintenance_search_vector;
ALTER TABLE maintenance DROP COLUMN IF EXISTS search_vector;
