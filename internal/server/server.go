package server

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"litemidgo/config"
	"litemidgo/internal/servicenow"
)

type Server struct {
	config     *config.Config
	snowClient *servicenow.Client
	httpServer *http.Server
}

type ProxyRequest struct {
	Agent   string      `json:"agent"`
	Topic   string      `json:"topic"`
	Name    string      `json:"name"`
	Source  string      `json:"source"`
	Payload interface{} `json:"payload"`
}

type ProxyResponse struct {
	Success   bool   `json:"success"`
	Message   string `json:"message"`
	SysID     string `json:"sys_id,omitempty"`
	Timestamp string `json:"timestamp"`
}

func NewServer(cfg *config.Config) *Server {
	snowClient := servicenow.NewClient(&cfg.ServiceNow)

	return &Server{
		config:     cfg,
		snowClient: snowClient,
	}
}

func (s *Server) Start() error {
	// Test ServiceNow connection before starting
	if err := s.snowClient.TestConnection(); err != nil {
		return fmt.Errorf("ServiceNow connection test failed: %w", err)
	}

	log.Printf("✓ ServiceNow connection established to %s", s.snowClient.GetInstanceURL())

	// Setup HTTP routes
	mux := http.NewServeMux()
	mux.HandleFunc("/health", s.handleHealth)
	mux.HandleFunc("/proxy/ecc_queue", s.handleECCQueueProxy)
	mux.HandleFunc("/", s.handleDefault)

	// Create HTTP server
	s.httpServer = &http.Server{
		Addr:         fmt.Sprintf("%s:%d", s.config.Server.Host, s.config.Server.Port),
		Handler:      mux,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	log.Printf("🚀 Starting LiteMIDgo server on %s:%d", s.config.Server.Host, s.config.Server.Port)
	log.Printf("📡 Available endpoints:")
	log.Printf("   - GET  /health - Health check")
	log.Printf("   - POST /proxy/ecc_queue - Proxy to ServiceNow ECC Queue")
	log.Printf("   - GET  / - Server information")

	return s.httpServer.ListenAndServe()
}

func (s *Server) Stop() error {
	if s.httpServer != nil {
		return s.httpServer.Close()
	}
	return nil
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Test ServiceNow connection
	if err := s.snowClient.TestConnection(); err != nil {
		response := ProxyResponse{
			Success:   false,
			Message:   fmt.Sprintf("ServiceNow connection failed: %v", err),
			Timestamp: time.Now().UTC().Format(time.RFC3339),
		}
		s.writeJSONResponse(w, http.StatusServiceUnavailable, response)
		return
	}

	response := ProxyResponse{
		Success:   true,
		Message:   "Service is healthy and ServiceNow connection is active",
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}
	s.writeJSONResponse(w, http.StatusOK, response)
}

func (s *Server) handleECCQueueProxy(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse request body
	var proxyReq ProxyRequest
	if err := json.NewDecoder(r.Body).Decode(&proxyReq); err != nil {
		response := ProxyResponse{
			Success:   false,
			Message:   fmt.Sprintf("Invalid JSON payload: %v", err),
			Timestamp: time.Now().UTC().Format(time.RFC3339),
		}
		s.writeJSONResponse(w, http.StatusBadRequest, response)
		return
	}

	// Validate required fields
	if proxyReq.Agent == "" {
		proxyReq.Agent = "litemidgo"
	}
	if proxyReq.Topic == "" {
		proxyReq.Topic = "endpointData"
	}
	if proxyReq.Name == "" {
		proxyReq.Name = "default"
	}
	if proxyReq.Source == "" {
		proxyReq.Source = r.RemoteAddr
	}

	// Create ECC Queue payload
	eccPayload := &servicenow.ECCQueuePayload{
		Agent:   proxyReq.Agent,
		Topic:   proxyReq.Topic,
		Name:    proxyReq.Name,
		Source:  proxyReq.Source,
		Payload: proxyReq.Payload,
	}

	// Send to ServiceNow
	resp, err := s.snowClient.SendToECCQueue(eccPayload)
	if err != nil {
		response := ProxyResponse{
			Success:   false,
			Message:   fmt.Sprintf("Failed to send to ServiceNow: %v", err),
			Timestamp: time.Now().UTC().Format(time.RFC3339),
		}
		s.writeJSONResponse(w, http.StatusInternalServerError, response)
		return
	}

	// Return success response
	response := ProxyResponse{
		Success:   true,
		Message:   "Successfully queued in ServiceNow ECC Queue",
		SysID:     resp.Result.SysID,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}
	s.writeJSONResponse(w, http.StatusOK, response)
}

func (s *Server) handleDefault(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	info := map[string]interface{}{
		"service":     "LiteMIDgo",
		"version":     "1.0.0",
		"description": "Lightweight ServiceNow MID Server proxy",
		"endpoints": map[string]string{
			"health":     "/health",
			"ecc_queue":  "/proxy/ecc_queue",
			"servicenow": s.snowClient.GetInstanceURL(),
		},
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	}

	s.writeJSONResponse(w, http.StatusOK, info)
}

func (s *Server) writeJSONResponse(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Printf("Failed to encode JSON response: %v", err)
	}
}
