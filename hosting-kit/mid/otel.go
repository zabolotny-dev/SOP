package mid

import (
	"hosting-kit/otel"
	"net/http"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel/trace"
)

func Otel(tracer trace.Tracer) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		injector := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := otel.InjectTracing(r.Context(), tracer)
			next.ServeHTTP(w, r.WithContext(ctx))
		})

		return otelhttp.NewHandler(injector, "request")
	}
}
