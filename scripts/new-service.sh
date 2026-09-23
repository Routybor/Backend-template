#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR=$(cd "$(dirname "$0")/.." && pwd)
cd "$ROOT_DIR"

NAME=${1:-}
if [[ -z "$NAME" ]]; then
  echo "Usage: $0 <name>   # e.g. orders (lowercase, dashes allowed)" >&2
  exit 1
fi

if ! [[ "$NAME" =~ ^[a-z][a-z0-9-]*$ ]]; then
  echo "Invalid service name: $NAME (lowercase letters, digits and dashes only)" >&2
  exit 1
fi

MODULE="${NAME}-service"
PASCAL=$(echo "$NAME" | sed -E 's/(^|-)([a-z])/\U\2/g')
DEST="backend/$MODULE"

if [[ -e "$DEST" ]]; then
  echo "Service already exists: $DEST" >&2
  exit 1
fi

echo "=== Generating $MODULE ==="

mkdir -p "$DEST/cmd/$NAME" "$DEST/internal/domain" "$DEST/internal/data" \
  "$DEST/internal/service" "$DEST/internal/infrastructure"

render() {
  sed -e "s/__NAME__/$NAME/g" \
      -e "s/__MODULE__/$MODULE/g" \
      -e "s/__PASCAL__/$PASCAL/g" \
      "$1" > "$2"
}

render templates/service/go.mod.tmpl "$DEST/go.mod"
render templates/service/Dockerfile.tmpl "$DEST/Dockerfile"
render templates/service/cmd/main.go.tmpl "$DEST/cmd/$NAME/main.go"
render templates/service/internal/domain/greeting.go.tmpl "$DEST/internal/domain/greeting.go"
render templates/service/internal/domain/errors.go.tmpl "$DEST/internal/domain/errors.go"
render templates/service/internal/domain/interfaces.go.tmpl "$DEST/internal/domain/interfaces.go"
render templates/service/internal/data/memory.go.tmpl "$DEST/internal/data/memory.go"
render templates/service/internal/service/greeting_service.go.tmpl "$DEST/internal/service/greeting_service.go"
render templates/service/internal/infrastructure/grpc_handler.go.tmpl "$DEST/internal/infrastructure/grpc_handler.go"
render templates/service/internal/infrastructure/http_router.go.tmpl "$DEST/internal/infrastructure/http_router.go"

render templates/service/proto.tmpl "backend/proto/$NAME.proto"

echo "=== Regenerating proto stubs ==="
bash scripts/gen-proto.sh

echo "=== Registering module ==="
(cd backend && go work use "./$MODULE")
(cd "backend/$MODULE" && go mod tidy && go build ./... && go vet ./...)

echo ""
echo "$MODULE created and builds."
echo ""
echo "Next steps:"
echo "  1. add to docker-compose.yml:"
echo ""
sed -e "s/__NAME__/$NAME/g" -e "s/__MODULE__/$MODULE/g" templates/service/compose.snippet.tmpl | sed 's/^/     /'
echo ""
echo "  2. wire an RPC client in backend/gateway/internal/service/ and register routes in app/routes.go"
echo "  3. see docs/ADDING_A_SERVICE.md for k8s manifests and the full checklist"
