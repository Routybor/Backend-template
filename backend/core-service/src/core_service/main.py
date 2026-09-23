import asyncio
import logging

import grpc
import uvicorn
from fastapi import FastAPI, HTTPException
from starlette.responses import JSONResponse

from common import grpcx
from common.observability import PrometheusMiddleware, health_response, metrics_response
from common.pb import item_pb2_grpc
from core_service.config import CoreConfig
from core_service.data import MemoryItemStore
from core_service.infrastructure import build_router
from core_service.infrastructure.grpc_handler import GrpcItemHandler
from core_service.service import ItemService

logging.basicConfig(level=logging.INFO, format="%(asctime)s %(levelname)s %(name)s %(message)s")
logger = logging.getLogger(__name__)


def create_http_app(svc: ItemService) -> FastAPI:
    app = FastAPI(
        title="core-service",
        docs_url=None,
        redoc_url=None,
        openapi_url=None,
    )
    app.add_middleware(PrometheusMiddleware)

    @app.exception_handler(HTTPException)
    async def http_exception_handler(request, exc: HTTPException):
        detail = exc.detail
        body = detail if isinstance(detail, dict) else {"error": str(detail)}
        return JSONResponse(body, status_code=exc.status_code, headers=getattr(exc, "headers", None))

    checks = {"store": svc.ping}

    @app.get("/health")
    async def health():
        return await health_response(checks)

    @app.get("/health/live")
    async def live():
        return {"status": "healthy"}

    @app.get("/health/ready")
    async def ready():
        return await health_response(checks)

    @app.get("/metrics")
    async def metrics():
        return metrics_response()

    app.include_router(build_router(svc))
    return app


async def serve() -> None:
    cfg = CoreConfig()

    svc = ItemService(MemoryItemStore())
    http_app = create_http_app(svc)

    grpc_server = grpcx.new_server()
    item_pb2_grpc.add_ItemServiceServicer_to_server(GrpcItemHandler(svc), grpc_server)
    grpc_server.add_insecure_port(f"[::]:{cfg.grpc_port}")
    await grpc_server.start()
    logger.info("grpc server listening", extra={"port": cfg.grpc_port})

    http_config = uvicorn.Config(
        http_app,
        host="0.0.0.0",
        port=cfg.port,
        log_level="info",
        access_log=False,
    )
    server = uvicorn.Server(http_config)

    try:
        await server.serve()
    finally:
        logger.info("shutting down core-service")
        await grpcx.stop(grpc_server, grace=15.0)
        logger.info("core-service stopped")


def main() -> None:
    asyncio.run(serve())


if __name__ == "__main__":
    main()
