#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR=$(cd "$(dirname "$0")/.." && pwd)
cd "$ROOT_DIR"
ENV_FILE=${ENV_FILE:-.env}
[[ -f "$ENV_FILE" ]] || { echo "Missing environment file: $ENV_FILE" >&2; exit 1; }
read_env() {
  awk -v key="$1" 'index($0, key "=") == 1 { print substr($0, length(key) + 2); exit }' "$ENV_FILE"
}

mkdir -p ./backups/postgres

BACKUP_UID=$(id -u)
BACKUP_GID=$(id -g)

if [[ ! -w ./backups || ! -w ./backups/postgres ]]; then
  echo "Repairing backup directory ownership for ${BACKUP_UID}:${BACKUP_GID}..."

  docker run --rm \
    --volume "$ROOT_DIR/backups:/backups" \
    postgres:17-alpine \
    chown "$BACKUP_UID:$BACKUP_GID" \
      /backups \
      /backups/postgres
fi

RUN_ID=${BACKUP_RUN_ID:-$(date -u +"%Y-%m-%d_%H-%M-%S")}
RETENTION_COUNT=${BACKUP_RETENTION_COUNT:-$(read_env BACKUP_RETENTION_COUNT)}
RETENTION_COUNT=${RETENTION_COUNT:-30}

BACKUP_ALLOW_EMPTY_POSTGRES=${BACKUP_ALLOW_EMPTY_POSTGRES:-$(read_env BACKUP_ALLOW_EMPTY_POSTGRES)}
BACKUP_SKIP_RETENTION=${BACKUP_SKIP_RETENTION:-false}

MANIFEST="backups/template-backup-${RUN_ID}.sha256"
MANIFEST_TMP="${MANIFEST}.tmp"

case "$RUN_ID" in
  *[!0-9A-Za-z_.-]*)
    echo "Invalid BACKUP_RUN_ID: $RUN_ID" >&2
    exit 1
    ;;
esac

if ! [[ "$RETENTION_COUNT" =~ ^[1-9][0-9]*$ ]]; then
  echo "BACKUP_RETENTION_COUNT must be a positive integer" >&2
  exit 1
fi

case "$BACKUP_SKIP_RETENTION" in
  true|false)
    ;;
  *)
    echo "BACKUP_SKIP_RETENTION must be true or false" >&2
    exit 1
    ;;
esac

postgres_files=(
  "backups/postgres/template-keycloak-${RUN_ID}.dump"
)

backup_files=("${postgres_files[@]}")

backup_complete=false

for file in "$MANIFEST" "${backup_files[@]}"; do
  if [[ -e "$file" ]]; then
    echo "Backup run already exists, refusing to overwrite it: $file" >&2
    exit 1
  fi
done

cleanup_failed_run() {
  if [[ "$backup_complete" != true ]]; then
    rm -f "$MANIFEST_TMP" "$MANIFEST"
    rm -f "${backup_files[@]}"
  fi
}

trap cleanup_failed_run EXIT INT TERM

echo "Running PostgreSQL backup..."

docker compose --env-file "$ENV_FILE" run --rm --no-deps \
  --user "$BACKUP_UID:$BACKUP_GID" \
  -e HOME=/tmp \
  -e "BACKUP_RUN_ID=$RUN_ID" \
  -e "BACKUP_ALLOW_EMPTY_POSTGRES=$BACKUP_ALLOW_EMPTY_POSTGRES" \
  keycloak-db-backup

for file in "${backup_files[@]}"; do
  if [[ ! -s "$file" ]]; then
    echo "Backup set is incomplete or empty: $file" >&2
    exit 1
  fi
done

checksum_files() {
  if command -v sha256sum >/dev/null 2>&1; then
    sha256sum "$@"
  else
    shasum -a 256 "$@"
  fi
}

(
  cd backups

  relative_files=()

  for file in "${backup_files[@]}"; do
    relative_files+=("${file#backups/}")
  done

  checksum_files "${relative_files[@]}"
) >"$MANIFEST_TMP"

chmod 600 "$MANIFEST_TMP" "${backup_files[@]}"
mv "$MANIFEST_TMP" "$MANIFEST"

BACKUP_MANIFEST="$MANIFEST" \
BACKUP_ALLOW_EMPTY_POSTGRES="$BACKUP_ALLOW_EMPTY_POSTGRES" \
ENV_FILE="$ENV_FILE" \
bash ./scripts/backup-verify.sh

backup_complete=true
trap - EXIT INT TERM

if [[ "$BACKUP_SKIP_RETENTION" != "true" ]]; then
  shopt -s nullglob

  manifests=(backups/template-backup-*.sha256)

  if ((${#manifests[@]} > RETENTION_COUNT)); then
    while IFS= read -r old_manifest; do
      while read -r _ relative_file; do
        case "$relative_file" in
          postgres/template-*.dump)
            rm -f "backups/$relative_file"
            ;;
        esac
      done <"$old_manifest"

      rm -f "$old_manifest" "${old_manifest%.sha256}.prerestore"
    done < <(
      ls -1t "${manifests[@]}" |
        tail -n "+$((RETENTION_COUNT + 1))"
    )
  fi
fi

echo "Backup set completed: $MANIFEST"
