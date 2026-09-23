#!/usr/bin/env bash
set -euo pipefail

ENV_FILE="${1:-.env}"
[[ -f "$ENV_FILE" ]] || { echo "Missing $ENV_FILE. Run 'make init'."; exit 1; }

required=(
  GRPC_PORT
  GATEWAY_PORT
  CORE_SERVICE_PORT
  KEYCLOAK_ADMIN
  KEYCLOAK_ADMIN_PASSWORD
  KEYCLOAK_URL
  KC_DB
  KEYCLOAK_REALM
  KEYCLOAK_CLIENT_ID
  KEYCLOAK_CLIENT_SECRET
  KEYCLOAK_DB_USER
  KEYCLOAK_DB_PASSWORD
  KEYCLOAK_DB_NAME
  BACKUP_RETENTION_COUNT
)
missing=()
for key in "${required[@]}"; do
  grep -Eq "^${key}=" "$ENV_FILE" || missing+=("$key")
done
if (( ${#missing[@]} > 0 )); then
  echo "Missing required env keys in $ENV_FILE: ${missing[*]}"
  exit 1
fi

if grep -Eq '^[A-Za-z_][A-Za-z0-9_]*=[^#]*[[:space:]][^#]*$' "$ENV_FILE"; then
  echo "Unquoted whitespace found in $ENV_FILE. Values with spaces must be quoted."
  exit 1
fi

read_env_value() {
  awk -v key="$1" 'index($0, key "=") == 1 { print substr($0, length(key) + 2); exit }' "$ENV_FILE"
}

if [[ $(read_env_value KC_DB) != "postgres" ]]; then
  echo "KC_DB must be postgres: the backup tooling requires a PostgreSQL-backed Keycloak."
  exit 1
fi

echo "Environment is valid: $ENV_FILE"
