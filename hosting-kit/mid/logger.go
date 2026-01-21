package mid

import (
	"hosting-kit/logger"
	"net/http"
	"time"

	"github.com/go-chi/chi/middleware"
)

func Logger(log *logger.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
			t1 := time.Now()

			next.ServeHTTP(ww, r)

			log.Info(r.Context(), "http request completed",
				"method", r.Method,
				"path", r.URL.Path,
				"status", ww.Status(),
				"duration", time.Since(t1).String(),
				"remote_ip", r.RemoteAddr,
			)
		})
	}
}
