-- Flags closed as superseded by the up migration stay closed: there is no
-- record of which ones were auto-closed beyond their resolution note.

DROP INDEX IF EXISTS idx_machine_flags_one_open_per_machine;

CREATE UNIQUE INDEX IF NOT EXISTS idx_machine_flags_open_requires_attention
    ON machine_flags(machine_serial_number)
    WHERE reason = 'requires_attention' AND status = 'open';
