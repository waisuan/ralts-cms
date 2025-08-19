-- Fix work_order_type index after field rename
-- Drop the old index that references worker_order_type
DROP INDEX IF EXISTS idx_maintenance_worker_order_type;

-- Create new index for work_order_type field
CREATE INDEX IF NOT EXISTS idx_maintenance_work_order_type ON maintenance (work_order_type);
