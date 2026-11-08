package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	httpDelivery "edu-platform/internal/delivery/http"
	"edu-platform/internal/delivery/http/handler"
	"edu-platform/internal/infrastructure/persistence/memory"
	lessonUC "edu-platform/internal/usecase/lesson"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))

	lessonRepo := memory.NewLessonRepository()

	lessonUseCase := lessonUC.NewUseCase(lessonRepo)

	homeHandler, err := handler.NewHomeHandler(lessonUseCase, "web/templates", logger)
	if err != nil {
		logger.Error("failed to initialize home handler", slog.String("error", err.Error()))
		os.Exit(1)
	}

	router := httpDelivery.NewRouter(homeHandler, "web/static")

	srv := &http.Server{
		Addr:         ":8080",
		Handler:      router,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		logger.Info("starting educational platform server", slog.String("addr", srv.Addr))
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("server error", slog.String("error", err.Error()))
			os.Exit(1)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	logger.Info("shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("forced shutdown", slog.String("error", err.Error()))
	}
}