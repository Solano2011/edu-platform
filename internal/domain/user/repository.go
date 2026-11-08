package user

import (
	"context"
	"time"
)

type Repository interface {
	FindByEmail(ctx context.Context, email Email) (*User, error)
	FindByID(ctx context.Context, id string) (*User, error)
	Save(ctx context.Context, user *User) error
}

type SessionRepository interface {
	Create(ctx context.Context, token string, userID string, duration time.Duration) error
	GetUserIDByToken(ctx context.Context, token string) (string, error)
	Delete(ctx context.Context, token string) error
}

type InviteRepository interface {
	Save(ctx context.Context, invite *Invite) error
	FindByCode(ctx context.Context, code string) (*Invite, error)
	FindAll(ctx context.Context) ([]*Invite, error)
	Update(ctx context.Context, invite *Invite) error
}
