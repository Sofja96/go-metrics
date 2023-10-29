package hash

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
)

func ComputeHmac256(key []byte, data []byte) (string, error) {
	if len(key) == 0 {
		return hex.EncodeToString(data), nil
	}
	h := hmac.New(sha256.New, key)
	h.Write(data)
	return hex.EncodeToString(h.Sum(nil)), nil

}
