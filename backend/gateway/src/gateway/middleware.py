import time
from collections import deque

from starlette.types import ASGIApp, Receive, Scope, Send

from common import requestid

SECURITY_HEADERS = (
    ("X-Content-Type-Options", "nosniff"),
    ("X-Frame-Options", "DENY"),
    ("X-XSS-Protection", "1; mode=block"),
    ("Strict-Transport-Security", "max-age=31536000; includeSubDomains"),
    ("Cache-Control", "no-store"),
)


class RequestIDMiddleware:
    def __init__(self, app: ASGIApp) -> None:
        self.app = app

    async def __call__(self, scope: Scope, receive: Receive, send: Send) -> None:
        if scope["type"] != "http":
            await self.app(scope, receive, send)
            return

        headers = dict(scope.get("headers") or [])
        rid = headers.get(b"x-request-id", b"").decode() or requestid.new_id()
        requestid.set_id(rid)

        async def send_wrapper(message) -> None:
            if message["type"] == "http.response.start":
                message.setdefault("headers", []).append(
                    (requestid.HEADER.lower().encode(), rid.encode())
                )
            await send(message)

        await self.app(scope, receive, send_wrapper)


class SecurityHeadersMiddleware:
    def __init__(self, app: ASGIApp) -> None:
        self.app = app

    async def __call__(self, scope: Scope, receive: Receive, send: Send) -> None:
        if scope["type"] != "http":
            await self.app(scope, receive, send)
            return

        async def send_wrapper(message) -> None:
            if message["type"] == "http.response.start":
                message.setdefault("headers", []).extend(
                    [(k.lower().encode(), v.encode()) for k, v in SECURITY_HEADERS]
                )
            await send(message)

        await self.app(scope, receive, send_wrapper)


class RateLimiter:
    def __init__(self, limit: int, window: float, sweep_interval: float = 60.0) -> None:
        self.limit = limit
        self.window = window
        self._hits: dict[str, deque] = {}
        self._sweep_interval = sweep_interval
        self._last_sweep = time.monotonic()

    def allow(self, key: str) -> bool:
        self._sweep()
        hits = self._hits.get(key)
        if hits is None:
            hits = self._hits[key] = deque()
        now = time.monotonic()
        while hits and now - hits[0] > self.window:
            hits.popleft()
        if len(hits) >= self.limit:
            return False
        hits.append(now)
        return True

    def _sweep(self) -> None:
        now = time.monotonic()
        if now - self._last_sweep < self._sweep_interval:
            return
        self._last_sweep = now
        for key, hits in list(self._hits.items()):
            if not hits or now - hits[-1] > self.window:
                del self._hits[key]


class RateLimitMiddleware:
    def __init__(self, app: ASGIApp, limit: int, window_sec: int) -> None:
        self.app = app
        self.limiter = RateLimiter(limit, float(window_sec))

    async def __call__(self, scope: Scope, receive: Receive, send: Send) -> None:
        if scope["type"] != "http":
            await self.app(scope, receive, send)
            return

        client = scope.get("client")
        ip = client[0] if client else ""
        if not self.limiter.allow(ip):
            await _abort(
                send, 429, [(b"content-type", b"application/json")],
                b'{"error":"rate limit exceeded"}',
            )
            return

        await self.app(scope, receive, send)


async def _abort(send, status: int, headers, body: bytes) -> None:
    await send({"type": "http.response.start", "status": status, "headers": headers})
    await send({"type": "http.response.body", "body": body})


STATE_CLOSED, STATE_OPEN, STATE_HALF_OPEN = 0, 1, 2


class CircuitBreaker:
    def __init__(self, threshold: int, timeout: float) -> None:
        self.threshold = threshold
        self.timeout = timeout
        self._state = STATE_CLOSED
        self._failures = 0
        self._last_failure = 0.0

    def allow(self) -> bool:
        if self._state != STATE_OPEN:
            return True
        if time.monotonic() - self._last_failure > self.timeout:
            self._state = STATE_HALF_OPEN
            return True
        return False

    def record_success(self) -> None:
        self._state = STATE_CLOSED
        self._failures = 0

    def record_failure(self) -> None:
        self._last_failure = time.monotonic()
        self._failures += 1
        if self._failures >= self.threshold:
            self._state = STATE_OPEN
