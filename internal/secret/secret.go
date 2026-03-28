package secret

import (
	"crypto/rand"
	"fmt"
)

const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func Generate(length int) (string, error) {
	if length <= 0 {
		return "", fmt.Errorf("secret: length must be positive")
	}
	raw := make([]byte, length)
	if _, err := rand.Read(raw); err != nil {
		return "", err
	}
	n := len(charset)
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[int(raw[i])%n]
	}
	return string(b), nil
}
