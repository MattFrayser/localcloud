package web

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strings"
	"time"

	"localcloud/internal/compute"
)

type Server struct {
	computeManager *compute.InstanceManager
	templates      *template.Template
	clients        map[chan []compute.Instance]bool
}

func NewServer(computeManager *compute.InstanceManager) (*Server, error) {
	templates := template.Must(template.ParseGlob("internal/web/templates/*.html"))

	return &Server{
		computeManager: computeManager,
		templates:      templates,
		clients:        make(map[chan []compute.Instance]bool),
	}, nil
}

func (s *Server) Start(addr string) error {
	// Create a new mux to handle all routes
	mux := http.NewServeMux()

	// Serve static files
	fs := http.FileServer(http.Dir("internal/web/static"))
	mux.Handle("/static/", http.StripPrefix("/static/", fs))

	// API endpoints
	mux.HandleFunc("/api/containers", s.handleContainers)
	mux.HandleFunc("/api/containers/", s.handleContainerLogs)
	mux.HandleFunc("/events", s.handleEvents)

	// Web interface
	mux.HandleFunc("/", s.handleIndex)

	// Wrap the mux with logging middleware
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Incoming request: %s %s", r.Method, r.URL.Path)
		mux.ServeHTTP(w, r)
	})

	log.Printf("Starting web interface on %s", addr)
	return http.ListenAndServe(addr, handler)
}

func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	s.templates.ExecuteTemplate(w, "index.html", nil)
}

func (s *Server) handleContainers(w http.ResponseWriter, r *http.Request) {
	instances := s.computeManager.ListInstances()
	json.NewEncoder(w).Encode(map[string]interface{}{
		"containers": instances,
	})
}

func (s *Server) handleContainerLogs(w http.ResponseWriter, r *http.Request) {
	// Extract container ID from URL
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 4 {
		http.Error(w, "Invalid container ID", http.StatusBadRequest)
		return
	}
	containerID := parts[3]

	// Remove any trailing path components (like /logs)
	if idx := strings.Index(containerID, "/"); idx != -1 {
		containerID = containerID[:idx]
	}

	switch {
	case r.Method == "GET" && strings.HasSuffix(r.URL.Path, "/logs"):
		// Handle logs retrieval
		logs, err := s.computeManager.GetLogs(r.Context(), containerID, time.Time{}, "100")
		if err != nil {
			http.Error(w, "Failed to get container logs", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte(logs))

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Server) handleEvents(w http.ResponseWriter, r *http.Request) {
	// Set headers for SSE
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	// Create a channel for this client
	clientChan := make(chan []compute.Instance)
	s.clients[clientChan] = true

	// Remove client when they disconnect
	defer func() {
		delete(s.clients, clientChan)
		close(clientChan)
	}()

	// Send initial state
	instances := s.computeManager.ListInstances()
	initialState := map[string]interface{}{
		"containers": instances,
	}
	data, _ := json.Marshal(initialState)
	fmt.Fprintf(w, "data: %s\n\n", data)
	w.(http.Flusher).Flush()

	// Keep connection alive
	for {
		select {
		case instances := <-clientChan:
			update := map[string]interface{}{
				"containers": instances,
			}
			data, _ := json.Marshal(update)
			fmt.Fprintf(w, "data: %s\n\n", data)
			w.(http.Flusher).Flush()
		case <-r.Context().Done():
			return
		}
	}
}
