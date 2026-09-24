package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
)

func NewSessionToken() (string, []byte, error) {
	randomBytes := make([]byte, 32)

	if _, err := rand.Read(randomBytes); err != nil {
		return "", nil, err
	}

	token := base64.RawURLEncoding.EncodeToString(randomBytes)

	return token, HashSessionToken(token), nil
}

func HashSessionToken(token string) []byte {
	tokenHash := sha256.Sum256([]byte(token))
	return tokenHash[:]
}
