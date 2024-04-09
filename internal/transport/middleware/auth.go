package middleware

import "errors"

type (
	key string
)

var (
	ErrInvalidToken           = errors.New("invalid token")
	ErrFailedIsAdminCheck     = errors.New("failed to check if user is admin")
	ErrKey                key = "errorkey"
	UidKey                key = "uidkey"
	IsAdminKey            key = "isadminkey"
)
