#!/bin/sh
set -e

BACKUP_FILE=${1}

if [ -z "$BACKUP_FILE" ]; then
  echo "Usage: $0 <backup_file>"
  echo "Example: CONFIRM_RESTORE=yes $0 /backups/postgres/template-keycloak-2026-09-23_12-00-00.dump"
  exit 1
fi

if [ ! -f "$BACKUP_FILE" ]; then
  echo "Error: Backup file not found: $BACKUP_FILE"
  exit 1
fi

if [ "${CONFIRM_RESTORE:-}" != "yes" ]; then
  echo "Refusing to restore without CONFIRM_RESTORE=yes."
  exit 1
fi

BACKUP_NAME=$(basename "$BACKUP_FILE")
case "$BACKUP_NAME" in
  template-keycloak-*)
    LABEL=keycloak
    ;;
  *)
    echo "Error: cannot determine database from backup name: $BACKUP_NAME" >&2
    exit 1
    ;;
esac

HOST="keycloak-db"
USER="${KEYCLOAK_DB_USER:-keycloak_user}"
PASSWORD="${KEYCLOAK_DB_PASSWORD:-keycloak_pass123}"
DB="${KEYCLOAK_DB_NAME:-keycloak_db}"

echo "Restoring $LABEL database from $BACKUP_FILE"
echo "Target: $HOST/$DB"

echo "Validating backup archive..."
case "$BACKUP_FILE" in
  *.dump)
    pg_restore --list "$BACKUP_FILE" >/dev/null
    echo "Running transactional pg_restore..."
    PGPASSWORD="$PASSWORD" pg_restore -h "$HOST" -U "$USER" -d "$DB" --clean --if-exists --no-owner --no-privileges --single-transaction --exit-on-error "$BACKUP_FILE"
    ;;
  *)
    echo "Error: unsupported PostgreSQL backup format: $BACKUP_FILE"
    exit 1
    ;;
esac

echo "Restore complete for $LABEL database"
