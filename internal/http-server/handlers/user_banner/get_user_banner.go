package userbanner

import (
	"Auth-Reg/internal/http-server/middleware/auth"
	"Auth-Reg/internal/http-server/middleware/auth/access"
	resp "Auth-Reg/internal/lib/api/response"
	"Auth-Reg/internal/storage"
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/go-chi/render"
)

type Response struct {
	resp.Response
}

type Banner interface {
	GetUserBanner(tag, feature int, use_last_version bool, admin bool) (map[string]interface{}, error)
}

func New(banner_log *slog.Logger, banner Banner) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		permission := r.Context().Value(auth.UserContextKey).(access.Access)

		if permission == access.NotAccess {
			banner_log.Info("unauthorized")
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		tag := r.URL.Query().Get("tag_id")
		feature := r.URL.Query().Get("feature_id")

		if tag == "" || feature == "" {
			banner_log.Info("required tag/feature")
			w.WriteHeader(http.StatusBadRequest)
			render.JSON(w, r, resp.Error("not set tag and/or feature"))
			return
		}

		tag_id, err := strconv.Atoi(tag)
		if err != nil {
			banner_log.Info("tag is not int")
			w.WriteHeader(http.StatusBadRequest)
			render.JSON(w, r, resp.Error("tag is not integer"))
			return
		}

		feature_id, err := strconv.Atoi(feature)
		if err != nil {
			banner_log.Info("feature is not int")
			w.WriteHeader(http.StatusBadRequest)
			render.JSON(w, r, resp.Error("feature is not integer"))
			return
		}

		var use_last_vers bool
		last := r.URL.Query().Get("use_last_revision")
		if last == "" {
			use_last_vers = false
		} else {
			use_last_vers, err = strconv.ParseBool(last)
			if err != nil {
				banner_log.Info("use_last_version is incorrect")
				w.WriteHeader(http.StatusBadRequest)
				render.JSON(w, r, resp.Error("use_last verison is incorrect"))
				return
			}
		}

		admin := false
		if permission == access.Admin {
			admin = true
		}

		banner, err := banner.GetUserBanner(tag_id, feature_id, use_last_vers, admin)
		if err != nil {
			if errors.Is(err, storage.ErrBannerNotFound) {
				banner_log.Info("banner not found")
				w.WriteHeader(http.StatusNotFound)
				return
			}
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
