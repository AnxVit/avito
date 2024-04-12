package main

import (
	"Auth-Reg/internal/config"
	"Auth-Reg/internal/http-server/handlers/banner"
	userbanner "Auth-Reg/internal/http-server/handlers/user_banner"
	"Auth-Reg/internal/http-server/middleware/auth"
	"Auth-Reg/internal/storage/cache"
	"Auth-Reg/internal/storage/postgres"
	"log/slog"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

const (
	envLocal = "local"
	envDev   = "dev"
	EnvProd  = "prod"
)

func main() {
	cfg := config.MustLoad()

	log := setupLogger(cfg.Env)
	log.Info("starting auth-reg server", slog.String("env", cfg.Env))

	db, err := postgres.New(cfg.DB)
	if err != nil {
		log.Error("failed to init db")
		os.Exit(1)
	}
	storage, err := cache.New(db)
	if err != nil {
		log.Error("failed to init cache")
		os.Exit(2)
	}

	router := chi.NewRouter()
	router.Use(middleware.RequestID)
	router.Use(middleware.RealIP)
	router.Use(middleware.Recoverer)
	router.Use(auth.MiddlewareAuth)

	router.Get("/user_banner", userbanner.New(log, storage))

	router.Get("/banner", banner.NewGet(log, db))
	router.Post("/banner", banner.NewPost(log, db))

	router.Patch("/banner/{id}", banner.NewPatch(log, db))
	router.Delete("/banner/{id}", banner.NewDelete(log, db))

	srv := &http.Server{
		Addr:         cfg.Server.Host + ":" + cfg.Server.Port,
		Handler:      router,
		ReadTimeout:  cfg.Timeout,
		WriteTimeout: cfg.Timeout,
	}

	if err := srv.ListenAndServe(); err != nil {
		log.Error("failed to start server")
	}

	log.Error("server stoped")
}

func setupLogger(env string) *slog.Logger {
	var log *slog.Logger
	switch env {
	case envLocal:
		log = slog.New(
			slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
		)
	case envDev:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
		)
	case EnvProd:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}),
		)
	}
	return log
}
