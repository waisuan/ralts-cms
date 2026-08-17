-- Migration 027: At most one open flag per machine.
--
-- A machine is either flagged or it is not; raising a second flag on an already
-- flagged machine now replaces the open one rather than stacking another row.
-- Enforced by a partial unique index so the invariant survives concurrent
-- writers, with the API doing an upsert against it.

-- Existing data may carry several open flags on one machine. Keep the newest and
-- close the rest, so history is preserved rather than deleted. resolved_by is
-- left NULL because no user made this decision.
WITH ranked AS (
    SELECT id,
           ROW_NUMBER() OVER (
               PARTITION BY machine_serial_number
               ORDER BY created_at DESC, id DESC
           ) AS rank
    FROM machine_flags
    WHERE status = 'open'
)
UPDATE machine_flags f
SET status = 'resolved',
    resolved_at = NOW(),
    resolution_note = 'Closed automatically: superseded by a newer flag on the same machine.'
FROM ranked
WHERE f.id = ranked.id AND ranked.rank > 1;

-- Subsumed by the broader index below.
DROP INDEX IF EXISTS idx_machine_flags_open_requires_attention;

CREATE UNIQUE INDEX IF NOT EXISTS idx_machine_flags_one_open_per_machine
    ON machine_flags(machine_serial_number)
    WHERE status = 'open';
