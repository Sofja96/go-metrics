package utils

import (
	"bytes"
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
