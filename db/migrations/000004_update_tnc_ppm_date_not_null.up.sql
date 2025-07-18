-- Make TNC and PPM date columns not nullable
ALTER TABLE machines ALTER COLUMN tnc_date SET NOT NULL;
ALTER TABLE machines ALTER COLUMN ppm_date SET NOT NULL;
