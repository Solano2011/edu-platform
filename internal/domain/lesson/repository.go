package lesson

import "context"

type Repository interface {
	FindPublished(ctx context.Context) ([]*Lesson, error)
	FindPublishedBySubject(ctx context.Context, subject Subject) ([]*Lesson, error)
	FindBySlug(ctx context.Context, slug string) (*Lesson, error)
	Save(ctx context.Context, lesson *Lesson) error
	Delete(ctx context.Context, slug string) error // Новый метод удаления по slug
}
