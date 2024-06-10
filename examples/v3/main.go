package main

import (
	"io"
	"net/http"

	"go.uber.org/fx"
	"go.uber.org/zap"
)

func main() {
	// Create Fx app.
	fx.New(
		fx.Provide(NewLogger),
		fx.Provide(NewHandler),
		fx.Invoke(RegisterHandler),
		fx.Invoke(StartServer),
	).Run()
}

func NewLogger() *zap.Logger {
	return zap.NewExample()
}

func NewHandler(logger *zap.Logger) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			logger.Info("[v3] - Handler received request")
			if _, err := io.Copy(w, r.Body); err != nil {
				logger.Warn("Failed to handle request", zap.Error(err))
			}
		},
	)
}

func RegisterHandler(logger *zap.Logger, h http.Handler) {
	logger.Info("Registering handler")
	http.Handle("/echo", h)
}

func StartServer(logger *zap.Logger) {
	logger.Info("Starting server")
	http.ListenAndServe(":8080", nil)
}
