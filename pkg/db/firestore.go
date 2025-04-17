package db

import (
	"context"
	"errors"
	"fmt"

	"cloud.google.com/go/firestore"
	"google.golang.org/api/iterator"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// QueryConstraint represents a Firestore query condition used to filter documents.
// It maps directly to Firestore's Where() method parameters.
type QueryConstraint struct {
	Path  string      // Field path (e.g., "stripeCustomerId")
	Op    string      // Operator (e.g., "==", "<", ">=", "in", "array-contains")
	Value interface{} // Value to compare against
}

// firestoreRepository implements Repository interface for Firestore database.
// It provides generic CRUD operations for any data type.
type firestoreRepository[T any] struct {
	client         *firestore.Client
	collectionName string
}

// NewFirestoreRepository creates a new instance of firestoreRepository for a specific type.
// It implements the Repository interface for the given type T.
//
// Parameters:
//   - client: Initialized Firestore client
//   - collectionName: Name of the Firestore collection where data will be stored
//
// Returns:
//   - Repository[T]: A repository instance for the specified type
func NewFirestoreRepository[T any](client *firestore.Client, collectionName string) DB[T] {
	return &firestoreRepository[T]{
		client:         client,
		collectionName: collectionName,
	}
}

// GetAll retrieves all documents from the collection with optional pagination.
//
// Parameters:
//   - ctx: Context for the database operation
//   - pageToken: Token representing the starting point for this page (empty for first page)
//   - pageSize: Maximum number of documents to retrieve (<=0 for no limit)
//
// Returns:
//   - []*T: Slice of document data
//   - string: Token for retrieving the next page (empty if no more pages)
//   - error: Any error encountered during the operation
func (r *firestoreRepository[T]) GetAll(ctx context.Context, pageToken string, pageSize int) ([]*T, string, error) {
	query := r.client.Collection(r.collectionName).OrderBy(firestore.DocumentID, firestore.Asc) // Order for consistent pagination
	if pageToken != "" {
		query = query.StartAfter(pageToken)
	}
	if pageSize > 0 {
		query = query.Limit(pageSize)
	}

	iter := query.Documents(ctx)
	defer iter.Stop()

	var results []*T
	var lastDocID string
	for {
		doc, err := iter.Next()
		if errors.Is(err, iterator.Done) {
			break
		}
		if err != nil {
			return nil, "", fmt.Errorf("failed to iterate documents: %w", err)
		}
		var data T
		if err := doc.DataTo(&data); err != nil {
			return nil, "", fmt.Errorf("failed to convert document data: %w", err)
		}

		results = append(results, &data)
		lastDocID = doc.Ref.ID // Store the ID of the last successfully processed doc
	}

	// Determine next page token (simply the ID of the last doc in this batch)
	// More robust pagination might involve cursors, but this is common.
	nextPageToken := ""
	// Only provide a next token if we potentially limited results and got some results
	if pageSize > 0 && len(results) == pageSize {
		nextPageToken = lastDocID
	}

	return results, nextPageToken, nil
}

// GetByID retrieves a single document by its ID.
//
// Parameters:
//   - ctx: Context for the database operation
//   - id: Unique identifier of the document
//
// Returns:
//   - *T: Document data
//   - error: NotFound error or any other error encountered
func (r *firestoreRepository[T]) GetByID(ctx context.Context, id string) (*T, error) {
	doc, err := r.client.Collection(r.collectionName).Doc(id).Get(ctx)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return nil, fmt.Errorf("document with id %s not found: %w", id, err) // Consider a specific ErrNotFound
		}

		return nil, fmt.Errorf("failed to get document %s: %w", id, err)
	}
	if !doc.Exists() { // Should be caught by the error check above, but good practice
		return nil, fmt.Errorf("document with id %s not found (exists=false)", id)
	}

	var result T
	if err := doc.DataTo(&result); err != nil {
		return nil, fmt.Errorf("failed to convert document data: %w", err)
	}

	return &result, nil
}

