-- Migration 023: Add assignedUserId FK on machines linking to users.
-- The personInCharge column is retained and still holds the display name for
-- the assignee, so CSV export, search_text (see migration 020) and existing
-- display code keep working. It is now populated two ways: derived server-side
-- from the assigned user's username when assignedUserId is set, or kept as
-- client-supplied free text when it is NULL (an assignee with no account yet,
-- who therefore receives no notifications).

ALTER TABLE machines
    ADD COLUMN IF NOT EXISTS "assignedUserId" BIGINT
    REFERENCES users(id) ON DELETE SET NULL;

CREATE INDEX IF NOT EXISTS idx_machines_assigned_user_id
    ON machines("assignedUserId");

-- Best-effort backfill: link existing rows whose personInCharge text exactly
-- matches a user's username (case-insensitive). Unmatched rows stay NULL.
UPDATE machines m
SET "assignedUserId" = u.id
FROM users u
WHERE m."assignedUserId" IS NULL
  AND m."personInCharge" IS NOT NULL
  AND m."personInCharge" <> ''
  AND LOWER(m."personInCharge") = LOWER(u.username);
