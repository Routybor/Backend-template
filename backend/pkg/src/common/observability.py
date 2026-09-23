import inspect
import logging
from time import perf_counter

from prometheus_client import Counter, Histogram, generate_latest
from starlette.concurrency import run_in_threadpool
from starlette.responses import JSONResponse, Response

logger = logging.getLogger(__name__)

http_requests = Counter(
    "http_server_requests_total",
    "HTTP requests by method and status",
    ["method", "status"],
)
http_duration = Histogram(
    "http_server_request_duration_seconds",
    "HTTP request duration",
    ["method"],
)

CONTENT_TYPE_LATEST = "text/plain; version=0.0.4; charset=utf-8"


class PrometheusMiddleware:
    def __init__(self, app) -> None:
        self.app = app

    async def __call__(self, scope, receive, send) -> None:
        if scope["type"] != "http":
            await self.app(scope, receive, send)
            return

        status = 500
        start = perf_counter()

        async def send_wrapper(message) -> None:
            nonlocal status
            if message["type"] == "http.response.start":
                status = message["status"]
            await send(message)

        try:
            await self.app(scope, receive, send_wrapper)
        finally:
            method = scope.get("method", "")
            http_requests.labels(method=method, status=str(status)).inc()
            http_duration.labels(method=method).observe(perf_counter() - start)


def metrics_response() -> Response:
    return Response(generate_latest(), media_type=CONTENT_TYPE_LATEST)


async def health_response(checks: dict) -> JSONResponse:
    results = {}
    healthy = True

    for name, check in checks.items():
        try:
            if inspect.iscoroutinefunction(check):
                await check()
            else:
                await run_in_threadpool(check)
            results[name] = "up"
        except Exception as exc:
            results[name] = "down"
            healthy = False
            logger.error("health check failed", extra={"dependency": name}, exc_info=exc)

    status = 200 if healthy else 503
    return JSONResponse(
        {"status": "healthy" if healthy else "unhealthy", "checks": results},
        status_code=status,
    )
