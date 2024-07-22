package hash

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
)

// ComputeHmac256 - вычисляет HMAC SHA-256 хэш для данных с использованием заданного ключа.
func ComputeHmac256(key []byte, data []byte) (string, error) {
	h := hmac.New(sha256.New, key)
	_, err := h.Write(data)
	if err != nil {
		return "", fmt.Errorf("error write data to hash writer: %w", err)
	}
	hashedData := h.Sum(nil)
	return hex.EncodeToString(hashedData), nil

}
