package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHealthz(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", healthzHandler)

	handler := Recover(Logging(BodyLimit(ContentType(mux))))

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

func TestHealthzHasNoCORSHeaders(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", healthzHandler)

	handler := Recover(Logging(BodyLimit(ContentType(mux))))

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	for name := range rr.Header() {
		if len(name) >= len("Access-Control-Allow-") && name[:len("Access-Control-Allow-")] == "Access-Control-Allow-" {
			t.Fatalf("unexpected CORS header %q present", name)
		}
	}
}
