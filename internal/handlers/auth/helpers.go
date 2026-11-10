package auth

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
)

const refreshTokenSize = 32

func generateRefreshToken() (string, error) {
	tokenBytes := make([]byte, refreshTokenSize)

	_, err := rand.Read(tokenBytes)

	if err != nil {
		return "", fmt.Errorf(
			"auth: generate refresh token: %w",
			err,
		)
	}

	token := base64.RawURLEncoding.EncodeToString(
		tokenBytes,
	)

	return token, nil
}
