// Package shortener предоставляет функции для генерации коротких идентификаторов URL.
package shortener

import (
	"crypto/rand"
	"math/big"
)

const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

// GenerateID генерирует случайный короткий идентификатор длиной 8 символов.
// Используется для создания коротких URL.
// Идентификатор состоит из букв латинского алфавита (в верхнем и нижнем регистре) и цифр.
// Возвращает строку с идентификатором и ошибку, если генерация случайного числа не удалась.
func GenerateID() (string, error) {
	n := 8
	id := make([]byte, n)

	for i := 0; i < n; i++ {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			return "", err
		}
		id[i] = charset[num.Int64()]
	}
	return string(id), nil
}
