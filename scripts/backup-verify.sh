#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR=$(cd "$(dirname "$0")/.." && pwd)
cd "$ROOT_DIR"
ENV_FILE=${ENV_FILE:-.env}
[[ -f "$ENV_FILE" ]] || { echo "Missing environment file: $ENV_FILE" >&2; exit 1; }
read_env() {
  awk -v key="$1" 'index($0, key "=") == 1 { print substr($0, length(key) + 2); exit }' "$ENV_FILE"
}
BACKUP_ALLOW_EMPTY_POSTGRES=${BACKUP_ALLOW_EMPTY_POSTGRES:-$(read_env BACKUP_ALLOW_EMPTY_POSTGRES)}

FULL_VERIFY=false
if [[ "${1:-}" == "--full" ]]; then
  FULL_VERIFY=true
elif [[ $# -gt 0 ]]; then
  echo "Usage: $0 [--full]" >&2
  exit 1
fi

if [[ -n "${BACKUP_MANIFEST:-}" ]]; then
  manifest_name=$(basename "$BACKUP_MANIFEST")
else
  shopt -s nullglob
  manifests=(backups/template-backup-*.sha256)
  if ((${#manifests[@]} == 0)); then
    echo "No complete template backup manifests found" >&2
    exit 1
  fi
  manifest_name=$(basename "$(ls -1t "${manifests[@]}" | head -n 1)")
fi

MANIFEST="$ROOT_DIR/backups/$manifest_name"
if [[ ! -s "$MANIFEST" ]]; then
  echo "Backup manifest is missing or empty: $MANIFEST" >&2
  exit 1
fi

verify_checksums() (
  cd "$ROOT_DIR/backups"
  if command -v sha256sum >/dev/null 2>&1; then
    sha256sum --check "$manifest_name"
  else
    shasum -a 256 --check "$manifest_name"
  fi
)

postgres_files=()
while read -r checksum relative_file; do
  if [[ ! "$checksum" =~ ^[0-9a-fA-F]{64}$ ]]; then
    echo "Invalid checksum entry in $MANIFEST" >&2
    exit 1
  fi

  case "$relative_file" in
    postgres/template-*.dump)
      postgres_files+=("$relative_file")
      ;;
    *)
      echo "Unsafe or unexpected backup path in manifest: $relative_file" >&2
      exit 1
      ;;
  esac
done <"$MANIFEST"

if ((${#postgres_files[@]} != 1)); then
  echo "Incomplete backup set: expected 1 PostgreSQL dump" >&2
  exit 1
fi

verify_checksums

postgres_container_files=()
for relative_file in "${postgres_files[@]}"; do
  postgres_container_files+=("/backups/postgres/$(basename "$relative_file")")
done

docker compose --env-file "$ENV_FILE" run --rm --no-deps \
  -e "BACKUP_ALLOW_EMPTY_POSTGRES=$BACKUP_ALLOW_EMPTY_POSTGRES" \
  --entrypoint /bin/sh keycloak-db-backup -c '
    set -eu
    for file do
      test -s "$file"
      pg_restore --list "$file" >/dev/null
      table_count=$(pg_restore --list "$file" | awk '\''$4 == "TABLE" { count++ } END { print count + 0 }'\'')
      if [ "$table_count" -eq 0 ]; then
        filename=$(basename "$file")
        label=${filename#template-}
        label=${label%%-*}
        case ",${BACKUP_ALLOW_EMPTY_POSTGRES:-}," in
          *",true"*|*",${label},"*)
            echo "Warning: $label archive has no user tables; explicitly allowed"
            ;;
          *)
            echo "Archive contains no user tables: $file" >&2
            exit 1
            ;;
        esac
      fi
    done
  ' sh "${postgres_container_files[@]}"

echo "Backup checksums and archive structure are valid: $manifest_name"

if [[ "$FULL_VERIFY" != true ]]; then
  exit 0
fi

suffix="${manifest_name#template-backup-}"
suffix="${suffix%.sha256}"
safe_suffix=$(printf '%s' "$suffix" | tr -c '0-9A-Za-z' '-')
pg_container="template-backup-verify-pg-${safe_suffix}"
pg_volume="${pg_container}-data"

cleanup_verify_resources() {
  docker rm -f "$pg_container" >/dev/null 2>&1 || true
  docker volume rm "$pg_volume" >/dev/null 2>&1 || true
}
trap cleanup_verify_resources EXIT INT TERM
cleanup_verify_resources

docker volume create "$pg_volume" >/dev/null
docker run -d --name "$pg_container" \
  -e POSTGRES_PASSWORD=verify-only \
  -v "$pg_volume:/var/lib/postgresql/data" \
  -v "$ROOT_DIR/backups/postgres:/backups:ro" \
  postgres:17-alpine >/dev/null

# Probe via TCP: during first-time init the image runs a temporary server that
# only listens on the unix socket; TCP readiness means the real server is up.
for _ in $(seq 1 60); do
  if docker exec "$pg_container" pg_isready -q -h 127.0.0.1 -U postgres; then
    break
  fi
  sleep 1
done
docker exec "$pg_container" pg_isready -q -h 127.0.0.1 -U postgres

index=0
for relative_file in "${postgres_files[@]}"; do
  index=$((index + 1))
  database="verify_${index}"
  archive="/backups/$(basename "$relative_file")"
  docker exec "$pg_container" createdb -h 127.0.0.1 -U postgres "$database"
  docker exec "$pg_container" pg_restore \
    -h 127.0.0.1 \
    -U postgres \
    -d "$database" \
    --no-owner \
    --no-privileges \
    --single-transaction \
    --exit-on-error \
    "$archive"
  table_count=$(docker exec "$pg_container" psql -At -h 127.0.0.1 -U postgres -d "$database" -c \
    "SELECT count(*) FROM pg_catalog.pg_tables WHERE schemaname NOT IN ('pg_catalog', 'information_schema');")
  if [[ "$table_count" -eq 0 ]]; then
    echo "Isolated PostgreSQL restore is structurally valid but contains no user tables: $relative_file"
  fi
  docker exec "$pg_container" dropdb -h 127.0.0.1 -U postgres "$database"
done

cleanup_verify_resources
trap - EXIT INT TERM
echo "Full isolated restore verification completed successfully: $manifest_name"
