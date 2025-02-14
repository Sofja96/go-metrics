package utils

import (
	"bytes"
	"crypto/hmac"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/hex"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"log"
	"net"
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

// GetLocalIP - возвращает локальный IP-адрес хоста
func GetLocalIP() (string, error) {
	iFaces, err := net.Interfaces()
	if err != nil {
		return "", fmt.Errorf("error getting network interfaces: %w", err)
	}

	for _, iFace := range iFaces {
		addresses, err := iFace.Addrs()
		if err != nil {
			return "", fmt.Errorf("error getting addresses for interface %s: %w", iFace.Name, err)
		}

		for _, addr := range addresses {
			if ipNet, ok := addr.(*net.IPNet); ok {
				if ipNet.IP.To4() != nil {
					return ipNet.IP.String(), nil
				}
			}
		}
	}

	return "", fmt.Errorf("no valid IPv4 address found")
}

// ComputeHmac256 - функция для вычисления HMAC-SHA256.
func ComputeHmac256(key []byte, data []byte) string {
	mac := hmac.New(sha256.New, key)
	mac.Write(data)
	hashedData := mac.Sum(nil)
	return hex.EncodeToString(hashedData)
}
