package main

import "net/http"

// GetHandler handles GET /flags/{key}. Implemented by the get ticket.
func (s *Server) GetHandler(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "not implemented", http.StatusNotImplemented)
}
