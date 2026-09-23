# UltimateTemplate

Go microservices + api gateway + keycloak template

## Stack

- **Gateway**: Go + Gin (JWT auth, gzip, rate limiting, circuit breaker)
- **Core Service**: Go + Gin + gRPC (CRUD API)
- **Communication**: gRPC between gateway and services
- **Auth**: Keycloak 26.x (OIDC), backed by PostgreSQL 17
- **Orchestration**: Docker Compose / Kubernetes

## Quick Start

### Docker Compose

```bash
make init    # creates .env from .env.example
make up      # builds (first time) and starts the stack
```

Access:
- Gateway: http://localhost:8080
- Keycloak: http://localhost:8180

### Kubernetes

```bash
# Configure
cp k8s/.env.example k8s/.env

# Deploy
.\scripts\deploy-k8s.ps1
```

## Make Commands

Run `make help` for the full list. Highlights:

| Command | Description |
|---------|-------------|
| `make init` | Create `.env` from `.env.example` |
| `make up` / `make down` / `make destroy` | Start / stop / stop + remove volumes |
| `make logs [SERVICE=gateway]` | Tail logs |
| `make health` | Check containers and public endpoints |
| `make test` | Run Go tests (gateway + core-service) |
| `make check` | Compose config validation + tests |
| `make k8s-deploy` / `make k8s-status` / `make k8s-undeploy` | Kubernetes lifecycle |

## Backups

Keycloak data lives in PostgreSQL (`keycloak-db`) and is backed up with
`pg_dump` into timestamped, checksummed sets under `backups/` (gitignored).

```bash
make backup                # dump keycloak DB + write sha256 manifest
make verify-backup         # verify checksums + archive structure
make verify-backup-full    # additionally restore into an isolated throwaway DB
make restore-db-all        # pick a backup date and restore (prompts, or CONFIRM_RESTORE=yes)
make bundle                # pack newest set into one portable .tar.gz
make restore-bundle BUNDLE=backups/template-backup-<id>.tar.gz CONFIRM_RESTORE=yes
```

Safety properties:
- Every restore requires `CONFIRM_RESTORE=yes` (or an interactive `yes`).
- A pre-restore snapshot of the current DB is taken automatically before any restore.
- Failed backup runs are cleaned up; retention prunes old sets (`BACKUP_RETENTION_COUNT`, default 30).

Note: the k8s manifests still run Keycloak with `dev-file` storage by default;
point `KEYCLOAK_DB` in `k8s/.env` at a PostgreSQL instance for parity with compose.

## Endpoints (via Gateway, requires JWT)

| Method | Path | Description |
|--------|------|-------------|
| GET | /health | Health check |
| GET | /items | List items |
| POST | /items | Create item |
| GET | /items/:id | Get item |
| DELETE | /items/:id | Delete item |

## Direct Access (no auth)

| Service | Port | Protocol |
|---------|------|----------|
| Gateway | 8080 | HTTP |
| Core Service | 8081 | HTTP |
| Core Service | 9091 | gRPC |

## Authentication

Get token:
```bash
curl -X POST http://localhost:8180/realms/microservices/protocol/openid-connect/token \
  -d "client_id=gateway&client_secret=gateway-secret&grant_type=client_credentials"
```

Use token:
```bash
curl -H "Authorization: Bearer <token>" http://localhost:8080/items
```

Default credentials:
- Client: `gateway` / `gateway-secret`
- User: `testuser` / `testuser`

## Environment Variables

### Docker Compose
- `.env`

### Kubernetes
- `k8s/.env`

See `k8s/ENV.md` for full variable reference.

## Scripts

```bash
# Kubernetes
.\scripts\deploy-k8s.ps1   # Deploy
.\scripts\undeploy-k8s.ps1 # Remove
.\scripts\status-k8s.ps1   # Check status
```

## Project Structure

```
.
├── Makefile                 # make init/up/backup/... command surface
├── docker-compose.yml
├── backend/
│   ├── gateway/              # API Gateway
│   │   ├── cmd/gateway/main.go
│   │   ├── internal/
│   │   │   ├── config/
│   │   │   ├── dto/
│   │   │   ├── handler/
│   │   │   ├── middleware/
│   │   │   └── service/
│   │   ├── go.mod
│   │   └── Dockerfile
│   └── core-service/         # CRUD Service
│       ├── cmd/core-service/main.go
│       ├── internal/
│       │   ├── config/
│       │   ├── dto/
│       │   ├── handler/
│       │   ├── repository/
│       │   └── service/
│       ├── go.mod
│       └── Dockerfile
├── k8s/                     # Kubernetes manifests
│   ├── base/                # Kustomize base
│   └── helm/                # Helm chart
├── scripts/                 # Deploy, health-check and backup scripts
├── backups/                 # Backup output (gitignored)
└── keycloak/                # Keycloak config
```

## Ports

| Service | Port | Protocol |
|---------|------|----------|
| Gateway | 8080 | HTTP |
| Keycloak | 8080 | HTTP |
| Core Service (HTTP) | 8081 | HTTP |
| Core Service (gRPC) | 9091 | gRPC |
