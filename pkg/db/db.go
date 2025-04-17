package db

import (
	"context"
)

// DB defines a generic data access interface for any type T.
// It provides standard CRUD operations and query capabilities with pagination support.
type DB[T any] interface {
	GetAll(ctx context.Context, pageToken string, pageSize int) ([]*T, string, error)
	GetByID(ctx context.Context, id string) (*T, error)
	GetByQuery(ctx context.Context, queries []QueryConstraint, pageToken string, pageSize int) ([]*T, string, error)
	Create(ctx context.Context, id string, data *T) (*T, error)
	Update(ctx context.Context, id string, data map[string]interface{}) (*T, error)
	Delete(ctx context.Context, id string) error
}
