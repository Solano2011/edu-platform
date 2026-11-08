package handler

import (
	"fmt"
	"html/template"
	"log/slog"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"

	lessonUC "edu-platform/internal/usecase/lesson"
)

type AdminHandler struct {
	lessonUseCase *lessonUC.UseCase
	adminTmpl     *template.Template
	logger        *slog.Logger
}

func NewAdminHandler(uc *lessonUC.UseCase, templatesDir string, logger *slog.Logger) (*AdminHandler, error) {
	tmpl, err := template.ParseFiles(
		filepath.Join(templatesDir, "layout.html"),
		filepath.Join(templatesDir, "admin_new.html"),
	)
	if err != nil {
		return nil, err
	}

	return &AdminHandler{
		lessonUseCase: uc,
		adminTmpl:     tmpl,
		logger:        logger,
	}, nil
}

type adminPageData struct {
	Title string
	Error string
}

// ServeNewForm отдает страницу с формой создания урока
func (h *AdminHandler) ServeNewForm(w http.ResponseWriter, r *http.Request) {
	data := adminPageData{
		Title: "Админка | Создать новый урок",
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	h.adminTmpl.ExecuteTemplate(w, "layout", data)
}

// HandleCreate обрабатывает POST-запрос отправки формы
func (h *AdminHandler) HandleCreate(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Ошибка чтения формы", http.StatusBadRequest)
		return
	}

	readingTime, _ := strconv.Atoi(r.FormValue("reading_time"))
	if readingTime <= 0 {
		readingTime = 10
	}

	// Автоматически генерируем slug из заголовка, если он пустой (простая транслитерация/замена пробелов)
	slug := r.FormValue("slug")
	if slug == "" {
		slug = strings.ToLower(strings.ReplaceAll(r.FormValue("title"), " ", "-"))
	}

	input := lessonUC.CreateLessonInput{
		Title:       r.FormValue("title"),
		Slug:        slug,
		Summary:     r.FormValue("summary"),
		Body:        r.FormValue("body"),
		Subject:     r.FormValue("subject"),
		ContentType: r.FormValue("content_type"),
		VideoURL:    r.FormValue("video_url"),
		ReadingTime: readingTime,
	}

	err := h.lessonUseCase.CreateLesson(r.Context(), input)
	if err != nil {
		h.logger.Error("failed to create lesson", slog.String("error", err.Error()))
		data := adminPageData{
			Title: "Админка | Создать новый урок",
			Error: "Ошибка при создании: " + err.Error(),
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		h.adminTmpl.ExecuteTemplate(w, "layout", data)
		return
	}

	// После успешного создания редиректим на главную или в раздел предмета
	http.Redirect(w, r, fmt.Sprintf("/subjects/%s", input.Subject), http.StatusSeeOther)
}

func (h *AdminHandler) HandleDelete(w http.ResponseWriter, r *http.Request) {
	slug := r.PathValue("slug")

	err := h.lessonUseCase.DeleteLesson(r.Context(), slug)
	if err != nil {
		h.logger.Error("failed to delete lesson", slog.String("slug", slug), slog.String("error", err.Error()))
		http.Error(w, "Не удалось удалить урок", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}
