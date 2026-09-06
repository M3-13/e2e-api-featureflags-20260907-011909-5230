package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func newTestServer() *Server {
	return NewServer(NewStore())
}

func TestCreateFlagCreated(t *testing.T) {
	s := newTestServer()

	body := `{"key":"neu","enabled":true}`
	req := httptest.NewRequest(http.MethodPost, "/flags", strings.NewReader(body))
	rr := httptest.NewRecorder()

	s.CreateHandler(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d", rr.Code)
	}

	var flag Flag
	if err := json.NewDecoder(rr.Body).Decode(&flag); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if flag.Key != "neu" {
		t.Fatalf("expected key neu, got %q", flag.Key)
	}
	if !flag.Enabled {
		t.Fatalf("expected enabled true, got false")
	}
}

func TestCreateFlagDuplicateKey(t *testing.T) {
	s := newTestServer()

	body := `{"key":"dup","enabled":false}`

	req1 := httptest.NewRequest(http.MethodPost, "/flags", strings.NewReader(body))
	rr1 := httptest.NewRecorder()
	s.CreateHandler(rr1, req1)
	if rr1.Code != http.StatusCreated {
		t.Fatalf("first create: expected 201, got %d", rr1.Code)
	}

	req2 := httptest.NewRequest(http.MethodPost, "/flags", strings.NewReader(body))
	rr2 := httptest.NewRecorder()
	s.CreateHandler(rr2, req2)
	if rr2.Code != http.StatusConflict {
		t.Fatalf("second create: expected 409, got %d", rr2.Code)
	}

	var errBody map[string]string
	if err := json.NewDecoder(rr2.Body).Decode(&errBody); err != nil {
		t.Fatalf("failed to decode error body: %v", err)
	}
	if errBody["error"] == "" {
		t.Fatalf("expected error field in response, got %v", errBody)
	}
}

func TestCreateFlagMissingKey(t *testing.T) {
	s := newTestServer()

	req := httptest.NewRequest(http.MethodPost, "/flags", strings.NewReader(`{"enabled":true}`))
	rr := httptest.NewRecorder()
	s.CreateHandler(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", rr.Code)
	}

	var errBody map[string]string
	if err := json.NewDecoder(rr.Body).Decode(&errBody); err != nil {
		t.Fatalf("failed to decode error body: %v", err)
	}
	if errBody["error"] == "" {
		t.Fatalf("expected error field in response, got %v", errBody)
	}
}

func TestCreateFlagEmptyKey(t *testing.T) {
	s := newTestServer()

	req := httptest.NewRequest(http.MethodPost, "/flags", strings.NewReader(`{"key":"","enabled":true}`))
	rr := httptest.NewRecorder()
	s.CreateHandler(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", rr.Code)
	}

	var errBody map[string]string
	if err := json.NewDecoder(rr.Body).Decode(&errBody); err != nil {
		t.Fatalf("failed to decode error body: %v", err)
	}
	if errBody["error"] == "" {
		t.Fatalf("expected error field in response, got %v", errBody)
	}
}

func TestCreateFlagInvalidJSON(t *testing.T) {
	s := newTestServer()

	req := httptest.NewRequest(http.MethodPost, "/flags", strings.NewReader(`{not json`))
	rr := httptest.NewRecorder()
	s.CreateHandler(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", rr.Code)
	}
}

func TestCreateFlagRolloutPercentTooHigh(t *testing.T) {
	s := newTestServer()

	req := httptest.NewRequest(http.MethodPost, "/flags", strings.NewReader(`{"key":"neu","enabled":true,"rollout_percent":101}`))
	rr := httptest.NewRecorder()
	s.CreateHandler(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", rr.Code)
	}

	var errBody map[string]string
	if err := json.NewDecoder(rr.Body).Decode(&errBody); err != nil {
		t.Fatalf("failed to decode error body: %v", err)
	}
	if errBody["error"] == "" {
		t.Fatalf("expected error field in response, got %v", errBody)
	}
}

func TestCreateFlagRolloutPercentNegative(t *testing.T) {
	s := newTestServer()

	req := httptest.NewRequest(http.MethodPost, "/flags", strings.NewReader(`{"key":"neu","enabled":true,"rollout_percent":-1}`))
	rr := httptest.NewRecorder()
	s.CreateHandler(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", rr.Code)
	}

	var errBody map[string]string
	if err := json.NewDecoder(rr.Body).Decode(&errBody); err != nil {
		t.Fatalf("failed to decode error body: %v", err)
	}
	if errBody["error"] == "" {
		t.Fatalf("expected error field in response, got %v", errBody)
	}
}

func TestCreateFlagInvalidKeySlash(t *testing.T) {
	s := newTestServer()

	req := httptest.NewRequest(http.MethodPost, "/flags", strings.NewReader(`{"key":"bad/key","enabled":true}`))
	rr := httptest.NewRecorder()
	s.CreateHandler(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", rr.Code)
	}

	var errBody map[string]string
	if err := json.NewDecoder(rr.Body).Decode(&errBody); err != nil {
		t.Fatalf("failed to decode error body: %v", err)
	}
	if errBody["error"] == "" {
		t.Fatalf("expected error field in response, got %v", errBody)
	}
}

func TestCreateFlagInvalidKeyCharacters(t *testing.T) {
	s := newTestServer()

	req := httptest.NewRequest(http.MethodPost, "/flags", strings.NewReader(`{"key":"bad key!","enabled":true}`))
	rr := httptest.NewRecorder()
	s.CreateHandler(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", rr.Code)
	}

	var errBody map[string]string
	if err := json.NewDecoder(rr.Body).Decode(&errBody); err != nil {
		t.Fatalf("failed to decode error body: %v", err)
	}
	if errBody["error"] == "" {
		t.Fatalf("expected error field in response, got %v", errBody)
	}
}
