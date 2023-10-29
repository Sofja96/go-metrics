package main

import (
	"github.com/Sofja96/go-metrics.git/internal/handlers"
	"github.com/Sofja96/go-metrics.git/internal/server/config"
	"github.com/Sofja96/go-metrics.git/internal/storage"
	"github.com/Sofja96/go-metrics.git/internal/storage/database"
	"github.com/Sofja96/go-metrics.git/internal/storage/memory"
	"log"
)

func main() {
	c := config.LoadConfig()
	config.ParseFlags(c)
	var store storage.Storage
	var err error
	if len(c.DatabaseDSN) == 0 {
		store, err = memory.NewInMemStorage(c.StoreInterval, c.FilePath, c.Restore)
		if err != nil {
			log.Print(err)
		}
	} else {
		store, err = database.NewStorage(c.DatabaseDSN)
	}
	if err != nil {
		log.Print(err)
	}

	e := handlers.CreateServer(c, store)
	log.Println("Running server on", c.Address)
	err = e.Start(c.Address)
	if err != nil {
		log.Fatal(err)
	}

}
