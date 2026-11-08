package lesson

import "errors"

var (
	ErrEmptyTitle     = errors.New("lesson title cannot be empty")
	ErrInvalidType    = errors.New("invalid content type")
	ErrInvalidSubject = errors.New("invalid subject")
	ErrLessonNotFound = errors.New("lesson not found")
)
