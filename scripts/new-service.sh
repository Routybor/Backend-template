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
PACKAGE=${MODULE//-/_}
PASCAL=$(echo "$NAME" | sed -E 's/(^|-)([a-z])/\U\2/g')
DEST="backend/$MODULE"

if [[ -e "$DEST" ]]; then
  echo "Service already exists: $DEST" >&2
  exit 1
fi

echo "=== Generating $MODULE ==="

PKG_DIR="$DEST/src/$PACKAGE"
mkdir -p "$PKG_DIR/internal/domain" "$PKG_DIR/internal/data" \
  "$PKG_DIR/internal/service" "$PKG_DIR/internal/infrastructure"

render() {
  sed -e "s/__NAME__/$NAME/g" \
      -e "s/__MODULE__/$MODULE/g" \
      -e "s/__PACKAGE__/$PACKAGE/g" \
      -e "s/__PASCAL__/$PASCAL/g" \
      "$1" > "$2"
}

render templates/service/pyproject.toml.tmpl "$DEST/pyproject.toml"
render templates/service/Dockerfile.tmpl "$DEST/Dockerfile"
render templates/service/src/__PACKAGE__/main.py.tmpl "$PKG_DIR/main.py"
render templates/service/src/__PACKAGE__/__init__.py.tmpl "$PKG_DIR/__init__.py"
render templates/service/src/__PACKAGE__/internal/config.py.tmpl "$PKG_DIR/internal/config.py"
render templates/service/src/__PACKAGE__/internal/domain/greeting.py.tmpl "$PKG_DIR/internal/domain/greeting.py"
render templates/service/src/__PACKAGE__/internal/domain/errors.py.tmpl "$PKG_DIR/internal/domain/errors.py"
render templates/service/src/__PACKAGE__/internal/domain/__init__.py.tmpl "$PKG_DIR/internal/domain/__init__.py"
render templates/service/src/__PACKAGE__/internal/data/memory.py.tmpl "$PKG_DIR/internal/data/memory.py"
render templates/service/src/__PACKAGE__/internal/data/__init__.py.tmpl "$PKG_DIR/internal/data/__init__.py"
render templates/service/src/__PACKAGE__/internal/service/greeting_service.py.tmpl "$PKG_DIR/internal/service/greeting_service.py"
render templates/service/src/__PACKAGE__/internal/service/__init__.py.tmpl "$PKG_DIR/internal/service/__init__.py"
render templates/service/src/__PACKAGE__/internal/infrastructure/grpc_handler.py.tmpl "$PKG_DIR/internal/infrastructure/grpc_handler.py"
render templates/service/src/__PACKAGE__/internal/infrastructure/http_router.py.tmpl "$PKG_DIR/internal/infrastructure/http_router.py"
render templates/service/src/__PACKAGE__/internal/infrastructure/__init__.py.tmpl "$PKG_DIR/internal/infrastructure/__init__.py"

render templates/service/proto.tmpl "backend/proto/$NAME.proto"

echo "=== Regenerating proto stubs ==="
bash scripts/gen-proto.sh

echo "=== Registering module ==="
sed -i "s/^members = \[\(.*\)\]/members = [\1, \"$MODULE\"]/" backend/pyproject.toml
(cd backend && uv sync --all-packages >/dev/null)
(cd "backend/$MODULE" && uv run --frozen python -c "import $PACKAGE.main" && echo "$MODULE imports ok")

echo ""
echo "$MODULE created and imports."
echo ""
echo "Next steps:"
echo "  1. add to docker-compose.yml:"
echo ""
sed -e "s/__NAME__/$NAME/g" -e "s/__MODULE__/$MODULE/g" -e "s/__PASCAL__/$PASCAL/g" templates/service/compose.snippet.tmpl | sed 's/^/     /'
echo ""
echo "  2. wire a client in backend/gateway/src/gateway/ (see items_client.py)"
echo "     and register routes in routes.py"
echo "  3. see docs/ADDING_A_SERVICE.md for k8s manifests and the full checklist"
