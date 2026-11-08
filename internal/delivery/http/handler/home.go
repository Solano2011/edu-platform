package handler

import (
	"errors"
	"html/template"
	"log/slog"
	"net/http"
	"path/filepath"

	"edu-platform/internal/delivery/http/middleware"
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

// Вспомогательная функция для получения статуса доступа
func getUserAuthStatus(r *http.Request) (isAuth bool, isAdmin bool) {
	u := middleware.UserFromContext(r.Context())
	if u != nil {
		return true, u.IsAdmin()
	}
	return false, false
}

type homePageData struct {
	Title           string
	IsAuthenticated bool
	IsAdmin         bool
}

type subjectPageData struct {
	Title           string
	SubjectName     string
	SubjectCode     string
	Lessons         []lessonUC.CardDTO
	IsAuthenticated bool
	IsAdmin         bool
}

type lessonPageData struct {
	Title           string
	Lesson          lessonUC.DetailDTO
	IsAuthenticated bool
	IsAdmin         bool
}

func (h *HomeHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	isAuth, isAdmin := getUserAuthStatus(r)
	data := homePageData{
		Title:           "EduPlatform | Обучающая платформа",
		IsAuthenticated: isAuth,
		IsAdmin:         isAdmin,
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	h.homeTmpl.ExecuteTemplate(w, "layout", data)
}

func (h *HomeHandler) ServeSubject(w http.ResponseWriter, r *http.Request) {
	subjParam := r.PathValue("subject")
	subjectName, lessons, err := h.lessonUseCase.GetLessonsBySubject(r.Context(), subjParam)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	isAuth, isAdmin := getUserAuthStatus(r)
	data := subjectPageData{
		Title:           subjectName + " | Конспекты и материалы",
		SubjectName:     subjectName,
		SubjectCode:     subjParam,
		Lessons:         lessons,
		IsAuthenticated: isAuth,
		IsAdmin:         isAdmin,
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	h.subjectTmpl.ExecuteTemplate(w, "layout", data)
}

func (h *HomeHandler) ServeLesson(w http.ResponseWriter, r *http.Request) {
	slug := r.PathValue("slug")
	l, err := h.lessonUseCase.GetLessonBySlug(r.Context(), slug)
	if err != nil {
		if errors.Is(err, lesson.ErrLessonNotFound) {
			http.NotFound(w, r)
			return
		}
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	isAuth, isAdmin := getUserAuthStatus(r)
	data := lessonPageData{
		Title:           l.Title + " | EduPlatform",
		Lesson:          l,
		IsAuthenticated: isAuth,
		IsAdmin:         isAdmin,
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	h.lessonTmpl.ExecuteTemplate(w, "layout", data)
}
