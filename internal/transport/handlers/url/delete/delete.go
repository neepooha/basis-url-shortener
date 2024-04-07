package delete

import (
	"errors"
	"log/slog"
	"net/http"
	resp "url_shortener/internal/lib/api/response"
	"url_shortener/internal/lib/logger/sl"
	"url_shortener/internal/storage"

	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
)

//go:generate go run github.com/vektra/mockery/v2@v2.28.2 --name=URLDeleter
type URLDeleter interface {
	DeleteURL(alias string) error
}

func New(log *slog.Logger, urlDeleter URLDeleter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.url.delete.New"

		// add to log op and reqID
		log := log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		// get alias from url
		alias := chi.URLParam(r, "alias")
		if alias == "" {
			log.Warn("alias is empty")
			render.JSON(w, r, resp.Error("invalid request"))
			return
		}
		log.Info("alias was get from url", slog.String("alias", alias))

		// delete URL by alias
		err := urlDeleter.DeleteURL(alias)
		if err != nil {
			if errors.Is(err, storage.ErrAliasNotFound) {
				log.Warn("url by alias was not found", slog.String("alias", alias))
				render.JSON(w, r, resp.Error("url by alias was not found"))
				return
			}
			log.Error("failed to delete url", sl.Err(err))
			render.JSON(w, r, resp.Error("internal error"))
			return
		}
		log.Info("delete alias", slog.String("alias", alias))

		// respone OK
		render.JSON(w, r, resp.OK())
	}
}
