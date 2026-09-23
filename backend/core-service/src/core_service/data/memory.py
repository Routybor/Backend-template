import threading

from core_service.domain import Item, ItemStore, NotFoundError


class MemoryItemStore(ItemStore):
    def __init__(self) -> None:
        self._items: dict[str, Item] = {}
        self._lock = threading.Lock()
        self._seed()

    def _seed(self) -> None:
        for i in (1, 2, 3):
            self._items[str(i)] = Item(str(i), f"Item {['One', 'Two', 'Three'][i - 1]}")

    def ping(self) -> None:
        return None

    def get_all(self) -> list[Item]:
        with self._lock:
            return list(self._items.values())

    def get_by_id(self, item_id: str) -> Item:
        with self._lock:
            item = self._items.get(item_id)
            if item is None:
                raise NotFoundError(f"item {item_id} not found")
            return item

    def create(self, name: str, description: str) -> Item:
        with self._lock:
            item_id = str(len(self._items) + 1)
            item = Item(item_id, name, description)
            self._items[item_id] = item
            return item

    def delete(self, item_id: str) -> None:
        with self._lock:
            if item_id not in self._items:
                raise NotFoundError(f"item {item_id} not found")
            del self._items[item_id]
