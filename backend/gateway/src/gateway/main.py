import logging
from contextlib import asynccontextmanager

import uvicorn
from fastapi import FastAPI, HTTPException
from starlette.responses import JSONResponse
from starlette.middleware.gzip import GZipMiddleware

from common.auth import TokenValidator
from common.observability import PrometheusMiddleware, metrics_response
from gateway import routes
from gateway.config import GatewayConfig
from gateway.items_client import ItemsClient
from gateway.middleware import (
    CircuitBreaker,
    RateLimitMiddleware,
    RequestIDMiddleware,
    SecurityHeadersMiddleware,
)

logging.basicConfig(level=logging.INFO, format="%(asctime)s %(levelname)s %(name)s %(message)s")
logger = logging.getLogger(__name__)


def create_app(cfg: GatewayConfig) -> FastAPI:
    @asynccontextmanager
    async def lifespan(app: FastAPI):
        app.state.validator = TokenValidator(cfg.keycloak_url, cfg.keycloak_realm)
        app.state.items_client = ItemsClient(cfg.core_service_grpc)
        app.state.circuit_breaker = CircuitBreaker(
            cfg.circuit_threshold, float(cfg.circuit_timeout_sec)
        )
        logger.info("gateway ready", extra={"port": cfg.gateway_port})
        yield
        await app.state.items_client.close()
        logger.info("gateway stopped")

    app = FastAPI(
        title="gateway",
        lifespan=lifespan,
        docs_url=None,
        redoc_url=None,
        openapi_url=None,
    )

    app.add_middleware(GZipMiddleware, minimum_size=500)
    app.add_middleware(RateLimitMiddleware, limit=cfg.rate_limit_requests, window_sec=cfg.rate_limit_window_sec)
    app.add_middleware(PrometheusMiddleware)
    app.add_middleware(SecurityHeadersMiddleware)
    app.add_middleware(RequestIDMiddleware)

    @app.exception_handler(HTTPException)
    async def http_exception_handler(request, exc: HTTPException):
        detail = exc.detail
        body = detail if isinstance(detail, dict) else {"error": str(detail)}
        return JSONResponse(body, status_code=exc.status_code, headers=getattr(exc, "headers", None))

    @app.get("/health")
    async def health():
        return {"status": "healthy"}

    @app.get("/metrics")
    async def metrics():
        return metrics_response()

    app.include_router(routes.router)
    return app


def main() -> None:
    cfg = GatewayConfig()
    app = create_app(cfg)
    uvicorn.run(
        app,
        host="0.0.0.0",
        port=cfg.gateway_port,
        log_level="info",
        access_log=False,
    )


if __name__ == "__main__":
    main()
