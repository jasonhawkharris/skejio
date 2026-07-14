package tourdates

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"skejio/backend/internal/auth"
	"skejio/backend/internal/db"
	"skejio/backend/internal/httpx"
)

// ScopedResource implements the common query shape shared by every resource
// owned by an artist (the caller, or someone they represent): List/Get/
// Delete/Update filtered against the caller's accessible artist ids.
// tourdates, tours, and riders all fit this shape - tourdates layers extra
// tour_id/rider_id list filtering and owned-FK patching on top of it.
type ScopedResource[T any] struct {
	DB      *pgxpool.Pool
	Table   string
	Columns string
	OrderBy string
	Scan    func(pgx.Row) (T, error)
}

// List returns rows owned by artistID (if given) or by any id in
// accessibleIDs otherwise, in OrderBy order. extraWhere is an additional SQL
// fragment (starting at placeholder $2, since $1 is the ownership filter)
// ANDed onto the query, with extraArgs as its parameters - pass "" and no
// args if there's nothing to add.
func (s ScopedResource[T]) List(ctx context.Context, accessibleIDs []string, artistID *uuid.UUID, extraWhere string, extraArgs ...any) ([]T, error) {
	query := "SELECT " + s.Columns + " FROM " + s.Table + " WHERE "
	var args []any
	if artistID != nil {
		query += "user_id = $1"
		args = append(args, *artistID)
	} else {
		query += "user_id = ANY($1::uuid[])"
		args = append(args, accessibleIDs)
	}
	query += extraWhere + " ORDER BY " + s.OrderBy
	args = append(args, extraArgs...)

	rows, err := s.DB.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	return db.ScanAll(rows, s.Scan)
}

// Get returns the row with the given id, if owned by an artist in
// accessibleIDs.
func (s ScopedResource[T]) Get(ctx context.Context, id uuid.UUID, accessibleIDs []string) (T, error) {
	row := s.DB.QueryRow(ctx,
		"SELECT "+s.Columns+" FROM "+s.Table+" WHERE id = $1 AND user_id = ANY($2::uuid[])", id, accessibleIDs)
	return s.Scan(row)
}

// Delete removes the row with the given id, if owned by an artist in
// accessibleIDs, reporting whether a row was actually deleted.
func (s ScopedResource[T]) Delete(ctx context.Context, id uuid.UUID, accessibleIDs []string) (bool, error) {
	tag, err := s.DB.Exec(ctx,
		"DELETE FROM "+s.Table+" WHERE id = $1 AND user_id = ANY($2::uuid[])", id, accessibleIDs)
	if err != nil {
		return false, err
	}
	return tag.RowsAffected() > 0, nil
}

// Update applies the fields accumulated in pb to the row with the given id,
// if owned by an artist in accessibleIDs.
func (s ScopedResource[T]) Update(ctx context.Context, pb *httpx.PatchBuilder, id uuid.UUID, accessibleIDs []string) (T, error) {
	where := fmt.Sprintf("id = $%d AND user_id = ANY($%d::uuid[])", pb.NextArg(), pb.NextArg()+1)
	query, args := pb.Build(s.Table, where, s.Columns, id, accessibleIDs)
	row := s.DB.QueryRow(ctx, query, args...)
	return s.Scan(row)
}

// ResolveListScope resolves the caller's accessible artist ids and, if an
// "artist_id" query param is present, narrows to just that artist - who must
// be in the accessible set, else a 404 (so as not to reveal whether that
// artist_id exists). A nil artistID return means "every accessible artist".
func ResolveListScope(w http.ResponseWriter, r *http.Request, db *pgxpool.Pool, noun string) (accessibleIDs []string, artistID *uuid.UUID, ok bool) {
	caller := auth.UserFromContext(r.Context())
	accessibleIDs, err := AccessibleArtistIDs(r.Context(), db, caller.ID)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "failed to list "+noun+"s")
		return nil, nil, false
	}

	param := r.URL.Query().Get("artist_id")
	if param == "" {
		return accessibleIDs, nil, true
	}

	id, err := uuid.Parse(param)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "artist_id must be a valid UUID")
		return nil, nil, false
	}
	if !ContainsID(accessibleIDs, id) {
		httpx.WriteError(w, http.StatusNotFound, "artist not found")
		return nil, nil, false
	}

	return accessibleIDs, &id, true
}

