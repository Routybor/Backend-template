from core_service.domain import InvalidInputError, Item, ItemStore


class ItemService:
    def __init__(self, store: ItemStore) -> None:
        self._store = store

    def get_all(self) -> list[Item]:
        return self._store.get_all()

    def create(self, name: str, description: str) -> Item:
        if not name or not name.strip():
            raise InvalidInputError("name is required")
        return self._store.create(name, description)

    def get(self, item_id: str) -> Item:
        return self._store.get_by_id(item_id)

    def delete(self, item_id: str) -> None:
        self._store.delete(item_id)

    def ping(self) -> None:
        self._store.ping()
