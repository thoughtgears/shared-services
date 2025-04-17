package router

import (
	"fmt"
)

// Run starts the HTTP server and includes graceful shutdown handling.
func (r *Router) Run() error {
	addr := fmt.Sprintf("%s:%s", r.host, r.port)

	if err := r.Engine.Run(addr); err != nil {
		return fmt.Errorf("run router: %w", err)
	}

	return nil
}
