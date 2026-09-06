package main

import "net/http"

// UpdateHandler handles PUT /flags/{key}. Implemented by the update ticket.
func (s *Server) UpdateHandler(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "not implemented", http.StatusNotImplemented)
}
