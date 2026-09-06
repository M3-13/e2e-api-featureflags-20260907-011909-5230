package main

import (
	"encoding/json"
	"net/http"
)

// ListHandler handles GET /flags, returning all stored flags as a JSON array.
func (s *Server) ListHandler(w http.ResponseWriter, r *http.Request) {
	flags := s.store.List()
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(flags); err != nil {
		http.Error(w, `{"error":"failed to encode response"}`, http.StatusInternalServerError)
	}
}
