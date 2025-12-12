-- Drop indexes first
DROP INDEX IF EXISTS idx_audit_logs_resource_type_created_at;
DROP INDEX IF EXISTS idx_audit_logs_resource;
DROP INDEX IF EXISTS idx_audit_logs_user_id;
DROP INDEX IF EXISTS idx_audit_logs_created_at;

-- Drop the table
DROP TABLE IF EXISTS audit_logs;

