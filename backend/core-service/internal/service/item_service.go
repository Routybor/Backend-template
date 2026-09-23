package service

import (
	"context"
	"fmt"
	"strings"

	"core-service/internal/domain"
)

type ItemService struct {
	store domain.ItemStore
}

func NewItemService(store domain.ItemStore) *ItemService {
	return &ItemService{store: store}
}

func (s *ItemService) GetAll(ctx context.Context) []domain.Item {
	return s.store.GetAll(ctx)
}

func (s *ItemService) Create(ctx context.Context, name, description string) (domain.Item, error) {
	if strings.TrimSpace(name) == "" {
		return domain.Item{}, fmt.Errorf("%w: name is required", domain.ErrInvalidInput)
	}
	return s.store.Create(ctx, name, description)
}

func (s *ItemService) Get(ctx context.Context, id string) (domain.Item, error) {
	return s.store.GetByID(ctx, id)
}

func (s *ItemService) Delete(ctx context.Context, id string) error {
	return s.store.Delete(ctx, id)
}

func (s *ItemService) Ping(ctx context.Context) error {
	return s.store.Ping(ctx)
}
