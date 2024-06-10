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

	echoHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger.Info("Received request for /echo")
		if _, err := io.Copy(w, r.Body); err != nil {
			logger.Warn("Failed to handle request", zap.Error(err))
		}
	})
	http.Handle("/echo", echoHandler)

	helloHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger.Info("Received request for /hello")
		body, err := io.ReadAll(r.Body)
		if err != nil {
			logger.Error("Failed to read request", zap.Error(err))
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		if _, err := fmt.Fprintf(w, "Hello, %s\n", body); err != nil {
			logger.Error("Failed to write response", zap.Error(err))
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
	})
	http.Handle("/hello", helloHandler)

	srv := &http.Server{Addr: ":8080"}

	ln, err := net.Listen("tcp", srv.Addr)
	if err != nil {
		logger.Error("Failed to listen", zap.Error(err))
		os.Exit(-1)
	}

	logger.Info("Starting HTTP server", zap.String("addr", srv.Addr))
	err = srv.Serve(ln)
	if err != nil {
		logger.Error("Failed to start HTTP server", zap.Error(err))
		os.Exit(-1)
	}
}
