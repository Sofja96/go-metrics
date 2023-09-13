package main

import (
	"fmt"
	"github.com/Sofja96/go-metrics.git/internal/handlers"
	"github.com/Sofja96/go-metrics.git/internal/server/config"
	"github.com/Sofja96/go-metrics.git/internal/storage"
	"log"
)

func main() {
	c := config.LoadConfig()
	config.ParseFlags(c)
	s := storage.NewMemStorage()
	e := handlers.CreateServer(s)
	fmt.Println("Running server on", c.Address)
	err := e.Start(c.Address)
	if err != nil {
		log.Fatal(err)
	}
}
