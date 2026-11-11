package handler

import (
	"errors"
	"html/template"
	"log/slog"
	"net/http"
	"path/filepath"

	"edu-platform/internal/domain/user"
	authUC "edu-platform/internal/usecase/auth"
)

type AuthHandler struct {
	authUseCase     *authUC.UseCase
	loginTmpl       *template.Template
	registerTmpl    *template.Template
	adminInviteTmpl *template.Template
	logger          *slog.Logger
}

func NewAuthHandler(uc *authUC.UseCase, templatesDir string, logger *slog.Logger) (*AuthHandler, error) {
	loginTmpl, err := template.ParseFiles(
		filepath.Join(templatesDir, "layout.html"),
		filepath.Join(templatesDir, "login.html"),
	)
	if err != nil {
		return nil, err
	}

	registerTmpl, err := template.ParseFiles(
		filepath.Join(templatesDir, "layout.html"),
		filepath.Join(templatesDir, "register.html"),
	)
	if err != nil {
		return nil, err
	}

	adminInviteTmpl, err := template.ParseFiles(
		filepath.Join(templatesDir, "layout.html"),
		filepath.Join(templatesDir, "admin_invites.html"),
	)
	if err != nil {
		return nil, err
	}

	return &AuthHandler{
		authUseCase:     uc,
		loginTmpl:       loginTmpl,
		registerTmpl:    registerTmpl,
		adminInviteTmpl: adminInviteTmpl,
		logger:          logger,
	}, nil
}

type authPageData struct {
	Title           string
	Error           string
	Code            string
	IsAuthenticated bool
	IsAdmin         bool
}

type invitePageData struct {
	Title           string
	Invites         []authUC.InviteDTO
	IsAuthenticated bool
	IsAdmin         bool
	Error           string
}

func (h *AuthHandler) ServeLogin(w http.ResponseWriter, r *http.Request) {
	data := authPageData{Title: "EduPlatform | Вход в систему"}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	h.loginTmpl.ExecuteTemplate(w, "layout", data)
}

func (h *AuthHandler) HandleLogin(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Некорректная форма", http.StatusBadRequest)
		return
	}

	email := r.FormValue("email")
	password := r.FormValue("password")

	token, _, err := h.authUseCase.Login(r.Context(), email, password)
	if err != nil {
		h.logger.Warn("failed login attempt", slog.String("email", email))
		data := authPageData{Title: "EduPlatform | Вход в систему", Error: "Неверный логин или пароль"}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusUnauthorized)
		h.loginTmpl.ExecuteTemplate(w, "layout", data)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    token,
		Path:     "/",
		MaxAge:   7 * 24 * 3600,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   false,
	})

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (h *AuthHandler) ServeRegister(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	data := authPageData{
		Title: "EduPlatform | Регистрация по приглашению",
		Code:  code,
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	h.registerTmpl.ExecuteTemplate(w, "layout", data)
}

func (h *AuthHandler) HandleRegister(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Некорректная форма", http.StatusBadRequest)
		return
	}

	code := r.FormValue("code")
	email := r.FormValue("email")
	password := r.FormValue("password")
	confirmPassword := r.FormValue("confirm_password")

	token, _, err := h.authUseCase.RegisterWithInvite(r.Context(), code, email, password, confirmPassword)
	if err != nil {
		h.logger.Warn("registration rejected", slog.String("email", email), slog.String("reason", err.Error()))

		var errMsg string
		switch {
		case errors.Is(err, user.ErrInviteNotFound):
			errMsg = "Недействительный код приглашения"
		case errors.Is(err, user.ErrInviteUsed):
			errMsg = "Этот код приглашения уже был использован"
		case errors.Is(err, user.ErrInviteExpired):
			errMsg = "Срок действия кода приглашения истек"
		case errors.Is(err, user.ErrPasswordsDoNotMatch):
			errMsg = "Пароли не совпадают"
		case errors.Is(err, user.ErrPasswordTooShort):
			errMsg = "Пароль должен содержать от 8 до 72 символов"
		case errors.Is(err, user.ErrInvalidEmail):
			errMsg = "Некорректный адрес электронной почты"
		case errors.Is(err, user.ErrUserAlreadyExists):
			errMsg = "Пользователь с таким email уже существует"
		default:
			errMsg = "Ошибка регистрации: проверьте правильность введенных данных"
		}

		data := authPageData{
			Title: "EduPlatform | Регистрация",
			Error: errMsg,
			Code:  code,
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusBadRequest)
		h.registerTmpl.ExecuteTemplate(w, "layout", data)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    token,
		Path:     "/",
		MaxAge:   7 * 24 * 3600,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   false,
	})

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (h *AuthHandler) HandleLogout(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("session_token")
	if err == nil {
		_ = h.authUseCase.Logout(r.Context(), cookie.Value)
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// ServeAdminInvites показывает список инвайтов (доступ только админу)
func (h *AuthHandler) ServeAdminInvites(w http.ResponseWriter, r *http.Request) {
	invites, err := h.authUseCase.ListInvites(r.Context())
	if err != nil {
		h.logger.Error("failed to load invites", slog.String("error", err.Error()))
		http.Error(w, "Не удалось загрузить инвайты", http.StatusInternalServerError)
		return
	}

	data := invitePageData{
		Title:           "Управление приглашениями учеников",
		Invites:         invites,
		IsAuthenticated: true,
		IsAdmin:         true,
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := h.adminInviteTmpl.ExecuteTemplate(w, "layout", data); err != nil {
		h.logger.Error("failed to render admin invites template", slog.String("error", err.Error()))
	}
}

// HandleCreateInvite создает новый инвайт по кнопке из админки
func (h *AuthHandler) HandleCreateInvite(w http.ResponseWriter, r *http.Request) {
	studentNote := r.FormValue("student_note")
	if studentNote == "" {
		studentNote = "Новый ученик"
	}

	_, err := h.authUseCase.CreateInvite(r.Context(), studentNote)
	if err != nil {
		http.Error(w, "Не удалось создать приглашение", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/admin/invites", http.StatusSeeOther)
}
