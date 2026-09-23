# Backend Template

Multi-variant backend template. Each framework variation lives in its own
branch, fully self-contained (own services, docker-compose, k8s manifests).

## Variants

| Branch | Stack | Status |
|--------|-------|--------|
| [`variant/go`](../../tree/variant/go) | Go + Gin gateway + gRPC core service + Keycloak + Docker/K8s | ready |
| [`variant/django`](../../tree/variant/django) | Django + Keycloak + Docker/K8s | wip |
| [`variant/fastapi`](../../tree/variant/fastapi) | FastAPI + Keycloak + Docker/K8s | wip |

## Usage

Clone a specific variant:

```bash
git clone -b variant/go <repo-url> backend-go
```

Or check out several variants side by side using worktrees:

```bash
git worktree add ../backend-go variant/go
git worktree add ../backend-django variant/django
git worktree add ../backend-fastapi variant/fastapi
```

## Contributing

- Shared fixes (Keycloak realm config, deploy scripts) should be cherry-picked
  across variant branches.
- Tag releases per variant: `go-v1.0.0`, `django-v0.1.0`, etc.
