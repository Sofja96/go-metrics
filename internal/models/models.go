package models

import "crypto/rsa"

// Metrics - структура метрик их идентификатор, тип и значение
type Metrics struct {
	ID    string   `json:"id"`              // имя метрики
	MType string   `json:"type"`            // параметр, принимающий значение gauge или counter
	Delta *int64   `json:"delta,omitempty"` // значение метрики в случае передачи counter
	Value *float64 `json:"value,omitempty"` // значение метрики в случае передачи gauge
}

type PostRequest struct {
	Key       string
	PublicKey *rsa.PublicKey
}
