from typing import Any, Awaitable, Callable

from fastapi import APIRouter, Depends, HTTPException, Request
from pydantic import BaseModel, Field

from gateway.auth import require_user
from gateway.errors import map_grpc_error

router = APIRouter(dependencies=[Depends(require_user)])


class CreateItemBody(BaseModel):
    name: str = Field(min_length=1, max_length=255)
    description: str = ""


async def with_breaker(request: Request, call: Callable[[], Awaitable[Any]]) -> Any:
    breaker = request.app.state.circuit_breaker
    if not breaker.allow():
        raise HTTPException(503, {"error": "service temporarily unavailable"})

    try:
        return await call()
    except HTTPException as exc:
        if exc.status_code >= 500:
            breaker.record_failure()
        else:
            breaker.record_success()
        raise
    except Exception as exc:
        breaker.record_failure()
        raise map_grpc_error(exc) from exc


@router.get("/items")
async def list_items(request: Request):
    async def call():
        return await request.app.state.items_client.get_all()

    items = await with_breaker(request, call)
    return {"items": [_item_dict(i) for i in items.items], "total": items.total}


@router.post("/items", status_code=201)
async def create_item(body: CreateItemBody, request: Request):
    async def call():
        return await request.app.state.items_client.create(body.name, body.description)

    item = await with_breaker(request, call)
    return _item_dict(item)


@router.get("/items/{item_id}")
async def get_item(item_id: str, request: Request):
    async def call():
        return await request.app.state.items_client.get_by_id(item_id)

    item = await with_breaker(request, call)
    return _item_dict(item)


@router.delete("/items/{item_id}", status_code=204)
async def delete_item(item_id: str, request: Request):
    async def call():
        return await request.app.state.items_client.delete(item_id)

    await with_breaker(request, call)


def _item_dict(item) -> dict:
    return {
        "id": item.id,
        "name": item.name,
        "description": item.description,
        "created_at": item.created_at,
        "updated_at": item.updated_at,
    }
