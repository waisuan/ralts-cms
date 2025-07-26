-- Drop index for PPM date field in machines table
DROP INDEX IF EXISTS idx_machines_ppm_date;

-- Drop index for work order type field in maintenance table
DROP INDEX IF EXISTS idx_maintenance_worker_order_type;
