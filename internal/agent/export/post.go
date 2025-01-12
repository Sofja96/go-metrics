package export

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/Sofja96/go-metrics.git/internal/agent/envs"

	"github.com/hashicorp/go-retryablehttp"

	"github.com/Sofja96/go-metrics.git/internal/agent/hash"
)

// Настройки повторной отправки по умолчанию.
const (
	retryMax     int           = 3               // максимальное количество
	retryWaitMin time.Duration = time.Second * 1 // минимальное время ожидания
	retryWaitMax time.Duration = time.Second * 5 // максимальное время ожидания
)

// PostQueries - функция для формирования метрик перед отправкой и запуска отправки метрик.
func PostQueries(ctx context.Context, cfg *envs.Config, chIn <-chan []byte, publicKey *rsa.PublicKey) {
	url := fmt.Sprintf("http://%s/updates/", cfg.Address)

	retryClient := retryablehttp.NewClient()
	retryClient.RetryMax = retryMax
	retryClient.RetryWaitMin = retryWaitMin
	retryClient.RetryWaitMax = retryWaitMax
	retryClient.Backoff = linearBackoff

	for {
		select {
		case <-ctx.Done():
			return
		case compressedData, ok := <-chIn:
			if !ok {
				log.Println("Канал данных закрыт. Завершаем Worker")
				return
			}
			err := PostBatch(retryClient, url, cfg.HashKey, compressedData, publicKey)
			if err != nil {
				log.Printf("Ошибка отправки метрик: %v", err)
			}

		}
	}
}

// PostBatch - функция отправки сжатых метрик на сервер.
func PostBatch(r *retryablehttp.Client, url string, key string, m []byte, publicKey *rsa.PublicKey) error {
	var dataToSend []byte
	var contentType string

	if publicKey != nil {
		encryptedData, err := EncryptWithPublicKey(m, publicKey)
		if err != nil {
			return fmt.Errorf("error encrypting data: %w", err)
		}
		dataToSend = encryptedData
		contentType = "application/octet-stream"
	} else {
		dataToSend = m
		contentType = "application/json"
	}

	req, err := retryablehttp.NewRequest("POST", url, dataToSend)
	if err != nil {
		return fmt.Errorf("error connection: %w", err)
	}

	req.Header.Add("content-type", contentType)
	req.Header.Add("content-encoding", "gzip")
	req.Header.Add("Accept-Encoding", "gzip")

	if len(key) != 0 {
		hmac, err := hash.ComputeHmac256([]byte(key), dataToSend)
		if err != nil {
			return fmt.Errorf("error compute hash data: %w", err)
		}
		req.Header.Add("HashSHA256", hmac)
	}
	resp, err := r.Do(req)
	if err != nil {
		return fmt.Errorf("error connection: %w", err)
	}
	defer resp.Body.Close()

	log.Printf("Response Status Code: %d", resp.StatusCode)
	log.Printf("Response Headers: %v", resp.Header)

	return nil
}

// EncryptWithPublicKey - функция для шифрования данных с использованием публичного ключа RSA.
func EncryptWithPublicKey(data []byte, publicKey *rsa.PublicKey) ([]byte, error) {
	chunkSize := publicKey.Size() - 2*sha256.Size - 2 // Максимальный размер блока

	var encryptedChunks []byte

	for start := 0; start < len(data); start += chunkSize {
		end := start + chunkSize
		if end > len(data) {
			end = len(data)
		}
		chunk := data[start:end]

		encryptedChunk, err := rsa.EncryptOAEP(sha256.New(), rand.Reader, publicKey, chunk, nil)
		if err != nil {
			return nil, fmt.Errorf("error encrypting chunk: %w", err)
		}

		encryptedChunks = append(encryptedChunks, encryptedChunk...)
	}

	return encryptedChunks, nil
}

// linearBackoff - расчитывает время ожижания между попытками отправки
func linearBackoff(min, max time.Duration, attemptNum int, resp *http.Response) time.Duration {
	sleepTime := min + min*time.Duration(2*attemptNum)
	return sleepTime
}
