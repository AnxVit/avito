package userbanner

import (
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/AnxVit/avito/internal/http-server/middleware/auth"
	"github.com/AnxVit/avito/internal/http-server/middleware/auth/access"
	resp "github.com/AnxVit/avito/internal/lib/api/response"
	"github.com/AnxVit/avito/internal/storage"

	"github.com/go-chi/render"
)

type Response struct {
	resp.Response
}

type Banner interface {
	GetUserBanner(tag, feature int, useLastVersion bool, admin bool) (map[string]interface{}, error)
}

func New(bannerLog *slog.Logger, banner Banner) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		permission := r.Context().Value(auth.UserContextKey).(access.Access) //nolint:forcetypeassert

		if permission == access.NotAccess {
			bannerLog.Info("unauthorized")
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		tag := r.URL.Query().Get("tag_id")
		feature := r.URL.Query().Get("feature_id")

		if tag == "" || feature == "" {
			bannerLog.Info("required tag/feature")
			w.WriteHeader(http.StatusBadRequest)
			render.JSON(w, r, resp.Error("not set tag and/or feature"))
			return
		}

		tagID, err := strconv.Atoi(tag)
		if err != nil {
			bannerLog.Info("tag is not int")
			w.WriteHeader(http.StatusBadRequest)
			render.JSON(w, r, resp.Error("tag is not integer"))
			return
		}

		featureID, err := strconv.Atoi(feature)
		if err != nil {
			bannerLog.Info("feature is not int")
			w.WriteHeader(http.StatusBadRequest)
			render.JSON(w, r, resp.Error("feature is not integer"))
			return
		}

		var lastVers bool
		last := r.URL.Query().Get("use_last_revision")
		if last == "" {
			lastVers = false
		} else {
			lastVers, err = strconv.ParseBool(last)
			if err != nil {
				bannerLog.Info("use_last_version is incorrect")
				w.WriteHeader(http.StatusBadRequest)
				render.JSON(w, r, resp.Error("use_last_version is incorrect"))
				return
			}
		}

		admin := false
		if permission == access.Admin {
			admin = true
		}

		banner, err := banner.GetUserBanner(tagID, featureID, lastVers, admin)
		if err != nil {
			if errors.Is(err, storage.ErrBannerNotFound) {
				bannerLog.Info("banner not found")
				w.WriteHeader(http.StatusNotFound)
				return
			}
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
