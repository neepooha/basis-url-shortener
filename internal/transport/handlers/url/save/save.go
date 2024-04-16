package save

import (
	"context"
	"errors"
	resp "github.com/neepooha/url_shortener/internal/lib/api/response"
	"github.com/neepooha/url_shortener/internal/lib/logger/sl"
	"github.com/neepooha/url_shortener/internal/lib/random"
	"github.com/neepooha/url_shortener/internal/storage"
	get "github.com/neepooha/url_shortener/internal/transport/middleware/context"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"github.com/go-playground/validator/v10"
)

type Request struct {
	URL   string `json:"url" validate:"required,url"`
	Alias string `json:"alias"`
}

type Response struct {
	resp.Response
	Alias string `json:"alias,omitempty"`
}

//go:generate go run github.com/vektra/mockery/v2@v2.42.2 --name=URLSaver
type URLSaver interface {
	SaveURL(ctx context.Context, urlToSave string, alias string) error
}

const aliasLength = 6

func New(log *slog.Logger, urlSaver URLSaver) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.url.save.New"

		// add to log op and reqID
		log := log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		if _, ok := get.UIDFromContext(r.Context()); !ok {
			if err, ok := get.ErrorFromContext(r.Context()); ok {
				log.Error("failed to get UID", sl.Err(err))
				render.JSON(w, r, resp.Error("Internal Error"))
				return
			}
			log.Info("user without logging")
			render.JSON(w, r, resp.Error("you are not logged into your account"))
			return
		}

		// decode json request
		var req Request
		err := render.DecodeJSON(r.Body, &req)
		if err != nil {
			log.Error("failed to decode request body", sl.Err(err))
			render.JSON(w, r, resp.Error("failed to decode request"))
			return
		}
		log.Info("request body decoded", slog.Any("request", req))

		// validate url
		if err := validator.New().Struct(req); err != nil {
			validateErr := err.(validator.ValidationErrors)
			log.Error("invalid request", sl.Err(err))
			render.JSON(w, r, resp.ValidationError(validateErr))
			return
		}

		// get alias from request or random
		alias := req.Alias
		if alias == "" {
			// TODO if new alias = old alias in db
			alias = random.NewRandomString(aliasLength)
		}

		// save url in DB
		err = urlSaver.SaveURL(r.Context(), req.URL, alias)
		if err != nil {
			if errors.Is(err, storage.ErrURLExists) {
				log.Warn("url already exists", slog.String("url", req.URL))
				render.JSON(w, r, resp.Error("url already exists"))
				return
			}
			log.Error("failed to add url", sl.Err(err))
			render.JSON(w, r, resp.Error("url already exists"))
			return
		}
		log.Info("url added")

		// response OK
		responseOK(w, r, alias)
	}
}

func responseOK(w http.ResponseWriter, r *http.Request, alias string) {
	render.JSON(w, r, Response{
		Response: resp.OK(),
		Alias:    alias,
	})
}
