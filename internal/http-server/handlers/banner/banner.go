package banner

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/AnxVit/avito/internal/domain/models"
	"github.com/AnxVit/avito/internal/http-server/middleware/auth"
	"github.com/AnxVit/avito/internal/http-server/middleware/auth/access"
	resp "github.com/AnxVit/avito/internal/lib/api/response"
	"github.com/AnxVit/avito/internal/storage"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
)

type Worker interface {
	GetBanner(tag, feature, limit, offset string) ([]models.BannerDB, error)
	PostBanner(banner *models.BannerPost) (int64, error)
	PatchBanner(id string, banner *models.BannerPatch) error
	DeleteBanner(id string) error
}

func NewGet(bannerLog *slog.Logger, getter Worker) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		permission := r.Context().Value(auth.UserContextKey).(access.Access) //nolint:forcetypeassert
		if permission == access.User {
			bannerLog.Info("don't have permission")
			w.WriteHeader(http.StatusForbidden)
			return
		}
		if permission == access.NotAccess {
			bannerLog.Info("unauthorized")
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		tag := r.URL.Query().Get("tag_id")
		feature := r.URL.Query().Get("feature_id")
		limit := r.URL.Query().Get("limit")
		offset := r.URL.Query().Get("offset")

		banner, err := getter.GetBanner(tag, feature, limit, offset)
		if err != nil {
			if errors.Is(err, storage.ErrNotAccess) {
				bannerLog.Info("not access")
				w.WriteHeader(http.StatusForbidden)
				return
			}
			bannerLog.Error("falied to get banner", slog.Attr{
				Key:   "error",
				Value: slog.StringValue(err.Error()),
			})
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		render.JSON(w, r, banner)
	}
}

func NewPost(bannerLog *slog.Logger, setter Worker) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		permission := r.Context().Value(auth.UserContextKey).(access.Access) //nolint:forcetypeassert
		if permission == access.User {
			bannerLog.Info("don't have permission")
			w.WriteHeader(http.StatusForbidden)
			return
		}
		if permission == access.NotAccess {
			bannerLog.Info("unauthorized")
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		var banner models.BannerPost
		err := json.NewDecoder(r.Body).Decode(&banner)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			render.JSON(w, r, resp.Error(err.Error()))
		}
		id, err := setter.PostBanner(&banner)
		if err != nil {
			bannerLog.Error("failed to post banner", slog.Attr{
				Key:   "error",
				Value: slog.StringValue(err.Error()),
			})
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusCreated)
		render.JSON(w, r, resp.ID(id))
	}
}

func NewPatch(bannerLog *slog.Logger, changer Worker) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		permission := r.Context().Value(auth.UserContextKey).(access.Access) //nolint:forcetypeassert
		if permission == access.User {
			bannerLog.Info("don't have permission")
			w.WriteHeader(http.StatusForbidden)
			return
		}
		if permission == access.NotAccess {
			bannerLog.Info("unauthorized")
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		id := chi.URLParam(r, "id")
		banner := models.BannerPatch{}

		err := json.NewDecoder(r.Body).Decode(&banner)
		if err != nil {
			bannerLog.Info("bad request", err)
			w.WriteHeader(http.StatusBadRequest)
			render.JSON(w, r, resp.Error(err.Error()))
			return
		}
		err = changer.PatchBanner(id, &banner)
		if err != nil {
			if errors.Is(err, storage.ErrBannerNotFound) {
				bannerLog.Info("banner not found")
				w.WriteHeader(http.StatusNotFound)
				return
			}
			bannerLog.Error("falied to patch banner", slog.Attr{
				Key:   "error",
				Value: slog.StringValue(err.Error()),
			})
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		render.JSON(w, r, resp.OK())
	}
}

func NewDelete(bannerLog *slog.Logger, deleter Worker) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		permission := r.Context().Value(auth.UserContextKey).(access.Access) //nolint:forcetypeassert
		if permission == access.User {
			bannerLog.Info("don't have permission")
			w.WriteHeader(http.StatusForbidden)
			return
		}
		if permission == access.NotAccess {
			bannerLog.Info("unauthorized")
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		id := chi.URLParam(r, "id")

		err := deleter.DeleteBanner(id)
		if err != nil {
			if errors.Is(err, storage.ErrBannerNotFound) {
				bannerLog.Info("banner not found")
				w.WriteHeader(http.StatusNotFound)
				return
			}
			bannerLog.Error("falied to delete banner", slog.Attr{
				Key:   "error",
				Value: slog.StringValue(err.Error()),
			})
			w.WriteHeader(http.StatusInternalServerError)
			render.JSON(w, r, resp.Error(err.Error()))
			return
		}
		w.WriteHeader(http.StatusNoContent)
		render.JSON(w, r, resp.OK())
	}
}
