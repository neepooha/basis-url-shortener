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

		alias := chi.URLParam(r, "alias")
		if alias == "" {
			log.Info("alias is empty")
			render.JSON(w, r, resp.Error("invalid request"))
			return
		}
		log.Info("alias was get from url", slog.String("alias", alias))

		err := urlDeleter.DeleteURL(alias)
		if err != nil {
			if errors.Is(err, storage.ErrAliasNotFound) {
				log.Info("url by alias was not found", slog.String("alias", alias))
				render.JSON(w, r, resp.Error("url by alias was not found"))
				return
			}
			log.Info("failed to delete url", sl.Err(err))
			render.JSON(w, r, resp.Error("internal error"))
			return
		}
		log.Info("delete alias", slog.String("alias", alias))

		render.JSON(w, r, resp.OK())
	}
}
