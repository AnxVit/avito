package main

import (
	"log/slog"
	"os"

	"github.com/AnxVit/avito/internal/config"
	"github.com/AnxVit/avito/internal/http-server/server"
	"github.com/AnxVit/avito/internal/storage/cache"
	"github.com/AnxVit/avito/internal/storage/postgres"
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

	repo, err := postgres.New(&cfg.DB)
	if err != nil {
		log.Error("failed to init db")
		os.Exit(1)
	}
	localcache, err := cache.New(repo)
	if err != nil {
		log.Error("failed to init cache")
		os.Exit(2)
	}

	srv := server.New(&cfg.Server, repo, localcache, log)
	if err := srv.Serve(); err != nil {
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
