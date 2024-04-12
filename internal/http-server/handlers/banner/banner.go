package banner

import (
	"Auth-Reg/internal/domain/models"
	"Auth-Reg/internal/http-server/middleware/auth"
	"Auth-Reg/internal/http-server/middleware/auth/access"
	resp "Auth-Reg/internal/lib/api/response"
	"Auth-Reg/internal/storage"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
)

type BannerGetter interface {
	GetBanner(tag, feature, limit, offset string) ([]models.BannerDB, error)
}

type BannerSetter interface {
	PostBanner(banner *models.BannerPost) (int64, error)
}

type BannerChanger interface {
	PatchBanner(id string, banner *models.BannerPatch) error
}

type BannerDeleter interface {
	DeleteBanner(id string) error
}

func NewGet(banner_log *slog.Logger, getter BannerGetter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		permission := r.Context().Value(auth.UserContextKey).(access.Access)
		if permission == access.User {
			banner_log.Info("don't have permission")
			w.WriteHeader(http.StatusForbidden)
			return
		}
		if permission == access.NotAccess {
			banner_log.Info("unauthorized")
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
				banner_log.Info("not access")
				w.WriteHeader(http.StatusForbidden)
				return
			}
			banner_log.Error("falied to get banner", slog.Attr{
				Key:   "error",
				Value: slog.StringValue(err.Error()),
			})
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		render.JSON(w, r, banner)
	}
}

func NewPost(banner_log *slog.Logger, setter BannerSetter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		permission := r.Context().Value(auth.UserContextKey).(access.Access)
		if permission == access.User {
			banner_log.Info("don't have permission")
			w.WriteHeader(http.StatusForbidden)
			return
		}
		if permission == access.NotAccess {
			banner_log.Info("unauthorized")
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
			banner_log.Error("failed to post banner", slog.Attr{
				Key:   "error",
				Value: slog.StringValue(err.Error()),
			})
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusCreated)
		render.JSON(w, r, resp.Id(id))
	}
}

func NewPatch(banner_log *slog.Logger, changer BannerChanger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		permission := r.Context().Value(auth.UserContextKey).(access.Access)
		if permission == access.User {
			banner_log.Info("don't have permission")
			w.WriteHeader(http.StatusForbidden)
			return
		}
		if permission == access.NotAccess {
			banner_log.Info("unauthorized")
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		id := chi.URLParam(r, "id")
		banner := models.BannerPatch{}
		err := json.NewDecoder(r.Body).Decode(&banner)

		if err != nil {
			banner_log.Info("bad request", err)
			w.WriteHeader(http.StatusBadRequest)
			render.JSON(w, r, resp.Error(err.Error()))
			return
		}
		err = changer.PatchBanner(id, &banner)
		if err != nil {
			if errors.Is(err, storage.ErrBannerNotFound) {
				banner_log.Info("banner not found")
				w.WriteHeader(http.StatusNotFound)
				return
			}
			banner_log.Error("falied to patch banner", slog.Attr{
				Key:   "error",
				Value: slog.StringValue(err.Error()),
			})
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		render.JSON(w, r, resp.OK())
	}
}

func NewDelete(banner_log *slog.Logger, deleter BannerDeleter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		permission := r.Context().Value(auth.UserContextKey).(access.Access)
		if permission == access.User {
			banner_log.Info("don't have permission")
			w.WriteHeader(http.StatusForbidden)
			return
		}
		if permission == access.NotAccess {
			banner_log.Info("unauthorized")
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		id := chi.URLParam(r, "id")

		err := deleter.DeleteBanner(id)
		if err != nil {
			if errors.Is(err, storage.ErrBannerNotFound) {
				banner_log.Info("banner not found")
				w.WriteHeader(http.StatusNotFound)
				return
			}
			banner_log.Error("falied to delete banner", slog.Attr{
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
