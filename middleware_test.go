package main

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestLoggingLogsMethodPathStatus(t *testing.T) {
	var buf bytes.Buffer
	old := log.Writer()
	log.SetOutput(&buf)
	defer log.SetOutput(old)

	handler := Logging(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
	}))

	req := httptest.NewRequest(http.MethodPost, "/flags?user=secret&token=hidden", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	out := buf.String()
	if !strings.Contains(out, "POST") {
		t.Fatalf("log output missing method: %q", out)
	}
	if !strings.Contains(out, "/flags") {
		t.Fatalf("log output missing path: %q", out)
	}
	if !strings.Contains(out, "201") {
		t.Fatalf("log output missing status: %q", out)
	}
	if strings.Contains(out, "secret") || strings.Contains(out, "hidden") || strings.Contains(out, "?user") {
		t.Fatalf("log output leaks query parameters: %q", out)
	}
}

func TestLoggingDefaultStatus(t *testing.T) {
	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer log.SetOutput(io.Discard)

	handler := Logging(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))

	req := httptest.NewRequest(http.MethodGet, "/flags", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if !strings.Contains(buf.String(), "200") {
		t.Fatalf("expected default 200 in log, got %q", buf.String())
	}
}

func TestRecoverReturns500JSON(t *testing.T) {
	handler := Recover(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("boom with internal detail")
	}))

	req := httptest.NewRequest(http.MethodGet, "/flags", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", rr.Code)
	}

	var body map[string]string
	if err := json.NewDecoder(rr.Body).Decode(&body); err != nil {
		t.Fatalf("failed to decode body: %v", err)
	}
	if _, ok := body["error"]; !ok {
		t.Fatalf("expected error key, got %v", body)
	}
	if strings.Contains(rr.Body.String(), "boom") || strings.Contains(rr.Body.String(), "panic") {
		t.Fatalf("response leaks internal details: %q", rr.Body.String())
	}
}

func TestRecoverLogOmitsPanicDetails(t *testing.T) {
	var buf bytes.Buffer
	old := log.Writer()
	log.SetOutput(&buf)
	defer log.SetOutput(old)
	oldFlags := log.Flags()
	log.SetFlags(0)
	defer log.SetFlags(oldFlags)

	handler := Recover(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("secret panic value 42")
	}))

	req := httptest.NewRequest(http.MethodGet, "/flags", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	out := buf.String()
	if !strings.Contains(out, "panic recovered") {
		t.Fatalf("expected 'panic recovered' in log, got %q", out)
	}
	if strings.Contains(out, "secret panic value 42") {
		t.Fatalf("log leaks panic details: %q", out)
	}
}

func TestBodyLimitRejectsLargeContentLength(t *testing.T) {
	handler := BodyLimit(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodPost, "/flags", nil)
	req.ContentLength = 1<<20 + 1
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("expected status 413, got %d", rr.Code)
	}
	var body map[string]string
	if err := json.NewDecoder(rr.Body).Decode(&body); err != nil {
		t.Fatalf("failed to decode body: %v", err)
	}
	if _, ok := body["error"]; !ok {
		t.Fatalf("expected error key, got %v", body)
	}
}

func TestBodyLimitAllowsSmallBody(t *testing.T) {
	handler := BodyLimit(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		io.Copy(io.Discard, r.Body)
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodPost, "/flags", strings.NewReader(`{"key":"x"}`))
	req.ContentLength = 11
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rr.Code)
	}
}

func TestBodyLimitRejectsOversizedChunkedBody(t *testing.T) {
	handler := BodyLimit(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		io.Copy(io.Discard, r.Body)
		w.WriteHeader(http.StatusOK)
	}))

	body := bytes.Repeat([]byte("a"), maxBodyBytes+1)
	req := httptest.NewRequest(http.MethodPost, "/flags", bytes.NewReader(body))
	req.ContentLength = -1 // unknown Content-Length (chunked)

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("expected status 413, got %d", rr.Code)
	}
	var errBody map[string]string
	if err := json.NewDecoder(rr.Body).Decode(&errBody); err != nil {
		t.Fatalf("failed to decode body: %v", err)
	}
	if _, ok := errBody["error"]; !ok {
		t.Fatalf("expected error key, got %v", errBody)
	}
}

func TestBodyLimitAcceptsBodyAtLimit(t *testing.T) {
	handler := BodyLimit(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		io.Copy(io.Discard, r.Body)
		w.WriteHeader(http.StatusOK)
	}))

	body := bytes.Repeat([]byte("a"), maxBodyBytes)
	req := httptest.NewRequest(http.MethodPost, "/flags", bytes.NewReader(body))
	req.ContentLength = -1 // unknown Content-Length (chunked)

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200 for body at limit, got %d", rr.Code)
	}
}

func TestBodyLimitPreservesBodyForHandler(t *testing.T) {
	handler := BodyLimit(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("handler read error: %v", err)
		}
		w.Write(got)
	}))

	payload := `{"key":"myflag"}`
	req := httptest.NewRequest(http.MethodPost, "/flags", strings.NewReader(payload))
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rr.Code)
	}
	if rr.Body.String() != payload {
		t.Fatalf("body not preserved: got %q want %q", rr.Body.String(), payload)
	}
}

func TestContentTypeRejectsNonJSON(t *testing.T) {
	handler := ContentType(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodPost, "/flags", strings.NewReader("x"))
	req.Header.Set("Content-Type", "text/plain")
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnsupportedMediaType {
		t.Fatalf("expected status 415, got %d", rr.Code)
	}
	var body map[string]string
	if err := json.NewDecoder(rr.Body).Decode(&body); err != nil {
		t.Fatalf("failed to decode body: %v", err)
	}
	if _, ok := body["error"]; !ok {
		t.Fatalf("expected error key, got %v", body)
	}
}

func TestContentTypeAcceptsJSON(t *testing.T) {
	handler := ContentType(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodPost, "/flags", strings.NewReader(`{"key":"x"}`))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rr.Code)
	}
}

func TestContentTypeAcceptsJSONWithCharset(t *testing.T) {
	handler := ContentType(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodPut, "/flags/x", strings.NewReader(`{"enabled":true}`))
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rr.Code)
	}
}

func TestContentTypeIgnoresGET(t *testing.T) {
	handler := ContentType(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/flags", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200 for GET, got %d", rr.Code)
	}
}
