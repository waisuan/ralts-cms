-- Revert maintenance table schema changes
-- 1. Make attachment field required again
-- 2. Rename work_order_type back to worker_order_type

ALTER TABLE maintenance ALTER COLUMN attachment SET NOT NULL;
ALTER TABLE maintenance RENAME COLUMN work_order_type TO worker_order_type;
