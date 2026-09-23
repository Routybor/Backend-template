import secrets
from contextvars import ContextVar

HEADER = "X-Request-ID"

_request_id: ContextVar[str] = ContextVar("request_id", default="")


def new_id() -> str:
    return secrets.token_hex(8)


def get() -> str:
    return _request_id.get()


def set_id(value: str) -> None:
    _request_id.set(value)
