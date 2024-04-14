package banner

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/AnxVit/avito/internal/domain/models"
	"github.com/AnxVit/avito/internal/http-server/middleware/auth"
	"github.com/AnxVit/avito/internal/http-server/middleware/auth/access"
	resp "github.com/AnxVit/avito/internal/lib/api/response"
	"github.com/AnxVit/avito/internal/storage"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"github.com/go-playground/validator/v10"
)

type Repository interface {
	GetBanner(tag, feature, limit, offset string) ([]models.BannerDB, error)
	PostBanner(banner *models.BannerPost) (int64, error)
	PatchBanner(id string, banner *models.BannerPatch) error
	DeleteBanner(id string) error
}

func NewGet(bannerLog *slog.Logger, getter Repository) http.HandlerFunc {
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

func NewPost(bannerLog *slog.Logger, setter Repository) http.HandlerFunc {
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
			if _, ok := err.(*json.SyntaxError); ok {
				bannerLog.Info("NewPost", slog.String("failed to unmarshall", err.Error()))
				render.JSON(w, r, resp.Error("unsupported type of value"))
				return
			}
			bannerLog.Info("NewPost", slog.String("failed to unmarshall", err.Error()))
			render.JSON(w, r, resp.Error("invalid body"))
			return
		}

		validate := validator.New()

		err = validate.Struct(banner)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			if _, ok := err.(*validator.InvalidValidationError); ok {
				bannerLog.Info("NewPost", slog.String("failed to validate", err.Error()))
				render.JSON(w, r, resp.Error(err.Error()))
				return
			}
			bannerLog.Info("NewPost", slog.String("failed to validate", err.Error()))
			render.JSON(w, r, resp.Error("invalid body"))
			return
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

func NewPatch(bannerLog *slog.Logger, changer Repository) http.HandlerFunc {
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
		if _, err := strconv.Atoi(id); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			bannerLog.Info("not correct id")
			render.JSON(w, r, resp.Error("not correct id"))
			return
		}
		banner := models.BannerPatch{}

		err := json.NewDecoder(r.Body).Decode(&banner)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			if _, ok := err.(*json.SyntaxError); ok {
				bannerLog.Info("NewPatch", slog.String("failed to unmarshall", err.Error()))
				render.JSON(w, r, resp.Error("unsupported type of value"))
				return
			}
			bannerLog.Info("NewPatch", slog.String("failed to unmarshall", err.Error()))
			render.JSON(w, r, resp.Error("invalid body"))
			return
		}
		if banner.Tag.Defined && banner.Tag.Value != nil {
			for _, val := range *banner.Tag.Value {
				if val <= 0 {
					bannerLog.Info("unsupported value: tag")
					w.WriteHeader(http.StatusBadRequest)
					render.JSON(w, r, resp.Error("unsupported type of value"))
					return
				}
			}
		}
		if banner.Feature.Defined && banner.Feature.Value != nil {
			if *banner.Feature.Value <= 0 {
				bannerLog.Info("unsupported value: feature")
				w.WriteHeader(http.StatusBadRequest)
				render.JSON(w, r, resp.Error("unsuported type of value"))
				return
			}
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

func NewDelete(bannerLog *slog.Logger, deleter Repository) http.HandlerFunc {
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
		if _, err := strconv.Atoi(id); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			bannerLog.Info("not correct id")
			render.JSON(w, r, resp.Error("not correct id"))
			return
		}

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
