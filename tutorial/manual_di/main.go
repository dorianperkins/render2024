package main

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"os"

	"go.uber.org/zap"
)

func main() {
	logger := zap.NewExample()

	handlers := make(map[string]http.Handler)
	echoHandler := NewEchoHandler(logger)
	handlers["/echo"] = echoHandler
	helloHandler := NewHelloHandler(logger)
	handlers["/hello"] = helloHandler

	mux := NewServeMux(handlers)
	srv := NewHTTPServer(logger, mux)

	ln, err := net.Listen("tcp", srv.Addr)
	if err != nil {
		os.Exit(-1)
	}

	logger.Info("Starting HTTP server", zap.String("addr", srv.Addr))
	err = srv.Serve(ln)
	if err != nil {
		logger.Error("Failed to start HTTP server", zap.Error(err))
		os.Exit(-1)
	}
}

// NewHTTPServer builds an HTTP server that will begin serving requests when the
// application starts.
func NewHTTPServer(logger *zap.Logger, mux *http.ServeMux) *http.Server {
	srv := &http.Server{Addr: ":8080", Handler: mux}
	return srv
}

// NewServeMux builds a new http.ServeMux and registers the provided handlers.
func NewServeMux(handlers map[string]http.Handler) *http.ServeMux {
	mux := http.NewServeMux()
	for pattern, handler := range handlers {
		mux.Handle(pattern, handler)
	}
	return mux
}

// EchoHandler is an http.Handler that copies its request body back to the
// response.
type EchoHandler struct {
	log *zap.Logger
}

// NewEchoHandler builds a new EchoHandler.
func NewEchoHandler(log *zap.Logger) *EchoHandler {
	return &EchoHandler{log: log}
}

// ServeHTTP handles an HTTP request to the /echo endpoint.
func (h *EchoHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.log.Info("Received request for /echo")
	if _, err := io.Copy(w, r.Body); err != nil {
		h.log.Warn("Failed to handle request", zap.Error(err))
	}
}

// HelloHandler is an HTTP handler that prints a greeting to the user.
type HelloHandler struct {
	log *zap.Logger
}

// NewHelloHandler builds a new HelloHandler.
func NewHelloHandler(log *zap.Logger) *HelloHandler {
	return &HelloHandler{log: log}
}

// ServeHTTP handles an HTTP request to the /hello endpoint.
func (h *HelloHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.log.Info("Received request for /hello")
	body, err := io.ReadAll(r.Body)
	if err != nil {
		h.log.Error("Failed to read request", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if _, err := fmt.Fprintf(w, "Hello, %s\n", body); err != nil {
		h.log.Error("Failed to write response", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}