// GetByQuery retrieves documents matching the specified query constraints with optional pagination.
// Multiple constraints are combined with logical AND.
//
// Parameters:
//   - ctx: Context for the database operation
//   - queries: Slice of QueryConstraint to filter the documents
//   - pageToken: Token representing the starting point for this page
//   - pageSize: Maximum number of documents to retrieve
//
// Returns:
//   - []*T: Slice of document data matching the query
//   - string: Token for retrieving the next page
//   - error: Any error encountered during the operation
func (r *firestoreRepository[T]) GetByQuery(ctx context.Context, queries []QueryConstraint, pageToken string, pageSize int) ([]*T, string, error) {
	fsQuery := r.client.Collection(r.collectionName).Query
	for _, q := range queries {
		fsQuery = fsQuery.Where(q.Path, q.Op, q.Value)
	}

	// Add ordering for consistent pagination if not already specified in queries
	// Note: Firestore requires the first OrderBy field to match the first range/inequality filter field if present.
	// This simple implementation assumes DocumentID ordering is sufficient or that queries include ordering.
	// A more robust implementation might need smarter OrderBy logic based on query constraints.
	fsQuery = fsQuery.OrderBy(firestore.DocumentID, firestore.Asc)

	if pageToken != "" {
		// Fetch the document snapshot for the page token to use StartAfter
		// This requires an extra read but is the standard way for non-cursor pagination
		docSnapshot, err := r.client.Collection(r.collectionName).Doc(pageToken).Get(ctx)
		if err != nil {
			return nil, "", fmt.Errorf("failed to get page token document %s: %w", pageToken, err)
		}
		fsQuery = fsQuery.StartAfter(docSnapshot) // Use snapshot for StartAfter
	}

	if pageSize > 0 {
		fsQuery = fsQuery.Limit(pageSize)
	}

	iter := fsQuery.Documents(ctx)
	defer iter.Stop()

	var results []*T
	var lastDocSnapshot *firestore.DocumentSnapshot // Store last snapshot for next page token
	for {
		doc, err := iter.Next()
		if errors.Is(err, iterator.Done) {
			break
		}
		if err != nil {
			return nil, "", fmt.Errorf("failed to iterate query documents: %w", err)
		}

		var data T
		if err := doc.DataTo(&data); err != nil {
			return nil, "", fmt.Errorf("failed to convert document data: %w", err)
		}

		results = append(results, &data)
		lastDocSnapshot = doc
	}

	// Use the last document's ID as the next page token
	nextPageToken := ""
	if pageSize > 0 && len(results) == pageSize && lastDocSnapshot != nil {
		nextPageToken = lastDocSnapshot.Ref.ID
	}

	return results, nextPageToken, nil
}

// Create adds a new document to the collection with the specified ID.
// If the document already exists, it will be overwritten.
//
// Parameters:
//   - ctx: Context for the database operation
//   - id: ID for the new document
//   - data: Data to store in the document
//
// Returns:
//   - *T: The created document data
//   - error: Any error encountered during creation
func (r *firestoreRepository[T]) Create(ctx context.Context, id string, data *T) (*T, error) {
	if _, err := r.client.Collection(r.collectionName).Doc(id).Set(ctx, data); err != nil {
		return nil, fmt.Errorf("failed to create document: %w", err)
	}

	doc, err := r.client.Collection(r.collectionName).Doc(id).Get(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get created document: %w", err)
	}

	if !doc.Exists() {
		return nil, fmt.Errorf("document with id %s not found after creation", doc.Ref.ID)
	}

	var result T
	if err := doc.DataTo(&result); err != nil {
		return nil, fmt.Errorf("failed to convert document data: %w", err)
	}

	return &result, nil
}

// Update modifies specific fields of an existing document.
// The document must exist, or an error will be returned.
//
// Parameters:
//   - ctx: Context for the database operation
//   - id: ID of the document to update
//   - data: Map of fields to update with their new values
//
// Returns:
//   - *T: The updated document data
//   - error: NotFound error or any other error encountered
func (r *firestoreRepository[T]) Update(ctx context.Context, id string, data map[string]interface{}) (*T, error) {
	_, err := r.client.Collection(r.collectionName).Doc(id).Set(ctx, data, firestore.MergeAll)
	if err != nil {
		return nil, fmt.Errorf("failed to update document %s: %w", id, err)
	}

	doc, err := r.client.Collection(r.collectionName).Doc(id).Get(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get updated document %s: %w", id, err)
	}

	if !doc.Exists() {
		return nil, fmt.Errorf("document with id %s not found after update", id)
	}

	var result T
	if err := doc.DataTo(&result); err != nil {
		return nil, fmt.Errorf("failed to convert document data: %w", err)
	}

	return &result, nil
}

// Delete removes a document from the collection.
// If the document does not exist, an error will be returned.
//
// Parameters:
//   - ctx: Context for the database operation
//   - id: ID of the document to delete
//
// Returns:
//   - error: NotFound error or any other error encountered
func (r *firestoreRepository[T]) Delete(ctx context.Context, id string) error {
	_, err := r.client.Collection(r.collectionName).Doc(id).Delete(ctx)
	if status.Code(err) == codes.NotFound {
		return fmt.Errorf("document with id %s not found: %w", id, err)
	}
	if err != nil {
		return fmt.Errorf("failed to delete document %s: %w", id, err)
	}

	return nil
}
