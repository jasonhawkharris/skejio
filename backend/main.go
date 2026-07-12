package main

import (
	"context"
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

func newRouter(app *App) *mux.Router {
	r := mux.NewRouter()
	r.HandleFunc("/", healthHandler).Methods(http.MethodGet)
	r.HandleFunc("/test", testHandler).Methods(http.MethodGet)
	r.HandleFunc("/tourdates", app.ListTourDates).Methods(http.MethodGet)
	r.HandleFunc("/tourdates", app.CreateTourDate).Methods(http.MethodPost)
	r.HandleFunc("/tourdates/{id}", app.GetTourDate).Methods(http.MethodGet)
	r.HandleFunc("/tourdates/{id}", app.PatchTourDate).Methods(http.MethodPatch)
	r.HandleFunc("/tourdates/{id}", app.DeleteTourDate).Methods(http.MethodDelete)
	r.HandleFunc("/users", app.ListUsers).Methods(http.MethodGet)
	r.HandleFunc("/users", app.CreateUser).Methods(http.MethodPost)
	r.HandleFunc("/users/{id}", app.GetUser).Methods(http.MethodGet)
	r.HandleFunc("/users/{id}", app.PatchUser).Methods(http.MethodPatch)
	r.HandleFunc("/users/{id}", app.DeleteUser).Methods(http.MethodDelete)
	return r
}

func main() {
	ctx := context.Background()
	pool := connectDB(ctx)
	defer pool.Close()

	app := &App{db: pool}
	r := newRouter(app)

	log.Println("listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}
