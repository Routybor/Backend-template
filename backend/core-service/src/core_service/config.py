from pydantic import Field

from common.config import EnvSettings


class CoreConfig(EnvSettings):
    port: int = Field(8081, gt=0)
    grpc_port: int = Field(9091, gt=0)
