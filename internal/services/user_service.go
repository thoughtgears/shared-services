package services

import (
	"context"
	"fmt"
	"reflect"
	"strings"

	"cloud.google.com/go/firestore"
	"github.com/google/uuid"

	"github.com/thoughtgears/shared-services/internal/db"
	"github.com/thoughtgears/shared-services/internal/models"
)

// UserService handles operations specific to users.
// It extends the UserService interface to include user-specific functionalities.
type UserService interface {
	GetByID(ctx context.Context, id string) (*models.User, error)
	Create(ctx context.Context, user *models.User) (*models.User, error)
	Update(ctx context.Context, id string, talent *models.User) (*models.User, error)
}

// userService is the concrete implementation of UserService.
// It uses a generic repository to perform CRUD operations on talent data.
// The repository is expected to be initialized with a specific data type (models.User).
type userService struct {
	datastore db.DB[models.User]
}

// NewUserService creates a new instance of userService.
// It initializes the service with a db for user data.
// This repository is expected to be a Firestore db.
// It is typically called during the initialization phase of the application.
//
// Parameters:
//   - datastore: DB for user data
//
// Returns:
//   - UserService: Instance of userService
func NewUserService(datastore db.DB[models.User]) UserService {
	return &userService{
		datastore: datastore,
	}
}

// GetByID retrieves a user by their unique ID.
// It returns the user object if found, or an error if not.
// This method is used to fetch user details.
// It is typically called when a user needs to be displayed.
func (u *userService) GetByID(ctx context.Context, id string) (*models.User, error) {
	query := []db.QueryConstraint{
		{
			Path:  "firebase_id",
			Op:    db.QueryOperatorEqual,
			Value: id,
		},
	}
	user, _, err := u.datastore.GetByQuery(ctx, query, "", 1)
	if err != nil {
		return nil, fmt.Errorf("error getting talent by ID: %w", err)
	}

	if len(user) == 0 {
		return nil, fmt.Errorf("user not found")
	}

	return user[0], nil
}

// Create handles the creation of a new user.
// It returns the created user object and an error if any occurs.
// This method is used to register a new user in the system.
// It is typically called when a new user is signing up.
func (u *userService) Create(ctx context.Context, user *models.User) (*models.User, error) {
	if user == nil {
		return nil, fmt.Errorf("user cannot be nil")
	}

	user.ID = uuid.NewString()
	userData := map[string]interface{}{
		"id":          user.ID,
		"first_name":  user.FirstName,
		"last_name":   user.LastName,
		"email":       user.Email,
		"phone":       user.Phone,
		"firebase_id": user.FirebaseID,
		"address": map[string]interface{}{
			"building_number": user.Address.BuildingNumber,
			"street":          user.Address.Street,
			"city":            user.Address.City,
			"postcode":        user.Address.PostCode,
			"country":         user.Address.Country,
		},
		"created_at": firestore.ServerTimestamp,
		"updated_at": firestore.ServerTimestamp,
	}

	createdUser, err := u.datastore.Create(ctx, user.ID, userData)
	if err != nil {
		return nil, fmt.Errorf("error creating user: %w", err)
	}

	return createdUser, nil
}

// Update modifies an existing talent's profile.
// It returns the updated talent object and an error if any occurs.
// This method is used to update a talent's profile information.
func (u *userService) Update(ctx context.Context, id string, user *models.User) (*models.User, error) {
	currentUserData, err := u.datastore.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("error getting user by ID: %w", err)
	}

	updates := buildUpdateMapFromUser(user)

	if len(updates) == 0 {
		return currentUserData, nil
	}

	updates["updated_at"] = firestore.ServerTimestamp

	updatedUser, err := u.datastore.Update(ctx, id, updates)
	if err != nil {
		return nil, fmt.Errorf("error updating user: %w", err)
	}

	return updatedUser, nil
}

// buildUpdateMapFromUser creates a map of fields to update from a User object
func buildUpdateMapFromUser(user *models.User) map[string]interface{} {
	if user == nil {
		return nil
	}

	updates := make(map[string]interface{})
	val := reflect.ValueOf(user).Elem()
	typ := val.Type()

	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		fieldType := typ.Field(i)

		// Skip unexported fields
		if !field.CanInterface() {
			continue
		}

		// Get the firestore tag name
		tag := fieldType.Tag.Get("firestore")
		if tag == "" || tag == "-" {
			continue
		}

		// Extract the base tag name without options
		tagParts := strings.Split(tag, ",")
		tagName := tagParts[0]

		// Handle nested structs recursively
		if field.Kind() == reflect.Struct {
			// Skip Time fields or other special structs
			if fieldType.Type.String() == "time.Time" {

				continue
			}

			nestedUpdates := buildNestedUpdateMap(field.Addr().Interface())
			if len(nestedUpdates) > 0 {
				updates[tagName] = nestedUpdates
			}

			continue
		}

		// Skip empty string values
		if field.Kind() == reflect.String && field.String() == "" {

			continue
		}

		// Skip zero values for other types
		if isZeroValue(field) {

			continue
		}

		updates[tagName] = field.Interface()
	}

	return updates
}

// buildNestedUpdateMap handles nested structures
func buildNestedUpdateMap(obj interface{}) map[string]interface{} {
	updates := make(map[string]interface{})
	val := reflect.ValueOf(obj).Elem()
	typ := val.Type()

	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		fieldType := typ.Field(i)

		if !field.CanInterface() {
			continue
		}

		tag := fieldType.Tag.Get("firestore")
		if tag == "" || tag == "-" {
			continue
		}

		tagParts := strings.Split(tag, ",")
		tagName := tagParts[0]

		// Skip empty string values
		if field.Kind() == reflect.String && field.String() == "" {
			continue
		}

		// Skip zero values
		if isZeroValue(field) {
			continue
		}

		updates[tagName] = field.Interface()
	}

	return updates
}

// isZeroValue checks if a reflect.Value is the zero value for its type
func isZeroValue(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.String:
		return v.String() == ""
	case reflect.Bool:
		return !v.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return v.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return v.Float() == 0
	case reflect.Slice, reflect.Map:
		return v.Len() == 0
	case reflect.Interface, reflect.Ptr:
		return v.IsNil()
	}

	return false
}
