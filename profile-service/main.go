package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
)

type profileResponse struct {
	ID       string `json:"id"`
	FullName string `json:"fullName"`
	Level    string `json:"level"`
	Service  string `json:"service"`
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/profile/", handleProfile)

	server := &http.Server{
		Addr:    ":8081",
		Handler: mux,
	}

	log.Println("profile-service listening on :8081")
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("profile-service failed: %v", err)
	}
}

func handleProfile(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	id := strings.TrimPrefix(r.URL.Path, "/profile/")
	if id == "" || strings.Contains(id, "/") {
		http.Error(w, "profile id is required", http.StatusBadRequest)
		return
	}

	response := profileResponse{
		ID:       id,
		FullName: "Test User " + id,
		Level:    "basic",
		Service:  "profile-service",
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
	}
}
