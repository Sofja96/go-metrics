package handlers

import (
	"github.com/Sofja96/go-metrics.git/internal/middleware"
	"github.com/Sofja96/go-metrics.git/internal/server/config"
	"github.com/Sofja96/go-metrics.git/internal/storage"
	"github.com/Sofja96/go-metrics.git/internal/storage/database"
	"github.com/Sofja96/go-metrics.git/internal/storage/memory"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
	"log"
)

type APIServer struct {
	//storage *storage.MemStorage
	echo    *echo.Echo
	address string
	logger  zap.SugaredLogger
	//config  *config.Config
	//db      *database.Postgres
}

func New() *APIServer {
	a := &APIServer{}
	c := config.LoadConfig()
	config.ParseFlags(c)

	a.address = c.Address
	a.echo = echo.New()
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

	logger, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}
	defer logger.Sync()
	//e := echo.New()
	//e.Use(middleware.WithLogging(a.logger))
	//e.Use(middleware.GzipMiddleware())
	//e.POST("/update/", UpdateJSON(store))
	//e.POST("/updates/", UpdatesBatch(store))
	//e.POST("/value/", ValueJSON(store))
	//e.GET("/", GetAllMetrics(store))
	//e.GET("/value/:typeM/:nameM", ValueMetric(store))
	//e.POST("/update/:typeM/:nameM/:valueM", Webhook(store))
	//e.GET("/ping", Ping(store))
	//if len(c.HashKey) != 0 {
	//	log.Println(c.HashKey)
	//	e.Use(middleware.HashMacMiddleware([]byte(c.HashKey)))
	//}
	key := c.HashKey
	a.logger = *logger.Sugar()
	a.echo.Use(middleware.WithLogging(a.logger))
	//	a.echo.Use(middleware.HashMiddleware([]byte(key)))
	log.Println(key)
	log.Println([]byte(key))
	a.echo.Use(middleware.GzipMiddleware())
	if len(key) != 0 {
		log.Println(key, "key")
		a.echo.Use(middleware.HashMiddleware([]byte(key)))
	}

	a.echo.POST("/update/", UpdateJSON(store))
	a.echo.POST("/updates/", UpdatesBatch(store))
	a.echo.POST("/value/", ValueJSON(store))
	a.echo.GET("/", GetAllMetrics(store))
	a.echo.GET("/value/:typeM/:nameM", ValueMetric(store))
	a.echo.POST("/update/:typeM/:nameM/:valueM", Webhook(store))
	a.echo.GET("/ping", Ping(store))
	log.Println(c.HashKey, "key")
	//a.echo.Use(middleware.HashMacMiddleware([]byte(c.HashKey)))
	//if len(key) != 0 {
	//	log.Println(key, "key")
	//a.echo.Use(middleware.HashMiddleware([]byte(key)))
	//}

	return a
}

func (a *APIServer) Start() error {
	//e := echo.New()
	err := a.echo.Start(a.address)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Running server on", a.address)

	return nil
}

func CreateServer(s storage.Storage) *echo.Echo {
	//	c := config.LoadConfig()
	var sugar zap.SugaredLogger
	//var key string
	//key := config.LoadConfig().HashKey
	//log.Println(key)
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
	e.POST("/updates/", UpdatesBatch(s))
	e.POST("/value/", ValueJSON(s))
	e.GET("/", GetAllMetrics(s))
	e.GET("/value/:typeM/:nameM", ValueMetric(s))
	e.POST("/update/:typeM/:nameM/:valueM", Webhook(s))
	e.GET("/ping", Ping(s))
	//log.Println(key)
	//e.Use(middleware.HashMacMiddleware([]byte(key)))
	//e.Use(middleware.HashMacMiddleware(([]byte(key))))
	//if len(key) != 0 {
	//	log.Println(key)
	//	e.Use(middleware.HashMacMiddleware(key))
	//}
	//e.Use(middleware.HashMacMiddleware())
	//if len(c.HashKey) > 0 {
	//	e.Use(middleware.HashMacMiddleware([]byte(c.HashKey)))
	//}
	return e
}

//func CreateServerHash(s storage.Storage, key string) *echo.Echo {
//	var sugar zap.SugaredLogger
//	//var key string
//	logger, err := zap.NewDevelopment()
//	if err != nil {
//		panic(err)
//	}
//	defer logger.Sync()
//	sugar = *logger.Sugar()
//	e := echo.New()
//	e.Use(middleware.WithLogging(sugar))
//	e.Use(middleware.GzipMiddleware())
//	if len(key) > 0 {
//		e.Use(middleware.HashMacMiddleware([]byte(key)))
//	}
//	e.POST("/update/", UpdateJSON(s))
//	e.POST("/updates/", UpdatesBatch(s))
//	e.POST("/value/", ValueJSON(s))
//	e.GET("/", GetAllMetrics(s))
//	e.GET("/value/:typeM/:nameM", ValueMetric(s))
//	e.POST("/update/:typeM/:nameM/:valueM", Webhook(s))
//	e.GET("/ping", Ping(s))
//	return e
//}
