import grpc

from common.pb import item_pb2, item_pb2_grpc
from core_service.domain import InvalidInputError, Item, NotFoundError
from core_service.service import ItemService


class GrpcItemHandler(item_pb2_grpc.ItemServiceServicer):
    def __init__(self, svc: ItemService) -> None:
        self._svc = svc

    async def GetAll(self, request, context):
        items = self._svc.get_all()
        return item_pb2.ListItemsResponse(
            items=[_to_pb(i) for i in items],
            total=len(items),
        )

    async def GetByID(self, request, context):
        try:
            item = self._svc.get(request.id)
        except NotFoundError:
            await context.abort(grpc.StatusCode.NOT_FOUND, "item not found")
        return _to_pb(item)

    async def Create(self, request, context):
        try:
            item = self._svc.create(request.name, request.description)
        except InvalidInputError as exc:
            await context.abort(grpc.StatusCode.INVALID_ARGUMENT, str(exc))
        return _to_pb(item)

    async def Delete(self, request, context):
        try:
            self._svc.delete(request.id)
        except NotFoundError:
            await context.abort(grpc.StatusCode.NOT_FOUND, "item not found")
        return item_pb2.DeleteResponse(success=True)

    async def Health(self, request, context):
        return item_pb2.HealthResponse(status="healthy")


def _to_pb(item: Item) -> item_pb2.Item:
    return item_pb2.Item(
        id=item.id,
        name=item.name,
        description=item.description,
        created_at=int(item.created_at.timestamp()),
        updated_at=int(item.updated_at.timestamp()),
    )
