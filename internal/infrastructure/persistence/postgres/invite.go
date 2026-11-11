package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"edu-platform/internal/domain/user"
)

type InviteRepository struct {
	pool *pgxpool.Pool
}

func NewInviteRepository(pool *pgxpool.Pool) *InviteRepository {
	return &InviteRepository{pool: pool}
}

func (r *InviteRepository) Save(ctx context.Context, inv *user.Invite) error {
	query := `
		INSERT INTO invites (code, student_note, role, is_used, used_by, created_at, expires_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (code) DO UPDATE SET
			is_used = EXCLUDED.is_used,
			used_by = EXCLUDED.used_by;
	`
	_, err := r.pool.Exec(ctx, query,
		inv.Code(),
		inv.StudentNote(),
		string(inv.Role()),
		inv.IsUsed(),
		inv.UsedBy(),
		inv.CreatedAt(),
		inv.ExpiresAt(),
	)
	return err
}

func (r *InviteRepository) Update(ctx context.Context, inv *user.Invite) error {
	return r.Save(ctx, inv)
}

func (r *InviteRepository) FindByCode(ctx context.Context, code string) (*user.Invite, error) {
	query := `
		SELECT code, student_note, role, is_used, used_by, created_at, expires_at
		FROM invites
		WHERE code = $1;
	`
	var (
		c           string
		studentNote string
		roleStr     string
		isUsed      bool
		usedBy      string
		createdAt   time.Time
		expiresAt   time.Time
	)

	err := r.pool.QueryRow(ctx, query, code).Scan(
		&c,
		&studentNote,
		&roleStr,
		&isUsed,
		&usedBy,
		&createdAt,
		&expiresAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, user.ErrInviteNotFound
		}
		return nil, err
	}

	return user.RestoreInvite(c, studentNote, user.Role(roleStr), isUsed, usedBy, createdAt, expiresAt), nil
}

func (r *InviteRepository) FindAll(ctx context.Context) ([]*user.Invite, error) {
	query := `
		SELECT code, student_note, role, is_used, used_by, created_at, expires_at
		FROM invites
		ORDER BY created_at DESC;
	`
	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var invites []*user.Invite
	for rows.Next() {
		var (
			code        string
			studentNote string
			roleStr     string
			isUsed      bool
			usedBy      string
			createdAt   time.Time
			expiresAt   time.Time
		)

		if err := rows.Scan(&code, &studentNote, &roleStr, &isUsed, &usedBy, &createdAt, &expiresAt); err != nil {
			return nil, err
		}

		invites = append(invites, user.RestoreInvite(code, studentNote, user.Role(roleStr), isUsed, usedBy, createdAt, expiresAt))
	}

	return invites, rows.Err()
}
