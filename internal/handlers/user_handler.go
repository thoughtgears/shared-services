package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/thoughtgears/shared-services/internal/models"
	"github.com/thoughtgears/shared-services/internal/services"
)

// UserHandler is a struct that contains services for handling user-related operations.
// It provides a unified interface for handling user operations in the system.
type UserHandler struct {
	service services.UserService
}

// NewUserHandler creates a new instance of UserHandler.
// It initializes the handler with the provided services.
// This function is used to set up the handler with the necessary services for user management.
// It is typically called during the initialization phase of the application.
func NewUserHandler(service services.UserService) *UserHandler {
	return &UserHandler{
		service: service,
	}
}

// RegisterRoutes registers the routes for user-related operations.
// It sets up the API endpoints for updating, retrieving user by ID for the frontend.
func (u *UserHandler) RegisterRoutes(router *gin.Engine) {
	// Talent routes
	users := router.Group("/v1/users")
	{
		users.GET("/:id", u.GetByID)
		users.POST("", u.Create)
		users.PUT("/:id", u.Update)
	}
}

// GetByID handles the GET request to retrieve a user by their unique ID.
// It returns the user object if found, or an error if not.
// This method is used to fetch user details.
func (u *UserHandler) GetByID(c *gin.Context) {
	id := c.Param("id")

	user, err := u.service.GetByID(c, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   err.Error(),
			"message": "Failed to retrieve user",
			"status":  http.StatusInternalServerError,
		})

		return
	}
	c.JSON(http.StatusOK, gin.H{
		"data":    user,
		"message": "User retrieved successfully",
		"status":  http.StatusOK,
	})
}

// Create handles the POST request to create a new user.
// It returns the created user object and an error if any occurs.
// This method is used to register a new user in the system.
func (u *UserHandler) Create(c *gin.Context) {
	var user models.User

	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   err.Error(),
			"message": "Invalid request payload",
			"status":  http.StatusBadRequest,
		})

		return
	}

	newUser, err := u.service.Create(c, &user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   err.Error(),
			"message": "Failed to create user",
			"status":  http.StatusInternalServerError,
		})

		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"data":    newUser,
		"message": "User created successfully",
		"status":  http.StatusCreated,
	})
}

// Update handles the PUT request to modify an existing user's profile.
// It returns the updated user object and an error if any occurs.
// This method is used to update a user's profile information.
func (u *UserHandler) Update(c *gin.Context) {
	id := c.Param("id")

	var user models.User

	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   err.Error(),
			"message": "Invalid request payload",
			"status":  http.StatusBadRequest,
		})

		return
	}

	updatedUser, err := u.service.Update(c, id, &user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   err.Error(),
			"message": "Failed to update user",
			"status":  http.StatusInternalServerError,
		})

		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":    updatedUser,
		"message": "User updated successfully",
		"status":  http.StatusOK,
	})
}
