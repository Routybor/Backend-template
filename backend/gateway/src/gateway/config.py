from pydantic import Field

from common.config import EnvSettings


class GatewayConfig(EnvSettings):
    gateway_port: int = Field(8080, gt=0)
    core_service_grpc: str = "localhost:9091"
    keycloak_url: str = "http://localhost:8180"
    keycloak_realm: str = "microservices"
    rate_limit_requests: int = Field(100, gt=0)
    rate_limit_window_sec: int = Field(60, gt=0)
    circuit_threshold: int = Field(5, gt=0)
    circuit_timeout_sec: int = Field(30, gt=0)
