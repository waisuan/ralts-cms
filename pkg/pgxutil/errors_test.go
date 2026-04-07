package pgxutil_test

import (
	"errors"
	"fmt"
	"testing"

	"ralts-cms/pkg/pgxutil"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/assert"
)

func TestIsUniqueViolation(t *testing.T) {
	t.Parallel()

	unique := &pgconn.PgError{Code: "23505"}
	other := &pgconn.PgError{Code: "23503"}

	t.Run("direct PgError 23505", func(t *testing.T) {
		t.Parallel()
		assert.True(t, pgxutil.IsUniqueViolation(unique))
	})

	t.Run("wrapped PgError", func(t *testing.T) {
		t.Parallel()
		assert.True(t, pgxutil.IsUniqueViolation(fmt.Errorf("repo: %w", unique)))
	})

	t.Run("other SQLSTATE", func(t *testing.T) {
		t.Parallel()
		assert.False(t, pgxutil.IsUniqueViolation(other))
	})

	t.Run("non-pg error", func(t *testing.T) {
		t.Parallel()
		assert.False(t, pgxutil.IsUniqueViolation(errors.New("duplicate key value violates unique constraint")))
	})
}
