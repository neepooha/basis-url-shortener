package delete

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"strings"
	resp "github.com/neepooha/url_shortener/internal/lib/api/response"
	"github.com/neepooha/url_shortener/internal/lib/logger/sl"

	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/render"
	"google.golang.org/grpc/metadata"
)

type Request struct {
	Email string `json:"email" validate:"required"`
	AppID int    `json:"app_id" validate:"required"`
}

type Response struct {
	resp.Response
}

type PermissionDeleter interface {
	DelAdmin(ctx context.Context, email string, appid int) (bool, error)
}

func exractToken(header http.Header) (string, error) {
	if len(header) == 0 {
		return "", errors.New("no headers in request")
	}
	authHeaders, ok := header["Authorization"]
	if !ok {
		return "", errors.New("no Authorization in header")
	}
	if len(authHeaders) != 1 {
		return "", errors.New("more than 1 header in request")
	}
	auth := authHeaders[0]
	const prefix = "Bearer "
	if !strings.HasPrefix(auth, prefix) {
		return "", errors.New(`missing "Bearer " prefix in "Authorization" header`)
	}
	if auth[len(prefix):] == "" {
		return "", errors.New(`missing token in "Authorization" header`)
	}
	return auth, nil
}

func New(log *slog.Logger, permProvider PermissionDeleter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.url.delete.New"

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

		token, err := exractToken(r.Header)
		if err != nil {
			log.Error("failed get JWT token", sl.Err(err))
			render.JSON(w, r, resp.Error(err.Error()))
			return
		}
		ctx := metadata.NewOutgoingContext(r.Context(), metadata.Pairs("Authorization", token))

		_, err = permProvider.DelAdmin(ctx, req.Email, req.AppID)
		if err != nil {
			errExpect := "grpc.DelAdmin: rpc error: code = InvalidArgument desc = invalid credentials"
			if err.Error() == errExpect {
				log.Error("Invalid credential", sl.Err(err))
				render.JSON(w, r, resp.Error("Invalid credential"))
				return
			}
			log.Error("error to set admin", sl.Err(err))
			render.JSON(w, r, resp.Error(err.Error()))
			return
		}
		log.Info("admin delete")

		// response OK
		render.JSON(w, r, Response{Response: resp.OK()})
	}
}
