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
	echo    *echo.Echo
	address string
	logger  zap.SugaredLogger
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
	key := c.HashKey
	a.logger = *logger.Sugar()
	a.echo.Use(middleware.WithLogging(a.logger))
	if len(key) != 0 {
		a.echo.Use(middleware.HashMacMiddleware([]byte(key)))
	}
	a.echo.Use(middleware.GzipMiddleware())
	a.echo.POST("/update/", UpdateJSON(store))
	a.echo.POST("/updates/", UpdatesBatch(store))
	a.echo.POST("/value/", ValueJSON(store))
	a.echo.GET("/", GetAllMetrics(store))
	a.echo.GET("/value/:typeM/:nameM", ValueMetric(store))
	a.echo.POST("/update/:typeM/:nameM/:valueM", Webhook(store))
	a.echo.GET("/ping", Ping(store))
	log.Println(c.HashKey, "key")
	return a
}

func (a *APIServer) Start() error {
	err := a.echo.Start(a.address)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Running server on", a.address)

	return nil
}

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
	e.POST("/updates/", UpdatesBatch(s))
	e.POST("/value/", ValueJSON(s))
	e.GET("/", GetAllMetrics(s))
	e.GET("/value/:typeM/:nameM", ValueMetric(s))
	e.POST("/update/:typeM/:nameM/:valueM", Webhook(s))
	e.GET("/ping", Ping(s))
	return e
}
