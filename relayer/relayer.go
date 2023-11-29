package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
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
	router.HandleFunc("/w3cli/whoami", func(w http.ResponseWriter, r *http.Request) {
		response, isJson, err := whoami()
		handleResponse(w, response, isJson, err)
	})
	router.HandleFunc("/w3cli/login", func(w http.ResponseWriter, r *http.Request) {
		response, isJson, inputError, err := login(r.URL.Query())
		if inputError {
			handleInputError(w, err)
			return
		}
		handleResponse(w, response, isJson, err)
	})
	router.HandleFunc("/w3cli/space-create", func(w http.ResponseWriter, r *http.Request) {
		response, isJson, err := spaceCreate()
		handleResponse(w, response, isJson, err)
	})
	router.HandleFunc("/w3cli/up", func(w http.ResponseWriter, r *http.Request) {
		response, isJson, inputError, err := up(r.URL.Query(), mgr.rewardsFilesPath)
		if inputError {
			handleInputError(w, err)
			return
		}
		handleResponse(w, response, isJson, err)
	})

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
			fmt.Printf("error while listening for HTTP requests: %s\n", err.Error())
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
func handleResponse(w http.ResponseWriter, response string, isJson bool, err error) {
	// Write out any errors
	if err != nil {
		w.Header().Add("Content-Type", "text/plain")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	var contentType string
	var bytes []byte
	if isJson {
		bytes, err = json.Marshal(response)
		contentType = "application/json"
	} else {
		bytes = []byte(response)
		contentType = "text/plain"
	}
	// Write the serialized response
	if err != nil {
		err = fmt.Errorf("error serializing response: %w", err)
		w.Header().Add("Content-Type", "text/plain")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
	} else {
		w.Header().Add("Content-Type", contentType)
		w.WriteHeader(http.StatusOK)
		w.Write(bytes)
	}
}

// Handles an error related to parsing the input parameters of a request
func handleInputError(w http.ResponseWriter, err error) {
	if err != nil {
		w.Header().Add("Content-Type", "text/plain")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
	}
}
