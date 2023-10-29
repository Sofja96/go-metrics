package middleware

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"github.com/labstack/echo/v4"
	"io"
	"net/http"
)

type hashedWriter struct {
	w  http.ResponseWriter
	mw io.Writer
	b  *bytes.Buffer
	k  []byte
}

func newHashedWriter(w http.ResponseWriter, key []byte) *hashedWriter {
	b := new(bytes.Buffer)
	mw := io.MultiWriter(w, b)
	return &hashedWriter{w, mw, b, key}
}

func (h *hashedWriter) Header() http.Header {
	return h.w.Header()
}

func (h *hashedWriter) Write(data []byte) (int, error) {
	return h.mw.Write(data)
}

func (h *hashedWriter) WriteHeader(statusCode int) {
	h.w.WriteHeader(statusCode)
}

func (h *hashedWriter) hash() (string, error) {
	return ComputeHmac256(h.k, h.b.Bytes())
}

func ComputeHmac256(key []byte, data []byte) (string, error) {
	if len(key) == 0 {
		return hex.EncodeToString(data), nil
	}
	h := hmac.New(sha256.New, key)
	h.Write(data)
	return hex.EncodeToString(h.Sum(nil)), nil

}

func HashMacMiddleware(key string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) (err error) {
			//cfg := config.LoadConfig().HashKey
			//config.ParseFlags(cfg)
			//key := config.LoadConfig().HashKey
			//config.ParseFlags(key)
			if len(key) == 0 {
				return c.String(http.StatusInternalServerError, "hashes are empty")
			}

			req := c.Request()
			rw := c.Response().Writer
			header := req.Header
			clientHash := header.Get("HashSHA256")
			if len(clientHash) == 0 {
				return c.String(http.StatusBadRequest, "empty hash")
			}

			clientHashBytes, err := io.ReadAll(req.Body)
			if err != nil {
				return fmt.Errorf("error occured on reading from io.Reader: %w", err)
			}
			clientHash, err = ComputeHmac256([]byte(key), clientHashBytes)
			if err != nil {
				return fmt.Errorf("error occured on calculating hash: %w", err)
			}
			if clientHash != clientHash {
				return c.String(http.StatusBadRequest, "hashes are not equal")
				//return fmt.Errorf("hashes are not equal")
			}

			hashedWriter := newHashedWriter(rw, []byte(key))
			hash, err := hashedWriter.hash()
			if err != nil {
				return c.String(http.StatusInternalServerError, "hashes are not equal")
			}
			hashedWriter.Header().Set("HashSHA256", hash)

			if err = next(c); err != nil {
				c.Error(err)
			}

			return err
		}
	}

}
