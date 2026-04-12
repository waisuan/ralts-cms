#!/usr/bin/env bash
set -euo pipefail

# ---------------------------------------------------------------------------
# migrate-data.sh -- Migrate Postgres data and S3 attachments from legacy
# roti to Ralts-CMS (Railway).
#
# Usage:
#   ./scripts/migrate-data.sh initial   # Full load (first time)
#   ./scripts/migrate-data.sh sync      # Incremental sync (daily cron)
#
# Required environment variables:
#   SOURCE_DATABASE_URL        -- legacy Postgres connection string
#   TARGET_DATABASE_URL        -- Railway Postgres connection string
#
# Optional (Postgres):
#   LAST_SYNC_FILE             -- path to persist last-sync timestamp
#                                 (default: /tmp/migrate_data_last_sync_ts)
#
# Optional (S3 -- set all to enable attachment sync):
#   SOURCE_S3_BUCKET           -- legacy AWS S3 bucket name
#   TARGET_S3_BUCKET           -- Railway storage bucket name
#   TARGET_S3_ENDPOINT         -- Railway S3 endpoint URL
#   TARGET_S3_ACCESS_KEY_ID    -- Railway bucket access key
#   TARGET_S3_SECRET_ACCESS_KEY -- Railway bucket secret key
#
#   Legacy AWS credentials are read from the standard AWS_ACCESS_KEY_ID,
#   AWS_SECRET_ACCESS_KEY, and AWS_DEFAULT_REGION env vars.
# ---------------------------------------------------------------------------

WORK_DIR=$(mktemp -d)
trap 'rm -rf "$WORK_DIR"' EXIT

LAST_SYNC_FILE="${LAST_SYNC_FILE:-/tmp/migrate_data_last_sync_ts}"

# -- Helpers ----------------------------------------------------------------

log()  { echo "[$(date -u '+%Y-%m-%d %H:%M:%S UTC')] $*"; }
die()  { log "ERROR: $*" >&2; exit 1; }

require_var() {
  local name="$1"
  if [[ -z "${!name:-}" ]]; then
    die "Required environment variable $name is not set"
  fi
}

src_psql() { psql --set=ON_ERROR_STOP=1 "$SOURCE_DATABASE_URL" "$@"; }
tgt_psql() { psql --set=ON_ERROR_STOP=1 "$TARGET_DATABASE_URL" "$@"; }

row_count() {
  local db_url="$1" table="$2"
  psql --set=ON_ERROR_STOP=1 "$db_url" -tAc "SELECT count(*) FROM $table"
}

validate_ts_format() {
  local ts="$1"
  if ! [[ "$ts" =~ ^[0-9]{4}-[0-9]{2}-[0-9]{2}\ [0-9]{2}:[0-9]{2}:[0-9]{2}$ ]]; then
    die "Invalid timestamp format: '$ts' (expected YYYY-MM-DD HH:MM:SS)"
  fi
}

# PL/pgSQL block that drops NOT NULL from every column in a temp table.
# Temp tables inherit NOT NULL from the source, but as staging tables we
# don't need those constraints -- the real table enforces them on INSERT.
drop_not_null_block() {
  local table="$1"
  cat <<PLPGSQL
DO \$\$ DECLARE r record;
BEGIN
  FOR r IN SELECT column_name FROM information_schema.columns
    WHERE table_name = '$table' AND table_schema LIKE 'pg_temp%'
  LOOP
    EXECUTE format('ALTER TABLE $table ALTER COLUMN %I DROP NOT NULL', r.column_name);
  END LOOP;
END \$\$;
PLPGSQL
}

# -- Column lists ---------------------------------------------------------------
# *_EXPORT includes id (CSV has it); *_IMPORT excludes id (let target assign).

MACHINES_EXPORT='id, "serialNumber", customer, state, "accountType", model, status,
  brand, district, "personInCharge", "reportedBy", "additionalNotes",
  attachment, "tncDate", "ppmDate", "createdAt", "updatedAt"'

MACHINES_EXPORT_FLAT='id, "serialNumber", customer, state, "accountType", model, status, brand, district, "personInCharge", "reportedBy", "additionalNotes", attachment, "tncDate", "ppmDate", "createdAt", "updatedAt"'

MACHINES_IMPORT='"serialNumber", customer, state, "accountType", model, status, brand, district, "personInCharge", "reportedBy", "additionalNotes", attachment, "tncDate", "ppmDate", "createdAt", "updatedAt"'

MAINTENANCE_EXPORT='id, "serialNumber", "workOrderNumber", "workOrderDate",
  "actionTaken", "reportedBy", "workOrderType", attachment,
  "createdAt", "updatedAt"'

