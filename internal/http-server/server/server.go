package server

import (
	"log/slog"
	"net/http"

	"github.com/AnxVit/avito/internal/config"
	"github.com/AnxVit/avito/internal/domain/models"
	"github.com/AnxVit/avito/internal/http-server/handlers/banner"
	userbanner "github.com/AnxVit/avito/internal/http-server/handlers/user_banner"
	"github.com/AnxVit/avito/internal/http-server/middleware/auth"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type Cache interface {
	GetUserBanner(tag, feature int, useLastReversion bool, admin bool) (map[string]interface{}, error)
}

type Repository interface {
	GetBanner(tag, feature, limit, offset string) ([]models.BannerDB, error)
	PostBanner(banner *models.BannerPost) (int64, error)
	PatchBanner(id string, banner *models.BannerPatch) error
	DeleteBanner(id string) error
}

type Server struct {
	server *http.Server
	Router *chi.Mux
}

func New(cfg *config.Server, repo Repository, localCache Cache, log *slog.Logger) *Server {
	router := chi.NewRouter()
	router.Use(middleware.RequestID)
	router.Use(middleware.RealIP)
	router.Use(middleware.Recoverer)
	router.Use(auth.MiddlewareAuth)

	router.Get("/user_banner", userbanner.New(log, localCache))

	router.Get("/banner", banner.NewGet(log, repo))
	router.Post("/banner", banner.NewPost(log, repo))

	router.Patch("/banner/{id}", banner.NewPatch(log, repo))
	router.Delete("/banner/{id}", banner.NewDelete(log, repo))

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
