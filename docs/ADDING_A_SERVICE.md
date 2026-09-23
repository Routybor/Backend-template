# Adding a Service

The fastest path is the generator:

```bash
make service NAME=orders
```

It stamps `backend/orders-service/` from `templates/service/`, adds the proto
to `backend/proto/`, regenerates stubs into `backend/pkg/src/common/pb`,
registers the module in the uv workspace (`backend/pyproject.toml` members),
syncs the lockfile, and verifies the imports.

## Generated layout

```
backend/<name>-service/
├── Dockerfile
├── pyproject.toml
└── src/<name>_service/
    ├── main.py                    grpc + http servers, graceful shutdown
    ├── config.py                  pydantic-settings, validated at startup
    └── internal/
        ├── domain/                entity, errors, store Protocol
        ├── data/                  in-memory store implementation
        ├── service/               business logic on ports only
        └── infrastructure/        gRPC handler adapter, health router
```

## Manual checklist

1. **Proto contract** — add `backend/proto/<name>.proto`, run `make proto`
   (imports: `from common.pb import <name>_pb2, <name>_pb2_grpc`).
2. **Module** — `backend/<name>-service/pyproject.toml` (name
   `<name>-service`, `backend-common = { workspace = true }` under
   `[tool.uv.sources]`), add the member to `backend/pyproject.toml`, then
   `uv sync --all-packages` in `backend/`.
3. **Dockerfile** — copy the pattern: `backend/` context, `uv sync --frozen
   --no-dev --package <module>`.
4. **Compose** — the generator prints a ready snippet; ports default to
   8082/9092, pick free ones for further services.
5. **Gateway wiring** — add a client in `backend/gateway/src/gateway/`
   (see `items_client.py`), map codes in `errors.py` if needed, and register
   routes in `routes.py`.
6. **K8s** — copy `k8s/base/core-service-{deployment,service}.yaml`, adjust
   names, ports, images, and add to `kustomization.yaml`.
7. **Tests** — service-layer tests with a fake store (see
   `core-service/tests/test_item_service.py`).

## Rules

- `backend/pkg` must never import service code; services never import each
  other — contracts only via `backend/proto`.
- No global state — clients and stores are constructor-injected in `main.py`.
- Domain errors are typed (`NotFoundError`, `InvalidInputError`), mapped to
  gRPC status codes in `infrastructure`, and to HTTP in the gateway.
- The shared `common` package stays framework-light: starlette-level only.
