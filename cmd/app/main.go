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

	delivery "edu-platform/internal/delivery/http"
	"edu-platform/internal/delivery/http/handler"
	"edu-platform/internal/domain/lesson"
	"edu-platform/internal/domain/user"
	pgRepo "edu-platform/internal/infrastructure/persistence/postgres"
	authUC "edu-platform/internal/usecase/auth"
	lessonUC "edu-platform/internal/usecase/lesson"
	"edu-platform/pkg/postgres"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	connString := os.Getenv("DATABASE_URL")
	if connString == "" {
		connString = "postgres://edu_user:edu_password_secure@127.0.0.1:5433/edu_platform?sslmode=disable"
	}

	logger.Info("connecting to postgresql pool...", "url", "localhost:5432/edu_platform")
	pool, err := postgres.NewPool(ctx, connString)
	if err != nil {
		logger.Error("failed to connect to postgresql", "error", err)
		os.Exit(1)
	}
	defer pool.Close()
	logger.Info("postgresql pool successfully initialized")

	userRepository := pgRepo.NewUserRepository(pool)
	sessionRepository := pgRepo.NewSessionRepository(pool)
	inviteRepository := pgRepo.NewInviteRepository(pool)
	lessonRepository := pgRepo.NewLessonRepository(pool)

	authUseCase := authUC.NewUseCase(userRepository, sessionRepository, inviteRepository)
	lessonUseCase := lessonUC.NewUseCase(lessonRepository)

	if err := seedInitialData(ctx, userRepository, lessonRepository, logger); err != nil {
		logger.Error("failed to seed initial data", "error", err)
		os.Exit(1)
	}

	templatesDir := "web/templates"
	staticDir := "web/static"

	homeHandler, err := handler.NewHomeHandler(lessonUseCase, templatesDir, logger)
	if err != nil {
		logger.Error("failed to initialize home handler", "error", err)
		os.Exit(1)
	}

	adminHandler, err := handler.NewAdminHandler(lessonUseCase, templatesDir, logger)
	if err != nil {
		logger.Error("failed to initialize admin handler", "error", err)
		os.Exit(1)
	}

	authHandler, err := handler.NewAuthHandler(authUseCase, templatesDir, logger)
	if err != nil {
		logger.Error("failed to initialize auth handler", "error", err)
		os.Exit(1)
	}

	router := delivery.NewRouter(homeHandler, adminHandler, authHandler, authUseCase, staticDir)

	srv := &http.Server{
		Addr:         ":8080",
		Handler:      router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	serverErrors := make(chan error, 1)
	go func() {
		logger.Info("server starting", "addr", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErrors <- err
		}
	}()

	select {
	case err := <-serverErrors:
		logger.Error("server error", "error", err)
		os.Exit(1)
	case <-ctx.Done():
		logger.Info("shutting down gracefully...")
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer shutdownCancel()

		if err := srv.Shutdown(shutdownCtx); err != nil {
			logger.Error("failed to shutdown gracefully", "error", err)
			_ = srv.Close()
		}
		logger.Info("server stopped successfully")
	}
}

func seedInitialData(
	ctx context.Context,
	userRepo user.Repository,
	lessonRepo lesson.Repository,
	logger *slog.Logger,
) error {
	adminEmail, _ := user.NewEmail("admin@edu.local")
	existingAdmin, _ := userRepo.FindByEmail(ctx, adminEmail)
	if existingAdmin == nil {
		logger.Info("seeding default admin account", "email", adminEmail.String())
		adminPass, err := user.NewPasswordHashFromPlain("admin123")
		if err != nil {
			return err
		}
		adminUser := user.NewUser("admin-root-id", adminEmail, adminPass, user.RoleAdmin)
		if err := userRepo.Save(ctx, adminUser); err != nil {
			return err
		}
	}

	published, err := lessonRepo.FindPublished(ctx)
	if err != nil {
		return err
	}

	if len(published) == 0 {
		logger.Info("seeding default lessons")

		l1, err := lesson.NewLesson(
			"1",
			"vostochnye-slavyane",
			"Восточные славяне в древности: расселение, быт и верования",
			"Разбираем карту расселения племенных союзов, хозяйственный уклад, подсечно-огневое земледелие и путь «из варяг в греки».",
			"Подробный конспект лекции по восточным славянам: родовая община, соседская община (вервь), пантеон богов.",
			lesson.SubjectHistory,
			lesson.TypeArticle,
			"",
			"",
			12,
		)
		if err == nil {
			_ = lessonRepo.Save(ctx, l1)
		}

		l2, err := lesson.NewLesson(
			"2",
			"reformy-izbrannoy-rady",
			"Реформы Избранной рады: централизация при Иване IV",
			"Видеоразбор Земского собора 1549 года, принятия Судебника 1550 года, отмены кормлений и создания стрелецкого войска.",
			"Анализ ключевых реформ 1550-х годов: Земский собор, отмена кормлений, стрелецкое войско.",
			lesson.SubjectHistory,
			lesson.TypeVideo,
			"https://www.youtube.com/watch?v=dQw4w9WgXcQ",
			"",
			25,
		)
		if err == nil {
			_ = lessonRepo.Save(ctx, l2)
		}

		l3, err := lesson.NewLesson(
			"3",
			"biosocialnaya-priroda-cheloveka",
			"Человек и общество: биосоциальная природа человека",
			"Человек и общество, экономика, политика, право и социальные отношения.",
			"Разбор потребностей, пирамида Маслоу, биологические и социальные факторы развития индивида.",
			lesson.SubjectSocial,
			lesson.TypeArticle,
			"",
			"",
			15,
		)
		if err == nil {
			_ = lessonRepo.Save(ctx, l3)
		}
	}

	return nil
}
