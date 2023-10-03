package main

import (
	"github.com/Sofja96/go-metrics.git/internal/handlers"
	"github.com/Sofja96/go-metrics.git/internal/server/config"
	"github.com/Sofja96/go-metrics.git/internal/storage"
	"log"
)

func main() {
	c := config.LoadConfig()
	config.ParseFlags(c)
	s := storage.NewMemStorage(c.StoreInterval, c.FilePath, c.Restore)
	if c.FilePath != "" {
		if c.Restore {
			//	storage.LoadStorageFromFile(s, c.FilePath)
			err := storage.LoadStorageFromFile(s, c.FilePath)
			if err != nil {
				log.Print(err)
			}
		}
		if c.StoreInterval != 0 {
			go func() {
				err := storage.Dump(s, c.FilePath, c.StoreInterval)
				if err != nil {
					log.Print(err)
				}
			}()
		}
	}
	e := handlers.CreateServer(s)
	log.Println("Running server on", c.Address)
	err := e.Start(c.Address)
	if err != nil {
		log.Fatal(err)
	}

}
