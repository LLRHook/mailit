package postgres

import (
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
)

// ErrNotFound is returned when a database query returns no rows.
var ErrNotFound = errors.New("record not found")

// notFound wraps pgx.ErrNoRows with a descriptive message.
func notFound(entity string) error {
	return fmt.Errorf("%s: %w", entity, ErrNotFound)
}

// isNoRows checks whether the error is pgx.ErrNoRows.
func isNoRows(err error) bool {
	return errors.Is(err, pgx.ErrNoRows)
}
