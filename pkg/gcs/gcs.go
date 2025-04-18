package gcs

import (
	"context"
	"errors"
	"fmt"
	"io"
	"time"

	"cloud.google.com/go/storage"
	"google.golang.org/api/iterator"
)

// FileInfo contains metadata about a stored file
// such as its path, size, content type, and last modified time.
type FileInfo struct {
	Path         string
	Size         int64
	ContentType  string
	LastModified time.Time
	Bucket       string
}

// GCSStorage is a struct that implements the Storage interface for Google Cloud Storage
// It provides methods for uploading, downloading, deleting files,
// and listing files in a Google Cloud Storage bucket.
type GCSStorage struct {
	client     *storage.Client
	bucketName string
}

// NewGCSStorage creates a new GCSStorage instance
// It initializes the GCS client and sets the bucket name and project ID.
func NewGCSStorage(client *storage.Client, bucketName string) (*GCSStorage, error) {
	return &GCSStorage{
		client:     client,
		bucketName: bucketName,
	}, nil
}

// Upload a file to GCS
// It takes a context, file path, content reader, and content type as parameters.
// It creates a new object in the specified bucket and writes the content to it.
// If the upload is successful, it returns nil.
// If there is an error, it returns the error.
// The content type is set to the specified value.
func (g *GCSStorage) Upload(ctx context.Context, path string, content io.Reader, contentType string) (*FileInfo, error) {
	bucket := g.client.Bucket(g.bucketName)
	obj := bucket.Object(path)
	wc := obj.NewWriter(ctx)
	wc.ContentType = contentType

	if _, err := io.Copy(wc, content); err != nil {
		if err := wc.Close(); err != nil {
			return nil, fmt.Errorf("failed to close writer after error: %w", err)
		}

		return nil, fmt.Errorf("failed to copy content to GCS: %w", err)
	}

	if err := wc.Close(); err != nil {
		return nil, fmt.Errorf("failed to close writer: %w", err)
	}

	attrs, err := obj.Attrs(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get object attributes: %w", err)
	}

	fileInfo := &FileInfo{
		Path:         attrs.Name,
		Size:         attrs.Size,
		ContentType:  attrs.ContentType,
		LastModified: attrs.Updated,
		Bucket:       g.bucketName,
	}

	return fileInfo, nil
}

// Download a file from GCS
// It takes a context and file path as parameters.
// It creates a new reader for the specified object in the bucket.
// If the download is successful, it returns the reader.
// If there is an error, it returns the error.
func (g *GCSStorage) Download(ctx context.Context, path string) (io.ReadCloser, error) {
	bucket := g.client.Bucket(g.bucketName)
	obj := bucket.Object(path)

	r, err := obj.NewReader(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create reader: %w", err)
	}

	return r, nil
}

// Delete a file from GCS
// It takes a context and file path as parameters.
// It creates a new object in the specified bucket and deletes it.
// If the deletion is successful, it returns nil.
// If there is an error, it returns the error.
func (g *GCSStorage) Delete(ctx context.Context, path string) error {
	bucket := g.client.Bucket(g.bucketName)
	obj := bucket.Object(path)

	if err := obj.Delete(ctx); err != nil {
		return fmt.Errorf("failed to delete object: %w", err)
	}

	return nil
}

// List files in a directory in GCS
// It takes a context and prefix as parameters.
// It creates a new iterator for the specified prefix in the bucket.
// It iterates through the objects and appends their metadata to a slice of FileInfo.
// If the listing is successful, it returns the slice of FileInfo.
// If there is an error, it returns the error.
func (g *GCSStorage) List(ctx context.Context, prefix string) ([]FileInfo, error) {
	bucket := g.client.Bucket(g.bucketName)

	var files []FileInfo
	it := bucket.Objects(ctx, &storage.Query{Prefix: prefix})

	for {
		attrs, err := it.Next()
		if errors.Is(err, iterator.Done) {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("error iterating through objects: %w", err)
		}

		files = append(files, FileInfo{
			Path:         attrs.Name,
			Size:         attrs.Size,
			ContentType:  attrs.ContentType,
			LastModified: attrs.Updated,
		})
	}

	return files, nil
}
