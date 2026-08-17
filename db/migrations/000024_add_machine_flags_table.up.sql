-- Migration 024: Persistent flags on machine records.
--
-- Reasons:
--   - missing_values:     admin-created; a required field is empty.
--   - other:              admin-created; free-text note explains the reason.
--   - requires_attention: reserved for system-raised PPM flags. No code path
--     creates one; the API rejects it as a manual reason.
--
-- A flag is a durable record with an open/resolved lifecycle so a badge can be
-- shown on the machine row until it is explicitly resolved by an admin or by
-- the machine's assignee.

CREATE TABLE IF NOT EXISTS machine_flags (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    machine_serial_number VARCHAR(200) NOT NULL
        REFERENCES machines("serialNumber") ON DELETE CASCADE ON UPDATE CASCADE,
    reason VARCHAR(30) NOT NULL,
    ppm_status VARCHAR(20),
    note TEXT,
    status VARCHAR(20) NOT NULL DEFAULT 'open',
    created_by BIGINT REFERENCES users(id) ON DELETE SET NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    resolved_by BIGINT REFERENCES users(id) ON DELETE SET NULL,
    resolved_at TIMESTAMP WITH TIME ZONE
);

-- At most one open requires_attention flag per machine, so any code raising that
-- reason is idempotent without read-then-write logic.
CREATE UNIQUE INDEX IF NOT EXISTS idx_machine_flags_open_requires_attention
    ON machine_flags(machine_serial_number)
    WHERE reason = 'requires_attention' AND status = 'open';

CREATE INDEX IF NOT EXISTS idx_machine_flags_machine
    ON machine_flags(machine_serial_number);

CREATE INDEX IF NOT EXISTS idx_machine_flags_open
    ON machine_flags(machine_serial_number)
    WHERE status = 'open';

CREATE INDEX IF NOT EXISTS idx_machine_flags_created_at
    ON machine_flags(created_at DESC);
