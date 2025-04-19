package middleware

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	firebase "firebase.google.com/go/v4"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	"google.golang.org/api/option"
)

// Global Firebase app instance to avoid recreating it for each request
var firebaseApp *firebase.App

// InitFirebase initializes the Firebase app on server startup
func InitFirebase(ctx context.Context) error {
	var err error
	opt := option.WithCredentialsFile("./secrets/firebase-service-account.json")
	firebaseApp, err = firebase.NewApp(ctx, nil, opt)
	if err != nil {
		return fmt.Errorf("failed to initialize Firebase app: %w", err)
	}

	return nil
}

// FirebaseAuth is middleware that validates Firebase auth tokens
// and adds the user information to the context.
// It uses the Firebase Admin SDK to verify the token and extract user claims.
// If the token is valid, it calls the next handler in the chain.
// If the token is invalid, it aborts the request with a 401 Unauthorized status.
// This middleware is typically used to protect routes that require authentication.
func FirebaseAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Ensure Firebase app is initialized
		if firebaseApp == nil {
			log.Error().Msg("Firebase app not initialized")
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"error":   "internal server error",
				"message": "Firebase client not initialized",
			})

			return
		}

		// Get the auth client
		ctx := c.Request.Context()
		client, err := firebaseApp.Auth(ctx)
		if err != nil {
			log.Error().Err(err).Msg("Failed to get Auth client")
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"error":   "internal server error",
				"message": "Failed to get Auth client",
			})

			return
		}

		// Extract and verify token
		authHeader := c.GetHeader("Authorization")
		idToken, err := extractToken(authHeader)
		if err != nil {
			log.Error().Err(err).Msg("Failed to extract token, invalid format")
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error":   "unauthorized",
				"message": "Invalid token format",
			})

			return
		}

		// Verify the token
		token, err := client.VerifyIDToken(ctx, idToken)
		if err != nil {
			log.Error().Err(err).Msgf("Failed to verify ID token: %v", token)
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error":   "unauthorized",
				"message": "Invalid token",
			})

			return
		}

		// Add the token claims to the context
		c.Set("user", token)
		c.Next()

	}
}

// extractToken extracts the token from the Authorization header.
func extractToken(authHeader string) (string, error) {
	if authHeader == "" {
		return "", errors.New("authorization header is required")
	}

	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		return "", errors.New("authorization header format must be 'Bearer {token}'")
	}

	return parts[1], nil
}
