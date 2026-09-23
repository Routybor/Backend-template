package data

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"core-service/internal/domain"
)

type MemoryItemStore struct {
	items map[string]domain.Item
	mu    sync.RWMutex
	idSeq atomic.Int64
}

func NewMemoryItemStore() *MemoryItemStore {
	s := &MemoryItemStore{items: make(map[string]domain.Item)}
	s.seed()
	return s
}

func (s *MemoryItemStore) seed() {
	now := time.Now()
	s.items["1"] = domain.Item{ID: "1", Name: "Item One", CreatedAt: now, UpdatedAt: now}
	s.items["2"] = domain.Item{ID: "2", Name: "Item Two", CreatedAt: now, UpdatedAt: now}
	s.items["3"] = domain.Item{ID: "3", Name: "Item Three", CreatedAt: now, UpdatedAt: now}
	s.idSeq.Store(3)
}

func (s *MemoryItemStore) Ping(ctx context.Context) error {
	return nil
}

func (s *MemoryItemStore) GetAll(ctx context.Context) []domain.Item {
	s.mu.RLock()
	defer s.mu.RUnlock()

	items := make([]domain.Item, 0, len(s.items))
	for _, item := range s.items {
		items = append(items, item)
	}
	return items
}

func (s *MemoryItemStore) GetByID(ctx context.Context, id string) (domain.Item, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	item, ok := s.items[id]
	if !ok {
		return domain.Item{}, fmt.Errorf("%w: item %s", domain.ErrNotFound, id)
	}
	return item, nil
}

func (s *MemoryItemStore) Create(ctx context.Context, name, description string) (domain.Item, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	id := s.idSeq.Add(1)
	now := time.Now()
	item := domain.Item{
		ID:          fmt.Sprintf("%d", id),
		Name:        name,
		Description: description,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	s.items[item.ID] = item
	return item, nil
}

func (s *MemoryItemStore) Delete(ctx context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.items[id]; !ok {
		return fmt.Errorf("%w: item %s", domain.ErrNotFound, id)
	}
	delete(s.items, id)
	return nil
}