MAINTENANCE_EXPORT_FLAT='id, "serialNumber", "workOrderNumber", "workOrderDate", "actionTaken", "reportedBy", "workOrderType", attachment, "createdAt", "updatedAt"'

MAINTENANCE_IMPORT='"serialNumber", "workOrderNumber", "workOrderDate", "actionTaken", "reportedBy", "workOrderType", attachment, "createdAt", "updatedAt"'

USERS_EXPORT_COLS='id, username, password, email, salt, approved, role, created_at'
USERS_IMPORT='username, password, email, salt, approved, role, created_at'

# -- Export -----------------------------------------------------------------

export_machines() {
  local where_clause="${1:-}"
  local query="SELECT $MACHINES_EXPORT FROM machines"
  [[ -n "$where_clause" ]] && query="$query WHERE $where_clause"
  query="$query ORDER BY id"

  log "Exporting machines..."
  src_psql -c "COPY ($query) TO STDOUT WITH CSV HEADER" > "$WORK_DIR/machines.csv"
  local count
  count=$(tail -n +2 "$WORK_DIR/machines.csv" | wc -l)
  log "Exported $count machines"
}

export_maintenance() {
  local where_clause="${1:-}"
  local query="SELECT $MAINTENANCE_EXPORT FROM maintenance"
  [[ -n "$where_clause" ]] && query="$query WHERE $where_clause"
  query="$query ORDER BY id"

  log "Exporting maintenance..."
  src_psql -c "COPY ($query) TO STDOUT WITH CSV HEADER" > "$WORK_DIR/maintenance.csv"
  local count
  count=$(tail -n +2 "$WORK_DIR/maintenance.csv" | wc -l)
  log "Exported $count maintenance records"
}

export_users() {
  log "Exporting users (always full)..."
  src_psql -c "COPY (SELECT $USERS_EXPORT_COLS FROM users ORDER BY id) TO STDOUT WITH CSV HEADER" \
    > "$WORK_DIR/users.csv"
  local count
  count=$(tail -n +2 "$WORK_DIR/users.csv" | wc -l)
  log "Exported $count users"
}

# -- Import -----------------------------------------------------------------
# Each import builds a single .sql file and runs it in one psql session so
# that the TEMP TABLE created at the start is visible to the \copy and INSERT.

import_machines() {
  log "Importing machines..."
  {
    echo "CREATE TEMP TABLE tmp_machines (LIKE machines INCLUDING DEFAULTS);"
    echo "ALTER TABLE tmp_machines DROP COLUMN IF EXISTS search_vector;"
    drop_not_null_block tmp_machines
    echo "\\copy tmp_machines($MACHINES_EXPORT_FLAT) FROM '$WORK_DIR/machines.csv' WITH CSV HEADER"
    cat <<SQL
INSERT INTO machines ($MACHINES_IMPORT)
SELECT $MACHINES_IMPORT FROM tmp_machines
ON CONFLICT ("serialNumber") DO UPDATE SET
  customer       = EXCLUDED.customer,
  state          = EXCLUDED.state,
  "accountType"  = EXCLUDED."accountType",
  model          = EXCLUDED.model,
  status         = EXCLUDED.status,
  brand          = EXCLUDED.brand,
  district       = EXCLUDED.district,
  "personInCharge"  = EXCLUDED."personInCharge",
  "reportedBy"      = EXCLUDED."reportedBy",
  "additionalNotes" = EXCLUDED."additionalNotes",
  attachment     = EXCLUDED.attachment,
  "tncDate"      = EXCLUDED."tncDate",
  "ppmDate"      = EXCLUDED."ppmDate",
  "createdAt"    = EXCLUDED."createdAt",
  "updatedAt"    = EXCLUDED."updatedAt";
DROP TABLE IF EXISTS tmp_machines;
SQL
  } > "$WORK_DIR/import_machines.sql"
  tgt_psql -f "$WORK_DIR/import_machines.sql"
  log "Machines import complete"
}

import_maintenance() {
  log "Importing maintenance..."
  {
    echo "CREATE TEMP TABLE tmp_maintenance (LIKE maintenance INCLUDING DEFAULTS);"
    echo "ALTER TABLE tmp_maintenance DROP COLUMN IF EXISTS search_vector;"
    drop_not_null_block tmp_maintenance
    echo "\\copy tmp_maintenance($MAINTENANCE_EXPORT_FLAT) FROM '$WORK_DIR/maintenance.csv' WITH CSV HEADER"
    cat <<SQL
INSERT INTO maintenance ($MAINTENANCE_IMPORT)
SELECT $MAINTENANCE_IMPORT FROM tmp_maintenance
ON CONFLICT ("serialNumber", "workOrderNumber") DO UPDATE SET
  "workOrderDate" = EXCLUDED."workOrderDate",
  "actionTaken"   = EXCLUDED."actionTaken",
  "reportedBy"    = EXCLUDED."reportedBy",
  "workOrderType" = EXCLUDED."workOrderType",
  attachment      = EXCLUDED.attachment,
  "createdAt"     = EXCLUDED."createdAt",
  "updatedAt"     = EXCLUDED."updatedAt";
DROP TABLE IF EXISTS tmp_maintenance;
SQL
  } > "$WORK_DIR/import_maintenance.sql"
  tgt_psql -f "$WORK_DIR/import_maintenance.sql"
  log "Maintenance import complete"
}

