package memory

import (
	"context"
	"sync"
	"time"

	"edu-platform/internal/domain/user"
)

type UserRepository struct {
	mu    sync.RWMutex
	users map[string]*user.User // key: email.String()
}

func NewUserRepository() *UserRepository {
	repo := &UserRepository{
		users: make(map[string]*user.User),
	}
	repo.seedAdmin()
	return repo
}

func (r *UserRepository) seedAdmin() {
	email, _ := user.NewEmail("admin@edu-platform.ru")
	// Пароль администратора по умолчанию: StrongAdminPass123!
	passHash, _ := user.NewPasswordHashFromPlain("StrongAdminPass123!")
	admin := user.NewUser("admin-root", email, passHash, user.RoleAdmin)
	r.users[email.String()] = admin
}

func (r *UserRepository) FindByEmail(ctx context.Context, email user.Email) (*user.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	u, exists := r.users[email.String()]
	if !exists {
		return nil, user.ErrUserNotFound
	}
	return u, nil
}

func (r *UserRepository) FindByID(ctx context.Context, id string) (*user.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, u := range r.users {
		if u.ID() == id {
			return u, nil
		}
	}
	return nil, user.ErrUserNotFound
}

func (r *UserRepository) Save(ctx context.Context, u *user.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.users[u.Email().String()]; exists {
		return user.ErrUserAlreadyExists
	}

	r.users[u.Email().String()] = u
	return nil
}

type sessionItem struct {
	userID    string
	expiresAt time.Time
}

type SessionRepository struct {
	mu       sync.RWMutex
	sessions map[string]sessionItem // key: token
}

func NewSessionRepository() *SessionRepository {
	return &SessionRepository{
		sessions: make(map[string]sessionItem),
	}
}

func (r *SessionRepository) Create(ctx context.Context, token string, userID string, duration time.Duration) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.sessions[token] = sessionItem{
		userID:    userID,
		expiresAt: time.Now().Add(duration),
	}
	return nil
}

func (r *SessionRepository) GetUserIDByToken(ctx context.Context, token string) (string, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	item, exists := r.sessions[token]
	if !exists {
		return "", user.ErrUnauthorized
	}

	if time.Now().After(item.expiresAt) {
		return "", user.ErrUnauthorized
	}

	return item.userID, nil
}

func (r *SessionRepository) Delete(ctx context.Context, token string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	delete(r.sessions, token)
	return nil
}

type InviteRepository struct {
	mu      sync.RWMutex
	invites map[string]*user.Invite // key: code
}

func NewInviteRepository() *InviteRepository {
	return &InviteRepository{
		invites: make(map[string]*user.Invite),
	}
}

func (r *InviteRepository) Save(ctx context.Context, inv *user.Invite) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.invites[inv.Code()] = inv
	return nil
}

func (r *InviteRepository) FindByCode(ctx context.Context, code string) (*user.Invite, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	inv, exists := r.invites[code]
	if !exists {
		return nil, user.ErrInviteNotFound
	}
	return inv, nil
}

func (r *InviteRepository) FindAll(ctx context.Context) ([]*user.Invite, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	list := make([]*user.Invite, 0, len(r.invites))
	for _, inv := range r.invites {
		list = append(list, inv)
	}
	return list, nil
}

func (r *InviteRepository) Update(ctx context.Context, inv *user.Invite) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.invites[inv.Code()]; !exists {
		return user.ErrInviteNotFound
	}
	r.invites[inv.Code()] = inv
	return nil
}
