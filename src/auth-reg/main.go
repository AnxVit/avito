package main

import (
	"Auth-Reg/internal/config"
	"Auth-Reg/internal/http-server/server"
	"Auth-Reg/internal/storage/cache"
	"Auth-Reg/internal/storage/postgres"
	"log/slog"
	"os"
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

	db, err := postgres.New(&cfg.DB)
	if err != nil {
		log.Error("failed to init db")
		os.Exit(1)
	}
	storage, err := cache.New(db)
	if err != nil {
		log.Error("failed to init cache")
		os.Exit(2)
	}

	srv := server.New(&cfg.Server, storage, log)
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
