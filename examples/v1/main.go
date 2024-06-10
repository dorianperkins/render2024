package main

import (
	"io"
	"net/http"

	"go.uber.org/zap"
)

func main() {
	// Create logger
	logger := zap.NewExample()

	// Create handler
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger.Info("[v1] Handler received request")
		if _, err := io.Copy(w, r.Body); err != nil {
			logger.Warn("Failed to handle request", zap.Error(err))
		}
	})

	// Register handler
	logger.Info("Registering handler")
	http.Handle("/echo", handler)

	// Start server
	logger.Info("Starting server")
	http.ListenAndServe(":8080", nil)
}
