package isadmin

import (
	"context"
	"github.com/neepooha/url_shortener/internal/lib/logger/sl"
	get "github.com/neepooha/url_shortener/internal/transport/middleware/context"
	"log/slog"
	"net/http"
)

type PermissionProvider interface {
	IsAdmin(ctx context.Context, userID uint64, appID int) (bool, error)
}

func New(log *slog.Logger, permProvider PermissionProvider) func(next http.Handler) http.Handler {
	const op = "middleware.IsAdmin.New"
	log = log.With(slog.String("op", op))
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			uid, ok := get.UIDFromContext(r.Context())
			if !ok {
				log.Error("failed to get UID user")
				ctx := context.WithValue(r.Context(), get.ErrKey, get.ErrFailedIsAdminCheck)
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}

			appID, ok := get.APPIDFromContext(r.Context())
			if !ok {
				log.Error("failed to get APPID")
				ctx := context.WithValue(r.Context(), get.ErrKey, get.ErrFailedIsAdminCheck)
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}

			isAdmin, err := permProvider.IsAdmin(r.Context(), uid, appID)
			if err != nil {
				log.Error("failed to check if user is admin", sl.Err(err))
				ctx := context.WithValue(r.Context(), get.ErrKey, get.ErrFailedIsAdminCheck)
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}
			ctx := context.WithValue(r.Context(), get.IsAdminKey, isAdmin)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
