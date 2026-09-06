package main

import "net/http"

// EvaluateHandler handles GET /flags/{key}/evaluate. Implemented by the evaluate ticket.
func (s *Server) EvaluateHandler(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "not implemented", http.StatusNotImplemented)
}
