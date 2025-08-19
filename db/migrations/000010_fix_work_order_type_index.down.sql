-- Revert work_order_type index fix
-- Drop the new index
DROP INDEX IF EXISTS idx_maintenance_work_order_type;

-- Recreate the old index (this will fail if worker_order_type column doesn't exist)
-- CREATE INDEX IF NOT EXISTS idx_maintenance_worker_order_type ON maintenance (worker_order_type);
