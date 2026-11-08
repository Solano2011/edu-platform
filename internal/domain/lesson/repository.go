package lesson

import "context"

type Repository interface {
	FindPublished(ctx context.Context) ([]*Lesson, error)
	FindPublishedBySubject(ctx context.Context, subject Subject) ([]*Lesson, error)
	FindBySlug(ctx context.Context, slug string) (*Lesson, error)
}