import_users() {
  log "Importing users..."
  {
    echo "CREATE TEMP TABLE tmp_users (LIKE users INCLUDING DEFAULTS);"
    drop_not_null_block tmp_users
    echo "\\copy tmp_users($USERS_EXPORT_COLS) FROM '$WORK_DIR/users.csv' WITH CSV HEADER"
    # NOTE: This UPSERT overwrites password+salt with the legacy values on every
    # sync run. Any transparent password upgrades performed by Ralts' dual-mode
    # auth will be reverted until syncing is turned off. This is intentional --
    # during the coexistence period Ralts is treated as read-only.
    cat <<SQL
INSERT INTO users ($USERS_IMPORT, status)
SELECT $USERS_IMPORT,
       CASE WHEN approved THEN 'approved' ELSE 'pending_approval' END
FROM tmp_users
ON CONFLICT (username) DO UPDATE SET
  password = EXCLUDED.password,
  email    = EXCLUDED.email,
  salt     = EXCLUDED.salt,
  approved = EXCLUDED.approved,
  role     = EXCLUDED.role,
  status   = EXCLUDED.status;
DROP TABLE IF EXISTS tmp_users;
SQL
  } > "$WORK_DIR/import_users.sql"
  tgt_psql -f "$WORK_DIR/import_users.sql"
  log "Users import complete"
}

# -- Post-load operations ---------------------------------------------------

reset_sequences() {
  log "Resetting sequences..."
  tgt_psql -tA <<'SQL'
SELECT setval('machines_id_seq',    GREATEST((SELECT COALESCE(MAX(id), 0) FROM machines),    1));
SELECT setval('maintenance_id_seq', GREATEST((SELECT COALESCE(MAX(id), 0) FROM maintenance), 1));
SELECT setval('users_id_seq',       GREATEST((SELECT COALESCE(MAX(id), 0) FROM users),       1));
SQL
  log "Sequences reset"
}

validate() {
  log "Validating row counts..."
  local src_machines src_maint src_users tgt_machines tgt_maint tgt_users

  src_machines=$(row_count "$SOURCE_DATABASE_URL" machines)
  src_maint=$(row_count "$SOURCE_DATABASE_URL" maintenance)
  src_users=$(row_count "$SOURCE_DATABASE_URL" users)

  tgt_machines=$(row_count "$TARGET_DATABASE_URL" machines)
  tgt_maint=$(row_count "$TARGET_DATABASE_URL" maintenance)
  tgt_users=$(row_count "$TARGET_DATABASE_URL" users)

  log "  machines:    source=$src_machines  target=$tgt_machines"
  log "  maintenance: source=$src_maint  target=$tgt_maint"
  log "  users:       source=$src_users  target=$tgt_users"

  local ok=true
  [[ "$src_machines" != "$tgt_machines" ]] && { log "  WARNING: machines count mismatch!"; ok=false; }
  [[ "$src_maint" != "$tgt_maint" ]]       && { log "  WARNING: maintenance count mismatch!"; ok=false; }
  [[ "$src_users" != "$tgt_users" ]]       && { log "  WARNING: users count mismatch!"; ok=false; }

  if $ok; then
    log "Row counts match"
  else
    log "Row count mismatches detected -- review above"
  fi
}

health_checks() {
  log "Running health checks..."

  local legacy_pw_count
  legacy_pw_count=$(tgt_psql -tAc "SELECT count(*) FROM users WHERE salt LIKE '\$2a\$%' OR salt LIKE '\$2b\$%'")
  log "  Legacy password hashes remaining: $legacy_pw_count"

  local sv_machines sv_maint
  sv_machines=$(tgt_psql -tAc "SELECT count(*) FROM machines WHERE search_vector IS NOT NULL")
  sv_maint=$(tgt_psql -tAc "SELECT count(*) FROM maintenance WHERE search_vector IS NOT NULL")
  log "  search_vector populated: machines=$sv_machines  maintenance=$sv_maint"
}

# -- S3 sync ----------------------------------------------------------------
# Two-step local relay: download from legacy AWS, upload to Railway bucket.
# aws s3 sync is idempotent -- only new/modified objects are transferred on
# the upload leg (compares size + last-modified).

