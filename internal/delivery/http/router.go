package http

import (
	"net/http"

	"edu-platform/internal/delivery/http/handler"
)

func NewRouter(homeHandler *handler.HomeHandler, adminHandler *handler.AdminHandler, staticDir string) http.Handler {
	mux := http.NewServeMux()

	fileServer := http.FileServer(http.Dir(staticDir))
	mux.Handle("GET /static/", http.StripPrefix("/static/", fileServer))

	// Публичные страницы
	mux.Handle("GET /{$}", homeHandler)
	mux.HandleFunc("GET /subjects/{subject}", homeHandler.ServeSubject)
	mux.HandleFunc("GET /lessons/{slug}", homeHandler.ServeLesson)

	// Панель репетитора (Админка)
	mux.HandleFunc("GET /admin/lessons/new", adminHandler.ServeNewForm)
	mux.HandleFunc("POST /admin/lessons", adminHandler.HandleCreate)
	mux.HandleFunc("POST /admin/lessons/{slug}/delete", adminHandler.HandleDelete)
	return mux
}
