package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
)

func NewOpaqueToken() (string, []byte, error) {
	randomBytes := make([]byte, 32)

	if _, err := rand.Read(randomBytes); err != nil {
		return "", nil, err
	}

	rawToken := base64.RawURLEncoding.EncodeToString(randomBytes)
	tokenHash := HashToken(rawToken)

	return rawToken, tokenHash, nil
}

func HashToken(rawToken string) []byte {
	hash := sha256.Sum256([]byte(rawToken))
	return hash[:]
}
