package db

import "github.com/jackc/pgx/v5"

// ScanAll drains rows into a slice using scan, closing rows and reporting
// any row-scan or iteration error. Returns an empty (never nil) slice when
// there are no rows, so callers can serialize it directly as JSON "[]".
func ScanAll[T any](rows pgx.Rows, scan func(pgx.Row) (T, error)) ([]T, error) {
	defer rows.Close()

	results := []T{}
	for rows.Next() {
		t, err := scan(rows)
		if err != nil {
			return nil, err
		}
		results = append(results, t)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return results, nil
}
