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
	r.HandleFunc("/login", app.Login).Methods(http.MethodPost)

	tourdates := r.PathPrefix("/tourdates").Subrouter()
	tourdates.Use(authMiddleware(app))
	tourdates.HandleFunc("", app.ListTourDates).Methods(http.MethodGet)
	tourdates.HandleFunc("", app.CreateTourDate).Methods(http.MethodPost)
	tourdates.HandleFunc("/{id}", app.GetTourDate).Methods(http.MethodGet)
	tourdates.HandleFunc("/{id}", app.PatchTourDate).Methods(http.MethodPatch)
	tourdates.HandleFunc("/{id}", app.DeleteTourDate).Methods(http.MethodDelete)

	logout := r.PathPrefix("/logout").Subrouter()
	logout.Use(authMiddleware(app))
	logout.HandleFunc("", app.Logout).Methods(http.MethodPost)

	r.HandleFunc("/users", app.CreateUser).Methods(http.MethodPost)

	usersByID := r.PathPrefix("/users").Subrouter()
	usersByID.Use(authMiddleware(app))
	usersByID.HandleFunc("/{id}", app.GetUser).Methods(http.MethodGet)
	usersByID.HandleFunc("/{id}", app.PatchUser).Methods(http.MethodPatch)
	usersByID.HandleFunc("/{id}", app.DeleteUser).Methods(http.MethodDelete)

	representatives := r.PathPrefix("/representatives").Subrouter()
	representatives.Use(authMiddleware(app))
	representatives.HandleFunc("", app.CreateRepresentative).Methods(http.MethodPost)
	representatives.HandleFunc("", app.ListMyRepresentatives).Methods(http.MethodGet)
	representatives.HandleFunc("/{id}", app.DeleteRepresentative).Methods(http.MethodDelete)

	representedArtists := r.PathPrefix("/represented-artists").Subrouter()
	representedArtists.Use(authMiddleware(app))
	representedArtists.HandleFunc("", app.ListRepresentedArtists).Methods(http.MethodGet)

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
