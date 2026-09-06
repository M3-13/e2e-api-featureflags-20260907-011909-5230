package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestListFlagsEmpty(t *testing.T) {
	s := NewServer(NewStore())

	req := httptest.NewRequest(http.MethodGet, "/flags", nil)
	rr := httptest.NewRecorder()
	s.ListHandler(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rr.Code)
	}

	if got := strings.TrimSpace(rr.Body.String()); got != "[]" {
		t.Fatalf("expected empty JSON array [], got %q", got)
	}

	var flags []Flag
	if err := json.NewDecoder(strings.NewReader(rr.Body.String())).Decode(&flags); err != nil {
		t.Fatalf("failed to decode response body: %v", err)
	}
	if flags == nil {
		t.Fatalf("expected non-nil empty slice, got nil")
	}
	if len(flags) != 0 {
		t.Fatalf("expected 0 flags, got %d", len(flags))
	}
}

func TestListFlagsMultiple(t *testing.T) {
	store := NewStore()
	store.mu.Lock()
	store.flags["flag-a"] = Flag{Key: "flag-a", Enabled: true, Description: "first", RolloutPercent: 100}
	store.flags["flag-b"] = Flag{Key: "flag-b", Enabled: false, Description: "second", RolloutPercent: 0}
	store.mu.Unlock()

	s := NewServer(store)

	req := httptest.NewRequest(http.MethodGet, "/flags", nil)
	rr := httptest.NewRecorder()
	s.ListHandler(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rr.Code)
	}

	var flags []Flag
	if err := json.NewDecoder(rr.Body).Decode(&flags); err != nil {
		t.Fatalf("failed to decode response body: %v", err)
	}
	if len(flags) != 2 {
		t.Fatalf("expected 2 flags, got %d", len(flags))
	}

	byKey := make(map[string]Flag, len(flags))
	for _, f := range flags {
		byKey[f.Key] = f
	}

	a, ok := byKey["flag-a"]
	if !ok {
		t.Fatalf("expected flag-a in list, got %v", flags)
	}
	if !a.Enabled || a.Description != "first" || a.RolloutPercent != 100 {
		t.Fatalf("flag-a fields mismatch: %+v", a)
	}

	b, ok := byKey["flag-b"]
	if !ok {
		t.Fatalf("expected flag-b in list, got %v", flags)
	}
	if b.Enabled || b.Description != "second" || b.RolloutPercent != 0 {
		t.Fatalf("flag-b fields mismatch: %+v", b)
	}
}

func TestStoreListEmptyReturnsEmptySlice(t *testing.T) {
	s := NewStore()
	got := s.List()
	if got == nil {
		t.Fatalf("expected empty non-nil slice, got nil")
	}
	if len(got) != 0 {
		t.Fatalf("expected 0 flags, got %d", len(got))
	}
}
