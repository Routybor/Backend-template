FROM python:3.14-slim AS builder

COPY --from=ghcr.io/astral-sh/uv:latest /uv /uvx /bin/

ENV UV_COMPILE_BYTECODE=1 UV_LINK_MODE=copy

WORKDIR /app

COPY pyproject.toml uv.lock ./
COPY pkg/ pkg/
COPY core-service/ core-service/
COPY gateway/pyproject.toml gateway/pyproject.toml

RUN uv sync --frozen --no-dev --no-editable --package core-service

FROM python:3.14-slim

WORKDIR /app

COPY --from=builder /app/.venv /app/.venv

ENV PATH="/app/.venv/bin:$PATH"

RUN addgroup --system app && adduser --system app --ingroup app

USER app

EXPOSE 8081 9091

ENV PORT=8081
ENV GRPC_PORT=9091

HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
    CMD python -c "import urllib.request; urllib.request.urlopen('http://localhost:8081/health')" || exit 1

ENTRYPOINT ["python", "-m", "core_service.main"]
