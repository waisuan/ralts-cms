-- Revert migration 020: Remove trigram search, restore full-text search
-- Note: This restores the search_vector infrastructure from migration 017.

-- =====================================================
-- MACHINES TABLE
-- =====================================================

DROP INDEX IF EXISTS idx_machines_trgm_search;
ALTER TABLE machines DROP COLUMN IF EXISTS search_text;

-- Restore search_vector column and trigger
ALTER TABLE machines ADD COLUMN IF NOT EXISTS search_vector tsvector;
CREATE INDEX IF NOT EXISTS idx_machines_search_vector ON machines USING GIN (search_vector);

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

CREATE TRIGGER machines_search_vector_update
    BEFORE INSERT OR UPDATE ON machines
    FOR EACH ROW EXECUTE FUNCTION machines_search_vector_update();

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
-- MAINTENANCE TABLE
-- =====================================================

DROP INDEX IF EXISTS idx_maintenance_trgm_search;
ALTER TABLE maintenance DROP COLUMN IF EXISTS search_text;

-- Restore search_vector column and trigger
ALTER TABLE maintenance ADD COLUMN IF NOT EXISTS search_vector tsvector;
CREATE INDEX IF NOT EXISTS idx_maintenance_search_vector ON maintenance USING GIN (search_vector);

CREATE OR REPLACE FUNCTION maintenance_search_vector_update() RETURNS trigger AS $$
BEGIN
    NEW.search_vector :=
        setweight(to_tsvector('english', COALESCE(NEW."workOrderNumber", '')), 'A') ||
        setweight(to_tsvector('english', COALESCE(NEW."reportedBy", '')), 'B') ||
        setweight(to_tsvector('english', COALESCE(NEW."workOrderType", '')), 'B');
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER maintenance_search_vector_update
    BEFORE INSERT OR UPDATE ON maintenance
    FOR EACH ROW EXECUTE FUNCTION maintenance_search_vector_update();

UPDATE maintenance SET search_vector =
    setweight(to_tsvector('english', COALESCE("workOrderNumber", '')), 'A') ||
    setweight(to_tsvector('english', COALESCE("reportedBy", '')), 'B') ||
    setweight(to_tsvector('english', COALESCE("workOrderType", '')), 'B');

DROP EXTENSION IF EXISTS pg_trgm;
