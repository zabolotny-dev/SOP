package mid

import (
	"hosting-kit/auth"
	"net/http"
)

func Authenticate(authClient auth.Client) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			cookie := r.Header.Get("Cookie")

			claims, err := authClient.Authenticate(r.Context(), cookie)
			if err != nil {
				http.Error(w, auth.ErrUnauthorized.Error(), http.StatusUnauthorized)
				return
			}

			ctx := auth.SetClaims(r.Context(), claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func AuthenticateOptional(authClient auth.Client) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cookie := r.Header.Get("Cookie")
			claims, err := authClient.Authenticate(r.Context(), cookie)

			ctx := r.Context()
			if err == nil {
				ctx = auth.SetClaims(r.Context(), claims)
			}

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
