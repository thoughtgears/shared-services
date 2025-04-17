package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/thoughtgears/shared-services/apps/user-api/services"
	"github.com/thoughtgears/shared-services/pkg/models"
)

// Handler is a struct that contains services for handling user-related operations.
// It provides a unified interface for handling user operations in the system.
type Handler struct {
	service services.UserService
}

// NewHandler creates a new instance of Handler.
// It initializes the handler with the provided services.
// This function is used to set up the handler with the necessary services for user management.
// It is typically called during the initialization phase of the application.
func NewHandler(service services.UserService) *Handler {
	return &Handler{
		service: service,
	}
}

// RegisterRoutes registers the routes for user-related operations.
// It sets up the API endpoints for updating, retrieving user by ID for the frontend.
func (h *Handler) RegisterRoutes(router *gin.Engine) {
	// Talent routes
	users := router.Group("/v1/users")
	{
		users.GET("/:id", h.GetByID)
		users.POST("")
		users.PUT("/:id", h.Update)
	}
}

// GetByID handles the GET request to retrieve a user by their unique ID.
// It returns the user object if found, or an error if not.
// This method is used to fetch user details.
func (h *Handler) GetByID(c *gin.Context) {
	id := c.Param("id")

	user, err := h.service.GetByID(c, id)
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
func (h *Handler) Create(c *gin.Context) {
	var user models.User

	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   err.Error(),
			"message": "Invalid request payload",
			"status":  http.StatusBadRequest,
		})

		return
	}

	newUser, err := h.service.Create(c, &user)
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
func (h *Handler) Update(c *gin.Context) {
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

	updatedUser, err := h.service.Update(c, id, &user)
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
