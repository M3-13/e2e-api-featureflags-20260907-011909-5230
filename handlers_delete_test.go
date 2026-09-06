package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestDeleteFlagRemovesFlag(t *testing.T) {
	store := NewStore()
	store.flags["mykey"] = Flag{Key: "mykey", Enabled: true}
	server := NewServer(store)

	req := httptest.NewRequest(http.MethodDelete, "/flags/mykey", nil)
	req.SetPathValue("key", "mykey")
	rr := httptest.NewRecorder()
	server.DeleteHandler(rr, req)

	if rr.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", rr.Code)
	}
	if rr.Body.Len() != 0 {
		t.Fatalf("expected empty body, got %q", rr.Body.String())
	}

	store.mu.RLock()
	_, exists := store.flags["mykey"]
	store.mu.RUnlock()
	if exists {
		t.Fatalf("expected flag to be removed from the store")
	}
}

func TestDeleteFlagUnknownKey(t *testing.T) {
	store := NewStore()
	server := NewServer(store)

	req := httptest.NewRequest(http.MethodDelete, "/flags/unknown", nil)
	req.SetPathValue("key", "unknown")
	rr := httptest.NewRecorder()
	server.DeleteHandler(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rr.Code)
	}

	var body map[string]string
	if err := json.NewDecoder(rr.Body).Decode(&body); err != nil {
		t.Fatalf("failed to decode response body: %v", err)
	}
	if body["error"] == "" {
		t.Fatalf("expected non-empty error field, got %q", body["error"])
	}
}
