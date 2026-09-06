package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"strings"
)

// maxBodyBytes is the maximum accepted request body size (1 MB).
const maxBodyBytes = 1 << 20

// statusRecorder wraps an http.ResponseWriter to capture the status code.
type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(code int) {
	r.status = code
	r.ResponseWriter.WriteHeader(code)
}

// writeJSONError writes a JSON error object of the form {"error": message}.
func writeJSONError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}

// Logging logs method, path (without query string) and status code for every
// request. It never logs query parameters or user identifiers.
func Logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(rec, r)
		log.Printf("%s %s %d", r.Method, r.URL.Path, rec.status)
	})
}

// Recover catches panics in downstream handlers and answers 500 with a JSON
// error object, without exposing stack traces or internal details.
func Recover(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.Print("panic recovered")
				writeJSONError(w, http.StatusInternalServerError, "internal server error")
			}
		}()
		next.ServeHTTP(w, r)
	})
}

// BodyLimit rejects POST and PUT requests whose body exceeds 1 MB with 413,
// and caps the body read via http.MaxBytesReader as a fallback. It detects
// oversized bodies even when Content-Length is unknown (e.g. chunked) by
// reading up to the limit and translating the MaxBytesReader error into a 413.
func BodyLimit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost && r.Method != http.MethodPut {
			next.ServeHTTP(w, r)
			return
		}

		if r.ContentLength > maxBodyBytes {
			writeJSONError(w, http.StatusRequestEntityTooLarge, "request body too large")
			return
		}

		// Allow reading one byte past the limit so that an exactly-1 MB body
		// reads cleanly to EOF, while anything larger raises a MaxBytesError.
		r.Body = http.MaxBytesReader(w, r.Body, maxBodyBytes+1)
		body, err := io.ReadAll(r.Body)
		if err != nil {
			var maxErr *http.MaxBytesError
			if errors.As(err, &maxErr) {
				writeJSONError(w, http.StatusRequestEntityTooLarge, "request body too large")
				return
			}
			writeJSONError(w, http.StatusBadRequest, "invalid request body")
			return
		}
		if len(body) > maxBodyBytes {
			writeJSONError(w, http.StatusRequestEntityTooLarge, "request body too large")
			return
		}
		r.Body = io.NopCloser(bytes.NewReader(body))
		next.ServeHTTP(w, r)
	})
}

// ContentType rejects POST and PUT requests whose Content-Type is not
// application/json with 415.
func ContentType(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost || r.Method == http.MethodPut {
			ct := r.Header.Get("Content-Type")
			if mediaType(ct) != "application/json" {
				writeJSONError(w, http.StatusUnsupportedMediaType, "content type must be application/json")
				return
			}
		}
		next.ServeHTTP(w, r)
	})
}

// mediaType strips parameters (e.g. "; charset=utf-8") from a Content-Type
// header value.
func mediaType(ct string) string {
	if i := strings.IndexByte(ct, ';'); i >= 0 {
		return strings.TrimSpace(ct[:i])
	}
	return strings.TrimSpace(ct)
}
