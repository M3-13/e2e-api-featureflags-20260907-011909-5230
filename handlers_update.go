package main

import (
	"encoding/json"
	"errors"
	"net/http"
)

// UpdateHandler handles PUT /flags/{key}. It decodes the JSON body
// (enabled, description, rollout_percent; missing fields are zero values),
// updates the flag and responds 200 with the updated flag JSON. An unknown
// key answers 404.
func (s *Server) UpdateHandler(w http.ResponseWriter, r *http.Request) {
	key := r.PathValue("key")

	var body struct {
		Enabled        bool   `json:"enabled"`
		Description    string `json:"description"`
		RolloutPercent int    `json:"rollout_percent"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "invalid JSON body"})
		return
	}

	if body.RolloutPercent < 0 || body.RolloutPercent > 100 {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "rollout_percent must be between 0 and 100"})
		return
	}

	updated, err := s.store.Update(key, Flag{
		Key:            key,
		Enabled:        body.Enabled,
		Description:    body.Description,
		RolloutPercent: body.RolloutPercent,
	})
	if err != nil {
		if errors.Is(err, ErrFlagNotFound) {
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

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(updated)
}
