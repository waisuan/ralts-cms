-- Migration 018: Add index for TNC date sorting
-- This index improves performance when sorting machines by TNC date

CREATE INDEX IF NOT EXISTS idx_machines_tnc_date ON machines ("tncDate");

