package middleware

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"github.com/labstack/echo/v4"
	"io"
	"log"
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

func (h *hashedWriter) hash() string {
	return ComputeHmac256(h.k, h.b.Bytes())
}

//func ComputeHmac256(key []byte, data []byte) (string, error) {
//	h := hmac.New(sha256.New, key)
//	h.Write(data)
//	hashedData := h.Sum(nil)
//	return hex.EncodeToString(hashedData), nil
//
//}

func ComputeHmac256(key []byte, data []byte) string {
	mac := hmac.New(sha256.New, key)
	mac.Write(data)
	hashedData := mac.Sum(nil)
	return hex.EncodeToString(hashedData)

}

func checkSign(key []byte, body []byte, hash string) error {
	clientHash := ComputeHmac256(key, body)
	if clientHash != hash {
		return fmt.Errorf("hashes are not equal")
	}
	return nil
}

//func checkSign(r *http.Request, bodyBytes []byte, key string) int {
//	if headerSign := r.Header.Get("HashSHA256"); headerSign != "" && key != "" {
//		sign := generateHMACSHA256(bodyBytes, key)
//		if sign != headerSign {
//			println("Sign hashes are not equal")
//			return http.StatusBadRequest
//		}
//	}
//	return http.StatusOK
//}

//func generateHMACSHA256(data []byte, key string) string {
//	h := hmac.New(sha256.New, []byte(key))
//	h.Write(data)
//	return fmt.Sprintf("%x", h.Sum(nil))
//}

//func checkHash(key []byte, reader io.Reader, hash string) error {
//	clientHashBytes, err := io.ReadAll(reader)
//	if err != nil {
//		return fmt.Errorf("error occured on reading from io.Reader: %w", err)
//	}
//	clientHash, err := ComputeHmac256(key, clientHashBytes)
//	if err != nil {
//		return fmt.Errorf("error occured on calculating hash: %w", err)
//	}
//	if clientHash != hash {
//		return fmt.Errorf("hashes are not equal")
//	}
//	return nil
//}

func HashMacMiddleware(key []byte) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) (err error) {
			clienthash := c.Request().Header.Get("Hashsha256")
			log.Println(clienthash, c.Request().Header)
			req := c.Request()
			//rw := c.Response().Writer
			reqBody := req.Body
			bodyBytes, err := io.ReadAll(reqBody)
			if err != nil {
				log.Println(bodyBytes, "body")
				return c.String(http.StatusInternalServerError, "hashes are empty")
			}

			err = checkSign(key, bodyBytes, clienthash)
			if err != nil {
				log.Println(req.Header)
				return c.String(http.StatusBadRequest, "hashes are not equal")
			}
			log.Println(req.Header)

			//if len(key) != 0 {
			//
			//}
			respSign := ComputeHmac256(key, bodyBytes)
			c.Response().Header().Set("HashSHA256", respSign)
			log.Println(respSign, "respsign")
			log.Println(c.Response().Header(), "header_response")

			//err = checkSign(key, bodyBytes, respSign)
			//if err != nil {
			//	log.Println(req.Header)
			//	return c.String(http.StatusBadRequest, "hashes are not equal")
			//}
			//log.Println(req.Header)

			//hashedWriter := newHashedWriter(rw, key)
			//hash := hashedWriter.hash()
			//log.Println(hash, "hash")
			//if err != nil {
			//	return c.String(http.StatusInternalServerError, "hashes are not equal")
			//}
			//hashedWriter.Header().Set("HashSHA256", hash)
			//log.Println(hashedWriter, "hashedWriter")
			//log.Println(hash)
			if err = next(c); err != nil {
				c.Error(err)
			}

			return err
		}
	}

}

func HashMiddleware(key []byte) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) (err error) {
			if _, ok := c.Request().Header["Hashsha256"]; ok {
				body, err := io.ReadAll(c.Request().Body)
				log.Println(c.Request().Header)
				//log.Println(body)
				if err != nil {
					return c.String(http.StatusInternalServerError, "hashes are not equal")

				}

				c.Request().Body = io.NopCloser(bytes.NewReader(body))

				err = checkSign(key, body, c.Request().Header.Get("Hashsha256"))
				log.Println(c.Request().Header)
				log.Println(key)
				//log.Println(body)
				//if err != nil {
				//	log.Println(c.Request().Header)
				//	log.Println(key)
				//	log.Println(body)
				//	return c.String(http.StatusBadRequest, "hashes are not equal")
				//}
			}
			log.Println(c.Request().Header)
			hashedWriter := newHashedWriter(c.Response(), key)

			hash := hashedWriter.hash()
			log.Println(hash, "hash")
			hashedWriter.Header().Set("HashSHA256", hash)
			log.Println(hashedWriter, "hashWriter")
			log.Println(c.Request().Header)
			if err = next(c); err != nil {
				c.Error(err)
			}

			return err
		}

	}

}
