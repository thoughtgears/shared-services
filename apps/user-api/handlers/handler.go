package handlers

// Handler is a struct that contains services for handling user-related operations.
// It provides a unified interface for handling user operations in the system.
type Handler struct {
}

// NewHandler creates a new instance of Handler.
// It initializes the handler with the provided services.
// This function is used to set up the handler with the necessary services for user management.
// It is typically called during the initialization phase of the application.
func NewHandler() *Handler {
	return &Handler{}
}

// RegisterRoutes registers the routes for user-related operations.
// It sets up the API endpoints for updating, retrieving user by ID for the frontend.
func (h *Handler) RegisterRoutes(router *gin.Engine) {
	// Talent routes
	users := router.Group("/v1/users")
	{
		users.GET("/:id", nil)
		users.PUT("/:id", nil)
	}
}
