package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func newGetServer() *Server {
	store := NewStore()
	store.mu.Lock()
	store.flags["feature-x"] = Flag{
		Key:            "feature-x",
		Enabled:        true,
		Description:    "a test flag",
		RolloutPercent: 50,
	}
	store.mu.Unlock()
	return NewServer(store)
}

func TestGetFlagExisting(t *testing.T) {
	server := newGetServer()

	mux := http.NewServeMux()
	mux.HandleFunc("GET /flags/{key}", server.GetHandler)

	req := httptest.NewRequest(http.MethodGet, "/flags/feature-x", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rr.Code)
	}

	var flag Flag
	if err := json.NewDecoder(rr.Body).Decode(&flag); err != nil {
		t.Fatalf("failed to decode response body: %v", err)
	}
	if flag.Key != "feature-x" {
		t.Fatalf("expected key feature-x, got %q", flag.Key)
	}
	if !flag.Enabled {
		t.Fatalf("expected enabled true, got false")
	}
	if flag.Description != "a test flag" {
		t.Fatalf("expected description %q, got %q", "a test flag", flag.Description)
	}
	if flag.RolloutPercent != 50 {
		t.Fatalf("expected rollout_percent 50, got %d", flag.RolloutPercent)
	}
}

func TestGetFlagUnknown(t *testing.T) {
	server := newGetServer()

	mux := http.NewServeMux()
	mux.HandleFunc("GET /flags/{key}", server.GetHandler)

	req := httptest.NewRequest(http.MethodGet, "/flags/missing", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d", rr.Code)
	}

	var body map[string]string
	if err := json.NewDecoder(rr.Body).Decode(&body); err != nil {
		t.Fatalf("failed to decode response body: %v", err)
	}
	if _, ok := body["error"]; !ok {
		t.Fatalf("expected error field in response body, got %v", body)
	}
}
