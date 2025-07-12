CREATE TABLE IF NOT EXISTS machines (
    id SERIAL PRIMARY KEY,
    serial_number VARCHAR(255) NOT NULL,
    customer VARCHAR(255),
    state VARCHAR(255),
    account_type VARCHAR(255),
    model VARCHAR(255),
    status VARCHAR(255),
    brand VARCHAR(255),
    district VARCHAR(255),
    person_in_charge VARCHAR(255),
    reported_by VARCHAR(255),
    additional_notes TEXT,
    attachment VARCHAR(255),
    ppm_status VARCHAR(255),
    tnc_date DATE,
    ppm_date DATE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_machines_serial_number ON machines (serial_number);