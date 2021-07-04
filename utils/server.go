package utils

import (
	"context"
	"fmt"
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
		initializeFunc Initialize        // Custom initialization function
		cleanupFunc    CleanUp           // Custom cleanup function
	}
)

// NewServer
// Create a new Application Server instance.
func NewServer(port int) *AppServer {

	router := mux.NewRouter()

	httpServer := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: router,
	}

	server := &AppServer{
		Server:      httpServer,
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


func (srv *AppServer) AddRoute(path, method string, handler http.HandlerFunc) error {

	srv.router().HandleFunc(path, srv.requestInterceptor(handler)).Methods(method)

	log.Printf("Added route %s %s", method, path)

	return nil
}

func (srv *AppServer) requestInterceptor(next http.HandlerFunc) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {

		log.Printf("Request %s %s", r.Method, r.RequestURI)

		next.ServeHTTP(w, r)

	}
}

func (srv *AppServer) Start() {

	sigChan := make(chan os.Signal, 1)

	signal.Notify(sigChan, syscall.SIGINT)  // Handling Ctrl + C
	signal.Notify(sigChan, syscall.SIGTERM) // Handling Docker stop

	log.Printf("Initializing resources...")

	if srv.initializeFunc != nil {

		err := srv.initializeFunc(srv)

		if err != nil {
			log.Printf("Failed to initialize resources, %s", err)
		}
	}

	log.Printf("Starting app server...")

	go func() {
		log.Printf("Listening on port %s. Ctrl+C to stop", srv.Addr)

		err := srv.ListenAndServe()

		if err != http.ErrServerClosed {
			log.Fatalf("Failed to start server, %s", err)
		}
	}()

	<-sigChan

	srv.prepareShutdown()

}

func (srv *AppServer) prepareShutdown() {

	log.Printf("Cleaning up resources...")

	if srv.cleanupFunc != nil {
		err := srv.cleanupFunc(srv)

		if err != nil {
			log.Printf("Error cleaning up resources ,%s", err)
		}
	}

	log.Printf("Shutting down app server...")

	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)

	err := srv.Shutdown(ctx)

	if err != nil {
		log.Fatalf("Error shutting down server, %s", err)
	} else {
		log.Printf("App server gracefully stopped")
	}
}

func (srv *AppServer) router() *mux.Router {

	return srv.Handler.(*mux.Router)
}

func (srv *AppServer) ResponseErrorEntityUnproc(response http.ResponseWriter, err error) {
	log.Printf("%s", err)
	response.Header().Set("Content-Type", "application/json")
	response.WriteHeader(http.StatusUnprocessableEntity)
	_,_ = response.Write([]byte(fmt.Sprintf("{\"error\":\"%s\"}", err)))
}

func (srv *AppServer) ResponseErrorServerErr(response http.ResponseWriter, err error) {
	log.Printf("%s", err)
	response.WriteHeader(http.StatusInternalServerError)
}

func (srv *AppServer) ResponseErrorNotfound(response http.ResponseWriter, err error) {
	log.Printf("%s", err)
	response.Header().Set("Content-Type", "application/json")
	response.WriteHeader(http.StatusNotFound)
	_,_ = response.Write([]byte(fmt.Sprintf("{\"error\":\"%s\"}", err)))
}
