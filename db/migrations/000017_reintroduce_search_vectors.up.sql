-- Migration 017: Re-introduce search vectors for machines and maintenance tables
-- This migration adds back full-text search functionality using the new camelCase column names

-- =====================================================
-- MACHINES TABLE SEARCH VECTOR
-- =====================================================

-- Add full-text search column to machines table
ALTER TABLE machines ADD COLUMN IF NOT EXISTS search_vector tsvector;

-- Create a GIN index for fast full-text search on machines
CREATE INDEX IF NOT EXISTS idx_machines_search_vector ON machines USING GIN (search_vector);

-- Create a function to update the machines search vector
CREATE OR REPLACE FUNCTION machines_search_vector_update() RETURNS trigger AS $$
BEGIN
    NEW.search_vector :=
        setweight(to_tsvector('english', COALESCE(NEW."serialNumber", '')), 'A') ||
        setweight(to_tsvector('english', COALESCE(NEW.customer, '')), 'B') ||
        setweight(to_tsvector('english', COALESCE(NEW.state, '')), 'C') ||
        setweight(to_tsvector('english', COALESCE(NEW."accountType", '')), 'C') ||
        setweight(to_tsvector('english', COALESCE(NEW.model, '')), 'B') ||
        setweight(to_tsvector('english', COALESCE(NEW.status, '')), 'C') ||
        setweight(to_tsvector('english', COALESCE(NEW.brand, '')), 'B') ||
        setweight(to_tsvector('english', COALESCE(NEW.district, '')), 'C') ||
        setweight(to_tsvector('english', COALESCE(NEW."personInCharge", '')), 'C') ||
        setweight(to_tsvector('english', COALESCE(NEW."reportedBy", '')), 'C');
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Create trigger to automatically update machines search vector
DROP TRIGGER IF EXISTS machines_search_vector_update ON machines;
CREATE TRIGGER machines_search_vector_update
    BEFORE INSERT OR UPDATE ON machines
    FOR EACH ROW EXECUTE FUNCTION machines_search_vector_update();

-- Update existing machines records with search vectors
UPDATE machines SET search_vector = 
    setweight(to_tsvector('english', COALESCE("serialNumber", '')), 'A') ||
    setweight(to_tsvector('english', COALESCE(customer, '')), 'B') ||
    setweight(to_tsvector('english', COALESCE(state, '')), 'C') ||
    setweight(to_tsvector('english', COALESCE("accountType", '')), 'C') ||
    setweight(to_tsvector('english', COALESCE(model, '')), 'B') ||
    setweight(to_tsvector('english', COALESCE(status, '')), 'C') ||
    setweight(to_tsvector('english', COALESCE(brand, '')), 'B') ||
    setweight(to_tsvector('english', COALESCE(district, '')), 'C') ||
    setweight(to_tsvector('english', COALESCE("personInCharge", '')), 'C') ||
    setweight(to_tsvector('english', COALESCE("reportedBy", '')), 'C');

-- =====================================================
-- MAINTENANCE TABLE SEARCH VECTOR
-- =====================================================

-- Add full-text search column to maintenance table
ALTER TABLE maintenance ADD COLUMN IF NOT EXISTS search_vector tsvector;

-- Create a GIN index for fast full-text search on maintenance
CREATE INDEX IF NOT EXISTS idx_maintenance_search_vector ON maintenance USING GIN (search_vector);

-- Create a function to update the maintenance search vector
CREATE OR REPLACE FUNCTION maintenance_search_vector_update() RETURNS trigger AS $$
BEGIN
    NEW.search_vector :=
        setweight(to_tsvector('english', COALESCE(NEW."workOrderNumber", '')), 'A') ||
        setweight(to_tsvector('english', COALESCE(NEW."reportedBy", '')), 'B') ||
        setweight(to_tsvector('english', COALESCE(NEW."workOrderType", '')), 'B');
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Create trigger to automatically update maintenance search vector
DROP TRIGGER IF EXISTS maintenance_search_vector_update ON maintenance;
CREATE TRIGGER maintenance_search_vector_update
    BEFORE INSERT OR UPDATE ON maintenance
    FOR EACH ROW EXECUTE FUNCTION maintenance_search_vector_update();

-- Update existing maintenance records with search vectors
UPDATE maintenance SET search_vector = 
    setweight(to_tsvector('english', COALESCE("workOrderNumber", '')), 'A') ||
    setweight(to_tsvector('english', COALESCE("reportedBy", '')), 'B') ||
    setweight(to_tsvector('english', COALESCE("workOrderType", '')), 'B');
