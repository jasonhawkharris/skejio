package api

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"

	"skejio/backend/internal/auth"
	"skejio/backend/internal/expenses"
	"skejio/backend/internal/financials"
	"skejio/backend/internal/representatives"
	"skejio/backend/internal/riders"
	"skejio/backend/internal/tourdates"
	"skejio/backend/internal/tours"
	"skejio/backend/internal/users"
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

// NewRouter wires every route. authH.DB is used to construct the auth
// middleware, since all handlers share the same underlying connection pool.
func NewRouter(authH *auth.Handler, tourdatesH *tourdates.Handler, usersH *users.Handler, repsH *representatives.Handler, financialsH *financials.Handler, expensesH *expenses.Handler, toursH *tours.Handler, ridersH *riders.Handler) *mux.Router {
	r := mux.NewRouter()
	r.HandleFunc("/", healthHandler).Methods(http.MethodGet)
	r.HandleFunc("/test", testHandler).Methods(http.MethodGet)
	r.HandleFunc("/login", authH.Login).Methods(http.MethodPost)

	tourdatesRouter := r.PathPrefix("/tourdates").Subrouter()
	tourdatesRouter.Use(auth.Middleware(authH.DB))
	tourdatesRouter.HandleFunc("", tourdatesH.List).Methods(http.MethodGet)
	tourdatesRouter.HandleFunc("", tourdatesH.Create).Methods(http.MethodPost)
	tourdatesRouter.HandleFunc("/{id}", tourdatesH.Get).Methods(http.MethodGet)
	tourdatesRouter.HandleFunc("/{id}", tourdatesH.Patch).Methods(http.MethodPatch)
	tourdatesRouter.HandleFunc("/{id}", tourdatesH.Delete).Methods(http.MethodDelete)
	tourdatesRouter.HandleFunc("/{id}/financials", financialsH.Get).Methods(http.MethodGet)
	tourdatesRouter.HandleFunc("/{id}/financials", financialsH.Create).Methods(http.MethodPost)
	tourdatesRouter.HandleFunc("/{id}/financials", financialsH.Patch).Methods(http.MethodPatch)
	tourdatesRouter.HandleFunc("/{id}/financials", financialsH.Delete).Methods(http.MethodDelete)
	tourdatesRouter.HandleFunc("/{id}/summary", financialsH.Summary).Methods(http.MethodGet)
	tourdatesRouter.HandleFunc("/{id}/expenses", expensesH.List).Methods(http.MethodGet)
	tourdatesRouter.HandleFunc("/{id}/expenses", expensesH.Create).Methods(http.MethodPost)
	tourdatesRouter.HandleFunc("/{id}/expenses/{expenseID}", expensesH.Get).Methods(http.MethodGet)
	tourdatesRouter.HandleFunc("/{id}/expenses/{expenseID}", expensesH.Patch).Methods(http.MethodPatch)
	tourdatesRouter.HandleFunc("/{id}/expenses/{expenseID}", expensesH.Delete).Methods(http.MethodDelete)

	toursRouter := r.PathPrefix("/tours").Subrouter()
	toursRouter.Use(auth.Middleware(authH.DB))
	toursRouter.HandleFunc("", toursH.List).Methods(http.MethodGet)
	toursRouter.HandleFunc("", toursH.Create).Methods(http.MethodPost)
	toursRouter.HandleFunc("/{id}", toursH.Get).Methods(http.MethodGet)
	toursRouter.HandleFunc("/{id}", toursH.Patch).Methods(http.MethodPatch)
	toursRouter.HandleFunc("/{id}", toursH.Delete).Methods(http.MethodDelete)

	ridersRouter := r.PathPrefix("/riders").Subrouter()
	ridersRouter.Use(auth.Middleware(authH.DB))
	ridersRouter.HandleFunc("", ridersH.List).Methods(http.MethodGet)
	ridersRouter.HandleFunc("", ridersH.Create).Methods(http.MethodPost)
	ridersRouter.HandleFunc("/{id}", ridersH.Get).Methods(http.MethodGet)
	ridersRouter.HandleFunc("/{id}", ridersH.Patch).Methods(http.MethodPatch)
	ridersRouter.HandleFunc("/{id}", ridersH.Delete).Methods(http.MethodDelete)

	logoutRouter := r.PathPrefix("/logout").Subrouter()
	logoutRouter.Use(auth.Middleware(authH.DB))
	logoutRouter.HandleFunc("", authH.Logout).Methods(http.MethodPost)

	r.HandleFunc("/sign-up", usersH.Create).Methods(http.MethodPost)

	usersByID := r.PathPrefix("/users").Subrouter()
	usersByID.Use(auth.Middleware(authH.DB))
	usersByID.HandleFunc("/{id}", usersH.Get).Methods(http.MethodGet)
	usersByID.HandleFunc("/{id}", usersH.Patch).Methods(http.MethodPatch)
	usersByID.HandleFunc("/{id}", usersH.Delete).Methods(http.MethodDelete)

	representativesRouter := r.PathPrefix("/representatives").Subrouter()
	representativesRouter.Use(auth.Middleware(authH.DB))
	representativesRouter.HandleFunc("", repsH.Create).Methods(http.MethodPost)
	representativesRouter.HandleFunc("", repsH.ListRepresentatives).Methods(http.MethodGet)
	representativesRouter.HandleFunc("/{id}", repsH.Delete).Methods(http.MethodDelete)

	representedArtistsRouter := r.PathPrefix("/represented-artists").Subrouter()
	representedArtistsRouter.Use(auth.Middleware(authH.DB))
	representedArtistsRouter.HandleFunc("", repsH.ListArtists).Methods(http.MethodGet)

	return r
}
