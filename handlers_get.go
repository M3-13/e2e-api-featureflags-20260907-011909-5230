package main

import (
	"encoding/json"
	"net/http"
)

// GetHandler handles GET /flags/{key}. It returns 200 with the flag JSON for a
// known key and 404 with a JSON error object for an unknown key.
func (s *Server) GetHandler(w http.ResponseWriter, r *http.Request) {
	key := r.PathValue("key")

	flag, ok := s.store.Get(key)
	if !ok {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "flag not found"})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(flag)
}
