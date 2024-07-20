package main

import (
	"github.com/Sofja96/go-metrics.git/internal/server/handlers"
	"log"
)

func main() {
	s := handlers.New()
	if err := s.Start(); err != nil {
		log.Fatal(err)
	}
}
