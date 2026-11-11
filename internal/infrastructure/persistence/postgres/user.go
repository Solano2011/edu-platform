package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"edu-platform/internal/domain/user"
)

type UserRepository struct {
	pool *pgxpool.Pool
}

func NewUserRepository(pool *pgxpool.Pool) *UserRepository {
	return &UserRepository{pool: pool}
}

func (r *UserRepository) Save(ctx context.Context, u *user.User) error {
	query := `
		INSERT INTO users (id, email, password_hash, role, created_at)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (id) DO UPDATE SET
			email = EXCLUDED.email,
			password_hash = EXCLUDED.password_hash,
			role = EXCLUDED.role;
	`
	_, err := r.pool.Exec(ctx, query,
		u.ID(),
		u.Email().String(),
		string(u.PasswordHash().Bytes()),
		string(u.Role()),
		u.CreatedAt(),
	)
	return err
}

func (r *UserRepository) FindByEmail(ctx context.Context, email user.Email) (*user.User, error) {
	query := `
		SELECT id, email, password_hash, role, created_at
		FROM users
		WHERE email = $1;
	`
	var (
		id           string
		userEmail    string
		passwordHash string
		roleStr      string
		createdAt    time.Time
	)

	err := r.pool.QueryRow(ctx, query, email.String()).Scan(
		&id,
		&userEmail,
		&passwordHash,
		&roleStr,
		&createdAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, user.ErrUserNotFound
		}
		return nil, err
	}

	return user.Restore(id, userEmail, passwordHash, user.Role(roleStr), createdAt), nil
}

func (r *UserRepository) FindByID(ctx context.Context, id string) (*user.User, error) {
	query := `
		SELECT id, email, password_hash, role, created_at
		FROM users
		WHERE id = $1;
	`
	var (
		userID       string
		userEmail    string
		passwordHash string
		roleStr      string
		createdAt    time.Time
	)

	err := r.pool.QueryRow(ctx, query, id).Scan(
		&userID,
		&userEmail,
		&passwordHash,
		&roleStr,
		&createdAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, user.ErrUserNotFound
		}
		return nil, err
	}

	return user.Restore(userID, userEmail, passwordHash, user.Role(roleStr), createdAt), nil
}
