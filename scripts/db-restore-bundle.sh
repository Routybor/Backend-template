#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR=$(cd "$(dirname "$0")/.." && pwd)
cd "$ROOT_DIR"
ENV_FILE=${ENV_FILE:-.env}
[[ -f "$ENV_FILE" ]] || { echo "Missing environment file: $ENV_FILE" >&2; exit 1; }

# Restore a full backup set from a single bundle produced by db-bundle.sh.
#
# Usage:
#   CONFIRM_RESTORE=yes scripts/db-restore-bundle.sh <bundle.tar.gz>

BUNDLE="${1:-}"

if [[ -z "$BUNDLE" ]]; then
  echo "Usage: CONFIRM_RESTORE=yes $0 <bundle.tar.gz>" >&2
  exit 1
fi

if [[ ! -s "$BUNDLE" ]]; then
  echo "Bundle not found or empty: $BUNDLE" >&2
  exit 1
fi

if [[ "${CONFIRM_RESTORE:-}" != "yes" ]]; then
  echo "Refusing to restore without CONFIRM_RESTORE=yes."
  exit 1
fi

bundle_name=$(basename "$BUNDLE")

case "$bundle_name" in
  template-backup-*.tar.gz) ;;
  *)
    echo "Unexpected bundle name: $bundle_name" >&2
    echo "Expected template-backup-<RUN_ID>.tar.gz" >&2
    exit 1
    ;;
esac

RUN_ID=${bundle_name#template-backup-}
RUN_ID=${RUN_ID%.tar.gz}

echo "=== Unpacking bundle $bundle_name (set: $RUN_ID) ==="
mkdir -p backups/postgres
while IFS= read -r member; do
  case "$member" in
    /*|../*|*/../*|*/..) echo "Unsafe archive member: $member" >&2; exit 1 ;;
  esac
done < <(tar -tzf "$BUNDLE")
if tar -tvzf "$BUNDLE" | awk 'substr($1, 1, 1) ~ /^[lh]$/ { found=1 } END { exit !found }'; then
  echo "Archive contains symbolic or hard links" >&2
  exit 1
fi
tar -xzf "$BUNDLE" -C backups

echo ""
echo "=== Verifying unpacked set ==="
BACKUP_MANIFEST="backups/template-backup-${RUN_ID}.sha256" ENV_FILE="$ENV_FILE" \
  bash ./scripts/backup-verify.sh

echo ""
echo "=== Restoring set $RUN_ID ==="
CONFIRM_RESTORE=yes ENV_FILE="$ENV_FILE" bash ./scripts/db-restore-all.sh "$RUN_ID"
