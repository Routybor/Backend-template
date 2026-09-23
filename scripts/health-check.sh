#!/usr/bin/env bash
set -Eeuo pipefail

ROOT_DIR=$(CDPATH= cd -- "$(dirname "$0")/.." && pwd)
cd "$ROOT_DIR"

ENV_FILE="${1:-.env}"
[[ -f "$ENV_FILE" ]] || { echo "Missing environment file: $ENV_FILE" >&2; exit 1; }
COMPOSE=(docker compose --env-file "$ENV_FILE" -f docker-compose.yml)

read_env() {
  awk -v key="$1" 'index($0, key "=") == 1 { print substr($0, length(key) + 2); exit }' "$ENV_FILE"
}

failed=false
services=(
  keycloak-db
  keycloak
  core-service
  gateway
)

echo "=== Template containers ==="
for service in "${services[@]}"; do
  container=$("${COMPOSE[@]}" ps -q "$service")
  if [[ -z "$container" ]]; then
    echo "MISSING  $service"
    failed=true
    continue
  fi

  status=$(docker inspect -f '{{.State.Status}}' "$container")
  health=$(docker inspect -f '{{if .State.Health}}{{.State.Health.Status}}{{else}}none{{end}}' "$container")
  if [[ "$status" != running || "$health" == unhealthy ]]; then
    echo "FAILED   $service (status=$status, health=$health)"
    failed=true
  else
    echo "READY    $service (health=$health)"
  fi
done

check_url() {
  local name=$1 url=$2
  if curl -fsS --max-time 5 "$url" >/dev/null; then
    echo "READY    $name ($url)"
  else
    echo "FAILED   $name ($url)"
    failed=true
  fi
}

echo "=== Public endpoints ==="
check_url "Gateway" "http://localhost:$(read_env GATEWAY_PORT)/health"
check_url "Core Service" "http://localhost:$(read_env CORE_SERVICE_PORT)/health"
check_url "Keycloak" "http://localhost:${KEYCLOAK_HTTP_PORT:-8180}/realms/$(read_env KEYCLOAK_REALM)"

[[ "$failed" == false ]] || exit 1
echo "Template health check passed"
