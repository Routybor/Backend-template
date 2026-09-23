#!/bin/sh
set -e

if [ "${CONFIRM_RESTORE:-}" != "yes" ]; then
  echo "Refusing to restore without CONFIRM_RESTORE=yes."
  exit 1
fi

if [ $# -eq 0 ]; then
  echo "Usage: $0 <backup_file>"
  echo ""
  echo "Available backups:"
  echo ""

  if [ -d "./backups/postgres" ]; then
    echo "PostgreSQL:"
    ls -1t ./backups/postgres/*.dump 2>/dev/null | head -5 || echo "  (no backups)"
  else
    echo "PostgreSQL: (no backups directory)"
  fi

  exit 0
fi

BACKUP_FILE=$1
ROOT_DIR=$(cd "$(dirname "$0")/.." && pwd)
cd "$ROOT_DIR"
ENV_FILE=${ENV_FILE:-.env}
if [ ! -f "$ENV_FILE" ]; then
  echo "ERROR: environment file not found: $ENV_FILE" >&2
  exit 1
fi

if [ ! -e "$BACKUP_FILE" ]; then
  echo "ERROR: Backup file not found: $BACKUP_FILE" >&2
  exit 1
fi

if [ ! -f "$BACKUP_FILE" ]; then
  echo "ERROR: Backup path is not a file (is it a directory?): $BACKUP_FILE" >&2
  exit 1
fi

BACKUP_NAME=$(basename "$BACKUP_FILE")
case "$BACKUP_NAME" in
  template-keycloak-*.dump)
    echo "Restoring PostgreSQL backup..."
    CONTAINER_PATH="/backups/postgres/$BACKUP_NAME"
    docker compose --env-file "$ENV_FILE" run --rm --no-deps -e CONFIRM_RESTORE=yes --entrypoint /bin/sh keycloak-db-backup /scripts/postgres-restore.sh "$CONTAINER_PATH"
    ;;
  *)
    echo "ERROR: unsupported backup filename: $BACKUP_NAME" >&2
    exit 1
    ;;
esac

echo "Restore completed for $BACKUP_FILE"
