package utils

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gofrs/uuid"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
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

func (srv *AppServer) Start() {
	// Backward-compatible start that listens for process signals only
	srv.startWithCancel(nil)
}

// StartContext starts the server and will stop when either the provided context is
// cancelled or termination signals are received. If ctx is nil, only signals are observed.
func (srv *AppServer) StartContext(ctx context.Context) {
	srv.startWithCancel(ctx)
}

func (srv *AppServer) startWithCancel(ctx context.Context) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT)  // Handling Ctrl + C
	signal.Notify(sigChan, syscall.SIGTERM) // Handling Docker stop

	srv.Logger().Info("init_start", nil)
	if srv.initializeFunc != nil {
		if err := srv.initializeFunc(srv); err != nil {
			srv.Logger().Error("init_failed", Fields{"error": err.Error()})
		}
	}

	srv.Logger().Info("server_starting", nil)
	// mark start time for uptime
	srv.startTime = time.Now()

	done := make(chan struct{})
	go func() {
		srv.Logger().Info("listening", Fields{"addr": srv.Addr})
		err := srv.ListenAndServe()
		if err != http.ErrServerClosed {
			// still fatal-exit on unexpected error
			log.Fatalf("Failed to start server, %s", err)
		}
		close(done)
	}()

	// Wait for either context cancellation, signal, or server close
	select {
	case <-done:
		// server closed
	case <-sigChan:
		// signal received
	case <-func() <-chan struct{} {
		if ctx == nil {
			// never triggers
			ch := make(chan struct{})
			return ch
		}
		return ctx.Done()
	}():
	}

	srv.prepareShutdown()
}

func (srv *AppServer) prepareShutdown() {

	srv.Logger().Info("cleanup_start", nil)

	if srv.cleanupFunc != nil {
		err := srv.cleanupFunc(srv)

		if err != nil {
			srv.Logger().Error("cleanup_error", Fields{"error": err.Error()})
		}
	}

	srv.Logger().Info("server_shutting_down", nil)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := srv.Shutdown(ctx)

	if err != nil {
		log.Fatalf("Error shutting down server, %s", err)
	} else {
		srv.Logger().Info("server_stopped", nil)
	}
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
	_, _ = response.Write([]byte(fmt.Sprintf("{\"error\":\"%s\"}", err)))
}

func (srv *AppServer) ResponseErrorServerErr(response http.ResponseWriter, err error) {
	srv.Logger().Error("error_internal_server", Fields{"error": err.Error()})
	response.Header().Set("Content-Type", "application/json")
	response.WriteHeader(http.StatusInternalServerError)
	_, _ = response.Write([]byte(fmt.Sprintf("{\"error\":\"%s\"}", err)))
}

func (srv *AppServer) ResponseErrorNotfound(response http.ResponseWriter, err error) {
	srv.Logger().Error("error_not_found", Fields{"error": err.Error()})
	response.Header().Set("Content-Type", "application/json")
	response.WriteHeader(http.StatusNotFound)
	_, _ = response.Write([]byte(fmt.Sprintf("{\"error\":\"%s\"}", err)))
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
