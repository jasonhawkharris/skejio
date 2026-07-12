package main

import (
	"context"
	"log"
	"net/http"

	"skejio/backend/internal/api"
	"skejio/backend/internal/auth"
	"skejio/backend/internal/db"
	"skejio/backend/internal/representatives"
	"skejio/backend/internal/tourdates"
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

	r := api.NewRouter(authH, tourdatesH, usersH, representativesH)

	log.Println("listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}
