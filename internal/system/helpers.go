package system

import (
	"crypto/rand"
	"encoding/base64"
)

func GenerateRandomState(length int) (string, error) {
	randomBytes := make([]byte, length)
	_, err := rand.Read(randomBytes)
	if err != nil {
		return "", err
	}
	state := base64.URLEncoding.EncodeToString(randomBytes)
	return state, nil
}
