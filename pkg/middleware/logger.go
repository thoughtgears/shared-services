package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// Logger returns a gin.HandlerFunc (middleware) that logs requests using
// the global zerolog logger (`log.Logger`).
//
// This middleware is designed to be added via `engine.Use()` and will log
// details about the request and response in a structured JSON format after
// the request has been processed by downstream handlers.
//
// It serves as a convenience wrapper around StructuredLogger, automatically
// providing the default global logger instance.
func Logger() gin.HandlerFunc {
	return StructuredLogger(&log.Logger)
}

// StructuredLogger returns a gin.HandlerFunc (middleware) that logs requests
// using a specific *zerolog.Logger instance provided as input.
//
// This allows for dependency injection of the logger, making it suitable for
// testing or using different logger configurations. The middleware should be
// added via `engine.Use()`.
//
// For each request, it performs the following steps:
//  1. Records the start time.
//  2. Calls `c.Next()` to allow downstream handlers to process the request.
//  3. After downstream processing, records the end time and calculates latency.
//  4. Gathers request details: Client IP, Method, Path (including query), Status Code, Body Size.
//  5. Extracts any errors added to the Gin context (`c.Errors`).
//  6. Determines the log level based on the response Status Code:
//     - >= 500: Error level
//     - >= 400: Warning level
//     - < 400: Info level
//  7. Logs a single structured JSON message including all gathered details using the provided logger instance.
//     The primary message of the log entry contains Gin's formatted private errors, if any.
//
// Parameters:
//   - logger: A pointer to the `zerolog.Logger` instance to use for logging.
//
// Returns:
//   - A `gin.HandlerFunc` to be used as middleware.
func StructuredLogger(logger *zerolog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {

		start := time.Now() // Start timer
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		// Process request by calling downstream handlers first
		c.Next()

		// ---- Logging executed AFTER the request ----

		// Prepare log parameters using gin's format
		param := gin.LogFormatterParams{}

		param.TimeStamp = time.Now() // Stop timer
		param.Latency = param.TimeStamp.Sub(start)
		// Optional: Truncate latency for readability if very long
		if param.Latency > time.Minute {
			param.Latency = param.Latency.Truncate(time.Second)
		}

		param.ClientIP = c.ClientIP()
		param.Method = c.Request.Method
		param.StatusCode = c.Writer.Status()
		// Capture errors attached to the context
		param.ErrorMessage = c.Errors.ByType(gin.ErrorTypePrivate).String()
		param.BodySize = c.Writer.Size()
		// Append query string to path if it exists
		if raw != "" {
			path = path + "?" + raw
		}
		param.Path = path

		// Determine log level based on status code
		status := c.Writer.Status() // Get status once for readability and efficiency
		var logEvent *zerolog.Event

		switch {
		case status >= 500: // Server errors (5xx)
			logEvent = logger.Error()
		case status >= 400: // Client errors (4xx) - This case is only reached if status < 500
			logEvent = logger.Warn()
		default: // Success, redirects, informational etc. (< 400)
			logEvent = logger.Info()
		}

		// Log structured event with relevant fields
		logEvent.Str("client_id", param.ClientIP).
			Str("method", param.Method).
			Int("status_code", param.StatusCode).
			Int("body_size", param.BodySize).
			Str("path", param.Path).
			Str("latency", param.Latency.String()).
			Msg(param.ErrorMessage)
	}
}
