package domain

import "context"

type ItemStore interface {
	GetAll(ctx context.Context) []Item
	GetByID(ctx context.Context, id string) (Item, error)
	Create(ctx context.Context, name, description string) (Item, error)
	Delete(ctx context.Context, id string) error
	Ping(ctx context.Context) error
}
