-- Revert TNC and PPM date columns back to nullable
ALTER TABLE machines ALTER COLUMN tnc_date DROP NOT NULL;
ALTER TABLE machines ALTER COLUMN ppm_date DROP NOT NULL;
