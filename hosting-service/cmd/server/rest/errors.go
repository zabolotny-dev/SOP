package rest

import (
	"hosting-kit/logger"
	"hosting-service/cmd/server/rest/gen"
	"net/http"

	"github.com/go-chi/render"
)

func makeResponseErrorHandler(log *logger.Logger) func(http.ResponseWriter, *http.Request, error) {
	return func(w http.ResponseWriter, r *http.Request, err error) {
		log.Error(r.Context(), "request failed",
			"error", err,
			"method", r.Method,
			"path", r.URL.Path,
		)

		render.Status(r, http.StatusInternalServerError)
		render.JSON(w, r, gen.StatusResponse{
			Message: "Internal server error",
		})
	}
}

func makeRequestErrorHandler(log *logger.Logger) func(http.ResponseWriter, *http.Request, error) {
	return func(w http.ResponseWriter, r *http.Request, err error) {
		log.Error(r.Context(), "request parsing failed", "error", err)
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, gen.StatusResponse{Message: "Bad Request"})
	}
}

func makeWrapperErrorHandler(log *logger.Logger) func(http.ResponseWriter, *http.Request, error) {
	return func(w http.ResponseWriter, r *http.Request, err error) {
		log.Error(r.Context(), "routing error", "error", err)
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, gen.StatusResponse{Message: err.Error()})
	}
}
