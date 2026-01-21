package mid

import (
	"hosting-kit/auth"
	"net/http"
)

func RequireAdmin() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims, err := auth.GetClaims(r.Context())
			if err != nil {
				http.Error(w, auth.ErrUnauthorized.Error(), http.StatusUnauthorized)
				return
			}

			if !claims.IsAdmin {
				http.Error(w, auth.ErrForbidden.Error(), http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
