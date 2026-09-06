package main

import (
	"encoding/json"
	"log"
	"net/http"
)

func healthzHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func main() {
	store := NewStore()
	server := NewServer(store)

	mux := http.NewServeMux()
	mux.HandleFunc("POST /flags", server.CreateHandler)
	mux.HandleFunc("GET /flags", server.ListHandler)
	mux.HandleFunc("GET /flags/{key}", server.GetHandler)
	mux.HandleFunc("PUT /flags/{key}", server.UpdateHandler)
	mux.HandleFunc("DELETE /flags/{key}", server.DeleteHandler)
	mux.HandleFunc("GET /flags/{key}/evaluate", server.EvaluateHandler)
	mux.HandleFunc("GET /healthz", healthzHandler)

	handler := Recover(Logging(BodyLimit(ContentType(mux))))

	log.Println("feature-flags-api listening on :8080")
	if err := http.ListenAndServe(":8080", handler); err != nil {
		log.Fatal(err)
	}
}
