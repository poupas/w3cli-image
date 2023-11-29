package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"sync"

	"github.com/gorilla/mux"
)

// Simple class for managing an HTTP server over a Unix socket
type relayManager struct {
	socketPath       string
	rewardsFilesPath string
	socket           net.Listener
	server           http.Server
	router           *mux.Router
}

// Creates a new relay manager
func newRelayManager(socketPath string, rewardsFilesPath string) *relayManager {
	// Create the router
	router := mux.NewRouter()

	// Create the manager
	mgr := &relayManager{
		socketPath:       socketPath,
		rewardsFilesPath: rewardsFilesPath,
		router:           router,
		server: http.Server{
			Handler: router,
		},
	}

	// Register the routes
	w3cliRouter := router.Host("w3cli").Subrouter()
	w3cliRouter.HandleFunc("/whoami", func(w http.ResponseWriter, r *http.Request) {
		response, isJson, err := whoami()
		handleResponse(r, w, response, isJson, err)
	}).Methods(http.MethodGet)
	w3cliRouter.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		response, isJson, inputError, err := login(r.URL.Query())
		if inputError {
			handleInputError(r, w, err)
			return
		}
		handleResponse(r, w, response, isJson, err)
	}).Methods(http.MethodGet)
	w3cliRouter.HandleFunc("/space-create", func(w http.ResponseWriter, r *http.Request) {
		response, isJson, err := spaceCreate()
		handleResponse(r, w, response, isJson, err)
	}).Methods(http.MethodGet)
	w3cliRouter.HandleFunc("/up", func(w http.ResponseWriter, r *http.Request) {
		response, isJson, inputError, err := up(r.URL.Query(), mgr.rewardsFilesPath)
		if inputError {
			handleInputError(r, w, err)
			return
		}
		handleResponse(r, w, response, isJson, err)
	}).Methods(http.MethodGet)

	return mgr
}

// Starts listening for incoming HTTP requests
func (m *relayManager) start(wg *sync.WaitGroup) error {
	// Create the socket
	socket, err := net.Listen("unix", m.socketPath)
	if err != nil {
		return fmt.Errorf("error creating socket: %w", err)
	}
	m.socket = socket

	// Start listening
	go func() {
		err := m.server.Serve(socket)
		if !errors.Is(err, http.ErrServerClosed) {
			fmt.Printf("Error while listening for HTTP requests: %s\n", err.Error())
		}
		wg.Done()
	}()

	return nil
}

// Stops the HTTP listener
func (m *relayManager) stop() error {
	err := m.server.Shutdown(context.Background())
	if err != nil {
		return fmt.Errorf("error stopping listener: %w", err)
	}
	return nil
}

// Handles a response to a request
func handleResponse(r *http.Request, w http.ResponseWriter, response string, isJson bool, err error) {
	// Write out any errors
	if err != nil {
		// Log the error
		logLine("Error handling '%s': %s'", r.URL.String(), err.Error())

		// Write the request
		w.Header().Add("Content-Type", "text/plain")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	// Get the content type
	var contentType string
	bytes := []byte(response)
	if isJson {
		contentType = "application/json"
	} else {
		contentType = "text/plain"
	}

	// Log the request
	logLine("Handled '%s'", r.URL.String())

	// Write the response
	w.Header().Add("Content-Type", contentType)
	w.WriteHeader(http.StatusOK)
	w.Write(bytes)
}

// Handles an error related to parsing the input parameters of a request
func handleInputError(r *http.Request, w http.ResponseWriter, err error) {
	if err != nil {
		// Log the error
		logLine("Invalid input from '%s': %s'", r.URL.String(), err.Error())

		// Write the response
		w.Header().Add("Content-Type", "text/plain")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
	}
}

// Log a line to the local console
func logLine(line string, args ...any) {
	log.Println(fmt.Sprintf(line, args...))
}
