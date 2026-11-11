package user

import "time"

type User struct {
	id           string
	email        Email
	passwordHash PasswordHash
	role         Role
	createdAt    time.Time
}

func NewUser(id string, email Email, passwordHash PasswordHash, role Role) *User {
	return &User{
		id:           id,
		email:        email,
		passwordHash: passwordHash,
		role:         role,
		createdAt:    time.Now(),
	}
}

func (u *User) ID() string                 { return u.id }
func (u *User) Email() Email               { return u.email }
func (u *User) PasswordHash() PasswordHash { return u.passwordHash }
func (u *User) Role() Role                 { return u.role }
func (u *User) CreatedAt() time.Time       { return u.createdAt }
func (u *User) IsAdmin() bool              { return u.role == RoleAdmin }

// Restore восстанавливает сущность User из постоянного хранилища без повторной валидации
func Restore(id string, email string, passwordHash string, role Role, createdAt time.Time) *User {
	return &User{
		id:           id,
		email:        Email{value: email},
		passwordHash: NewPasswordHashFromHash([]byte(passwordHash)),
		role:         role,
		createdAt:    createdAt,
	}
}
