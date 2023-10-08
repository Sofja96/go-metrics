package main

import (
	"github.com/Sofja96/go-metrics.git/internal/handlers"
	"log"
)

//func main() {
//	c := config.LoadConfig()
//	config.ParseFlags(c)
//	//var DB *database.Dbinstance
//	////db      *database.DBConnection
//	//DB = database.DBConnection(c.DatabaseDSN)
//	//DB := database.DBConnection(c.DbDSN)
//	s := storage.NewMemStorage(c.StoreInterval, c.FilePath, c.Restore)
//	if c.FilePath != "" {
//		if c.Restore {
//			err := storage.LoadStorageFromFile(s, c.FilePath)
//			if err != nil {
//				log.Print(err)
//			}
//		}
//		if c.StoreInterval != 0 {
//			go func() {
//				err := storage.Dump(s, c.FilePath, c.StoreInterval)
//				if err != nil {
//					log.Print(err)
//				}
//			}()
//		}
//	}
//	e := handlers.CreateServer(s)
//	log.Println("Running server on", c.Address)
//	err := e.Start(c.Address)
//	if err != nil {
//		log.Fatal(err)
//	}
//
//}

func main() {
	s := handlers.New()
	if err := s.Start(); err != nil {
		log.Fatal(err)
	}
}
