package main

import (
	"auth/internal/config"
	"auth/internal/domain/models"
	"auth/internal/http-server/handlers/url/deleteuser"
	"auth/internal/http-server/handlers/url/forgotpassword"
	"auth/internal/http-server/handlers/url/login"
	"auth/internal/http-server/handlers/url/register"
	"auth/internal/http-server/handlers/url/restorepassword"
	"auth/internal/http-server/handlers/url/restoreuser"
	"auth/internal/http-server/handlers/url/updateuser"
	"auth/internal/http-server/middleware/authorization"
	"auth/internal/http-server/middleware/logger"
	"auth/internal/lib/logger/sl"
	authService "auth/internal/services/auth"
	linksService "auth/internal/services/links"
	"auth/internal/storage/postgres"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"log/slog"
	"net/http"
	"os"
)

const (
	envLocal = "local"
	envProd  = "prod"
)

func main() {
	// Config init
	cfg := config.MustLoad()

	// Client init
	client := &http.Client{}

	// Logger init
	log := setupLogger(cfg.Env)
	log.With(slog.Any("config", cfg)).Info("start")

	// Storage init
	storage, err := postgres.New(cfg.Path)
	if err != nil {
		log.Error("failed to init storage", sl.Err(err))
		os.Exit(1)
	}

	// Services init
	auth := authService.New(storage, cfg.TokenTtl)
	links := linksService.New(storage)

	// Router init
	router := chi.NewRouter()

	// middlewares
	router.Use(middleware.RequestID)
	router.Use(logger.New(log))
	router.Use(middleware.Recoverer)
	router.Use(middleware.URLFormat)

	// Handlers
	registerHandler := register.New(log, auth)
	loginHandler := login.New(log, auth, cfg.TokenTtl, cfg.SecretKey)
	restorePasswordHandler := restorepassword.New(log, auth, storage)
	forgotPasswordHandler := forgotpassword.New(log, links, auth, client, cfg.LinkTtl, cfg.ApiKey, cfg.Name, cfg.Email)
	updateUserHandler := authorization.New(updateuser.New(log, auth), log, cfg.SecretKey, auth, []models.UserRole{models.JobSeeker, models.Admin})
	deleteUserHandler := authorization.New(deleteuser.New(log, auth), log, cfg.SecretKey, auth, []models.UserRole{models.JobSeeker, models.Admin})
	restoreUserHandler := authorization.New(restoreuser.New(log, auth), log, cfg.SecretKey, auth, []models.UserRole{models.JobSeeker, models.Admin})

	// TODO: по-хорошему надо сделать отдельный хэндлер регистрации для работодателя
	router.Post("/api/auth/register", registerHandler)
	router.Get("/api/auth/login", loginHandler)
	router.Put("/api/auth/restore-password", restorePasswordHandler)
	router.Post("/api/auth/forgot-password", forgotPasswordHandler)
	router.Post("/api/auth/refresh-tokens", restorepassword.New(log, auth, storage))
	router.Put("/api/auth/update-user", updateUserHandler)
	router.Put("/api/auth/delete-user", deleteUserHandler)
	router.Put("/api/auth/restore-user", restoreUserHandler)

	log.Info("starting server", slog.String("address", cfg.Address))

	// Server init
	srv := &http.Server{
		Addr:         cfg.Address,
		Handler:      router,
		ReadTimeout:  cfg.HttpServer.Timeout,
		WriteTimeout: cfg.HttpServer.Timeout,
		IdleTimeout:  cfg.HttpServer.IdleTimeout,
	}

	if err := srv.ListenAndServe(); err != nil {
		log.Error("failed to start server")
	}

	log.Error("server stopped")

	// TODO: сделать подтверждение телефона и почты
}

func setupLogger(env string) *slog.Logger {
	var log *slog.Logger

	switch env {
	case envLocal:
		log = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	case envProd:
		log = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	}

	return log
}
