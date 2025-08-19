-- Fix maintenance table schema to match production data
-- 1. Rename worker_order_type to work_order_type 
-- 2. Make attachment field nullable

ALTER TABLE maintenance RENAME COLUMN worker_order_type TO work_order_type;
ALTER TABLE maintenance ALTER COLUMN attachment DROP NOT NULL;
