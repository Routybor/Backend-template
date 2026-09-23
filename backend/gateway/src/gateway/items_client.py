import grpc

from common import grpcx
from common.pb import item_pb2, item_pb2_grpc

CALL_TIMEOUT = 10.0


class ItemsClient:
    def __init__(self, addr: str) -> None:
        self._channel = grpcx.client_channel(addr)
        self._stub = item_pb2_grpc.ItemServiceStub(self._channel)

    async def get_all(self):
        return await self._stub.GetAll(
            item_pb2.GetAllRequest(),
            timeout=CALL_TIMEOUT,
            metadata=grpcx.outgoing_metadata(),
        )

    async def get_by_id(self, item_id: str):
        return await self._stub.GetByID(
            item_pb2.GetByIDRequest(id=item_id),
            timeout=CALL_TIMEOUT,
            metadata=grpcx.outgoing_metadata(),
        )

    async def create(self, name: str, description: str):
        return await self._stub.Create(
            item_pb2.CreateItemRequest(name=name, description=description),
            timeout=CALL_TIMEOUT,
            metadata=grpcx.outgoing_metadata(),
        )

    async def delete(self, item_id: str):
        return await self._stub.Delete(
            item_pb2.DeleteRequest(id=item_id),
            timeout=CALL_TIMEOUT,
            metadata=grpcx.outgoing_metadata(),
        )

    async def close(self) -> None:
        await self._channel.close()
