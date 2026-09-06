package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func newEvaluateServer() *Server {
	store := NewStore()
	store.flags["always-on"] = Flag{Key: "always-on", Enabled: true, RolloutPercent: 100}
	store.flags["always-off"] = Flag{Key: "always-off", Enabled: true, RolloutPercent: 0}
	store.flags["half"] = Flag{Key: "half", Enabled: true, RolloutPercent: 50}
	return NewServer(store)
}

func evaluateRequest(t *testing.T, server *Server, target string) *httptest.ResponseRecorder {
	t.Helper()
	mux := http.NewServeMux()
	mux.HandleFunc("GET /flags/{key}/evaluate", server.EvaluateHandler)
	handler := Recover(Logging(BodyLimit(ContentType(mux))))

	req := httptest.NewRequest(http.MethodGet, target, nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	return rr
}

func TestEvaluateIsDeterministic(t *testing.T) {
	server := newEvaluateServer()

	first := evaluateRequest(t, server, "/flags/half/evaluate?user=alice")
	second := evaluateRequest(t, server, "/flags/half/evaluate?user=alice")

	if first.Code != http.StatusOK || second.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d and %d", first.Code, second.Code)
	}

	var a, b map[string]bool
	if err := json.NewDecoder(first.Body).Decode(&a); err != nil {
		t.Fatalf("failed to decode first response: %v", err)
	}
	if err := json.NewDecoder(second.Body).Decode(&b); err != nil {
		t.Fatalf("failed to decode second response: %v", err)
	}

	if a["result"] != b["result"] {
		t.Fatalf("expected same result for same key/user, got %v and %v", a["result"], b["result"])
	}
}

func TestEvaluateRollout100AlwaysTrue(t *testing.T) {
	server := newEvaluateServer()

	for _, user := range []string{"alice", "bob", "carol"} {
		rr := evaluateRequest(t, server, "/flags/always-on/evaluate?user="+user)
		if rr.Code != http.StatusOK {
			t.Fatalf("user %s: expected 200, got %d", user, rr.Code)
		}
		var body map[string]bool
		if err := json.NewDecoder(rr.Body).Decode(&body); err != nil {
			t.Fatalf("user %s: failed to decode: %v", user, err)
		}
		if body["result"] != true {
			t.Fatalf("user %s: expected true, got %v", user, body["result"])
		}
	}
}

func TestEvaluateRollout0AlwaysFalse(t *testing.T) {
	server := newEvaluateServer()

	for _, user := range []string{"alice", "bob", "carol"} {
		rr := evaluateRequest(t, server, "/flags/always-off/evaluate?user="+user)
		if rr.Code != http.StatusOK {
			t.Fatalf("user %s: expected 200, got %d", user, rr.Code)
		}
		var body map[string]bool
		if err := json.NewDecoder(rr.Body).Decode(&body); err != nil {
			t.Fatalf("user %s: failed to decode: %v", user, err)
		}
		if body["result"] != false {
			t.Fatalf("user %s: expected false, got %v", user, body["result"])
		}
	}
}

func TestEvaluateMissingUserReturns400(t *testing.T) {
	server := newEvaluateServer()

	rr := evaluateRequest(t, server, "/flags/half/evaluate")

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}

	var body map[string]string
	if err := json.NewDecoder(rr.Body).Decode(&body); err != nil {
		t.Fatalf("failed to decode: %v", err)
	}
	if body["error"] == "" {
		t.Fatalf("expected an error field in the response body")
	}
}

func TestEvaluateUnknownKeyReturns404(t *testing.T) {
	server := newEvaluateServer()

	rr := evaluateRequest(t, server, "/flags/does-not-exist/evaluate?user=alice")

	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rr.Code)
	}

	var body map[string]string
	if err := json.NewDecoder(rr.Body).Decode(&body); err != nil {
		t.Fatalf("failed to decode: %v", err)
	}
	if body["error"] == "" {
		t.Fatalf("expected an error field in the response body")
	}
}
