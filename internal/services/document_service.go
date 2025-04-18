package services

import (
	"bytes"
	"context"
	"fmt"

	"cloud.google.com/go/firestore"
	"github.com/google/uuid"

	"github.com/thoughtgears/shared-services/internal/db"
	"github.com/thoughtgears/shared-services/internal/gcs"
	"github.com/thoughtgears/shared-services/internal/models"
)

// DocumentService handles operations specific to documents.
// It extends the DocumentService interface to include document-specific functionalities.
// This interface defines the methods that can be used to interact with documents in the system.
// It abstracts the underlying gcs implementation, and database interactions.
// The methods include creating, updating, deleting, and retrieving documents.
type DocumentService interface {
	GetByID(ctx context.Context, id string) (*models.Document, error)
	GetAllByUserID(ctx context.Context, userID string) ([]*models.Document, error)
	Create(ctx context.Context, userID string, documentType models.DocumentType, content []byte) (*models.Document, error)
	Update(ctx context.Context, id string, content []byte) (*models.Document, error)
	Delete(ctx context.Context, id string) error
}

// documentService is the concrete implementation of DocumentService.
// It uses a storage service to perform CRUD operations on document data.
// The storage service is expected to be a GCS or S3 storage service.
// The db is expected to be a Firestore db.
type documentService struct {
	storage gcs.Storage
	db      db.DB[models.Document]
}

// NewDocumentService creates a new instance of documentService.
// It initializes the service with a gcs service and a db for document data.
func NewDocumentService(storage gcs.Storage, db db.DB[models.Document]) DocumentService {
	return &documentService{
		storage: storage,
		db:      db,
	}
}

// GetByID retrieves a document by its unique ID.
// It returns the document object if found, or an error if not.
// This method is used to fetch document details.
func (d *documentService) GetByID(ctx context.Context, id string) (*models.Document, error) {
	document, err := d.db.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get document by ID: %w", err)
	}

	return document, nil
}

// GetAllByUserID retrieves all documents associated with a specific user ID.
// It returns a slice of document objects and an error if any occurs.
func (d *documentService) GetAllByUserID(ctx context.Context, userID string) ([]*models.Document, error) {
	query := []db.QueryConstraint{
		{
			Path:  "user_id",
			Op:    db.QueryOperatorEqual,
			Value: userID,
		},
	}

	documents, _, err := d.db.GetByQuery(ctx, query, "", 100)
	if err != nil {
		return nil, fmt.Errorf("failed to get documents by user ID: %w", err)
	}

	return documents, nil
}

// Create handles the creation of a new document.
// It returns the created document object and an error if any occurs.
// It uploads the document to the gcs service and saves the metadata in the database.
func (d *documentService) Create(ctx context.Context, userID string, documentType models.DocumentType, content []byte) (*models.Document, error) {
	data := bytes.NewReader(content)
	documentID := uuid.NewString()
	documentName := uuid.NewString()

	fileExtension, err := DetectFileType(content)
	if err != nil {
		return nil, fmt.Errorf("failed to detect file type: %w", err)
	}

	ext := GetStandardizedExtension(fileExtension.Extension)
	path := fmt.Sprintf("documents/%s/%s.%s", userID, documentName, ext)

	fileInfo, err := d.storage.Upload(ctx, path, data, fileExtension.MimeType)
	if err != nil {
		return nil, fmt.Errorf("failed to upload document: %w", err)
	}

	document := map[string]interface{}{
		"id":           documentID,
		"user_id":      userID,
		"name":         documentName,
		"size":         fileInfo.Size,
		"type":         documentType,
		"content_type": fileExtension.MimeType,
		"path":         path,
		"bucket":       fileInfo.Bucket,
		"created_at":   firestore.ServerTimestamp,
		"updated_at":   firestore.ServerTimestamp,
	}

	createdDocument, err := d.db.Create(ctx, documentID, document)
	if err != nil {
		return nil, fmt.Errorf("failed to create document: %w", err)
	}

	return createdDocument, nil
}

// Update handles the update of an existing document.
// It returns the updated document object and an error if any occurs.
// It uploads the updated document to the gcs service and updates the metadata in the database.
func (d *documentService) Update(ctx context.Context, id string, content []byte) (*models.Document, error) {
	data := bytes.NewReader(content)
	documentName := uuid.NewString()

	fileExtension, err := DetectFileType(content)
	if err != nil {
		return nil, fmt.Errorf("failed to detect file type: %w", err)
	}

	ext := GetStandardizedExtension(fileExtension.Extension)
	path := fmt.Sprintf("documents/%s/%s.%s", id, documentName, ext)

	fileInfo, err := d.storage.Upload(ctx, path, data, fileExtension.MimeType)
	if err != nil {
		return nil, fmt.Errorf("failed to upload document: %w", err)
	}

	document := map[string]interface{}{
		"name":         documentName,
		"size":         fileInfo.Size,
		"content_type": fileExtension.MimeType,
		"path":         path,
		"updated_at":   firestore.ServerTimestamp,
	}

	updatedDocument, err := d.db.Update(ctx, id, document)
	if err != nil {
		return nil, fmt.Errorf("failed to update document: %w", err)
	}

	return updatedDocument, nil
}

// Delete handles the deletion of a document.
// It removes the document from the gcs service and deletes the metadata from the database.
// It returns an error if any occurs during the process.
func (d *documentService) Delete(ctx context.Context, id string) error {
	document, err := d.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get document by ID: %w", err)
	}

	err = d.storage.Delete(ctx, document.Path)
	if err != nil {
		return fmt.Errorf("failed to delete document from gcs: %w", err)
	}

	err = d.db.Delete(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to delete document from database: %w", err)
	}

	return nil
}
