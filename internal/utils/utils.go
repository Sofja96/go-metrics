package utils

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"log"
	"os"
)

func GetDataFromFile(path string) *bytes.Buffer {
	data, err := os.ReadFile(path)
	if err != nil {
		log.Fatal(err)
	}
	return bytes.NewBuffer(data)
}

func FloatPtr(f float64) *float64 {
	return &f
}

func IntPtr(i int64) *int64 {
	return &i
}

func GenerateRsaKeyPair() (*rsa.PrivateKey, *rsa.PublicKey) {
	key, _ := rsa.GenerateKey(rand.Reader, 4096)
	return key, &key.PublicKey
}

func PrivateToString(key *rsa.PrivateKey) string {
	bytes := x509.MarshalPKCS1PrivateKey(key)
	private := pem.EncodeToMemory(
		&pem.Block{

			Type:  "RSA PRIVATE KEY",
			Bytes: bytes,
		},
	)
	return string(private)
}

func PublicToString(key *rsa.PublicKey) (string, error) {
	bytes, err := x509.MarshalPKIXPublicKey(key)
	if err != nil {
		return ``, err
	}

	public := pem.EncodeToMemory(
		&pem.Block{
			Type:  "RSA PUBLIC KEY",
			Bytes: bytes,
		},
	)

	return string(public), nil
}

func ExportToFile(data string, file string) error {
	return os.WriteFile(file, []byte(data), 0644)
}

// ReadConfigFromFile - универсальная функция для чтения конфигурации из файла и ее парсинга в любой тип
func ReadConfigFromFile[T any](filePath string) (*T, error) {
	if filePath == "" {
		return nil, nil
	}

	fcontent, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config T

	err = json.Unmarshal(fcontent, &config)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal config file: %w", err)
	}

	return &config, nil
}
