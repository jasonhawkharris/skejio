# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project overview

Skejio ("Organize your tour!") is a tour-management API. A Go backend (`backend/`) exposes a REST API backed by Postgres: artists manage their own tour dates, and can grant managers/agents/labels delegated access to the artists they represent.

## Commands

Working directory for everything below: `backend/`.

**Run / build**
- `go run ./cmd/server` — starts the API on `:8080` (loads `backend/.env` for `DATABASE_URL`)
- `go build ./cmd/server` — builds a `server` binary
- `go vet ./...`

**Tests**
- `go test -p 1 ./...` — always use `-p 1`. All domain packages share one physical `skejio_test` Postgres database; without `-p 1`, `go test` runs each package's test binary in parallel and they race truncates/inserts against each other, producing flaky failures.
- `go test -p 1 ./internal/tourdates/... -run TestCreateTourDate` — run a single test
- Override which database tests hit with the `TEST_DATABASE_URL` env var (defaults to `postgres://<user>@localhost:5432/skejio_test?sslmode=disable`)

**Migrations** (via `golang-migrate`, `brew install golang-migrate`)
- `migrate create -ext sql -dir migrations -seq <name>` — scaffold a new migration (paired `.up.sql`/`.down.sql`)
- `migrate -database "postgres://<user>@localhost:5432/skejio?sslmode=disable" -path migrations up` — apply to the dev database
- Apply the same migration to `skejio_test` too — tests won't see new columns/tables otherwise

**Local Postgres setup** — installed via Homebrew (`postgresql@17`), running as a `brew services` background service, `trust` auth for local connections (no password). Two databases exist: `skejio` (dev) and `skejio_test` (tests) — both must be migrated independently.

## Architecture

**Package layout** — `cmd/server` is the sole entrypoint; everything else lives under `internal/`:

- `internal/api` — `NewRouter()` wires every route. It's the only package that imports all of `auth`, `tourdates`, `users`, `representatives`, `financials`, and `expenses`. Both `cmd/server` and `internal/testutil` build the router through this one function, so tests exercise real production routing.
- `internal/auth` — opaque session-token auth (not JWT). `POST /login` verifies a bcrypt hash and inserts a random token into a `sessions` table (24h expiry, `ON DELETE CASCADE` from `users`). `auth.Middleware` looks up the token on every request to a protected route and attaches `auth.CurrentUser{ID, UserType, Token}` to the request context, read back via `auth.UserFromContext(ctx)`. `POST /logout` deletes the session row.
- `internal/tourdates`, `internal/users`, `internal/representatives`, `internal/financials`, `internal/expenses` — one package per resource. Each owns its model, its `Handler{DB *pgxpool.Pool}`, and that handler's methods (`List`, `Get`, `Create`, `Patch`, `Delete` as applicable) — there's no shared "App" struct; each `Handler` gets the pool independently.
- `internal/httpx`, `internal/dberr`, `internal/password`, `internal/db` — small leaf packages: JSON response helpers, Postgres error-code classification (e.g. unique-violation detection), bcrypt hashing, and connection setup, respectively.
- `internal/testutil` — shared test harness. Builds the real router against `skejio_test` and exposes fixture helpers (`CreateTestUserOfType`, `CreateAndLoginTestUser`, `CreateTestTourDate`, `AddRepresentative`, `TruncateTables`, ...) so no domain package duplicates setup. Every domain test file is an *external* test package (e.g. `package tourdates_test`) whose `TestMain` is just `func TestMain(m *testing.M) { testutil.Run(m) }` — this has to be an external package because `internal/api` (which `testutil` depends on) imports the domain package itself, so an internal (white-box) test package would create an import cycle.

**Authorization model** — every resource is scoped to the authenticated caller. Access denied on someone else's resource returns 404, not 403, so existence isn't leaked:

