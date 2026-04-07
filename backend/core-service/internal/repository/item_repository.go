package repository

import (
	"fmt"
	"sync"
	"sync/atomic"

	"core-service/dto"
)

type ItemRepository struct {
	items map[string]dto.Item
	mu    sync.RWMutex
	idSeq atomic.Int64
}

func NewItemRepository() *ItemRepository {
	repo := &ItemRepository{
		items: make(map[string]dto.Item),
	}
	repo.seed()
	return repo
}

func (r *ItemRepository) seed() {
	r.items["1"] = dto.Item{ID: "1", Name: "Item One"}
	r.items["2"] = dto.Item{ID: "2", Name: "Item Two"}
	r.items["3"] = dto.Item{ID: "3", Name: "Item Three"}
	r.idSeq.Store(3)
}

func (r *ItemRepository) GetAll() []dto.Item {
	r.mu.RLock()
	defer r.mu.RUnlock()

	items := make([]dto.Item, 0, len(r.items))
	for _, item := range r.items {
		items = append(items, item)
	}
	return items
}

func (r *ItemRepository) Create(name string) dto.Item {
	r.mu.Lock()
	defer r.mu.Unlock()

	id := r.idSeq.Add(1)
	item := dto.Item{
		ID:   fmt.Sprintf("%d", id),
		Name: name,
	}
	r.items[item.ID] = item
	return item
}

func (r *ItemRepository) GetByID(id string) (dto.Item, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	item, ok := r.items[id]
	return item, ok
}

func (r *ItemRepository) Delete(id string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.items[id]; ok {
		delete(r.items, id)
		return true
	}
	return false
}
