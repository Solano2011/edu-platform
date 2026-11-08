package http

import (
	"net/http"

	"edu-platform/internal/delivery/http/handler"
)

func NewRouter(homeHandler *handler.HomeHandler, staticDir string) http.Handler {
	mux := http.NewServeMux()

	fileServer := http.FileServer(http.Dir(staticDir))
	mux.Handle("GET /static/", http.StripPrefix("/static/", fileServer))

	mux.Handle("GET /{$}", homeHandler)
	mux.HandleFunc("GET /subjects/{subject}", homeHandler.ServeSubject)
	mux.HandleFunc("GET /lessons/{slug}", homeHandler.ServeLesson)

	return mux
}
