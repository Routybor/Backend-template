import pytest

from core_service.domain import InvalidInputError, Item, ItemStore, NotFoundError
from core_service.service import ItemService


class FakeStore(ItemStore):
    def __init__(self):
        self.items: dict[str, Item] = {}
        self._next = 0

    def ping(self):
        return None

    def get_all(self):
        return list(self.items.values())

    def get_by_id(self, item_id):
        item = self.items.get(item_id)
        if item is None:
            raise NotFoundError(f"item {item_id} not found")
        return item

    def create(self, name, description):
        self._next += 1
        item = Item(str(self._next), name, description)
        self.items[item.id] = item
        return item

    def delete(self, item_id):
        if item_id not in self.items:
            raise NotFoundError(f"item {item_id} not found")
        del self.items[item_id]


@pytest.fixture
def svc():
    return ItemService(FakeStore())


def test_create_requires_name(svc):
    with pytest.raises(InvalidInputError):
        svc.create("   ", "desc")


def test_create_get_round_trip(svc):
    created = svc.create("thing", "desc")
    got = svc.get(created.id)
    assert (got.name, got.description) == ("thing", "desc")


def test_get_not_found(svc):
    with pytest.raises(NotFoundError):
        svc.get("404")


def test_delete_not_found(svc):
    with pytest.raises(NotFoundError):
        svc.delete("404")


def test_delete(svc):
    created = svc.create("thing", "")
    svc.delete(created.id)
    with pytest.raises(NotFoundError):
        svc.get(created.id)
