package gzip

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
)

// Compress принимает данные в любом формате (protobuf или JSON) и сжимает их в gzip
func Compress[T any](metrics []T) ([]byte, error) {
	var b bytes.Buffer
	js, err := json.Marshal(metrics)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal JSON: %w", err)
	}

	gz, err := gzip.NewWriterLevel(&b, gzip.BestSpeed)
	if err != nil {
		return nil, fmt.Errorf("failed to create gzip writer: %w", err)
	}

	_, err = gz.Write(js)
	if err != nil {
		return nil, fmt.Errorf("failed to write gzip data: %w", err)
	}

	if err := gz.Close(); err != nil {
		return nil, fmt.Errorf("failed to close gzip writer: %w", err)
	}

	return b.Bytes(), nil
}

// Decompress распаковывает Gzip-сжатые данные
func Decompress(compressedData []byte) ([]byte, error) {
	reader, err := gzip.NewReader(bytes.NewReader(compressedData))
	if err != nil {
		return nil, err
	}
	defer reader.Close()

	var decompressedData bytes.Buffer
	_, err = io.Copy(&decompressedData, reader)
	if err != nil {
		return nil, err
	}

	return decompressedData.Bytes(), nil
}
