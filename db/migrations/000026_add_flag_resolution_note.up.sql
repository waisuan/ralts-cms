-- Migration 026: Optional note recorded when a flag is resolved.
--
-- Whoever clears a flag can say what they did about it. The note is optional —
-- most flags are self-explanatory — and is kept alongside resolved_by and
-- resolved_at as part of the flag's audit trail.

ALTER TABLE machine_flags ADD COLUMN IF NOT EXISTS resolution_note TEXT;
