package lesson

import (
	"context"
	"edu-platform/internal/domain/lesson"
	"html/template"
)

type UseCase struct {
	repo lesson.Repository
}

func NewUseCase(repo lesson.Repository) *UseCase {
	return &UseCase{repo: repo}
}

func (uc *UseCase) GetLessonsBySubject(ctx context.Context, subjectCode string) (string, []CardDTO, error) {
	subj := lesson.Subject(subjectCode)
	if !subj.IsValid() {
		return "", nil, lesson.ErrInvalidSubject
	}

	lessons, err := uc.repo.FindPublishedBySubject(ctx, subj)
	if err != nil {
		return "", nil, err
	}

	dtos := make([]CardDTO, 0, len(lessons))
	for _, l := range lessons {
		dtos = append(dtos, CardDTO{
			Slug:          l.Slug(),
			Title:         l.Title(),
			Summary:       l.Summary(),
			Subject:       l.Subject().DisplayName(),
			ContentType:   string(l.ContentType()),
			HasVideo:      l.VideoURL() != "",
			ReadingTime:   l.ReadingTime(),
			DateFormatted: l.PublishedAt().Format("02.01.2006"),
		})
	}

	return subj.DisplayName(), dtos, nil
}

func (uc *UseCase) GetLessonBySlug(ctx context.Context, slug string) (DetailDTO, error) {
	l, err := uc.repo.FindBySlug(ctx, slug)
	if err != nil {
		return DetailDTO{}, err
	}

	return DetailDTO{
		Slug:          l.Slug(),
		Title:         l.Title(),
		Summary:       l.Summary(),
		BodyHTML:      template.HTML(l.Body()),
		SubjectCode:   string(l.Subject()),
		SubjectName:   l.Subject().DisplayName(),
		HasVideo:      l.VideoURL() != "",
		VideoURL:      l.VideoURL(),
		ReadingTime:   l.ReadingTime(),
		DateFormatted: l.PublishedAt().Format("02.01.2006"),
	}, nil
}
