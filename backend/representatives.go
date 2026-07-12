package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// RepresentedUser is a user seen through a representative relationship: the
// artist's representative (from the artist's point of view) or the artist
// being represented (from the representative's point of view).
type RepresentedUser struct {
	RelationshipID uuid.UUID `json:"relationship_id"`
	UserID         uuid.UUID `json:"user_id"`
	Name           string    `json:"name"`
	Email          string    `json:"email"`
	UserType       string    `json:"user_type"`
	CreatedAt      time.Time `json:"created_at"`
}

// CreateRepresentative lets an artist grant a MANAGER/AGENT/LABEL user access
// to their own tourdates. Only the artist can grant this - the representative
// cannot add themselves.
func (a *App) CreateRepresentative(w http.ResponseWriter, r *http.Request) {
	caller := userFromContext(r.Context())
	if caller.UserType != "ARTIST" {
		writeError(w, http.StatusForbidden, "only artists can grant representative access")
		return
	}

	var req struct {
		RepresentativeID uuid.UUID `json:"representative_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	if req.RepresentativeID == uuid.Nil {
		writeError(w, http.StatusBadRequest, "representative_id is required")
		return
	}
	if req.RepresentativeID == caller.ID {
		writeError(w, http.StatusBadRequest, "cannot add yourself as your own representative")
		return
	}

	var repName, repEmail, repUserType string
	err := a.db.QueryRow(r.Context(), "SELECT name, email, user_type FROM users WHERE id = $1", req.RepresentativeID).
		Scan(&repName, &repEmail, &repUserType)
	if errors.Is(err, pgx.ErrNoRows) {
		writeError(w, http.StatusBadRequest, "representative_id does not reference an existing user")
		return
	} else if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create representative")
		return
	}
	if repUserType != "MANAGER" && repUserType != "AGENT" && repUserType != "LABEL" {
		writeError(w, http.StatusBadRequest, "representative_id must reference a MANAGER, AGENT, or LABEL user")
		return
	}

	var relationshipID uuid.UUID
	var createdAt time.Time
	err = a.db.QueryRow(r.Context(),
		"INSERT INTO artist_representatives (artist_id, representative_id) VALUES ($1, $2) RETURNING id, created_at",
		caller.ID, req.RepresentativeID,
	).Scan(&relationshipID, &createdAt)
	if isUniqueViolation(err) {
		writeError(w, http.StatusConflict, "this representative has already been added")
		return
	} else if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create representative")
		return
	}

	writeJSON(w, http.StatusCreated, RepresentedUser{
		RelationshipID: relationshipID,
		UserID:         req.RepresentativeID,
		Name:           repName,
		Email:          repEmail,
		UserType:       repUserType,
		CreatedAt:      createdAt,
	})
}

// ListMyRepresentatives returns the representatives the caller (an artist)
// has granted access to.
func (a *App) ListMyRepresentatives(w http.ResponseWriter, r *http.Request) {
	caller := userFromContext(r.Context())
	if caller.UserType != "ARTIST" {
		writeError(w, http.StatusForbidden, "only artists have representatives")
		return
	}

	rows, err := a.db.Query(r.Context(), `
		SELECT ar.id, u.id, u.name, u.email, u.user_type, ar.created_at
		FROM artist_representatives ar
		JOIN users u ON u.id = ar.representative_id
		WHERE ar.artist_id = $1
		ORDER BY ar.created_at`, caller.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list representatives")
		return
	}
	defer rows.Close()

	representatives := []RepresentedUser{}
	for rows.Next() {
		var ru RepresentedUser
		if err := rows.Scan(&ru.RelationshipID, &ru.UserID, &ru.Name, &ru.Email, &ru.UserType, &ru.CreatedAt); err != nil {
			writeError(w, http.StatusInternalServerError, "failed to read representatives")
			return
		}
		representatives = append(representatives, ru)
	}
	if err := rows.Err(); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to read representatives")
		return
	}

	writeJSON(w, http.StatusOK, representatives)
}

// ListRepresentedArtists returns the artists the caller (a MANAGER/AGENT/
// LABEL) represents.
func (a *App) ListRepresentedArtists(w http.ResponseWriter, r *http.Request) {
	caller := userFromContext(r.Context())
	if caller.UserType != "MANAGER" && caller.UserType != "AGENT" && caller.UserType != "LABEL" {
		writeError(w, http.StatusForbidden, "only managers, agents, and labels represent artists")
		return
	}

	rows, err := a.db.Query(r.Context(), `
		SELECT ar.id, u.id, u.name, u.email, u.user_type, ar.created_at
		FROM artist_representatives ar
		JOIN users u ON u.id = ar.artist_id
		WHERE ar.representative_id = $1
		ORDER BY ar.created_at`, caller.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list represented artists")
		return
	}
	defer rows.Close()

	artists := []RepresentedUser{}
	for rows.Next() {
		var ru RepresentedUser
		if err := rows.Scan(&ru.RelationshipID, &ru.UserID, &ru.Name, &ru.Email, &ru.UserType, &ru.CreatedAt); err != nil {
			writeError(w, http.StatusInternalServerError, "failed to read represented artists")
			return
		}
		artists = append(artists, ru)
	}
	if err := rows.Err(); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to read represented artists")
		return
	}

	writeJSON(w, http.StatusOK, artists)
}

// DeleteRepresentative revokes a relationship. Either party - the artist or
// the representative - may revoke it. An id that doesn't belong to the
// caller on either side 404s, so as not to reveal it exists.
func (a *App) DeleteRepresentative(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(mux.Vars(r)["id"])
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}
	caller := userFromContext(r.Context())

	tag, err := a.db.Exec(r.Context(),
		"DELETE FROM artist_representatives WHERE id = $1 AND (artist_id = $2 OR representative_id = $2)",
		id, caller.ID,
	)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to delete representative")
		return
	}
	if tag.RowsAffected() == 0 {
		writeError(w, http.StatusNotFound, "representative relationship not found")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
