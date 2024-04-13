package auth

import (
	"context"
	"net/http"

	"github.com/AnxVit/avito/internal/http-server/middleware/auth/access"
)

type tokenKey uint

const (
	UserContextKey tokenKey = 1
)

func MiddlewareAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("token")

		var acc access.Access
		if token == "" {
			acc = access.NotAccess
		} else {
			acc = access.GetAccess(token)
		}
		ctx := context.WithValue(r.Context(), UserContextKey, acc)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
