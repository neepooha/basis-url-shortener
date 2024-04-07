package redirect

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

//go:generate go run github.com/vektra/mockery/v2@v2.28.2 --name=URLGetter
type URLGetter interface {
	GetURL(alias string) (string, error)
}

func New(log *slog.Logger, urlGetter URLGetter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.url.redirect.New"

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

		// get resURL by alias
		resURL, err := urlGetter.GetURL(alias)
		if err != nil {
			if errors.Is(err, storage.ErrURLNotFound) {
				log.Warn("wrong alias", slog.String("alias", alias))
				render.JSON(w, r, resp.Error("wrong alias"))
				return
			}
			log.Error("failed to get url", sl.Err(err))
			render.JSON(w, r, resp.Error("internal error"))
			return
		}
		log.Info("got url", slog.String("url", resURL))

		// redirect to resURL
		http.Redirect(w, r, resURL, http.StatusFound)
	}
}
