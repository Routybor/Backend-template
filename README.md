# UltimateTemplate

Go microservices + api gateway + keycloak template

## Stack

- **Gateway**: Go + Gin (JWT auth, gzip, rate limiting, circuit breaker)
- **Core Service**: Go + Gin (CRUD API)
- **Auth**: Keycloak 26.x (OIDC)
- **Orchestration**: Docker Compose / Kubernetes

## Quick Start

### Docker Compose

```bash
docker-compose up --build -d
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

## Endpoints

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| GET | /health | No | Health check |
| GET | /items | JWT | List items |
| POST | /items | JWT | Create item |
| GET | /items/:id | JWT | Get item |
| PUT | /items/:id | JWT | Update item |
| DELETE | /items/:id | JWT | Delete item |

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
├── scripts/                 # Deploy scripts
└── keycloak/                # Keycloak config
```

## Ports

| Service | Internal | External |
|---------|----------|----------|
| Gateway | 8080 | 8080 |
| Keycloak | 8080 | 8180 |
| Core Service | 8081 | - |