// ResolveCreateOwner returns the artist id a newly-created resource should
// be owned by: requestedArtistID if given - which must be the caller
// themselves or an artist they represent (403 otherwise) - or the caller
// themselves if requestedArtistID is uuid.Nil.
func ResolveCreateOwner(w http.ResponseWriter, r *http.Request, db *pgxpool.Pool, requestedArtistID uuid.UUID, noun string) (uuid.UUID, bool) {
	caller := auth.UserFromContext(r.Context())
	if requestedArtistID == uuid.Nil {
		return caller.ID, true
	}
	if requestedArtistID == caller.ID {
		return requestedArtistID, true
	}

	accessibleIDs, err := AccessibleArtistIDs(r.Context(), db, caller.ID)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "failed to create "+noun)
		return uuid.Nil, false
	}
	if !ContainsID(accessibleIDs, requestedArtistID) {
		httpx.WriteError(w, http.StatusForbidden, "you do not have access to create "+noun+"s for this artist")
		return uuid.Nil, false
	}

	return requestedArtistID, true
}

// HandleGet implements the common Get shape for an artist-scoped resource:
// parse the "id" route param, resolve the caller's accessible artist ids,
// and look up the row - 404ing (rather than 403ing) if it doesn't exist or
// isn't accessible, so as not to reveal that it exists.
func HandleGet[T any](w http.ResponseWriter, r *http.Request, res ScopedResource[T], noun string) {
	id, err := uuid.Parse(mux.Vars(r)["id"])
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid id")
		return
	}
	caller := auth.UserFromContext(r.Context())
	accessibleIDs, err := AccessibleArtistIDs(r.Context(), res.DB, caller.ID)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "failed to read "+noun)
		return
	}

	t, err := res.Get(r.Context(), id, accessibleIDs)
	if errors.Is(err, pgx.ErrNoRows) {
		httpx.WriteError(w, http.StatusNotFound, noun+" not found")
		return
	} else if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "failed to read "+noun)
		return
	}

	httpx.WriteJSON(w, http.StatusOK, t)
}

// HandleDelete implements the common Delete shape for an artist-scoped
// resource: parse the "id" route param, resolve the caller's accessible
// artist ids, and delete the row - 404ing if it doesn't exist or isn't
// accessible, same as HandleGet.
func HandleDelete[T any](w http.ResponseWriter, r *http.Request, res ScopedResource[T], noun string) {
	id, err := uuid.Parse(mux.Vars(r)["id"])
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid id")
		return
	}
	caller := auth.UserFromContext(r.Context())
	accessibleIDs, err := AccessibleArtistIDs(r.Context(), res.DB, caller.ID)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "failed to delete "+noun)
		return
	}

	found, err := res.Delete(r.Context(), id, accessibleIDs)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "failed to delete "+noun)
		return
	}
	if !found {
		httpx.WriteError(w, http.StatusNotFound, noun+" not found")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// HandleList implements the common List shape for an artist-scoped resource
// with no additional filters: resolve accessible scope (see
// ResolveListScope) and list every row in it. Resources with extra
// query-param filters (tourdates, for tour_id/rider_id) call
// ResolveListScope and res.List directly instead.
func HandleList[T any](w http.ResponseWriter, r *http.Request, res ScopedResource[T], noun string) {
	accessibleIDs, artistID, ok := ResolveListScope(w, r, res.DB, noun)
	if !ok {
		return
	}

	items, err := res.List(r.Context(), accessibleIDs, artistID, "")
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "failed to list "+noun+"s")
		return
	}

	httpx.WriteJSON(w, http.StatusOK, items)
}
