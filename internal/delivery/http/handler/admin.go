package handler

import (
	"errors"
	"fmt"
	"html/template"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"edu-platform/internal/domain/lesson"
	lessonUC "edu-platform/internal/usecase/lesson"
)

type AdminHandler struct {
	lessonUseCase *lessonUC.UseCase
	adminNewTmpl  *template.Template
	adminEditTmpl *template.Template
	logger        *slog.Logger
}

func NewAdminHandler(uc *lessonUC.UseCase, templatesDir string, logger *slog.Logger) (*AdminHandler, error) {
	newTmpl, err := template.ParseFiles(
		filepath.Join(templatesDir, "layout.html"),
		filepath.Join(templatesDir, "admin_new.html"),
	)
	if err != nil {
		return nil, err
	}

	editTmpl, err := template.ParseFiles(
		filepath.Join(templatesDir, "layout.html"),
		filepath.Join(templatesDir, "admin_edit.html"),
	)
	if err != nil {
		return nil, err
	}

	return &AdminHandler{
		lessonUseCase: uc,
		adminNewTmpl:  newTmpl,
		adminEditTmpl: editTmpl,
		logger:        logger,
	}, nil
}

type adminPageData struct {
	Title  string
	Error  string
	Lesson lessonUC.EditLessonDTO
}

func (h *AdminHandler) ServeNewForm(w http.ResponseWriter, r *http.Request) {
	data := adminPageData{
		Title: "Админка | Создать новый урок",
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	h.adminNewTmpl.ExecuteTemplate(w, "layout", data)
}

func (h *AdminHandler) HandleCreate(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(32 << 20); err != nil {
		http.Error(w, "Ошибка чтения формы или слишком большой файл", http.StatusBadRequest)
		return
	}

	var imageRelativePath string
	file, header, err := r.FormFile("lesson_image")
	if err == nil {
		defer file.Close()
		uploadDir := "./web/static/uploads"
		if err := os.MkdirAll(uploadDir, 0755); err == nil {
			ext := filepath.Ext(header.Filename)
			safeFilename := fmt.Sprintf("%d%s", time.Now().UnixNano(), ext)
			dstPath := filepath.Join(uploadDir, safeFilename)

			dst, err := os.Create(dstPath)
			if err == nil {
				defer dst.Close()
				if _, err := io.Copy(dst, file); err == nil {
					imageRelativePath = "/static/uploads/" + safeFilename
				}
			}
		}
	}

	readingTime, _ := strconv.Atoi(r.FormValue("reading_time"))
	if readingTime <= 0 {
		readingTime = 10
	}

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
		ImageURL:    imageRelativePath,
		ReadingTime: readingTime,
	}

	err = h.lessonUseCase.CreateLesson(r.Context(), input)
	if err != nil {
		h.logger.Error("failed to create lesson", slog.String("error", err.Error()))
		data := adminPageData{
			Title: "Админка | Создать новый урок",
			Error: "Ошибка при создании: " + err.Error(),
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		h.adminNewTmpl.ExecuteTemplate(w, "layout", data)
		return
	}

	http.Redirect(w, r, fmt.Sprintf("/subjects/%s", input.Subject), http.StatusSeeOther)
}

// ServeEditForm отдает форму редактирования с заполненными полями
func (h *AdminHandler) ServeEditForm(w http.ResponseWriter, r *http.Request) {
	slug := r.PathValue("slug")

	editDTO, err := h.lessonUseCase.GetLessonForEdit(r.Context(), slug)
	if err != nil {
		if errors.Is(err, lesson.ErrLessonNotFound) {
			http.NotFound(w, r)
			return
		}
		http.Error(w, "Внутренняя ошибка сервера", http.StatusInternalServerError)
		return
	}

	data := adminPageData{
		Title:  "Редактирование | " + editDTO.Title,
		Lesson: editDTO,
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	h.adminEditTmpl.ExecuteTemplate(w, "layout", data)
}

// HandleUpdate обрабатывает сохранение отредактированного урока
func (h *AdminHandler) HandleUpdate(w http.ResponseWriter, r *http.Request) {
	slug := r.PathValue("slug")

	if err := r.ParseMultipartForm(32 << 20); err != nil {
		http.Error(w, "Ошибка чтения формы", http.StatusBadRequest)
		return
	}

	var imageRelativePath string
	file, header, err := r.FormFile("lesson_image")
	if err == nil {
		defer file.Close()
		uploadDir := "./web/static/uploads"
		if err := os.MkdirAll(uploadDir, 0755); err == nil {
			ext := filepath.Ext(header.Filename)
			safeFilename := fmt.Sprintf("%d%s", time.Now().UnixNano(), ext)
			dstPath := filepath.Join(uploadDir, safeFilename)

			dst, err := os.Create(dstPath)
			if err == nil {
				defer dst.Close()
				if _, err := io.Copy(dst, file); err == nil {
					imageRelativePath = "/static/uploads/" + safeFilename
				}
			}
		}
	}

	readingTime, _ := strconv.Atoi(r.FormValue("reading_time"))
	if readingTime <= 0 {
		readingTime = 10
	}

	input := lessonUC.UpdateLessonInput{
		Slug:        slug,
		Title:       r.FormValue("title"),
		Summary:     r.FormValue("summary"),
		Body:        r.FormValue("body"),
		Subject:     r.FormValue("subject"),
		ContentType: r.FormValue("content_type"),
		VideoURL:    r.FormValue("video_url"),
		ImageURL:    imageRelativePath,
		ReadingTime: readingTime,
	}

	err = h.lessonUseCase.UpdateLesson(r.Context(), input)
	if err != nil {
		h.logger.Error("failed to update lesson", slog.String("slug", slug), slog.String("error", err.Error()))
		data := adminPageData{
			Title: "Редактирование | Ошибка",
			Error: "Ошибка при сохранении: " + err.Error(),
			Lesson: lessonUC.EditLessonDTO{
				Slug:        slug,
				Title:       input.Title,
				Summary:     input.Summary,
				Body:        input.Body,
				Subject:     input.Subject,
				ContentType: input.ContentType,
				VideoURL:    input.VideoURL,
				ReadingTime: input.ReadingTime,
			},
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		h.adminEditTmpl.ExecuteTemplate(w, "layout", data)
		return
	}

	http.Redirect(w, r, fmt.Sprintf("/lessons/%s", slug), http.StatusSeeOther)
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