- `tourdates` are owned by `user_id`. A caller may act on their own tourdates plus any artist's tourdates they represent (see below). `tourdates.AccessibleArtistIDs()` (exported so `internal/financials` can reuse it) computes that set (caller + everyone they represent) and every query filters on it via `user_id = ANY($n::uuid[])`. `GET /tourdates` merges across all accessible artists unless `?artist_id=` narrows to one (which must be in the accessible set, or 404s). `POST /tourdates` defaults ownership to the caller, or to an explicit `artist_id` in the body if the caller represents that artist (403 otherwise).
- `financials` is one-to-one with `tourdates` (`tourdate_id` is `UNIQUE`, FK `ON DELETE CASCADE`) and kept in its own table rather than more nullable columns on `tourdates`, since it's a distinct concern (money) likely to grow its own fields. Nested under the tourdate instead of being its own top-level resource, since it's never listed independently: `GET/POST/PATCH/DELETE /tourdates/{id}/financials`. Access follows the same rule as the tourdate itself (owner or representative; 404 otherwise). `POST` 409s if a row already exists for that tourdate (use `PATCH` instead). `fee`/`tips` are nullable `NUMERIC(10,2)`.
- `expenses` is one-to-*many* with `tourdates` (plain FK, no uniqueness) — a tourdate can have any number of expense line items. `GET/POST /tourdates/{id}/expenses` (list/create) and `GET/PATCH/DELETE /tourdates/{id}/expenses/{expenseID}` (item), same owner-or-representative access rule as `financials`. `category` is `CHECK`-constrained to `TRAVEL`/`LODGING`/`CREW`/`GEAR`/`OTHER` (validated in Go too, same pattern as `users.user_type`); `amount` is required `NUMERIC(10,2)`; `description` is nullable.
- `GET /tourdates/{id}/summary` (`financials.Handler.Summary`) computes `fee`, `tips`, `total_expenses` (`SUM` over `expenses`), and `net` (`fee + tips - total_expenses`) on every read rather than storing it, so it can't drift out of sync with the underlying `financials`/`expenses` rows. `fee`/`tips` are `nil` if no `financials` row exists yet; `total_expenses` is `0` (never `nil`) if there are no expenses.
- `users` are strictly self-only: `GET/PATCH/DELETE /users/{id}` 404 for any id but the caller's own. There is no "list all users" endpoint. Account creation is unauthenticated: `POST /sign-up` takes `name`/`email`/`password`/`user_type` (any of `ARTIST`/`MANAGER`/`AGENT`/`CREW`/`LABEL`), hashes the password, and 409s on a duplicate email.
- `artist_representatives` is a many-to-many join table (`artist_id`, `representative_id`, both FK → `users` `ON DELETE CASCADE`, `UNIQUE(artist_id, representative_id)`) granting a `MANAGER`/`AGENT`/`LABEL` user access to an `ARTIST`'s tourdates. Endpoints (`internal/representatives`):
  - `POST /representatives` — only an `ARTIST` may call this (403 otherwise); body `{"representative_id": "<uuid>"}` must reference an existing `MANAGER`/`AGENT`/`LABEL` user (400 otherwise); rejects granting to yourself and duplicate grants (409)
  - `GET /representatives` — artist-only; lists the representatives the caller has granted access to
  - `GET /represented-artists` — `MANAGER`/`AGENT`/`LABEL`-only; lists the artists the caller represents
  - `DELETE /representatives/{id}` — either party (the artist or the representative) may revoke; an id the caller isn't part of 404s
- `users.user_type` is DB-`CHECK`-constrained to `ARTIST`/`MANAGER`/`AGENT`/`CREW`/`LABEL`, and validated again in Go before the query runs so an invalid value gets a clean 400 instead of a raw constraint-violation error.

**Request/response conventions**:
- `PATCH` handlers decode into `map[string]json.RawMessage` so an omitted field is left unchanged while an explicit JSON `null` (only meaningful on nullable columns, e.g. `tourdates.state`) clears it.
- `tourdates.Date` wraps `time.Time` with custom `MarshalJSON`/`UnmarshalJSON` to serialize as a bare `YYYY-MM-DD` string, matching the Postgres `DATE` column (no time-of-day/timezone).
- Errors are always `{"error": "<message>"}`, written via `httpx.WriteError`.

## Basic Instructions

Always grill me with questions about implementation. No question is too small.
