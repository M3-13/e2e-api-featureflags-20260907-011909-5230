package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

// newTestHandler wires the full production middleware chain around the real
// mux, matching main.go, so auth behaviour is exercised end to end.
func newTestHandler() http.Handler {
	store := NewStore()
	server := NewServer(store)

	mux := http.NewServeMux()
	mux.HandleFunc("POST /flags", server.CreateHandler)
	mux.HandleFunc("GET /flags", server.ListHandler)
	mux.HandleFunc("GET /flags/{key}", server.GetHandler)
	mux.HandleFunc("PUT /flags/{key}", server.UpdateHandler)
	mux.HandleFunc("DELETE /flags/{key}", server.DeleteHandler)
	mux.HandleFunc("GET /flags/{key}/evaluate", server.EvaluateHandler)
	mux.HandleFunc("GET /healthz", healthzHandler)

	return Recover(Logging(BodyLimit(ContentType(RequireAuth(mux)))))
}

func TestFlagsWithoutKeyReturns401(t *testing.T) {
	handler := newTestHandler()

	req := httptest.NewRequest(http.MethodGet, "/flags", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d", rr.Code)
	}
	assertUnauthorizedBody(t, rr)
}

func TestFlagsWrongKeyReturns401(t *testing.T) {
	t.Setenv("API_KEY", "correct-secret")
	handler := newTestHandler()

	req := httptest.NewRequest(http.MethodGet, "/flags", nil)
	req.Header.Set("X-API-Key", "wrong-secret")
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d", rr.Code)
	}
	assertUnauthorizedBody(t, rr)
}

func TestFlagsCorrectKeyPassesThrough(t *testing.T) {
	t.Setenv("API_KEY", "correct-secret")
	handler := newTestHandler()

	req := httptest.NewRequest(http.MethodGet, "/flags", nil)
	req.Header.Set("X-API-Key", "correct-secret")
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rr.Code)
	}

	var body []Flag
	if err := json.NewDecoder(rr.Body).Decode(&body); err != nil {
		t.Fatalf("failed to decode response body: %v", err)
	}
}

func TestHealthzWithoutKeyReturns200(t *testing.T) {
	handler := newTestHandler()

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rr.Code)
	}

	var body map[string]string
	if err := json.NewDecoder(rr.Body).Decode(&body); err != nil {
		t.Fatalf("failed to decode response body: %v", err)
	}
	if body["status"] != "ok" {
		t.Fatalf("expected status ok, got %q", body["status"])
	}
}

func TestFlagsWithoutConfiguredKeyAlways401(t *testing.T) {
	t.Setenv("API_KEY", "")
	handler := newTestHandler()

	req := httptest.NewRequest(http.MethodGet, "/flags", nil)
	req.Header.Set("X-API-Key", "")
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401 when no API_KEY configured, got %d", rr.Code)
	}
}

func assertUnauthorizedBody(t *testing.T, rr *httptest.ResponseRecorder) {
	t.Helper()
	var body map[string]string
	if err := json.NewDecoder(rr.Body).Decode(&body); err != nil {
		t.Fatalf("failed to decode response body: %v", err)
	}
	if body["error"] != "unauthorized" {
		t.Fatalf("expected error unauthorized, got %q", body["error"])
	}
}
