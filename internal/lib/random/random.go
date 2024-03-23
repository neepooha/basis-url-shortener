package random

import (
	"crypto/rand"
	"encoding/base64"
)

func NewRandomString(size int) string {
	b := GenerateRandomBytes(size)
	return base64.URLEncoding.EncodeToString(b)[:size]
}

func GenerateRandomBytes(n int) []byte {
	b := make([]byte, n)
	rand.Read(b)
	return b
}
