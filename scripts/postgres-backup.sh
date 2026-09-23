#!/bin/sh
set -eu

BACKUP_DIR=${BACKUP_DIR:-/backups/postgres}
BACKUP_READY_TIMEOUT=${BACKUP_READY_TIMEOUT:-120}
BACKUP_RUN_ID=${BACKUP_RUN_ID:-$(date -u +"%Y-%m-%d_%H-%M-%S")}
BACKUP_ALLOW_EMPTY_POSTGRES=${BACKUP_ALLOW_EMPTY_POSTGRES:-}

case "$BACKUP_RUN_ID" in
  *[!0-9A-Za-z_.-]*)
    echo "Invalid BACKUP_RUN_ID: $BACKUP_RUN_ID" >&2
    exit 1
    ;;
esac

mkdir -p "$BACKUP_DIR"

HOST="keycloak-db"
USER="${KEYCLOAK_DB_USER:-keycloak_user}"
PASSWORD="${KEYCLOAK_DB_PASSWORD:-keycloak_pass123}"
DB="${KEYCLOAK_DB_NAME:-keycloak_db}"
LABEL="keycloak"

TARGET="$BACKUP_DIR/template-${LABEL}-${BACKUP_RUN_ID}.dump"
TEMP_TARGET="${TARGET}.tmp"

ELAPSED=0

echo "Waiting for PostgreSQL $HOST/$DB..."

until pg_isready -q -h "$HOST" -U "$USER" -d "$DB"; do
  if [ "$ELAPSED" -ge "$BACKUP_READY_TIMEOUT" ]; then
    echo "PostgreSQL $HOST/$DB was not ready after ${BACKUP_READY_TIMEOUT}s" >&2
    exit 1
  fi
  sleep 2
  ELAPSED=$((ELAPSED + 2))
done

echo "Backing up $LABEL ($HOST/$DB) to $TARGET"

if ! PGPASSWORD="$PASSWORD" pg_dump -h "$HOST" -U "$USER" -d "$DB" -Fc -f "$TEMP_TARGET"; then
  rm -f "$TEMP_TARGET"
  exit 1
fi

test -s "$TEMP_TARGET"
pg_restore --list "$TEMP_TARGET" >/dev/null

TABLE_COUNT=$(pg_restore --list "$TEMP_TARGET" | awk '$4 == "TABLE" { count++ } END { print count + 0 }')
if [ "$TABLE_COUNT" -eq 0 ]; then
  case ",${BACKUP_ALLOW_EMPTY_POSTGRES}," in
    *",true"*|*",${LABEL},"*)
      echo "Warning: $LABEL has no user tables; allowed by BACKUP_ALLOW_EMPTY_POSTGRES"
      ;;
    *)
      echo "Backup validation failed for $LABEL: archive contains no user tables" >&2
      rm -f "$TEMP_TARGET"
      exit 1
      ;;
  esac
fi

mv "$TEMP_TARGET" "$TARGET"

echo "PostgreSQL backup complete: $TARGET"
