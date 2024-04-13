package auth

import (
	"context"
	"log/slog"
	"net/http"
	"strings"
	"url_shortener/internal/lib/logger/sl"
	get "url_shortener/internal/transport/middleware/context"

	"github.com/golang-jwt/jwt"
)

func New(log *slog.Logger, appSecret string) func(next http.Handler) http.Handler {
	const op = "middleware.auth.New"
	log = log.With(slog.String("op", op))
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			tokenStr := extractBearerToken(r)
			log.Debug("got JWT token", slog.String("jwt-token", tokenStr))
			if tokenStr == "" {
				next.ServeHTTP(w, r)
				return
			}

			tokenParsed, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) { return []byte(appSecret), nil })
			if err != nil {
				log.Warn("failed to parse token", sl.Err(err))
				ctx := context.WithValue(r.Context(), get.ErrKey, get.ErrInvalidToken)
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}

			claims := tokenParsed.Claims.(jwt.MapClaims)
			log.Info("user authorized", slog.Any("claims", claims))
			ctx := context.WithValue(r.Context(), get.UidKey, uint64(claims["uid"].(float64)))

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func extractBearerToken(r *http.Request) string {
	authHeader := r.Header.Get("Authorization")
	splitToken := strings.Split(authHeader, "Bearer ")
	if len(splitToken) != 2 {
		return ""
	}
	return splitToken[1]
}
