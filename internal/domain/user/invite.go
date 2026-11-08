package user

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"time"
)

var (
	ErrInviteNotFound = errors.New("инвайт-код не найден")
	ErrInviteUsed     = errors.New("инвайт-код уже был использован")
	ErrInviteExpired  = errors.New("срок действия инвайт-кода истёк")
)

type Invite struct {
	code        string
	studentNote string // Имя или пометка ученика, для которого создан код
	role        Role
	isUsed      bool
	usedBy      string // ID созданного пользователя
	createdAt   time.Time
	expiresAt   time.Time
}

func NewInvite(studentNote string, role Role, duration time.Duration) (*Invite, error) {
	bytes := make([]byte, 16) // 32-символьный криптографический hex-код
	if _, err := rand.Read(bytes); err != nil {
		return nil, err
	}

	return &Invite{
		code:        hex.EncodeToString(bytes),
		studentNote: studentNote,
		role:        role,
		isUsed:      false,
		createdAt:   time.Now(),
		expiresAt:   time.Now().Add(duration),
	}, nil
}

func (i *Invite) Code() string         { return i.code }
func (i *Invite) StudentNote() string  { return i.studentNote }
func (i *Invite) Role() Role           { return i.role }
func (i *Invite) IsUsed() bool         { return i.isUsed }
func (i *Invite) UsedBy() string       { return i.usedBy }
func (i *Invite) CreatedAt() time.Time { return i.createdAt }
func (i *Invite) ExpiresAt() time.Time { return i.expiresAt }

func (i *Invite) MarkAsUsed(userID string) error {
	if i.isUsed {
		return ErrInviteUsed
	}
	if time.Now().After(i.expiresAt) {
		return ErrInviteExpired
	}
	i.isUsed = true
	i.usedBy = userID
	return nil
}

func (i *Invite) IsValid() error {
	if i.isUsed {
		return ErrInviteUsed
	}
	if time.Now().After(i.expiresAt) {
		return ErrInviteExpired
	}
	return nil
}
