CREATE TABLE IF NOT EXISTS maintenance (
    id SERIAL PRIMARY KEY,
    machine_serial_number VARCHAR(255) NOT NULL,
    work_order_number VARCHAR(255) NOT NULL,
    work_order_date DATE NOT NULL,
    action_taken TEXT NOT NULL,
    reported_by VARCHAR(255) NOT NULL,
    worker_order_type VARCHAR(255) NOT NULL,
    attachment VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_maintenance ON maintenance (machine_serial_number, work_order_number);