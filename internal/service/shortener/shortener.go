package shortener

import (
	"crypto/rand"
	"encoding/base64"
)

func GenerateID() (string, error) {
	n := 8
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}

	return base64.URLEncoding.EncodeToString(b)[:n], nil
}
