package http

import (
	"net/http"

	"edu-platform/internal/delivery/http/handler"
	"edu-platform/internal/delivery/http/middleware"
	authUC "edu-platform/internal/usecase/auth"
)

func NewRouter(
	homeHandler *handler.HomeHandler,
	adminHandler *handler.AdminHandler,
	authHandler *handler.AuthHandler,
	authUseCase *authUC.UseCase,
	staticDir string,
) http.Handler {
	mux := http.NewServeMux()

	fileServer := http.FileServer(http.Dir(staticDir))
	mux.Handle("GET /static/", http.StripPrefix("/static/", fileServer))

	// Лимитер попыток: 5 запросов в минуту для защиты от перебора
	authLimiter := middleware.NewIPRateLimiter(5.0/60.0, 5.0)

	// Аутентификация и регистрация по инвайтам
	mux.HandleFunc("GET /login", authHandler.ServeLogin)
	mux.HandleFunc("POST /login", authLimiter.Limit(authHandler.HandleLogin))
	mux.HandleFunc("GET /register", authHandler.ServeRegister)
	mux.HandleFunc("POST /register", authLimiter.Limit(authHandler.HandleRegister))
	mux.HandleFunc("POST /logout", authHandler.HandleLogout)

	// Публичные страницы
	mux.Handle("GET /{$}", homeHandler)
	mux.HandleFunc("GET /subjects/{subject}", homeHandler.ServeSubject)
	mux.HandleFunc("GET /lessons/{slug}", homeHandler.ServeLesson)

	// Защищенное управление уроками (только для роли admin)
	mux.HandleFunc("GET /admin/lessons/new", middleware.RequireAdmin(adminHandler.ServeNewForm))
	mux.HandleFunc("POST /admin/lessons", middleware.RequireAdmin(adminHandler.HandleCreate))
	mux.HandleFunc("GET /admin/lessons/{slug}/edit", middleware.RequireAdmin(adminHandler.ServeEditForm))
	mux.HandleFunc("POST /admin/lessons/{slug}/edit", middleware.RequireAdmin(adminHandler.HandleUpdate))
	mux.HandleFunc("POST /admin/lessons/{slug}/delete", middleware.RequireAdmin(adminHandler.HandleDelete))

	// Защищенное управление инвайтами учеников (только для роли admin)
	mux.HandleFunc("GET /admin/invites", middleware.RequireAdmin(authHandler.ServeAdminInvites))
	mux.HandleFunc("POST /admin/invites", middleware.RequireAdmin(authHandler.HandleCreateInvite))

	return middleware.AuthMiddleware(authUseCase)(mux)
}
