package mid

import (
	"hosting-kit/otel"
	"log"
	"net/http"
	"time"
)

const limitMs = 20

func Performance(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		startTime := time.Now()

		next.ServeHTTP(w, r)

		duration := time.Since(startTime)

		if duration.Milliseconds() > limitMs {
			log.Printf("Slow request detected: [%s] %s %s %dms", otel.GetTraceID(r.Context()), r.Method, r.RequestURI, duration.Milliseconds())
		}
	})
}
