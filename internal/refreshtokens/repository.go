// Package refreshtokens persists opaque refresh tokens (hashed) for sliding sessions.
package refreshtokens

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

//go:generate mockgen -destination=../refreshtokens/mock_refreshtokens_repository.go -package=refreshtokens -source=repository.go

// ErrNotFound is returned when no active refresh token matches the hash.
var ErrNotFound = errors.New("refresh token not found or revoked")

// Token is a row used by the server after lookup (never the raw client string).
type Token struct {
	ID     uuid.UUID
	UserID int64
}

type db struct {
	client *pgxpool.Pool
}

// Repository stores and rotates opaque refresh tokens.
type Repository interface {
	Create(ctx context.Context, userID int64, tokenHash []byte, expiresAt time.Time) error
	// GetActiveByHash returns a token row if the hash matches a non-revoked, unexpired row.
	GetActiveByHash(ctx context.Context, tokenHash []byte) (*Token, error)
	// Rotate revokes the old row and inserts a new token in one transaction; returns the new id.
	Rotate(ctx context.Context, oldID uuid.UUID, userID int64, newHash []byte, newExpires time.Time) (newID uuid.UUID, err error)
	RevokeByID(ctx context.Context, id uuid.UUID) error
	RevokeAllForUser(ctx context.Context, userID int64) error
}

// NewRepository creates a token repository.
func NewRepository(client *pgxpool.Pool) Repository {
	return &db{client: client}
}

func (r *db) Create(ctx context.Context, userID int64, tokenHash []byte, expiresAt time.Time) error {
	const q = `
		INSERT INTO refresh_tokens (user_id, token_hash, expires_at)
		VALUES ($1, $2, $3)
	`
	_, err := r.client.Exec(ctx, q, userID, tokenHash, expiresAt)
	if err != nil {
		return fmt.Errorf("create refresh token: %w", err)
	}
	return nil
}

func (r *db) GetActiveByHash(ctx context.Context, tokenHash []byte) (*Token, error) {
	const q = `
		SELECT id, user_id
		FROM refresh_tokens
		WHERE token_hash = $1
		  AND revoked_at IS NULL
		  AND expires_at > NOW()
	`
	var t Token
	err := r.client.QueryRow(ctx, q, tokenHash).Scan(&t.ID, &t.UserID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get refresh token: %w", err)
	}
	return &t, nil
}

func (r *db) Rotate(ctx context.Context, oldID uuid.UUID, userID int64, newHash []byte, newExpires time.Time) (uuid.UUID, error) {
	tx, err := r.client.Begin(ctx)
	if err != nil {
		return uuid.Nil, fmt.Errorf("begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	const ins = `
		INSERT INTO refresh_tokens (user_id, token_hash, expires_at)
		VALUES ($1, $2, $3)
		RETURNING id
	`
	var newID uuid.UUID
	if err := tx.QueryRow(ctx, ins, userID, newHash, newExpires).Scan(&newID); err != nil {
		return uuid.Nil, fmt.Errorf("insert refresh token: %w", err)
	}

	const upd = `
		UPDATE refresh_tokens
		SET revoked_at = NOW(), replaced_by = $1
		WHERE id = $2 AND revoked_at IS NULL
	`
	tag, err := tx.Exec(ctx, upd, newID, oldID)
	if err != nil {
		return uuid.Nil, fmt.Errorf("revoke old refresh token: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return uuid.Nil, ErrNotFound
	}

	if err := tx.Commit(ctx); err != nil {
		return uuid.Nil, fmt.Errorf("commit: %w", err)
	}
	return newID, nil
}

func (r *db) RevokeByID(ctx context.Context, id uuid.UUID) error {
	const q = `
		UPDATE refresh_tokens
		SET revoked_at = NOW()
		WHERE id = $1 AND revoked_at IS NULL
	`
	tag, err := r.client.Exec(ctx, q, id)
	if err != nil {
		return fmt.Errorf("revoke refresh token: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return nil
	}
	return nil
}

func (r *db) RevokeAllForUser(ctx context.Context, userID int64) error {
	const q = `
		UPDATE refresh_tokens
		SET revoked_at = NOW()
		WHERE user_id = $1 AND revoked_at IS NULL
	`
	_, err := r.client.Exec(ctx, q, userID)
	if err != nil {
		return fmt.Errorf("revoke all refresh tokens: %w", err)
	}
	return nil
}
