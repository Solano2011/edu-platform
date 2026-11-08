package auth

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"time"

	"edu-platform/internal/domain/user"
)

type UseCase struct {
	userRepo    user.Repository
	sessionRepo user.SessionRepository
	inviteRepo  user.InviteRepository
}

func NewUseCase(userRepo user.Repository, sessionRepo user.SessionRepository, inviteRepo user.InviteRepository) *UseCase {
	return &UseCase{
		userRepo:    userRepo,
		sessionRepo: sessionRepo,
		inviteRepo:  inviteRepo,
	}
}

type InviteDTO struct {
	Code        string
	StudentNote string
	IsUsed      bool
	UsedBy      string
	ExpiresAt   string
}

// CreateInvite — метод для репетитора (создает инвайт на 30 дней)
func (uc *UseCase) CreateInvite(ctx context.Context, studentNote string) (*user.Invite, error) {
	inv, err := user.NewInvite(studentNote, user.RoleStudent, 30*24*time.Hour)
	if err != nil {
		return nil, err
	}

	if err := uc.inviteRepo.Save(ctx, inv); err != nil {
		return nil, err
	}

	return inv, nil
}

func (uc *UseCase) ListInvites(ctx context.Context) ([]InviteDTO, error) {
	invites, err := uc.inviteRepo.FindAll(ctx)
	if err != nil {
		return nil, err
	}

	dtos := make([]InviteDTO, 0, len(invites))
	for _, inv := range invites {
		dtos = append(dtos, InviteDTO{
			Code:        inv.Code(),
			StudentNote: inv.StudentNote(),
			IsUsed:      inv.IsUsed(),
			UsedBy:      inv.UsedBy(),
			ExpiresAt:   inv.ExpiresAt().Format("02.01.2006 15:04"),
		})
	}
	return dtos, nil
}

// RegisterWithInvite проверяет валидность инвайта, создает ученика и сжигает инвайт
func (uc *UseCase) RegisterWithInvite(ctx context.Context, inviteCode, emailRaw, passwordRaw, confirmPasswordRaw string) (string, *user.User, error) {
	// 1. Проверяем валидность кода
	inv, err := uc.inviteRepo.FindByCode(ctx, inviteCode)
	if err != nil {
		return "", nil, user.ErrInviteNotFound
	}
	if err := inv.IsValid(); err != nil {
		return "", nil, err
	}

	// 2. Проверяем пароли
	if passwordRaw != confirmPasswordRaw {
		return "", nil, user.ErrPasswordsDoNotMatch
	}

	// 3. Валидируем Email
	email, err := user.NewEmail(emailRaw)
	if err != nil {
		return "", nil, err
	}

	existing, _ := uc.userRepo.FindByEmail(ctx, email)
	if existing != nil {
		return "", nil, user.ErrUserAlreadyExists
	}

	// 4. Хешируем bcrypt
	passHash, err := user.NewPasswordHashFromPlain(passwordRaw)
	if err != nil {
		return "", nil, err
	}

	// 5. Создаем пользователя с ролью из инвайта (RoleStudent)
	userID := time.Now().Format("20060102150405")
	newUser := user.NewUser(userID, email, passHash, inv.Role())

	if err := uc.userRepo.Save(ctx, newUser); err != nil {
		return "", nil, err
	}

	// 6. Помечаем инвайт как использованный
	if err := inv.MarkAsUsed(newUser.ID()); err != nil {
		return "", nil, err
	}
	_ = uc.inviteRepo.Update(ctx, inv)

	// 7. Сразу логиним ученика
	tokenBytes := make([]byte, 32)
	_, _ = rand.Read(tokenBytes)
	token := hex.EncodeToString(tokenBytes)

	_ = uc.sessionRepo.Create(ctx, token, newUser.ID(), 7*24*time.Hour)

	return token, newUser, nil
}

// Существующие методы Login, Authenticate, Logout остаются без изменений...
func (uc *UseCase) Login(ctx context.Context, emailRaw, passwordRaw string) (string, *user.User, error) {
	email, err := user.NewEmail(emailRaw)
	if err != nil {
		return "", nil, user.ErrInvalidCredentials
	}

	u, err := uc.userRepo.FindByEmail(ctx, email)
	if err != nil {
		return "", nil, user.ErrInvalidCredentials
	}

	if !u.PasswordHash().Compare(passwordRaw) {
		return "", nil, user.ErrInvalidCredentials
	}

	tokenBytes := make([]byte, 32)
	_, _ = rand.Read(tokenBytes)
	token := hex.EncodeToString(tokenBytes)

	err = uc.sessionRepo.Create(ctx, token, u.ID(), 7*24*time.Hour)
	if err != nil {
		return "", nil, err
	}

	return token, u, nil
}

func (uc *UseCase) Authenticate(ctx context.Context, token string) (*user.User, error) {
	if token == "" {
		return nil, user.ErrUnauthorized
	}

	userID, err := uc.sessionRepo.GetUserIDByToken(ctx, token)
	if err != nil {
		return nil, err
	}

	return uc.userRepo.FindByID(ctx, userID)
}

func (uc *UseCase) Logout(ctx context.Context, token string) error {
	if token == "" {
		return nil
	}
	return uc.sessionRepo.Delete(ctx, token)
}
