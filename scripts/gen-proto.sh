#!/usr/bin/env bash
set -euo pipefail

cd "$(dirname "$0")/.."
cd backend

GRPCIO_TOOLS_VERSION=${GRPCIO_TOOLS_VERSION:-1.84.0}

docker run --rm -v "$PWD":/src -w /src python:3.14-slim sh -c "
  set -eu
  pip install --no-cache-dir --quiet grpcio-tools==$GRPCIO_TOOLS_VERSION
  mkdir -p pkg/src/common/pb
  python -m grpc_tools.protoc -I proto \
    --python_out=pkg/src/common/pb \
    --pyi_out=pkg/src/common/pb \
    --grpc_python_out=pkg/src/common/pb \
    proto/*.proto
  for f in pkg/src/common/pb/*_pb2_grpc.py; do
    sed -i -E 's/^import ([a-z0-9_]+_pb2) as /from . import \1 as /' \"\$f\"
  done
"

echo "proto generated into backend/pkg/src/common/pb"
