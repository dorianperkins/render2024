package main

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"

	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"
	"go.uber.org/zap"
)

func main() {
	fx.New(
		// Replace the default Fx logger with our own logger.
		fx.WithLogger(func(log *zap.Logger) fxevent.Logger {
			return &fxevent.ZapLogger{Logger: log}
		}),
		// Register the HTTP server, handlers, and logger.
		// NOTE: The order of these provides does not matter.
		fx.Provide(
			NewHTTPServer,
			AsRoute(NewEchoHandler),
			AsRoute(NewHelloHandler),
			zap.NewExample,
		),
		// Invoke a function which takes the http.Server as a dependency such
		// that Fx will be required to execute NewHTTPServer.
		fx.Invoke(func(*http.Server) {}),
	).Run()
}

// AsRoute annotates the given constructor to state that it provides a route to
// the "routes" value group.
func AsRoute(f any) any {
	return fx.Annotate(
		f,
		fx.As(new(Route)),
		fx.ResultTags(`group:"routes"`),
	)
}

// Params are the params for constructing the HTTP server.
type HTTPServerParams struct {
	fx.In

	L      fx.Lifecycle
	Logger *zap.Logger
	Routes []Route `group:"routes"`
}

// Result is the result of constructing the HTTP server.
type HTTPServerResult struct {
	fx.Out

	Server *http.Server
	Mux    *http.ServeMux
}

// NewHTTPServer builds an HTTP server that will begin serving requests when the
// Fx application starts.
func NewHTTPServer(p HTTPServerParams) HTTPServerResult {
	mux := http.NewServeMux()
	for _, route := range p.Routes {
		mux.Handle(route.Pattern(), route)
	}

	srv := &http.Server{Addr: ":8080", Handler: mux}
	p.L.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			ln, err := net.Listen("tcp", srv.Addr)
			if err != nil {
				return err
			}
			p.Logger.Info("Starting HTTP server", zap.String("addr", srv.Addr))
			go srv.Serve(ln)
			return nil
		},
		OnStop: func(ctx context.Context) error {
			return srv.Shutdown(ctx)
		},
	})

	return HTTPServerResult{
		Server: srv,
		Mux:    mux,
	}
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
	h.log.Info("Received request for handler", zap.String("handler", h.Pattern()))
	if _, err := io.Copy(w, r.Body); err != nil {
		h.log.Warn("Failed to handle request", zap.Error(err))
	}
}

// Pattern is the path at which this handler is registered.
func (*EchoHandler) Pattern() string {
	return "/echo"
}

// Route is an http.Handler that stores the pattern under which it will be
// registered on the mux.
type Route interface {
	http.Handler

	// Pattern reports the path at which this is registered.
	Pattern() string
}

// HelloHandler is an HTTP handler that prints a greeting to the user.
type HelloHandler struct {
	log *zap.Logger
}

// NewHelloHandler builds a new HelloHandler.
func NewHelloHandler(log *zap.Logger) *HelloHandler {
	return &HelloHandler{log: log}
}

// Pattern is the path at which this handler is registered.
func (*HelloHandler) Pattern() string {
	return "/hello"
}

// ServeHTTP handles an HTTP request to the /hello endpoint.
func (h *HelloHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.log.Info("Received request for handler", zap.String("handler", h.Pattern()))
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
