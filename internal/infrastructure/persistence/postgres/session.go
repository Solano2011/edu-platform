package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"edu-platform/internal/domain/user"
)

type SessionRepository struct {
	pool *pgxpool.Pool
}

func NewSessionRepository(pool *pgxpool.Pool) *SessionRepository {
	return &SessionRepository{pool: pool}
}

func (r *SessionRepository) Create(ctx context.Context, token string, userID string, duration time.Duration) error {
	query := `
		INSERT INTO sessions (token, user_id, expires_at)
		VALUES ($1, $2, $3)
		ON CONFLICT (token) DO UPDATE SET
			expires_at = EXCLUDED.expires_at;
	`
	expiresAt := time.Now().Add(duration)
	_, err := r.pool.Exec(ctx, query, token, userID, expiresAt)
	return err
}

func (r *SessionRepository) GetUserIDByToken(ctx context.Context, token string) (string, error) {
	query := `
		SELECT user_id
		FROM sessions
		WHERE token = $1 AND expires_at > NOW();
	`
	var userID string
	err := r.pool.QueryRow(ctx, query, token).Scan(&userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", user.ErrUnauthorized
		}
		return "", err
	}

	return userID, nil
}

func (r *SessionRepository) Delete(ctx context.Context, token string) error {
	query := `DELETE FROM sessions WHERE token = $1;`
	_, err := r.pool.Exec(ctx, query, token)
	return err
}
