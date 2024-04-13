package server

import (
	"log/slog"
	"net/http"

	"github.com/AnxVit/avito/internal/config"
	"github.com/AnxVit/avito/internal/http-server/handlers/banner"
	userbanner "github.com/AnxVit/avito/internal/http-server/handlers/user_banner"
	"github.com/AnxVit/avito/internal/http-server/middleware/auth"
	"github.com/AnxVit/avito/internal/storage/cache"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type Server struct {
	server *http.Server
	Router *chi.Mux
}

func New(cfg *config.Server, storage *cache.Cache, log *slog.Logger) *Server {
	router := chi.NewRouter()
	router.Use(middleware.RequestID)
	router.Use(middleware.RealIP)
	router.Use(middleware.Recoverer)
	router.Use(auth.MiddlewareAuth)

	router.Get("/user_banner", userbanner.New(log, storage))

	router.Get("/banner", banner.NewGet(log, storage.DB))
	router.Post("/banner", banner.NewPost(log, storage.DB))

	router.Patch("/banner/{id}", banner.NewPatch(log, storage.DB))
	router.Delete("/banner/{id}", banner.NewDelete(log, storage.DB))

	srv := &http.Server{
		Addr:         cfg.Host + ":" + cfg.Port,
		Handler:      router,
		ReadTimeout:  cfg.Timeout,
		WriteTimeout: cfg.Timeout,
	}
	return &Server{
		server: srv,
		Router: router,
	}
}

func (s *Server) Serve() error {
	return s.server.ListenAndServe()
}
