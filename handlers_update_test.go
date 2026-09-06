package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestUpdateHandlerSuccess(t *testing.T) {
	s := NewStore()
	s.mu.Lock()
	s.flags["myflag"] = Flag{Key: "myflag", Enabled: false, Description: "old", RolloutPercent: 0}
	s.mu.Unlock()

	server := NewServer(s)

	body := `{"enabled":true,"description":"updated","rollout_percent":50}`
	req := httptest.NewRequest(http.MethodPut, "/flags/myflag", strings.NewReader(body))
	req.SetPathValue("key", "myflag")
	rr := httptest.NewRecorder()

	server.UpdateHandler(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rr.Code)
	}

	var got Flag
	if err := json.Unmarshal(rr.Body.Bytes(), &got); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if got.Key != "myflag" {
		t.Errorf("expected key %q, got %q", "myflag", got.Key)
	}
	if !got.Enabled {
		t.Errorf("expected enabled true, got false")
	}
	if got.Description != "updated" {
		t.Errorf("expected description %q, got %q", "updated", got.Description)
	}
	if got.RolloutPercent != 50 {
		t.Errorf("expected rollout_percent 50, got %d", got.RolloutPercent)
	}
}

func TestUpdateHandlerNotFound(t *testing.T) {
	s := NewStore()
	server := NewServer(s)

	body := `{"enabled":true}`
	req := httptest.NewRequest(http.MethodPut, "/flags/missing", strings.NewReader(body))
	req.SetPathValue("key", "missing")
	rr := httptest.NewRecorder()

	server.UpdateHandler(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d", rr.Code)
	}
}
