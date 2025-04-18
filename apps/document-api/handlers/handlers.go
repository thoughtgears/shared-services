package handlers

import (
	"errors"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"

	"github.com/thoughtgears/shared-services/apps/document-api/services"
	"github.com/thoughtgears/shared-services/pkg/models"
)

// Handler is a struct that contains services for handling document-related operations.
// It provides a unified interface for handling document operations in the system.
type Handler struct {
	service services.DocumentService
}

// NewHandler creates a new instance of Handler.
// It initializes the handler with the provided services.
// This function is used to set up the handler with the necessary services for document management.
// It is typically called during the initialization phase of the application.
func NewHandler(service services.DocumentService) *Handler {
	return &Handler{
		service: service,
	}
}

// RegisterRoutes registers the routes for user-related operations.
// It sets up the API endpoints for updating, retrieving user by ID for the frontend.
func (h *Handler) RegisterRoutes(router *gin.Engine) {
	// Talent routes
	documents := router.Group("/v1/documents")
	{
		documents.GET("", h.GetAllByUserID) // Get all documents by user ID
		documents.GET("/:id", h.GetByID)    // Get document by ID
		documents.POST("", h.Create)
		documents.PUT("/:id", h.Update)
		documents.DELETE("/:id", h.Delete)
	}
}

// GetByID handles the GET request to retrieve a document by its unique ID.
// It returns the document object if found, or an error if not.
// This method is used to fetch document details.
func (h *Handler) GetByID(c *gin.Context) {
	id := c.Param("id")

	document, err := h.service.GetByID(c, id)
	if err != nil {
		log.Info().Err(err).Msg("Failed to get document by ID")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   err.Error(),
			"message": "Failed to retrieve document",
			"status":  http.StatusInternalServerError,
		})

		return
	}
	c.JSON(http.StatusOK, gin.H{
		"data":    document,
		"message": "Document retrieved successfully",
		"status":  http.StatusOK,
	})
}

// GetAllByUserID handles the GET request to retrieve all documents associated with a specific user ID.
// It returns a slice of document objects and an error if any occurs.
// This method is used to fetch all documents for a user.
func (h *Handler) GetAllByUserID(c *gin.Context) {
	userID := c.Query("user_id")

	documents, err := h.service.GetAllByUserID(c, userID)
	if err != nil {
		log.Info().Err(err).Msg("Failed to get documents by user ID")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   err.Error(),
			"message": "Failed to retrieve documents",
			"status":  http.StatusInternalServerError,
		})

		return
	}
	c.JSON(http.StatusOK, gin.H{
		"data":    documents,
		"message": "Documents retrieved successfully",
		"status":  http.StatusOK,
	})
}

// Create handles the POST request to create a new document.
// It returns the created document object and an error if any occurs.
// This method is used to upload a new document to the system.
func (h *Handler) Create(c *gin.Context) {
	userID := c.PostForm("user_id")
	if userID == "" {
		log.Error().Err(errors.New("user_id is required")).Msg("form field user_id is empty")
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "user_id is required",
			"message": "Missing required field: user_id",
			"status":  http.StatusBadRequest,
		})

		return
	}

	documentTypeStr := c.PostForm("document_type")
	documentType, err := models.ParseDocumentType(documentTypeStr)
	if err != nil {
		log.Error().Err(err).Msg("Invalid document type")
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   err.Error(),
			"message": "Invalid document type",
			"status":  http.StatusBadRequest,
		})

		return
	}

	file, err := c.FormFile("file")
	if err != nil {
		log.Error().Err(err).Msg("Failed to get file from form")
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   err.Error(),
			"message": "No file was uploaded or invalid file",
			"status":  http.StatusBadRequest,
		})

		return
	}

	openedFile, err := file.Open()
	if err != nil {
		log.Error().Err(err).Msg("Failed to open uploaded file")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   err.Error(),
			"message": "Failed to read uploaded file",
			"status":  http.StatusInternalServerError,
		})

		return
	}
	defer openedFile.Close()

	content, err := io.ReadAll(openedFile)
	if err != nil {
		log.Error().Err(err).Msg("Failed to read file content")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   err.Error(),
			"message": "Failed to read file content",
			"status":  http.StatusInternalServerError,
		})

		return
	}

	newDocument, err := h.service.Create(c, userID, documentType, content)
	if err != nil {
		log.Error().Err(err).Msg("Failed to create document")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   err.Error(),
			"message": "Failed to create document",
			"status":  http.StatusInternalServerError,
		})

		return
	}

	c.JSON(http.StatusAccepted, gin.H{
		"data":    newDocument,
		"message": "Document created successfully",
		"status":  http.StatusAccepted,
	})
}

// Update handles the PUT request to update an existing document.
// It returns the updated document object and an error if any occurs.
// This method is used to modify an existing document in the system.
func (h *Handler) Update(c *gin.Context) {
	id := c.Param("id")

	file, err := c.FormFile("file")
	if err != nil {
		log.Error().Err(err).Msg("Failed to get file from form")
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   err.Error(),
			"message": "No file was uploaded or invalid file",
			"status":  http.StatusBadRequest,
		})

		return
	}

	openedFile, err := file.Open()
	if err != nil {
		log.Error().Err(err).Msg("Failed to open uploaded file")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   err.Error(),
			"message": "Failed to read uploaded file",
			"status":  http.StatusInternalServerError,
		})

		return
	}
	defer openedFile.Close()

	content, err := io.ReadAll(openedFile)
	if err != nil {
		log.Error().Err(err).Msg("Failed to read file content")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   err.Error(),
			"message": "Failed to read file content",
			"status":  http.StatusInternalServerError,
		})

		return
	}

	document, err := h.service.Update(c, id, content)
	if err != nil {
		log.Error().Err(err).Msg("Failed to update document")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   err.Error(),
			"message": "Failed to update document",
			"status":  http.StatusInternalServerError,
		})

		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":    document,
		"message": "Document updated successfully",
	})
}

// Delete handles the DELETE request to remove a document by its unique ID.
// It returns a success message and an error if any occurs.
// This method is used to delete a document from the system.
func (h *Handler) Delete(c *gin.Context) {
	id := c.Param("id")

	err := h.service.Delete(c, id)
	if err != nil {
		log.Error().Err(err).Msg("Failed to delete document")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   err.Error(),
			"message": "Failed to delete document",
			"status":  http.StatusInternalServerError,
		})

		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Document deleted successfully",
	})
}
