package main

import (
	"encoding/json"
	"net/http"
)

// DeleteHandler handles DELETE /flags/{key}. It removes the flag and responds
// 204 without a body; an unknown key responds 404 with a JSON error object.
func (s *Server) DeleteHandler(w http.ResponseWriter, r *http.Request) {
	key := r.PathValue("key")

	if err := s.store.Delete(key); err != nil {
		if err == ErrFlagNotFound {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]string{"error": "flag not found"})
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "internal server error"})
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
