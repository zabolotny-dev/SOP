package mid

import (
	"hosting-kit/logger"
	"net/http"
	"time"
)

const limitMs = 20

func Performance(log *logger.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			startTime := time.Now()

			next.ServeHTTP(w, r)

			duration := time.Since(startTime)

			if duration.Milliseconds() > limitMs {
				log.Warn(r.Context(), "slow request detected",
					"method", r.Method,
					"uri", r.RequestURI,
					"duration_ms", duration.Milliseconds(),
					"limit_ms", limitMs,
				)
			}
		})
	}
}
