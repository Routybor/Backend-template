# Adding a Service

The fastest path is the generator:

```bash
make service NAME=orders
```

It stamps `backend/orders-service/` from `templates/service/`, adds the proto
to `backend/proto/`, regenerates stubs into `backend/pkg/pb`, registers the
module in `backend/go.work`, and verifies the build.

## Generated layout

```
backend/<name>-service/
├── Dockerfile
├── go.mod
├── cmd/<name>/main.go
└── internal/
    ├── domain/          entity, errors, store port
    ├── data/            in-memory store implementation
    ├── service/         business logic on ports only
    └── infrastructure/  gRPC handler adapter, config, health router
```

## Manual checklist

1. **Proto contract** — add `backend/proto/<name>.proto`, run `make proto`.
2. **Module** — `backend/<name>-service/go.mod` (module `<name>-service`), then
   `go work use ./<name>-service` in `backend/`.
3. **Dockerfile** — copy the pattern: `backend/` context, generated minimal
   `go.work` with only `./pkg` + the service.
4. **Compose** — the generator prints a ready snippet; ports default to
   8082/9092, pick free ones for further services.
5. **Gateway wiring** — add a client in `backend/gateway/internal/service/`
   (see `items_client.go`), map errors in `handlers/errors.go` if needed, and
   register routes in `gateway/internal/app/routes.go`.
6. **K8s** — copy `k8s/base/core-service-{deployment,service}.yaml`, adjust
   names, ports, images, and add to `kustomization.yaml`.
7. **Tests** — service-layer tests with a fake store (see
   `core-service/internal/service/item_service_test.go`).

## Rules

- `backend/pkg` must never import service code.
- Services never import each other — contracts live in `backend/proto`.
- No global state: clients and stores are constructor-injected in `cmd/`.
- Domain errors are sentinels (`domain.ErrNotFound`, ...) mapped to gRPC
  status codes in `infrastructure`, and to HTTP in the gateway.
