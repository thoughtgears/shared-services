package gcs

import (
	"context"
	"io"
)

// Storage is an interface for a gcs service
// that provides methods for uploading, downloading,
// deleting files, and listing files in a gcs system.
// It abstracts the underlying gcs implementation,
// allowing for different gcs backends (e.g., S3, local filesystem, GCS).
type Storage interface {
	Upload(ctx context.Context, path string, content io.Reader, contentType string) (*FileInfo, error)
	Download(ctx context.Context, path string) (io.ReadCloser, error)
	Delete(ctx context.Context, path string) error
	List(ctx context.Context, prefix string) ([]FileInfo, error)
}
