package auth

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"url_shortener/internal/lib/logger/sl"
	authKeys "url_shortener/internal/transport/middleware"

	"github.com/golang-jwt/jwt"
)

type PermissionProvider interface {
	IsAdmin(ctx context.Context, userID uint64) (bool, error)
}

func New(log *slog.Logger, appSecret string, permProvider PermissionProvider) func(next http.Handler) http.Handler {
	const op = "middleware.auth.New"
	log = log.With(slog.String("op", op))
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			tokenStr := extractBearerToken(r)
			if tokenStr == "" {
				next.ServeHTTP(w, r)
				return
			}

			tokenParsed, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) { return []byte(appSecret), nil })
			if err != nil {
				log.Warn("failed to parse token", sl.Err(err))
				ctx := context.WithValue(r.Context(), authKeys.ErrKey, authKeys.ErrInvalidToken)
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}

			claims := tokenParsed.Claims.(jwt.MapClaims)
			log.Info("user authorized", slog.Any("claims", claims))
			isAdmin, err := permProvider.IsAdmin(r.Context(), uint64(claims["uid"].(float64)))
			if err != nil {
				log.Error("failed to check if user is admin", sl.Err(err))
				ctx := context.WithValue(r.Context(), authKeys.ErrKey, authKeys.ErrFailedIsAdminCheck)
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}

			ctx := context.WithValue(r.Context(), authKeys.UidKey, uint64(claims["uid"].(float64)))
			ctx = context.WithValue(ctx, authKeys.IsAdminKey, isAdmin)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func extractBearerToken(r *http.Request) string {
	authHeader := r.Header.Get("Authorization")
	fmt.Println(authHeader)
	splitToken := strings.Split(authHeader, "Bearer ")
	if len(splitToken) != 2 {
		return ""
	}
	fmt.Println(splitToken[1])
	return splitToken[1]
}

func UIDFromContext(ctx context.Context) (uint64, bool) {
	uid, ok := ctx.Value(authKeys.UidKey).(uint64)
	return uid, ok
}

func IsAdminFromContext(ctx context.Context) (bool, bool) {
	uid, ok := ctx.Value(authKeys.IsAdminKey).(bool)
	return uid, ok
}

func ErrorFromContext(ctx context.Context) (error, bool) {
	err, ok := ctx.Value(authKeys.ErrKey).(error)
	return err, ok
}
