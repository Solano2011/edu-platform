package user

import (
	"net/mail"
	"strings"
	"unicode/utf8"

	"golang.org/x/crypto/bcrypt"
)

type Role string

const (
	RoleAdmin   Role = "admin"
	RoleStudent Role = "student"
)

func (r Role) IsValid() bool {
	return r == RoleAdmin || r == RoleStudent
}

// Email инкапсулирует строгую проверку формата RFC 5322 и нормализацию
type Email struct {
	value string
}

func NewEmail(raw string) (Email, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" || len(trimmed) > 254 {
		return Email{}, ErrInvalidEmail
	}

	addr, err := mail.ParseAddress(trimmed)
	if err != nil || addr.Address != trimmed {
		return Email{}, ErrInvalidEmail
	}

	return Email{value: strings.ToLower(addr.Address)}, nil
}

func (e Email) String() string {
	return e.value
}

// PasswordHash инкапсулирует безопасную работу с паролем через bcrypt
type PasswordHash struct {
	hash []byte
}

func NewPasswordHashFromPlain(plain string) (PasswordHash, error) {
	length := utf8.RuneCountInString(plain)
	if length < 8 {
		return PasswordHash{}, ErrPasswordTooShort
	}
	// bcrypt отсекает всё после 72 байт, поэтому ограничиваем ввод явно
	if len([]byte(plain)) > 72 {
		return PasswordHash{}, ErrPasswordTooLong
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(plain), 12)
	if err != nil {
		return PasswordHash{}, err
	}

	return PasswordHash{hash: hash}, nil
}

func NewPasswordHashFromHash(hash []byte) PasswordHash {
	return PasswordHash{hash: hash}
}

func (p PasswordHash) Compare(plain string) bool {
	return bcrypt.CompareHashAndPassword(p.hash, []byte(plain)) == nil
}

func (p PasswordHash) Bytes() []byte {
	return p.hash
}
