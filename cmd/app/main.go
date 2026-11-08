package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	httpDelivery "edu-platform/internal/delivery/http"
	"edu-platform/internal/delivery/http/handler"
	"edu-platform/internal/infrastructure/persistence/memory"
	authUC "edu-platform/internal/usecase/auth"
	lessonUC "edu-platform/internal/usecase/lesson"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	// 1. Инициализация Infrastructure слоя (In-Memory хранилища)
	lessonRepo := memory.NewLessonRepository()
	userRepo := memory.NewUserRepository()
	sessionRepo := memory.NewSessionRepository()
	inviteRepo := memory.NewInviteRepository()

	// 2. Инициализация UseCase слоя (Бизнес-логика)
	lessonUseCase := lessonUC.NewUseCase(lessonRepo)
	authUseCase := authUC.NewUseCase(userRepo, sessionRepo, inviteRepo)

	// 3. Инициализация Delivery слоя (HTTP-хендлеры и шаблоны)
	templatesDir := "web/templates"
	staticDir := "web/static"

	homeHandler, err := handler.NewHomeHandler(lessonUseCase, templatesDir, logger)
	if err != nil {
		logger.Error("failed to initialize home handler", slog.String("error", err.Error()))
		os.Exit(1)
	}

	adminHandler, err := handler.NewAdminHandler(lessonUseCase, templatesDir, logger)
	if err != nil {
		logger.Error("failed to initialize admin handler", slog.String("error", err.Error()))
		os.Exit(1)
	}

	authHandler, err := handler.NewAuthHandler(authUseCase, templatesDir, logger)
	if err != nil {
		logger.Error("failed to initialize auth handler", slog.String("error", err.Error()))
		os.Exit(1)
	}

	// 4. Сборка маршрутизатора
	router := httpDelivery.NewRouter(homeHandler, adminHandler, authHandler, authUseCase, staticDir)

	// 5. Конфигурация HTTP-сервера с тайм-аутами
	srv := &http.Server{
		Addr:         ":8080",
		Handler:      router,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// 6. Запуск сервера в отдельной горутине
	go func() {
		logger.Info("starting educational platform server", slog.String("addr", srv.Addr))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("server failed to start", slog.String("error", err.Error()))
			os.Exit(1)
		}
	}()

	// 7. Ожидание сигналов прерывания для Graceful Shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("server forced to shutdown", slog.String("error", err.Error()))
		os.Exit(1)
	}

	logger.Info("server exited cleanly")
}
