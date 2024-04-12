package auth

import (
	"Auth-Reg/internal/http-server/middleware/auth/access"
	"context"
	"net/http"
)

type tokenKey uint

const (
	UserContextKey tokenKey = 1
)

func MiddlewareAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("token")

		var ac access.Access
		if token == "" {
			ac = access.NotAccess
		}
		ac = access.GetAccess(token)
		ctx := context.WithValue(r.Context(), UserContextKey, ac)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
