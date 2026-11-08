package lesson

import (
	"context"
	"edu-platform/internal/domain/lesson"
	"html/template"
	"time"
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
		ImageURL:      l.ImageURL(),
		ReadingTime:   l.ReadingTime(),
		DateFormatted: l.PublishedAt().Format("02.01.2006"),
	}, nil
}

func (uc *UseCase) CreateLesson(ctx context.Context, input CreateLessonInput) error {
	subj := lesson.Subject(input.Subject)
	cType := lesson.ContentType(input.ContentType)

	id := time.Now().Format("20060102150405")

	newLesson, err := lesson.NewLesson(
		id,
		input.Slug,
		input.Title,
		input.Summary,
		input.Body,
		subj,
		cType,
		input.VideoURL,
		input.ImageURL,
		input.ReadingTime,
	)
	if err != nil {
		return err
	}

	return uc.repo.Save(ctx, newLesson)
}

func (uc *UseCase) DeleteLesson(ctx context.Context, slug string) error {
	return uc.repo.Delete(ctx, slug)
}

func (uc *UseCase) GetLessonForEdit(ctx context.Context, slug string) (EditLessonDTO, error) {
	l, err := uc.repo.FindBySlug(ctx, slug)
	if err != nil {
		return EditLessonDTO{}, err
	}

	return EditLessonDTO{
		Slug:        l.Slug(),
		Title:       l.Title(),
		Summary:     l.Summary(),
		Body:        l.Body(),
		Subject:     string(l.Subject()),
		ContentType: string(l.ContentType()),
		VideoURL:    l.VideoURL(),
		ImageURL:    l.ImageURL(),
		ReadingTime: l.ReadingTime(),
	}, nil
}

func (uc *UseCase) UpdateLesson(ctx context.Context, input UpdateLessonInput) error {
	l, err := uc.repo.FindBySlug(ctx, input.Slug)
	if err != nil {
		return err
	}

	subj := lesson.Subject(input.Subject)
	cType := lesson.ContentType(input.ContentType)

	err = l.Update(
		input.Title,
		input.Summary,
		input.Body,
		subj,
		cType,
		input.VideoURL,
		input.ImageURL,
		input.ReadingTime,
	)
	if err != nil {
		return err
	}

	return uc.repo.Update(ctx, l)
}
