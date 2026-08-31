// Package main implements the song project as a magic link authentication server
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"time"

	"azzurrotech/song/pkg/auth"
)

func main() {
	port := flag.String("port", "8080", "Port to listen on")
	secretKey := flag.String("secret", "default-secret-key-32-chars-long-123", "Secret key for encryption")
	help := flag.Bool("help", false, "Show help message")
	version := flag.Bool("version", false, "Show version information")

	flag.Parse()

	if *help {
		fmt.Println("Usage: song [options]")
		fmt.Println("  --port     Set the port to listen on (default: 8080)")
		fmt.Println("  --secret   Secret key for encryption (must be at least 32 chars)")
		fmt.Println("  --help     Show this help message")
		fmt.Println("  --version  Show version information")
		return
	}

	if *version {
		fmt.Println("Song Server v1.0.0")
		fmt.Println("Magic Link Authentication Service")
		fmt.Println("Copyright 2025 Azzurro Technology Inc.")
		return
	}

	authService, err := auth.NewAuthService(*secretKey)
	if err != nil {
		log.Fatalf("Failed to create auth service: %v", err)
	}

	server := &Server{authService: authService}
	if err := server.Start(*port); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}

type Server struct {
	authService *auth.AuthService
}

func (s *Server) Start(port string) error {
	fmt.Printf("Starting SONG server on port %s\n", port)
	fmt.Println("Magic Link Authentication Service")

	// Set up authentication routes
	http.HandleFunc("/api/auth/generate", s.generateMagicLinkHandler)
	http.HandleFunc("/api/auth/validate", s.validateMagicLinkHandler)
	http.HandleFunc("/api/auth/revoke", s.revokeMagicLinkHandler)

	// Set up health check route
	http.HandleFunc("/health", s.healthCheckHandler)

	return http.ListenAndServe(":"+port, http.DefaultServeMux)
}

func (s *Server) generateMagicLinkHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req auth.GenerateLinkRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON: "+err.Error(), http.StatusBadRequest)
		return
	}

	if req.UserID == "" {
		http.Error(w, "user_id is required", http.StatusBadRequest)
		return
	}

	magicLink, err := s.authService.GenerateMagicLink(req.UserID, req.DeviceInfo)
	if err != nil {
		http.Error(w, "Failed to generate magic link: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(auth.GenerateLinkResponse{
		MagicLink: magicLink.Token,
		ExpiresAt: magicLink.Expiry,
	})
}

func (s *Server) validateMagicLinkHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req auth.ValidateLinkRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON: "+err.Error(), http.StatusBadRequest)
		return
	}

	if req.Link == "" || req.UserID == "" {
		http.Error(w, "link and user_id are required", http.StatusBadRequest)
		return
	}

	magicLink, err := s.authService.ValidateMagicLink(req.Link, req.UserID, req.DeviceInfo)
	if err != nil {
		http.Error(w, "Invalid magic link: "+err.Error(), http.StatusUnauthorized)
		return
	}

	if magicLink.Used {
		http.Error(w, "Magic link already used", http.StatusNotFound)
		return
	}

	if time.Now().After(magicLink.Expiry) {
		s.authService.MarkLinkAsUsed(req.Link)
		http.Error(w, "Magic link expired", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(auth.ValidateLinkResponse{
		Valid:     true,
		UserID:    magicLink.UserID,
		ExpiresAt: magicLink.Expiry,
	})
}

func (s *Server) revokeMagicLinkHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Link string `json:"link"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON: "+err.Error(), http.StatusBadRequest)
		return
	}

	if req.Link == "" {
		http.Error(w, "link is required", http.StatusBadRequest)
		return
	}

	if err := s.authService.MarkLinkAsUsed(req.Link); err != nil {
		http.Error(w, "Failed to revoke magic link: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Magic link revoked successfully",
	})
}

func (s *Server) healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now(),
		"service":   "song-auth",
		"version":   "1.0.0",
	})
}
