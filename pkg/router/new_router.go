package router

import (
	"github.com/gin-gonic/gin"

	"github.com/thoughtgears/shared-services/pkg/middleware"
)

type Router struct {
	Engine *gin.Engine
	host   string
	port   string
}

// NewRouter creates and configures a new Router instance with middleware and configuration.
//
// It initializes a new Gin Engine using gin.New() (instead of gin.Default() to allow
// for explicit middleware selection). It sets the Gin mode to ReleaseMode if
// config.Debug is false.
//
// Middleware added includes:
//   - A custom structured logger (via middleware.Logger()).
//   - Gin's default recovery middleware to handle panics gracefully.
//
// It clears any default trusted proxies using SetTrustedProxies(nil), which is often
// suitable when running behind a known reverse proxy or load balancer.
//
// Parameters:
//   - local: If the application is running locally its set to true.
//   - port: A pointer to a string representing the port to run the server on.
//
// Returns:
//   - A pointer to the configured *Router instance, ready to be run.
func NewRouter(local bool, port *string) *Router {
	var newRouter Router

	if local {
		gin.SetMode(gin.DebugMode)
		newRouter.host = "127.0.0.1"
	}

	// Set default port to 8080 if not provided
	if port == nil {
		newRouter.port = "8080"
	} else {
		newRouter.port = *port
	}

	newRouter.Engine = gin.New()
	newRouter.Engine.Use(middleware.Logger())
	newRouter.Engine.Use(gin.Recovery())

	// Explicitly clear trusted proxies (important for security depending on deployment)
	// If behind a trusted proxy (like Cloudflare), you might configure this differently.
	// --- Trusted Proxies ---
	// Consider setting trusted proxies if behind LB/Reverse Proxy
	// err := Engine.SetTrustedProxies([]string{"192.168.1.100", "10.0.0.0/8"})
	// if err != nil {
	//     log.Fatalf("Failed to set trusted proxies: %v", err)
	// }
	// For now, clearing them might be fine depending on your setup.
	_ = newRouter.Engine.SetTrustedProxies(nil)

	return &newRouter
}
