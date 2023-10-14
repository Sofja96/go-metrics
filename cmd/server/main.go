package main

import (
	"github.com/Sofja96/go-metrics.git/internal/handlers"
	"github.com/Sofja96/go-metrics.git/internal/server/config"
	"github.com/Sofja96/go-metrics.git/internal/storage"
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
		store, err = memory.NewPostgresqlStorage(c.DatabaseDSN)
	}
	if err != nil {
		log.Print(err)
	}

	//store = memory.NewMemStorage(c.StoreInterval, c.FilePath, c.Restore)
	//if c.DatabaseDSN != "" {
	//	store, _ = database.NewStorage(c.DatabaseDSN)
	//	if err != nil {
	//		log.Fatalf("Could not init postgres repository: %s", err.Error())
	//	}
	//} else if c.FilePath != "" {
	//	if c.Restore {
	//		err, _ := memory.LoadStorageFromFile(store, c.FilePath)
	//		if err != nil {
	//			log.Print(err)
	//		}
	//	} else if c.StoreInterval != 0 {
	//		go func() {
	//			err := memory.Dump(store, c.FilePath, c.StoreInterval)
	//			if err != nil {
	//				log.Print(err)
	//			}
	//		}()
	//	}
	//}
	//} else {
	//	store = memory.NewMemStorage(c.StoreInterval, c.FilePath, c.Restore)
	//}
	////var DB *database.Dbinstance
	//////db      *database.DBConnection
	////DB = database.DBConnection(c.DatabaseDSN)
	////DB := database.DBConnection(c.DbDSN)
	//s := storage.NewMemStorage(c.StoreInterval, c.FilePath, c.Restore)
	//if c.FilePath != "" {
	//	if c.Restore {
	//		err := storage.LoadStorageFromFile(s, c.FilePath)
	//		if err != nil {
	//			log.Print(err)
	//		}
	//	}
	//	if c.StoreInterval != 0 {
	//		go func() {
	//			err := storage.Dump(s, c.FilePath, c.StoreInterval)
	//			if err != nil {
	//				log.Print(err)
	//			}
	//		}()
	//	}
	//}
	e := handlers.CreateServer(store)
	log.Println("Running server on", c.Address)
	err = e.Start(c.Address)
	if err != nil {
		log.Fatal(err)
	}

}

//func main() {
//	s := handlers.New()
//	if err := s.Start(); err != nil {
//		log.Fatal(err)
//	}
//}
