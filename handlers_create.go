package main

import "net/http"

// CreateHandler handles POST /flags. Implemented by the create ticket.
func (s *Server) CreateHandler(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "not implemented", http.StatusNotImplemented)
}
