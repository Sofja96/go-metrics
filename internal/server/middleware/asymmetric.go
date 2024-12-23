package middleware

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/labstack/echo/v4"
)

// LoadPrivateKey - функция для загрузки приватного ключа из файла
func LoadPrivateKey(path string) (*rsa.PrivateKey, error) {
	if path == "" {
		return nil, nil
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("error reading private key file: %w", err)
	}

	block, _ := pem.Decode(data)
	if block == nil || block.Type != "RSA PRIVATE KEY" {
		return nil, fmt.Errorf("invalid PEM format or missing private key")
	}

	privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("error parsing private key: %w", err)
	}
	return privateKey, nil
}

// DecryptMiddleware - middleware для дешифровки данных, зашифрованных публичным ключом
func DecryptMiddleware(privateKey *rsa.PrivateKey) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) (err error) {
			contentAsymmetric := c.Request().Header.Get("Content-Type")
			if contentAsymmetric != "application/octet-stream" {
				return next(c)
			}

			log.Printf("content-type: %s", contentAsymmetric)

			body, err := io.ReadAll(c.Request().Body)
			if err != nil {
				log.Printf("error reading request body: %v", err)
				return c.String(http.StatusInternalServerError, "error reading request body")
			}

			decryptedData, err := DecryptWithPrivateKey(body, privateKey)
			if err != nil {
				log.Printf("error decrypting data: %v", err)
				return c.String(http.StatusInternalServerError, "error decrypting data")
			}

			c.Request().Body = io.NopCloser(bytes.NewReader(decryptedData))

			c.Request().Header.Set("Content-Type", "application/json")

			if err = next(c); err != nil {
				c.Error(err)
			}

			return err
		}
	}
}

// DecryptWithPrivateKey - функция для дешифровки данных с использованием приватного ключа
func DecryptWithPrivateKey(data []byte, privateKey *rsa.PrivateKey) ([]byte, error) {
	chunkSize := privateKey.Size()
	var decryptedData []byte

	for start := 0; start < len(data); start += chunkSize {
		end := start + chunkSize
		if end > len(data) {
			end = len(data)
		}
		chunk := data[start:end]

		decryptedChunk, err := rsa.DecryptOAEP(sha256.New(), rand.Reader, privateKey, chunk, nil)
		if err != nil {
			return nil, fmt.Errorf("error decrypting chunk: %w", err)
		}

		decryptedData = append(decryptedData, decryptedChunk...)
	}

	return decryptedData, nil
}
