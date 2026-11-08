package user

import "errors"

var (
	ErrInvalidEmail        = errors.New("invalid email address format")
	ErrPasswordTooShort    = errors.New("password must be at least 8 characters long")
	ErrPasswordTooLong     = errors.New("password cannot exceed 72 bytes")
	ErrPasswordsDoNotMatch = errors.New("passwords do not match")
	ErrUserNotFound        = errors.New("user not found")
	ErrUserAlreadyExists   = errors.New("user already exists")
	ErrInvalidCredentials  = errors.New("invalid email or password")
	ErrUnauthorized        = errors.New("unauthorized")
	ErrForbidden           = errors.New("forbidden")
)
