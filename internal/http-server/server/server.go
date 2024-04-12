package server

import (
	"Auth-Reg/internal/config"
	"Auth-Reg/internal/http-server/handlers/banner"
	userbanner "Auth-Reg/internal/http-server/handlers/user_banner"
	"Auth-Reg/internal/http-server/middleware/auth"
	"Auth-Reg/internal/storage/cache"
	"log/slog"
	"net/http"

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
