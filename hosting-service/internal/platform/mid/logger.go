package mid

import (
	"context"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/google/uuid"
)

const correlationIDHeader = "X-Request-ID"

type correlationIDKey struct{}

func Logger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		correlationId := r.Header.Get(correlationIDHeader)
		if correlationId == "" {
			correlationId = uuid.New().String()
		}
		w.Header().Set(correlationIDHeader, correlationId)

		ctx := context.WithValue(r.Context(), correlationIDKey{}, correlationId)
		ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
		startTime := time.Now()

		if strings.HasPrefix(r.RequestURI, "/api") {
			log.Printf("Request started: [%s] %s %s", correlationId, r.Method, r.RequestURI)
		}

		next.ServeHTTP(ww, r.WithContext(ctx))

		duration := time.Since(startTime)

		if strings.HasPrefix(r.RequestURI, "/api") {
			log.Printf("Request finished: [%s] %s %s with status %d in %dms",
				correlationId,
				r.Method,
				r.RequestURI,
				ww.Status(),
				duration.Milliseconds(),
			)
		}
	})
}

func GetCorrelationID(ctx context.Context) string {
	if id, ok := ctx.Value(correlationIDKey{}).(string); ok {
		return id
	}
	return ""
}
