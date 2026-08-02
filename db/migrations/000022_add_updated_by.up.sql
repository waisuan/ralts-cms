-- Migration 022: Add updatedBy column to machines and maintenance
-- Tracks the username of the last authenticated user who created/updated the record.
-- Nullable: existing rows and system/anonymous writes have no known editor.

ALTER TABLE machines ADD COLUMN IF NOT EXISTS "updatedBy" VARCHAR(200);
ALTER TABLE maintenance ADD COLUMN IF NOT EXISTS "updatedBy" VARCHAR(200);
