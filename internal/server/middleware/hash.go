package middleware

import (
	"bytes"
	"fmt"
	"io"
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/Sofja96/go-metrics.git/internal/utils"
)

type hashedWriter struct {
	http.ResponseWriter
	mw io.Writer
	b  *bytes.Buffer
	k  []byte
}

func newHashedWriter(w http.ResponseWriter, key []byte) *hashedWriter {
	b := new(bytes.Buffer)
	mw := io.MultiWriter(w, b)
	return &hashedWriter{w, mw, b, key}
}

func (h *hashedWriter) Write(data []byte) (int, error) {
	return h.mw.Write(data)
}

func (h *hashedWriter) hash() string {
	return utils.ComputeHmac256(h.k, h.b.Bytes())
}

func checkSign(key []byte, body []byte, hash string) error {
	clientHash := utils.ComputeHmac256(key, body)
	if clientHash != hash {
		return fmt.Errorf("hashes are not equal")
	}
	return nil
}

// HashMacMiddleware - метод цифровой подписи передаваемых данных.
func HashMacMiddleware(key []byte) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) (err error) {
			if _, ok := c.Request().Header["Hashsha256"]; ok {
				bodyBytes, err := io.ReadAll(c.Request().Body)
				if err != nil {
					return c.String(http.StatusInternalServerError, "hashes are empty")
				}

				c.Request().Body = io.NopCloser(bytes.NewReader(bodyBytes))

				err = checkSign(key, bodyBytes, c.Request().Header.Get("Hashsha256"))
				if err != nil {
					return c.String(http.StatusBadRequest, "hashes are not equal")
				}

			}

			hashedWriter := newHashedWriter(c.Response(), key)

			hash := hashedWriter.hash()
			hashedWriter.Header().Set("HashSHA256", hash)
			if err = next(c); err != nil {
				c.Error(err)
			}
			return err
		}
	}

}
