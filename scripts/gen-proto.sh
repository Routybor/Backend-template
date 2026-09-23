#!/usr/bin/env bash
set -euo pipefail

cd "$(dirname "$0")/.."
cd backend

docker run --rm -v "$PWD":/src -w /src golang:1.26-alpine sh -c '
  set -eu
  apk add --no-cache protoc
  go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.36.10
  go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.5.1
  mkdir -p pkg/pb
  protoc -I proto \
    --go_out=pkg/pb --go_opt=paths=source_relative \
    --go-grpc_out=pkg/pb --go-grpc_opt=paths=source_relative \
    proto/*.proto
'

echo "proto generated into backend/pkg/pb"
