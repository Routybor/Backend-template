#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR=$(cd "$(dirname "$0")/.." && pwd)
cd "$ROOT_DIR"
ENV_FILE=${ENV_FILE:-.env}
[[ -f "$ENV_FILE" ]] || { echo "Missing environment file: $ENV_FILE" >&2; exit 1; }

REQUESTED_RUN_ID="${1:-}"

DATES_FILE=$(mktemp)
COMPLETE_DATES=$(mktemp)

cleanup() {
  rm -f "$DATES_FILE" "$COMPLETE_DATES"
}

trap cleanup EXIT INT TERM

for file in backups/postgres/template-keycloak-*.dump; do
  [[ -s "$file" ]] || continue

  date_part=$(
    basename "$file" |
      sed -n 's/^template-keycloak-\([0-9_-]*\)\.dump$/\1/p'
  )

  if [[ -n "$date_part" ]]; then
    echo "$date_part" >>"$DATES_FILE"
  fi
done

if [[ ! -s "$DATES_FILE" ]]; then
  echo "No PostgreSQL backups found in backups/postgres."
  exit 1
fi

while IFS= read -r date_part; do
  all_found=true

  if [[ ! -s "backups/postgres/template-keycloak-${date_part}.dump" ]]; then
    all_found=false
  fi

  if [[ ! -s "backups/template-backup-${date_part}.sha256" ]]; then
    all_found=false
  fi

  if [[ "$all_found" == true ]]; then
    echo "$date_part" >>"$COMPLETE_DATES"
  fi
done < <(sort -ur "$DATES_FILE")

if [[ ! -s "$COMPLETE_DATES" ]]; then
  echo "No complete backup sets found."
  echo "Each set must contain the PostgreSQL dump and its sha256 manifest."
  exit 1
fi

SORTED_DATES=()

while IFS= read -r date_part; do
  SORTED_DATES+=("$date_part")
done < <(sort -r "$COMPLETE_DATES")

format_date() {
  local date_value=$1

  printf '%s %s:%s:%s UTC\n' \
    "${date_value:0:10}" \
    "${date_value:11:2}" \
    "${date_value:14:2}" \
    "${date_value:17:2}"
}

DISPLAY_DATES=()

for date_value in "${SORTED_DATES[@]}"; do
  DISPLAY_DATES+=("$(format_date "$date_value")")
done

echo ""
echo "=== Database Restore (select by date) ==="
echo ""
echo "Available backup dates (newest first):"
echo ""

if [[ -n "$REQUESTED_RUN_ID" ]]; then
  selected_date=""

  for date_value in "${SORTED_DATES[@]}"; do
    if [[ "$date_value" == "$REQUESTED_RUN_ID" ]]; then
      selected_date="$date_value"
      break
    fi
  done

  if [[ -z "$selected_date" ]]; then
    echo "Requested backup set not found or incomplete: $REQUESTED_RUN_ID" >&2
    exit 1
  fi

  selected_display_date="$(format_date "$selected_date")"
else
  PS3=$'\nSelect backup date: '

  select selected_display_date in "${DISPLAY_DATES[@]}"; do
    if [[ -n "$selected_display_date" ]]; then
      selected_index=$((REPLY - 1))
      selected_date="${SORTED_DATES[$selected_index]}"
      break
    fi

    echo "Invalid selection. Try again."
  done
fi

echo ""
echo "Selected: $selected_display_date ($selected_date)"
echo ""
echo "Files to restore for $selected_date:"
echo ""

pg_dump_file="backups/postgres/template-keycloak-${selected_date}.dump"

echo "  [PG]   $pg_dump_file"

echo ""
echo "=== Validating selected archives ==="

manifest="backups/template-backup-${selected_date}.sha256"
[[ -s "$manifest" ]] || { echo "Backup manifest is missing: $manifest" >&2; exit 1; }
BACKUP_MANIFEST="$manifest" ENV_FILE="$ENV_FILE" bash ./scripts/backup-verify.sh

echo ""
echo "--- Confirm Restore ---"
echo "WARNING: This will overwrite the current database."
echo ""

if [[ "${CONFIRM_RESTORE:-}" == "yes" ]]; then
  echo "CONFIRM_RESTORE=yes is set; proceeding without prompt."
else
  confirm=""
  read -r -p "Type 'yes' to proceed: " confirm || true

  if [[ "$confirm" != "yes" ]]; then
    echo "Aborted."
    exit 0
  fi
fi

echo ""
echo "=== Creating pre-restore backup ==="

pre_restore_run_id=$(date -u +"%Y-%m-%d_%H-%M-%S")

# Skip retention here: pruning could delete the very set we are about to restore.
BACKUP_RUN_ID="$pre_restore_run_id" \
BACKUP_SKIP_RETENTION=true \
BACKUP_ALLOW_EMPTY_POSTGRES="keycloak" \
ENV_FILE="$ENV_FILE" \
bash ./scripts/db-backup.sh

# Mark this set as a pre-restore snapshot so tooling (db-bundle.sh) skips it by default.
touch "backups/template-backup-${pre_restore_run_id}.prerestore"

container_path="/backups/postgres/$(basename "$pg_dump_file")"

echo ""
echo "=== Restoring PostgreSQL: keycloak ==="

docker compose --env-file "$ENV_FILE" run --rm --no-deps \
  -e CONFIRM_RESTORE=yes \
  --entrypoint /bin/sh \
  keycloak-db-backup \
  /scripts/postgres-restore.sh \
  "$container_path"

echo ""
echo "=== Restore from $selected_date complete ==="
