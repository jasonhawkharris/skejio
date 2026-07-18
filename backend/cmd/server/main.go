package main

import (
	"context"
	"log"
	"net/http"

	"skejio/backend/internal/api"
	"skejio/backend/internal/auth"
	"skejio/backend/internal/db"
	"skejio/backend/internal/expenses"
	"skejio/backend/internal/financials"
	"skejio/backend/internal/merch"
	"skejio/backend/internal/merchvariants"
	"skejio/backend/internal/representatives"
	"skejio/backend/internal/riders"
	"skejio/backend/internal/tourdates"
	"skejio/backend/internal/tours"
	"skejio/backend/internal/users"
)

func main() {
	ctx := context.Background()
	pool := db.Connect(ctx)
	defer pool.Close()

	authH := &auth.Handler{DB: pool}
	tourdatesH := &tourdates.Handler{DB: pool}
	usersH := &users.Handler{DB: pool}
	representativesH := &representatives.Handler{DB: pool}
	financialsH := &financials.Handler{DB: pool}
	expensesH := &expenses.Handler{DB: pool}
	toursH := &tours.Handler{DB: pool}
	ridersH := &riders.Handler{DB: pool}
	merchH := &merch.Handler{DB: pool}
	merchVariantsH := &merchvariants.Handler{DB: pool}

	r := api.NewRouter(authH, tourdatesH, usersH, representativesH, financialsH, expensesH, toursH, ridersH, merchH, merchVariantsH)

	log.Println("listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}
