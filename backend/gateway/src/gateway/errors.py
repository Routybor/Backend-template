import grpc
from fastapi import HTTPException

UNMAPPED_CODES = frozenset(
    {
        grpc.StatusCode.UNKNOWN,
        grpc.StatusCode.INTERNAL,
        grpc.StatusCode.DATA_LOSS,
        grpc.StatusCode.UNIMPLEMENTED,
        grpc.StatusCode.PERMISSION_DENIED,
    }
)


def map_grpc_error(exc: Exception) -> HTTPException:
    code = getattr(exc, "code", lambda: None)()
    details = getattr(exc, "details", lambda: "")()

    if code == grpc.StatusCode.NOT_FOUND:
        return HTTPException(404, {"error": "not_found", "message": details})
    if code == grpc.StatusCode.INVALID_ARGUMENT:
        return HTTPException(400, {"error": "invalid_request", "message": details})
    if code == grpc.StatusCode.UNAVAILABLE:
        return HTTPException(502, {"error": "service_unavailable"})
    if code == grpc.StatusCode.DEADLINE_EXCEEDED:
        return HTTPException(504, {"error": "gateway_timeout"})
    return HTTPException(500, {"error": "internal_error"})
