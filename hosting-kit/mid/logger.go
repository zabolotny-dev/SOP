package mid

import (
	"hosting-kit/otel"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/middleware"
)

func Logger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

		t1 := time.Now()

		next.ServeHTTP(ww, r)

		traceID := otel.GetTraceID(r.Context())

		if traceID == "" || traceID == "00000000000000000000000000000000" {
			traceID = "no-trace"
		}

		log.Printf("[%s] %s %s %d %s", traceID, r.Method, r.URL.Path, ww.Status(), time.Since(t1))
	})
}
