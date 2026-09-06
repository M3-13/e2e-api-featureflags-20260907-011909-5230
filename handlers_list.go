package main

import "net/http"

// ListHandler handles GET /flags. Implemented by the list ticket.
func (s *Server) ListHandler(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "not implemented", http.StatusNotImplemented)
}
