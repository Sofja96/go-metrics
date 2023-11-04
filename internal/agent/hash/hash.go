package hash

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
)

func ComputeHmac256(key []byte, data []byte) string {
	h := hmac.New(sha256.New, key)
	h.Write(data)
	hashedData := h.Sum(nil)
	return hex.EncodeToString(hashedData)

}