s3_sync() {
  require_var SOURCE_S3_BUCKET
  require_var TARGET_S3_BUCKET
  require_var TARGET_S3_ENDPOINT
  require_var TARGET_S3_ACCESS_KEY_ID
  require_var TARGET_S3_SECRET_ACCESS_KEY

  log "Syncing S3 attachments..."
  local s3_work="$WORK_DIR/s3"
  mkdir -p "$s3_work"

  aws s3 sync "s3://$SOURCE_S3_BUCKET" "$s3_work" \
    --exclude "*.sql" --exclude "*.zip"

  local src_count
  src_count=$(find "$s3_work" -type f | wc -l)
  log "  Downloaded $src_count files from source"

  if [[ "$src_count" -eq 0 ]]; then
    log "  WARNING: No files downloaded from source -- skipping upload"
    return
  fi

  AWS_ACCESS_KEY_ID="$TARGET_S3_ACCESS_KEY_ID" \
  AWS_SECRET_ACCESS_KEY="$TARGET_S3_SECRET_ACCESS_KEY" \
    aws s3 sync "$s3_work" "s3://$TARGET_S3_BUCKET" \
    --endpoint-url "$TARGET_S3_ENDPOINT" \
    --region auto

  log "  Uploaded to target bucket"
  log "S3 sync complete"
}

maybe_s3_sync() {
  if [[ -n "${SOURCE_S3_BUCKET:-}" ]]; then
    s3_sync
  else
    log "S3 sync skipped (SOURCE_S3_BUCKET not set)"
  fi
}

# -- Modes ------------------------------------------------------------------

run_initial() {
  log "=== INITIAL MIGRATION ==="
  require_var SOURCE_DATABASE_URL
  require_var TARGET_DATABASE_URL

  export_machines
  export_maintenance
  export_users

  import_machines
  import_maintenance
  import_users

  reset_sequences
  validate
  health_checks
  maybe_s3_sync

  log "=== INITIAL MIGRATION COMPLETE ==="
}

run_sync() {
  log "=== INCREMENTAL SYNC ==="
  require_var SOURCE_DATABASE_URL
  require_var TARGET_DATABASE_URL

  local last_sync_ts
  if [[ -f "$LAST_SYNC_FILE" ]]; then
    last_sync_ts=$(cat "$LAST_SYNC_FILE")
  else
    last_sync_ts=$(date -u -d '24 hours ago' '+%Y-%m-%d %H:%M:%S')
  fi
  validate_ts_format "$last_sync_ts"
  log "Syncing changes since: $last_sync_ts"

  export_machines "\"updatedAt\" > '$last_sync_ts'"
  export_maintenance "\"updatedAt\" > '$last_sync_ts'"
  export_users

  import_machines
  import_maintenance
  import_users

  validate
  health_checks
  maybe_s3_sync

  date -u '+%Y-%m-%d %H:%M:%S' > "$LAST_SYNC_FILE"
  log "Sync timestamp saved to $LAST_SYNC_FILE"
  log "=== INCREMENTAL SYNC COMPLETE ==="
}

# -- Main -------------------------------------------------------------------

case "${1:-}" in
  initial) run_initial ;;
  sync)    run_sync ;;
  *)
    echo "Usage: $0 {initial|sync}"
    echo ""
    echo "Modes:"
    echo "  initial  Full migration (first time)"
    echo "  sync     Incremental sync (daily cron)"
    echo ""
    echo "Environment variables (Postgres -- required):"
    echo "  SOURCE_DATABASE_URL          Legacy Postgres connection string"
    echo "  TARGET_DATABASE_URL          Railway Postgres connection string"
    echo "  LAST_SYNC_FILE               Path to persist sync timestamp (default: /tmp/migrate_data_last_sync_ts)"
    echo ""
    echo "Environment variables (S3 -- optional, set all to enable):"
    echo "  SOURCE_S3_BUCKET             Legacy AWS S3 bucket name"
    echo "  TARGET_S3_BUCKET             Railway storage bucket name"
    echo "  TARGET_S3_ENDPOINT           Railway S3 endpoint URL"
    echo "  TARGET_S3_ACCESS_KEY_ID      Railway bucket access key"
    echo "  TARGET_S3_SECRET_ACCESS_KEY  Railway bucket secret key"
    echo "  AWS_ACCESS_KEY_ID            Legacy AWS access key (standard AWS var)"
    echo "  AWS_SECRET_ACCESS_KEY        Legacy AWS secret key (standard AWS var)"
    echo "  AWS_DEFAULT_REGION           Legacy AWS region (standard AWS var)"
    exit 1
    ;;
esac
