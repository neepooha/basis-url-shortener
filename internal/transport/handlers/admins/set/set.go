package save

import (
	"context"
	"log/slog"
	"net/http"
	resp "url_shortener/internal/lib/api/response"
	"url_shortener/internal/lib/logger/sl"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
)

type Request struct {
	Email string `json:"email" validate:"required"`
}

type Response struct {
	resp.Response
}

type PermissionSetter interface {
	SetAdmin(ctx context.Context, email string) (bool, error)
}

func New(log *slog.Logger, permProvider PermissionSetter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.admins.set.New"

		// add to log op and reqID
		log := log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		// decode json request
		var req Request
		err := render.DecodeJSON(r.Body, &req)
		if err != nil {
			log.Error("failed to decode request body", sl.Err(err))
			render.JSON(w, r, resp.Error("failed to decode request"))
			return
		}
		log.Info("request body decoded", slog.Any("request", req))

		_, err = permProvider.SetAdmin(r.Context(), req.Email)
		if err != nil {
			errExpect := "grpc.SetAdmin: rpc error: code = InvalidArgument desc = invalid credentials"
			if err.Error() == errExpect {
				log.Error("Invalid credential", sl.Err(err))
				render.JSON(w, r, resp.Error("Invalid credential"))
				return
			}
			log.Error("error to set admin", sl.Err(err))
			render.JSON(w, r, resp.Error("error"))
			return
		}
		log.Info("user set to admin")

		// response OK
		render.JSON(w, r, Response{Response: resp.OK()})
	}
}
