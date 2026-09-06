package main

import (
	"encoding/json"
	"net/http"
)

// EvaluateHandler handles GET /flags/{key}/evaluate?user={id}.
func (s *Server) EvaluateHandler(w http.ResponseWriter, r *http.Request) {
	key := r.PathValue("key")
	user := r.URL.Query().Get("user")

	if user == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "user parameter is required"})
		return
	}

	result, err := s.store.Evaluate(key, user)
	if err != nil {
		status := http.StatusInternalServerError
		msg := "internal server error"
		if err == ErrFlagNotFound {
			status = http.StatusNotFound
			msg = "flag not found"
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		json.NewEncoder(w).Encode(map[string]string{"error": msg})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{"result": result})
}
