from gateway.errors import map_grpc_error
from gateway.middleware import CircuitBreaker, RateLimiter


class FakeRpcError(Exception):
    def __init__(self, code, details=""):
        super().__init__(details)
        self._code = code
        self._details = details

    def code(self):
        return self._code

    def details(self):
        return self._details


def test_grpc_error_mapping():
    import grpc

    cases = {
        grpc.StatusCode.NOT_FOUND: 404,
        grpc.StatusCode.INVALID_ARGUMENT: 400,
        grpc.StatusCode.UNAVAILABLE: 502,
        grpc.StatusCode.DEADLINE_EXCEEDED: 504,
        grpc.StatusCode.INTERNAL: 500,
    }
    for code, expected in cases.items():
        exc = map_grpc_error(FakeRpcError(code, "boom"))
        assert exc.status_code == expected


def test_rate_limiter_window():
    limiter = RateLimiter(limit=2, window=60.0, sweep_interval=60.0)
    assert limiter.allow("1.2.3.4")
    assert limiter.allow("1.2.3.4")
    assert not limiter.allow("1.2.3.4")
    assert limiter.allow("5.6.7.8")


def test_circuit_breaker_opens_and_recovers():
    breaker = CircuitBreaker(threshold=2, timeout=60.0)
    breaker.record_failure()
    breaker.record_failure()
    assert not breaker.allow()

    breaker._last_failure -= 120.0
    assert breaker.allow()
    breaker.record_success()
    assert breaker.allow()
