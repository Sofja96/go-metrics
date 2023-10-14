package handlers

import (
	"github.com/Sofja96/go-metrics.git/internal/middleware"
	"github.com/Sofja96/go-metrics.git/internal/storage"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

// import (
//
//	//_ "github.com/Sofja96/go-metrics.git/internal/database"
//	"github.com/Sofja96/go-metrics.git/internal/storage/database"
//	"github.com/Sofja96/go-metrics.git/internal/storage/memory"
//
//	//"github.com/Sofja96/go-metrics.git/store/database"
//	"github.com/Sofja96/go-metrics.git/internal/storage"
//	//"github.com/Sofja96/go-metrics.git/store/database"
//	"github.com/Sofja96/go-metrics.git/internal/middleware"
//	config "github.com/Sofja96/go-metrics.git/internal/server/config"
//	//"github.com/Sofja96/go-metrics.git/store/storage/database"
//	//"github.com/Sofja96/go-metrics.git/store/storage/memory"
//	"github.com/labstack/echo/v4"
//	"go.uber.org/zap"
//	"log"
//
// )
//
//	type APIServer struct {
//		//storage *storage.Storage
//		storage *memory.MemStorage
//		echo    *echo.Echo
//		address string
//		logger  zap.SugaredLogger
//		config  *config.Config
//		db      *database.Postgres
//		store   *storage.Storage
//	}
//
//	func New() *APIServer {
//		a := &APIServer{}
//		c := config.LoadConfig()
//		config.ParseFlags(c)
//
//		a.address = c.Address
//		//a.config = &conf
//		//	a.db = database.NewStorage(c.DatabaseDSN)
//		log.Println("DSN", c.DatabaseDSN)
//		a.echo = echo.New()
//		//err :=
//		//if len(c.DatabaseDSN) == 0 {
//		//	a.storage = storage.NewMemStorage(c.StoreInterval, c.FilePath, c.Restore)
//		//} else {
//		//	a.db = database.NewStorage(c.DatabaseDSN)
//		//}
//		//if err != nil {
//		//	return nil, fmt.Errorf("error on creating storage: %w", err)
//		//}
//
//		//var s storage.Storage
//		//if len(c.DatabaseDSN) == 0 {
//		//	s = memory.NewMemStorage(c.StoreInterval, c.FilePath, c.Restore)
//		//	//	s, err = storage.NewInMemoryStorage(filename, restore, timeout)
//		//} else {
//		//	s, _ = storage.NewPostgresqlStorage(c.DatabaseDSN)
//		//}
//		//if err != nil {
//		//	return nil, fmt.Errorf("error on creating storage: %w", err)
//		//}
//		var store storage.Storage
//		if c.DatabaseDSN != "" {
//			store, _ = database.NewStorage(c.DatabaseDSN)
//			//if err != nil {
//			//	log.Fatalf("Could not init postgres repository: %s", err.Error())
//			//}
//		} else if c.FilePath != "" {
//			if c.Restore {
//				err := memory.LoadStorageFromFile(a.storage, c.FilePath)
//				if err != nil {
//					log.Print(err)
//				}
//			} else if c.StoreInterval != 0 {
//				go func() {
//					err := memory.Dump(a.storage, c.FilePath, c.StoreInterval)
//					if err != nil {
//						log.Print(err)
//					}
//				}()
//			}
//		} else {
//			store = memory.NewMemStorage(c.StoreInterval, c.FilePath, c.Restore)
//		}
//
//		//a.storage = memory.NewMemStorage(c.StoreInterval, c.FilePath, c.Restore)
//		//if c.FilePath != "" {
//		//	if c.Restore {
//		//		err := memory.LoadStorageFromFile(a.storage, c.FilePath)
//		//		if err != nil {
//		//			log.Print(err)
//		//		}
//		//	}
//		//	if c.StoreInterval != 0 {
//		//		go func() {
//		//			err := memory.Dump(a.storage, c.FilePath, c.StoreInterval)
//		//			if err != nil {
//		//				log.Print(err)
//		//			}
//		//		}()
//		//	}
//		//}
//
//		//if a.db.DB != nil {
//		//	var mutex = &sync.Mutex{}
//		//	//mutex.Lock()
//		//	//err :=
//		//	database.SaveMetricsInStorage(a.storage, a.db)
//		//	//if err != nil {
//		//	//		log.Print(err)
//		//	//	}
//		//	//mutex.Unlock()
//		//	if c.StoreInterval != 0 {
//		//		mutex.Lock()
//		//		go database.Dump(a.storage, a.db, c.StoreInterval)
//		//		mutex.Unlock()
//		//	}
//		//} else if c.FilePath != "" {
//		//	if c.Restore {
//		//		err := storage.LoadStorageFromFile(a.storage, c.FilePath)
//		//		if err != nil {
//		//			log.Print(err)
//		//		}
//		//	}
//		//	if c.StoreInterval != 0 {
//		//		go func() {
//		//			var mutex = &sync.Mutex{}
//		//			mutex.Lock()
//		//			err := storage.Dump(a.storage, c.FilePath, c.StoreInterval)
//		//			mutex.Unlock()
//		//			if err != nil {
//		//				log.Print(err)
//		//			}
//		//		}()
//		//	}
//		//}
//
//		logger, err := zap.NewDevelopment()
//		if err != nil {
//			panic(err)
//		}
//		defer logger.Sync()
//
//		a.logger = *logger.Sugar()
//		a.echo.Use(middleware.WithLogging(a.logger))
//		a.echo.Use(middleware.GzipMiddleware())
//		a.echo.POST("/update/", UpdateJSON(store))
//		a.echo.POST("/value/", ValueJSON(store))
//		//a.echo.GET("/", AllMetrics(store))
//		a.echo.GET("/", GetAllMetrics(store))
//		a.echo.GET("/value/:typeM/:nameM", ValueMetric(store))
//		a.echo.POST("/update/:typeM/:nameM/:valueM", Webhook(store))
//		a.echo.GET("/ping", Ping(store))
//		//a.echo.GET("/ping", PingDB((*database2.Postgres)(a.db)))
//
//		return a
//	}
//
//	func (a *APIServer) Start() error {
//		err := a.echo.Start("localhost:8080")
//		if err != nil {
//			log.Fatal(err)
//		}
//		log.Println("Running server on", a.address)
//
//		return nil
//	}
//
// //func CreateServer(s *memory.MemStorage) *echo.Echo {
// //	var sugar zap.SugaredLogger
// //	logger, err := zap.NewDevelopment()
// //	if err != nil {
// //		panic(err)
// //	}
// //	defer logger.Sync()
// //	sugar = *logger.Sugar()
// //	e := echo.New()
// //	e.Use(middleware.WithLogging(sugar))
// //	e.Use(middleware.GzipMiddleware())
// //	e.POST("/update/", UpdateJSON(s))
// //	e.POST("/value/", ValueJSON(s))
// //	e.GET("/", GetAllMetrics(s))
// //	e.GET("/value/:typeM/:nameM", ValueMetric(s))
// //	e.POST("/update/:typeM/:nameM/:valueM", Webhook(s))
// //	return e
// //}
func CreateServer(s storage.Storage) *echo.Echo {
	var sugar zap.SugaredLogger
	logger, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}
	defer logger.Sync()
	sugar = *logger.Sugar()
	e := echo.New()
	e.Use(middleware.WithLogging(sugar))
	e.Use(middleware.GzipMiddleware())
	e.POST("/update/", UpdateJSON(s))
	e.POST("/value/", ValueJSON(s))
	e.GET("/", GetAllMetrics(s))
	e.GET("/value/:typeM/:nameM", ValueMetric(s))
	e.POST("/update/:typeM/:nameM/:valueM", Webhook(s))
	e.GET("/ping", Ping(s))
	return e
}
