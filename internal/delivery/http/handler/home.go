package handler

import (
	"errors"
	"html/template"
	"log/slog"
	"net/http"
	"path/filepath"

	"edu-platform/internal/domain/lesson"
	lessonUC "edu-platform/internal/usecase/lesson"
)

type HomeHandler struct {
	lessonUseCase *lessonUC.UseCase
	homeTmpl      *template.Template
	subjectTmpl   *template.Template
	lessonTmpl    *template.Template
	logger        *slog.Logger
}

func NewHomeHandler(uc *lessonUC.UseCase, templatesDir string, logger *slog.Logger) (*HomeHandler, error) {
	homeTmpl, err := template.ParseFiles(
		filepath.Join(templatesDir, "layout.html"),
		filepath.Join(templatesDir, "home.html"),
	)
	if err != nil {
		return nil, err
	}

	subjectTmpl, err := template.ParseFiles(
		filepath.Join(templatesDir, "layout.html"),
		filepath.Join(templatesDir, "subject.html"),
	)
	if err != nil {
		return nil, err
	}

	lessonTmpl, err := template.ParseFiles(
		filepath.Join(templatesDir, "layout.html"),
		filepath.Join(templatesDir, "lesson.html"),
	)
	if err != nil {
		return nil, err
	}

	return &HomeHandler{
		lessonUseCase: uc,
		homeTmpl:      homeTmpl,
		subjectTmpl:   subjectTmpl,
		lessonTmpl:    lessonTmpl,
		logger:        logger,
	}, nil
}

type homePageData struct {
	Title string
}

type subjectPageData struct {
	Title       string
	SubjectName string
	Lessons     []lessonUC.CardDTO
}

type lessonPageData struct {
	Title  string
	Lesson lessonUC.DetailDTO
}

func (h *HomeHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	data := homePageData{
		Title: "Учебная платформа | Выбор направления",
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := h.homeTmpl.ExecuteTemplate(w, "layout", data); err != nil {
		h.logger.Error("failed to render home template", slog.String("error", err.Error()))
	}
}

func (h *HomeHandler) ServeSubject(w http.ResponseWriter, r *http.Request) {
	subjectCode := r.PathValue("subject")

	subjectName, lessons, err := h.lessonUseCase.GetLessonsBySubject(r.Context(), subjectCode)
	if err != nil {
		if errors.Is(err, lesson.ErrInvalidSubject) {
			http.NotFound(w, r)
			return
		}
		h.logger.Error("failed to get lessons by subject", slog.String("error", err.Error()))
		http.Error(w, "Внутренняя ошибка сервера", http.StatusInternalServerError)
		return
	}

	data := subjectPageData{
		Title:       subjectName + " | Учебные материалы",
		SubjectName: subjectName,
		Lessons:     lessons,
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := h.subjectTmpl.ExecuteTemplate(w, "layout", data); err != nil {
		h.logger.Error("failed to render subject template", slog.String("error", err.Error()))
	}
}

// ServeLesson отдает страницу чтения конкретного урока и просмотра видео
func (h *HomeHandler) ServeLesson(w http.ResponseWriter, r *http.Request) {
	slug := r.PathValue("slug")

	detail, err := h.lessonUseCase.GetLessonBySlug(r.Context(), slug)
	if err != nil {
		if errors.Is(err, lesson.ErrLessonNotFound) {
			http.NotFound(w, r)
			return
		}
		h.logger.Error("failed to get lesson by slug", slog.String("slug", slug), slog.String("error", err.Error()))
		http.Error(w, "Внутренняя ошибка сервера", http.StatusInternalServerError)
		return
	}

	data := lessonPageData{
		Title:  detail.Title + " | " + detail.SubjectName,
		Lesson: detail,
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := h.lessonTmpl.ExecuteTemplate(w, "layout", data); err != nil {
		h.logger.Error("failed to render lesson template", slog.String("error", err.Error()))
	}
}
