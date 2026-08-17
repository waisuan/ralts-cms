-- Migration 025: In-app notifications inbox.
--
-- Rows are created when:
--   * a machine's assigned_user changes (type = 'assigned')
--   * a machine flag is created (type = 'flagged')
--
-- The frontend polls unread-count/list for a bell + /inbox page. Recipients
-- see only their own rows (repo scopes by user_id from the JWT context).

CREATE TABLE IF NOT EXISTS notifications (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    type VARCHAR(30) NOT NULL,
    machine_serial_number VARCHAR(200)
        REFERENCES machines("serialNumber") ON DELETE CASCADE ON UPDATE CASCADE,
    flag_id UUID REFERENCES machine_flags(id) ON DELETE CASCADE,
    title VARCHAR(255) NOT NULL,
    body TEXT,
    actor_user_id BIGINT REFERENCES users(id) ON DELETE SET NULL,
    read_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Recent-first listing per recipient.
CREATE INDEX IF NOT EXISTS idx_notifications_user_created_at
    ON notifications(user_id, created_at DESC);

-- Fast unread badge counts.
CREATE INDEX IF NOT EXISTS idx_notifications_user_unread
    ON notifications(user_id)
    WHERE read_at IS NULL;
