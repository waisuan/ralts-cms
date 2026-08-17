DROP INDEX IF EXISTS idx_machines_assigned_user_id;
ALTER TABLE machines DROP COLUMN IF EXISTS "assignedUserId";
