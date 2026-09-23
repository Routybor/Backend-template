#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR=$(cd "$(dirname "$0")/.." && pwd)
cd "$ROOT_DIR"

# Bundle one complete backup set (manifest + all dumps it references) into a
# single portable archive: backups/template-backup-<RUN_ID>.tar.gz
#
# Usage:
#   scripts/db-bundle.sh            # bundle the newest backup set
#   scripts/db-bundle.sh <RUN_ID>   # bundle a specific set, e.g. 2026-09-23_12-00-00

RUN_ID="${1:-}"

if [[ -z "$RUN_ID" ]]; then
  shopt -s nullglob
  manifests=(backups/template-backup-*.sha256)

  if ((${#manifests[@]} == 0)); then
    echo "No backup manifests found in backups/." >&2
    echo "Run 'make backup' first." >&2
    exit 1
  fi

  # Pick the newest set that is NOT a pre-restore snapshot (those are near-empty
  # captures of the local DB taken automatically before a restore).
  while IFS= read -r manifest; do
    candidate=$(basename "$manifest" .sha256)
    candidate=${candidate#template-backup-}

    if [[ -e "backups/template-backup-${candidate}.prerestore" ]]; then
      continue
    fi

    RUN_ID="$candidate"
    break
  done < <(ls -1t "${manifests[@]}")

  if [[ -z "$RUN_ID" ]]; then
    echo "Only pre-restore snapshots found; pass an explicit RUN_ID." >&2
    echo "  make bundle RUN_ID=<date>" >&2
    exit 1
  fi
fi

MANIFEST="backups/template-backup-${RUN_ID}.sha256"

if [[ ! -s "$MANIFEST" ]]; then
  echo "Manifest not found or empty: $MANIFEST" >&2
  exit 1
fi

# The manifest lists every dump/archive that belongs to this set, relative to
# backups/. Bundle those plus the manifest itself.
bundle_members=("template-backup-${RUN_ID}.sha256")

while read -r _ relative_file; do
  [[ -n "$relative_file" ]] || continue
  bundle_members+=("$relative_file")
done <"$MANIFEST"

BUNDLE="template-backup-${RUN_ID}.tar.gz"
BUNDLE_TMP="${BUNDLE}.tmp"

(
  cd backups

  for member in "${bundle_members[@]}"; do
    if [[ ! -s "$member" ]]; then
      echo "Cannot bundle: missing or empty file backups/$member" >&2
      exit 1
    fi
  done

  rm -f "$BUNDLE_TMP"
  tar -czf "$BUNDLE_TMP" "${bundle_members[@]}"
  chmod 600 "$BUNDLE_TMP"
  mv "$BUNDLE_TMP" "$BUNDLE"
)

echo "Bundle created: backups/${BUNDLE}"
echo "Copy it anywhere and restore with:"
echo "  make restore-bundle BUNDLE=backups/${BUNDLE} CONFIRM_RESTORE=yes"
