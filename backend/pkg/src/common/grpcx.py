import logging
from time import perf_counter

import grpc
from prometheus_client import Counter, Histogram

from common import requestid

logger = logging.getLogger(__name__)

server_requests = Counter(
    "grpc_server_requests_total",
    "gRPC server requests by method and code",
    ["method", "code"],
)
server_duration = Histogram(
    "grpc_server_request_duration_seconds",
    "gRPC server request duration",
    ["method"],
)


def client_channel(addr: str) -> grpc.aio.Channel:
    return grpc.aio.insecure_channel(addr)


def outgoing_metadata() -> tuple[tuple[str, str], ...]:
    rid = requestid.get()
    if rid:
        return ((requestid.HEADER.lower(), rid),)
    return ()


def new_server() -> grpc.aio.Server:
    return grpc.aio.server(interceptors=[ObservabilityInterceptor()])


async def stop(server: grpc.aio.Server, grace: float) -> None:
    await server.stop(grace=grace)


class ObservabilityInterceptor(grpc.aio.ServerInterceptor):
    async def intercept_service(self, continuation, handler_call_details):
        handler = await continuation(handler_call_details)
        if handler is None or handler.unary_unary is None:
            return handler

        behavior = handler.unary_unary
        method = handler_call_details.method

        async def wrapped(request, context):
            for key, value in context.invocation_metadata() or ():
                if key.lower() == requestid.HEADER.lower() and value:
                    requestid.set_id(value)
                    break

            start = perf_counter()
            try:
                response = await behavior(request, context)
            except Exception as exc:
                code = _status_code(exc)
                if code == grpc.StatusCode.UNKNOWN:
                    logger.exception(
                        "unhandled error in grpc handler",
                        extra={"method": method, "request_id": requestid.get()},
                    )
                _observe(method, code, perf_counter() - start)
                raise

            _observe(method, grpc.StatusCode.OK, perf_counter() - start)
            return response

        return grpc.unary_unary_rpc_method_handler(
            wrapped,
            request_deserializer=handler.request_deserializer,
            response_serializer=handler.response_serializer,
        )


def _status_code(exc: Exception) -> grpc.StatusCode:
    code = getattr(exc, "code", None)
    return code() if callable(code) else grpc.StatusCode.UNKNOWN


def _observe(method: str, code: grpc.StatusCode, duration: float) -> None:
    server_requests.labels(method=method, code=code.name).inc()
    server_duration.labels(method=method).observe(duration)
