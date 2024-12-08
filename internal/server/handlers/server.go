package handlers

import (
	"github.com/Sofja96/go-metrics.git/internal/server/config"
	middleware2 "github.com/Sofja96/go-metrics.git/internal/server/middleware"
	"github.com/Sofja96/go-metrics.git/internal/server/storage"
	"github.com/Sofja96/go-metrics.git/internal/server/storage/database"
	"github.com/Sofja96/go-metrics.git/internal/server/storage/memory"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
	"log"
)

// APIServer - структура настроек API сервера.
type APIServer struct {
	echo    *echo.Echo
	address string
	logger  zap.SugaredLogger
}

// New - создает, инициализурет и конфигурирует новый экземпляр ApiServer.
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
	a.echo.Use(middleware2.WithLogging(a.logger))
	if len(key) != 0 {
		a.echo.Use(middleware2.HashMacMiddleware([]byte(key)))
	}
	a.echo.Use(middleware2.GzipMiddleware())
	a.echo.POST("/update/", UpdateJSON(store))
	a.echo.POST("/updates/", UpdatesBatch(store))
	a.echo.POST("/value/", ValueJSON(store))
	a.echo.GET("/", GetAllMetrics(store))
	a.echo.GET("/value/:typeM/:nameM", ValueMetric(store))
	a.echo.POST("/update/:typeM/:nameM/:valueM", Webhook(store))
	a.echo.GET("/ping", Ping(store))
	return a
}

// Start - запускает сервер на заданном адресе.
func (a *APIServer) Start() error {
	err := a.echo.Start(a.address)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Running server on", a.address)

	return nil
}
