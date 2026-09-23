from datetime import datetime
from typing import Protocol


class ItemStore(Protocol):
    def get_all(self) -> list["Item"]: ...
    def get_by_id(self, item_id: str) -> "Item": ...
    def create(self, name: str, description: str) -> "Item": ...
    def delete(self, item_id: str) -> None: ...
    def ping(self) -> None: ...


class Item:
    __slots__ = ("id", "name", "description", "created_at", "updated_at")

    def __init__(self, item_id: str, name: str, description: str = "") -> None:
        self.id = item_id
        self.name = name
        self.description = description
        self.created_at = datetime.now(tz=None)
        self.updated_at = self.created_at
