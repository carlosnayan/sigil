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
	n := len(charset)
	threshold := 256 - (256 % n)
	b := make([]byte, length)
	for i := range b {
		for {
			var x [1]byte
			if _, err := rand.Read(x[:]); err != nil {
				return "", err
			}
			v := int(x[0])
			if v < threshold {
				b[i] = charset[v%n]
				break
			}
		}
	}
	return string(b), nil
}
