package service

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"core-service/internal/domain"
)

type fakeStore struct {
	items map[string]domain.Item
	next  int
}

func (f *fakeStore) Ping(ctx context.Context) error { return nil }

func (f *fakeStore) GetAll(ctx context.Context) []domain.Item {
	var out []domain.Item
	for _, it := range f.items {
		out = append(out, it)
	}
	return out
}

func (f *fakeStore) GetByID(ctx context.Context, id string) (domain.Item, error) {
	it, ok := f.items[id]
	if !ok {
		return domain.Item{}, fmt.Errorf("%w: item %s", domain.ErrNotFound, id)
	}
	return it, nil
}

func (f *fakeStore) Create(ctx context.Context, name, description string) (domain.Item, error) {
	f.next++
	it := domain.Item{ID: fmt.Sprintf("%d", f.next), Name: name, Description: description}
	f.items[it.ID] = it
	return it, nil
}

func (f *fakeStore) Delete(ctx context.Context, id string) error {
	if _, ok := f.items[id]; !ok {
		return fmt.Errorf("%w: item %s", domain.ErrNotFound, id)
	}
	delete(f.items, id)
	return nil
}

func newTestService() *ItemService {
	return NewItemService(&fakeStore{items: map[string]domain.Item{}})
}

func TestCreateRequiresName(t *testing.T) {
	svc := newTestService()

	_, err := svc.Create(context.Background(), "   ", "desc")
	if !errors.Is(err, domain.ErrInvalidInput) {
		t.Fatalf("expected ErrInvalidInput, got %v", err)
	}
}

func TestCreateGetRoundTrip(t *testing.T) {
	svc := newTestService()

	created, err := svc.Create(context.Background(), "thing", "desc")
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	got, err := svc.Get(context.Background(), created.ID)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if got.Name != "thing" || got.Description != "desc" {
		t.Fatalf("round trip mismatch: %+v", got)
	}
}

func TestGetNotFound(t *testing.T) {
	svc := newTestService()

	_, err := svc.Get(context.Background(), "404")
	if !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestDeleteNotFound(t *testing.T) {
	svc := newTestService()

	err := svc.Delete(context.Background(), "404")
	if !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestDelete(t *testing.T) {
	svc := newTestService()

	created, err := svc.Create(context.Background(), "thing", "")
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	if err := svc.Delete(context.Background(), created.ID); err != nil {
		t.Fatalf("delete: %v", err)
	}

	if _, err := svc.Get(context.Background(), created.ID); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("expected ErrNotFound after delete, got %v", err)
	}
}
