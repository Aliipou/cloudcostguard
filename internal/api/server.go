package api

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/Aliipou/cloudcostguard/internal/config"
	"github.com/Aliipou/cloudcostguard/internal/engine"
	"github.com/Aliipou/cloudcostguard/internal/report"
	"github.com/Aliipou/cloudcostguard/web"
)

// Server is the HTTP API server for CloudCostGuard.
type Server struct {
	configFile string
	mux        *http.ServeMux
	port       int
}

// NewServer creates a new API server.
func NewServer(configFile string, port int) *Server {
	s := &Server{
		configFile: configFile,
		mux:        http.NewServeMux(),
		port:       port,
	}
	s.routes()
	return s
}

func (s *Server) routes() {
	s.mux.HandleFunc("/api/health", s.handleHealth)
	s.mux.HandleFunc("/api/scan", s.handleScan)
	s.mux.HandleFunc("/api/providers", s.handleProviders)
	s.mux.HandleFunc("/", s.handleDashboard)
}

// Start begins listening on the configured port.
func (s *Server) Start() error {
	addr := fmt.Sprintf(":%d", s.port)
	log.Printf("CloudCostGuard dashboard available at http://localhost%s\n", addr)
	return http.ListenAndServe(addr, s.mux)
}

// Handler returns the underlying http.Handler (useful for testing).
func (s *Server) Handler() http.Handler {
	return s.mux
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func (s *Server) handleProviders(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	providers := []map[string]string{
		{"id": "aws", "name": "Amazon Web Services"},
		{"id": "azure", "name": "Microsoft Azure"},
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(providers)
}

func (s *Server) handleScan(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	provider := r.URL.Query().Get("provider")
	if provider == "" {
		provider = "aws"
	}
	scanType := r.URL.Query().Get("type")
	if scanType == "" {
		scanType = "all"
	}

	cfg, err := config.Load(s.configFile)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to load config: "+err.Error())
		return
	}
	cfg.Provider = provider

	eng, err := engine.New(cfg, engine.Options{
		ScanType:    scanType,
		MinSavings:  0,
		Concurrency: 5,
		Verbose:     false,
	})
	if err != nil {
		writeError(w, http.StatusBadRequest, "failed to initialize scan engine: "+err.Error())
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Minute)
	defer cancel()

	findings, err := eng.Scan(ctx)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "scan failed: "+err.Error())
		return
	}

	result := report.ToJSON(findings)
	w.Header().Set("Content-Type", "application/json")
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	enc.Encode(result)
}

func (s *Server) handleDashboard(w http.ResponseWriter, r *http.Request) {
	// Serve files from the embedded web directory.
	http.FileServer(http.FS(web.StaticFS)).ServeHTTP(w, r)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{"error": msg})
}
