package server

import (
	"context"
	"encoding/json"
	"feedexampleredis/internal/app"
	"feedexampleredis/internal/storage"
	"fmt"
	"github.com/gorilla/mux"
	"log/slog"
	"net"
	"net/http"
)

// Server represents the API server.
type Server struct {
	cfg app.Config
	r   *storage.Redis
	v6  *storage.MMDB
}

// NewServer creates a new Server instance.
func NewServer(cfg app.Config, r *storage.Redis, v6 *storage.MMDB) *Server {
	return &Server{
		cfg: cfg,
		r:   r,
		v6:  v6,
	}
}

// authenticateMiddleware checks for the presence and validity of a TOKEN header.
func (s *Server) authenticateMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("TOKEN")
		for _, validToken := range s.cfg.LocalAPIAuthTokens {
			if token == validToken {
				next.ServeHTTP(w, r)
				return
			}
		}
		http.Error(w, "Forbidden", http.StatusForbidden)
	})
}

// handleContext is the handler for the /v2/context/{ipAddress} endpoint.
func (s *Server) handleContext(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	ipAddress := vars["ipAddress"]

	slog.Info("received request", "ip_address", ipAddress)

	// Validate the IP address
	if ipAddress == "" {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	parsedIP := net.ParseIP(ipAddress)
	if parsedIP == nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	var response []byte
	if parsedIP.To4() != nil {
		// Query redis for the IP context
		ipContext, err := s.r.GetByIP(r.Context(), ipAddress)
		if err != nil {
			http.Error(w, "Not Found", http.StatusNotFound)
			return
		}

		// If there is no ip in the context, return a 404
		if ipContext == nil {
			http.Error(w, "Not Found", http.StatusNotFound)
			return
		}

		if ipContext.IP == "" {
			http.Error(w, "Not Found", http.StatusNotFound)
			return
		}

		// Return the IP context as JSON
		response, err = json.Marshal(ipContext)
		if err != nil {
			slog.Error("error marshalling IP context", "error", err.Error())
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
	} else {
		// Query the MMDB for the IP context
		ipContext, err := s.v6.GetIP(parsedIP)
		if err != nil {
			http.Error(w, "Not Found", http.StatusNotFound)
			return
		}

		// If there is no ip in the context, return a 404
		if ipContext == nil {
			http.Error(w, "Not Found", http.StatusNotFound)
			return
		}

		if ipContext.Network == "" {
			http.Error(w, "Not Found", http.StatusNotFound)
			return
		}

		// Return the IP context as JSON
		response, err = json.Marshal(ipContext)
		if err != nil {
			slog.Error("error marshalling IP context", "error", err.Error())
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
	}

	// set the content type
	w.Header().Set("Content-Type", "application/json")

	w.WriteHeader(http.StatusOK)
	w.Write(response)
}

// Start starts the API server.
func (s *Server) Start(ctx context.Context) error {
	r := mux.NewRouter()
	r.Handle("/v2/context/{ipAddress}", s.authenticateMiddleware(http.HandlerFunc(s.handleContext))).Methods("GET")
	address := fmt.Sprintf(":%d", s.cfg.Port)
	srv := &http.Server{
		Addr:    address,
		Handler: r,
	}
	go func() {
		slog.Info("Starting HTTP server", "address", address)
		if err := srv.ListenAndServe(); err != nil {
			slog.Error("error starting HTTP server", "error", err.Error())
		}
	}()

	<-ctx.Done()

	return ctx.Err()
}

// StartTLS starts the API server with TLS (HTTPS).
func (s *Server) StartTLS(ctx context.Context) error {
	r := mux.NewRouter()
	r.Handle("/v2/context/{ipAddress}", s.authenticateMiddleware(http.HandlerFunc(s.handleContext))).Methods("GET")
	address := fmt.Sprintf(":%d", s.cfg.Port)
	srv := &http.Server{
		Addr:    address,
		Handler: r,
	}

	go func() {
		slog.Info("Starting HTTPS server", "address", address)
		// Start the HTTPS server
		if err := srv.ListenAndServeTLS(s.cfg.CertFile, s.cfg.KeyFile); err != nil {
			slog.Error("error starting HTTPS server", "error", err.Error())
		}
	}()

	<-ctx.Done()
	return ctx.Err()
}
