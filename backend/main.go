package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func testHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Hey, Jason. This is a test",
	})
}

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/", healthHandler).Methods(http.MethodGet)
	r.HandleFunc("/test", testHandler).Methods(http.MethodGet)

	log.Println("listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}
