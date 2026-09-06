package main

import "net/http"

// DeleteHandler handles DELETE /flags/{key}. Implemented by the delete ticket.
func (s *Server) DeleteHandler(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "not implemented", http.StatusNotImplemented)
}
