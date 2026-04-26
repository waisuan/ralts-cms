#!/usr/bin/env bash
set -euo pipefail

# Railway Postgres (optional Railway bucket) → Cloudflare R2. Env vars: railway-cron.toml
# Set BACKUP_DEBUG=1 for bash trace (xtrace). OOM kills leave no log — check deployment exit code (137).

log()  { echo "[$(date -u '+%Y-%m-%d %H:%M:%S UTC')] $*" >&2; }
die()  { log "ERROR: $*"; exit 1; }

[[ "${BACKUP_DEBUG:-0}" == "1" ]] && set -x

err_trap() {
  local s=$?
  log "ERROR: command failed (exit $s) at line ${BASH_LINENO[0]}"
  exit "$s"
}
trap err_trap ERR

log "backup-to-r2.sh started"

WORK_DIR=$(mktemp -d)
trap 'rm -rf "$WORK_DIR"' EXIT
log "work dir: $WORK_DIR"

require_var() {
  local name="$1"
  if [[ -z "${!name:-}" ]]; then
    die "Required environment variable $name is not set"
  fi
}

resolve_database_url() {
  if [[ -n "${DATABASE_URL:-}" ]]; then
    echo "$DATABASE_URL"
    return
  fi
  if [[ -n "${BACKUP_DATABASE_URL:-}" ]]; then
    echo "$BACKUP_DATABASE_URL"
    return
  fi
  if [[ -n "${TARGET_DATABASE_URL:-}" ]]; then
    echo "$TARGET_DATABASE_URL"
    return
  fi
  die "Set DATABASE_URL, BACKUP_DATABASE_URL, or TARGET_DATABASE_URL"
}

r2_aws() {
  AWS_ACCESS_KEY_ID="$BACKUP_R2_ACCESS_KEY_ID" \
  AWS_SECRET_ACCESS_KEY="$BACKUP_R2_SECRET_ACCESS_KEY" \
    aws "$@" --endpoint-url "$BACKUP_R2_ENDPOINT" --region auto
}

railway_s3_aws() {
  AWS_ACCESS_KEY_ID="$TARGET_S3_ACCESS_KEY_ID" \
  AWS_SECRET_ACCESS_KEY="$TARGET_S3_SECRET_ACCESS_KEY" \
    aws "$@" --endpoint-url "$TARGET_S3_ENDPOINT" --region auto
}

backup_postgres() {
  local db_url="$1"
  local stamp key dump

  stamp=$(date -u '+%Y%m%dT%H%M%SZ')
  dump="$WORK_DIR/ralts-postgres-$stamp.dump"
  key="${BACKUP_PREFIX:+$BACKUP_PREFIX/}postgres/ralts-postgres-$stamp.dump"

  log "Dumping PostgreSQL (custom format)..."
  pg_dump --format=custom --no-owner --no-acl "$db_url" -f "$dump"

  log "Uploading s3://$BACKUP_R2_BUCKET/$key"
  r2_aws s3 cp "$dump" "s3://$BACKUP_R2_BUCKET/$key"
  rm -f "$dump"

  log "Postgres backup uploaded"
}

maybe_backup_object_storage() {
  if [[ "${BACKUP_SKIP_OBJECT_STORAGE:-0}" == "1" ]]; then
    log "Object storage backup skipped (BACKUP_SKIP_OBJECT_STORAGE=1)"
    return
  fi

  if [[ -z "${TARGET_S3_BUCKET:-}" || -z "${TARGET_S3_ENDPOINT:-}" \
     || -z "${TARGET_S3_ACCESS_KEY_ID:-}" || -z "${TARGET_S3_SECRET_ACCESS_KEY:-}" ]]; then
    log "Object storage backup skipped (Railway S3 vars not fully set)"
    return
  fi

  log "Syncing Railway object storage to R2 (local staging)..."
  local s3_work="$WORK_DIR/s3-objects"
  mkdir -p "$s3_work"

  railway_s3_aws s3 sync "s3://$TARGET_S3_BUCKET" "$s3_work" --only-show-errors

  local src_count
  src_count=$(find "$s3_work" -type f | wc -l)
  if [[ "$src_count" -eq 0 ]]; then
    log "  WARNING: No files downloaded from bucket -- skipping R2 object upload"
    return
  fi
  log "  Staged $src_count files"

  local dest_prefix="${BACKUP_PREFIX:+$BACKUP_PREFIX/}objects"
  r2_aws s3 sync "$s3_work" "s3://$BACKUP_R2_BUCKET/$dest_prefix" --only-show-errors
  log "Object storage backup complete"
}

log "=== RAILWAY → R2 BACKUP ==="

require_var BACKUP_R2_BUCKET
require_var BACKUP_R2_ENDPOINT
require_var BACKUP_R2_ACCESS_KEY_ID
require_var BACKUP_R2_SECRET_ACCESS_KEY

db_url=$(resolve_database_url)
backup_postgres "$db_url"
maybe_backup_object_storage

log "=== BACKUP COMPLETE ==="
