package utils

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gofrs/uuid"
	"github.com/gorilla/mux"
)

type (
	Initialize func(srv *AppServer) (err error)
	CleanUp    func(srv *AppServer) (err error)

	// AppServer
	// Application Server object that controls the application state and life cycle.
	// It's based on the http Server from net/http package, and offers the ability to register HTTP routes.
	// The router is Handler uses the gorilla/mux implementation
	// Initialization, Shutdown and Cleanup is managed by the AppServer. Custom functions for initialization and
	// cleanup are provided so the life cycle of other objects can be added to it.
	AppServer struct {
		*http.Server
		initializeFunc Initialize // Custom initialization function
		cleanupFunc    CleanUp    // Custom cleanup function
		startTime      time.Time  // server start time for uptime reporting
		Version        string     // application version for health endpoint
		logger         Logger     // structured logger implementation
	}
)

// NewServer
// Create a new Application Server instance.
func NewServer(port int) *AppServer {

	router := mux.NewRouter()

	httpServer := &http.Server{
		Addr:              fmt.Sprintf(":%d", port),
		Handler:           router,
		ReadTimeout:       5 * time.Second,
		ReadHeaderTimeout: 2 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	server := &AppServer{
		Server: httpServer,
		// default version when not provided by main/env
		Version: "dev",
		logger:  NewStdLogger(),
	}

	return server
}

// NewServerWithInitialization
// Create a new Application Server instance, with Custom initialization and cleanup functions
func NewServerWithInitialization(port int, initializeFunc Initialize, cleanupFunc CleanUp) *AppServer {

	server := NewServer(port)

	server.initializeFunc = initializeFunc
	server.cleanupFunc = cleanupFunc

	return server
}

func (srv *AppServer) Vars(r *http.Request) map[string]string {
	return mux.Vars(r)
}

// Logger returns the configured structured logger
func (srv *AppServer) Logger() Logger {
	if srv.logger == nil {
		srv.logger = NewStdLogger()
	}
	return srv.logger
}

// SetLogger allows replacing the default logger implementation
func (srv *AppServer) SetLogger(l Logger) {
	srv.logger = l
}

func (srv *AppServer) AddRoute(path, method string, handler http.HandlerFunc) error {

	srv.router().HandleFunc(path, srv.requestInterceptor(handler)).Methods(method)

	srv.Logger().Info("route_added", Fields{"method": method, "path": path})

	return nil
}

// AddStaticRoute registers a static file server for the given path prefix and directory
func (srv *AppServer) AddStaticRoute(pathPrefix, dir string) {
	fileServer := http.FileServer(http.Dir(dir))
	srv.router().PathPrefix(pathPrefix).Handler(http.StripPrefix(pathPrefix, fileServer))

	srv.Logger().Info("static_route_added", Fields{"path_prefix": pathPrefix, "directory": dir})
}

func (srv *AppServer) requestInterceptor(next http.HandlerFunc) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		// Request ID propagation: use incoming X-Request-ID if present, else generate
		reqID := r.Header.Get("X-Request-ID")
		if reqID == "" {
			if id, err := uuid.NewV4(); err == nil {
				reqID = id.String()
			} else {
				reqID = "unknown"
			}
		}
		w.Header().Set("X-Request-ID", reqID)
		srv.Logger().Info("request", Fields{"request_id": reqID, "method": r.Method, "path": r.RequestURI})

		next.ServeHTTP(w, r)

		dur := time.Since(start)
		srv.Logger().Info("completed", Fields{"request_id": reqID, "method": r.Method, "path": r.RequestURI, "duration_ms": dur.Milliseconds()})
	}
}

// Start starts the server and blocks until ctx is cancelled, a termination signal is
// received, or an unexpected server error occurs. Pass context.Background() for
// signal-only shutdown.
func (srv *AppServer) Start(ctx context.Context) error {
	return srv.startWithCancel(ctx)
}

func (srv *AppServer) startWithCancel(ctx context.Context) error {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT)  // Handling Ctrl + C
	signal.Notify(sigChan, syscall.SIGTERM) // Handling Docker stop

	srv.Logger().Info("init_start", nil)
	if srv.initializeFunc != nil {
		if err := srv.initializeFunc(srv); err != nil {
			return fmt.Errorf("initialization failed: %w", err)
		}
	}

	srv.Logger().Info("server_starting", nil)
	srv.startTime = time.Now()

	errChan := make(chan error, 1)
	go func() {
		srv.Logger().Info("listening", Fields{"addr": srv.Addr})
		_, _ = fmt.Fprintf(os.Stdout, "\n  App running at http://localhost%s\n\n", srv.Addr)
		err := srv.ListenAndServe()
		if err == http.ErrServerClosed {
			errChan <- nil
		} else {
			errChan <- err
		}
	}()

	select {
	case err := <-errChan:
		if err != nil {
			return fmt.Errorf("server error: %w", err)
		}
	case <-sigChan:
	case <-ctx.Done():
	}

	return srv.prepareShutdown()
}

func (srv *AppServer) prepareShutdown() error {
	srv.Logger().Info("cleanup_start", nil)

	if srv.cleanupFunc != nil {
		if err := srv.cleanupFunc(srv); err != nil {
			srv.Logger().Error("cleanup_error", Fields{"error": err.Error()})
		}
	}

	srv.Logger().Info("server_shutting_down", nil)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		return fmt.Errorf("shutdown error: %w", err)
	}
	srv.Logger().Info("server_stopped", nil)
	return nil
}

func (srv *AppServer) router() *mux.Router {

	return srv.Handler.(*mux.Router)
}

// StartTime returns the time the server was started.
func (srv *AppServer) StartTime() time.Time {
	return srv.startTime
}

func (srv *AppServer) ResponseErrorEntityUnproc(response http.ResponseWriter, err error) {
	srv.Logger().Error("error_unprocessable_entity", Fields{"error": err.Error()})
	response.Header().Set("Content-Type", "application/json")
	response.WriteHeader(http.StatusUnprocessableEntity)
	_, _ = fmt.Fprintf(response, "{\"error\":\"%s\"}", err)
}

func (srv *AppServer) ResponseErrorServerErr(response http.ResponseWriter, err error) {
	srv.Logger().Error("error_internal_server", Fields{"error": err.Error()})
	response.Header().Set("Content-Type", "application/json")
	response.WriteHeader(http.StatusInternalServerError)
	_, _ = fmt.Fprintf(response, "{\"error\":\"%s\"}", err)
}

func (srv *AppServer) ResponseErrorNotfound(response http.ResponseWriter, err error) {
	srv.Logger().Error("error_not_found", Fields{"error": err.Error()})
	response.Header().Set("Content-Type", "application/json")
	response.WriteHeader(http.StatusNotFound)
	_, _ = fmt.Fprintf(response, "{\"error\":\"%s\"}", err)
}

// RespondJSON writes a JSON response with the given status code. It ensures headers are set before body
// and centralizes JSON encoding and error handling.
func (srv *AppServer) RespondJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	enc := json.NewEncoder(w)
	if err := enc.Encode(v); err != nil {
		// we cannot change status code here as headers are already written; log the error
		srv.Logger().Error("json_encode_error", Fields{"error": err.Error()})
	}
}
