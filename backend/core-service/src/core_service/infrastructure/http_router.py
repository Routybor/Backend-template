from datetime import datetime

from fastapi import APIRouter, HTTPException
from pydantic import BaseModel, Field

from core_service.domain import InvalidInputError, NotFoundError
from core_service.service import ItemService


class CreateItemBody(BaseModel):
    name: str = Field(min_length=1, max_length=255)
    description: str = ""


class ItemOut(BaseModel):
    id: str
    name: str
    description: str = ""
    created_at: datetime
    updated_at: datetime


class ListOut(BaseModel):
    items: list[ItemOut]
    total: int


def build_router(svc: ItemService) -> APIRouter:
    router = APIRouter()

    @router.get("/items", response_model=ListOut)
    async def list_items():
        items = svc.get_all()
        return ListOut(items=[_to_out(i) for i in items], total=len(items))

    @router.post("/items", response_model=ItemOut, status_code=201)
    async def create_item(body: CreateItemBody):
        try:
            item = svc.create(body.name, body.description)
        except InvalidInputError as exc:
            raise HTTPException(400, {"error": "invalid_request", "message": str(exc)}) from None
        return _to_out(item)

    @router.get("/items/{item_id}", response_model=ItemOut)
    async def get_item(item_id: str):
        try:
            item = svc.get(item_id)
        except NotFoundError:
            raise HTTPException(404, {"error": "not_found", "message": "item not found"}) from None
        return _to_out(item)

    @router.delete("/items/{item_id}", status_code=204)
    async def delete_item(item_id: str):
        try:
            svc.delete(item_id)
        except NotFoundError:
            raise HTTPException(404, {"error": "not_found", "message": "item not found"}) from None

    return router


def _to_out(item) -> ItemOut:
    return ItemOut(
        id=item.id,
        name=item.name,
        description=item.description,
        created_at=item.created_at,
        updated_at=item.updated_at,
    )
