-- Add index for PPM date field in machines table
CREATE INDEX IF NOT EXISTS idx_machines_ppm_date ON machines (ppm_date);

-- Add index for work order type field in maintenance table
CREATE INDEX IF NOT EXISTS idx_maintenance_worker_order_type ON maintenance (worker_order_type);
