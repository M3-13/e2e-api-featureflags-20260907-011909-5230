package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"regexp"
)

var keyPattern = regexp.MustCompile(`^[A-Za-z0-9._-]+$`)

// CreateHandler handles POST /flags.
func (s *Server) CreateHandler(w http.ResponseWriter, r *http.Request) {
	var f Flag
	if err := json.NewDecoder(r.Body).Decode(&f); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "invalid JSON body"})
		return
	}

	if f.Key == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "key is required"})
		return
	}

	if len(f.Key) > 128 || !keyPattern.MatchString(f.Key) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "invalid key"})
		return
	}

	if f.RolloutPercent < 0 || f.RolloutPercent > 100 {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "rollout_percent must be between 0 and 100"})
		return
	}

	created, err := s.store.Create(f)
	if err != nil {
		if errors.Is(err, ErrFlagExists) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusConflict)
			json.NewEncoder(w).Encode(map[string]string{"error": "flag already exists"})
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "internal error"})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(created)
}
