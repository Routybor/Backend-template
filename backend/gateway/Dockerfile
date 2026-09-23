FROM python:3.14-slim AS builder

COPY --from=ghcr.io/astral-sh/uv:latest /uv /uvx /bin/

ENV UV_COMPILE_BYTECODE=1 UV_LINK_MODE=copy

WORKDIR /app

COPY pyproject.toml uv.lock ./
COPY pkg/ pkg/
COPY gateway/ gateway/
COPY core-service/pyproject.toml core-service/pyproject.toml

RUN uv sync --frozen --no-dev --no-editable --package gateway

FROM python:3.14-slim

WORKDIR /app

COPY --from=builder /app/.venv /app/.venv

ENV PATH="/app/.venv/bin:$PATH"

RUN addgroup --system app && adduser --system app --ingroup app

USER app

EXPOSE 8080

ENV GATEWAY_PORT=8080
ENV CORE_SERVICE_GRPC=core-service:9091
ENV KEYCLOAK_URL=http://keycloak:8080
ENV KEYCLOAK_REALM=microservices

HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
    CMD python -c "import urllib.request; urllib.request.urlopen('http://localhost:8080/health')" || exit 1

ENTRYPOINT ["python", "-m", "gateway.main"]
